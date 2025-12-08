package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ConnectionsHandler struct {
	envURL       string
	clientID     string
	clientSecret string
}

type ConnectionRequest struct {
	Enabled bool `json:"enabled"`
}

type ConnectionResponse struct {
	Success    bool               `json:"success"`
	Message    string             `json:"message,omitempty"`
	Connection *ConnectionDetails `json:"connection,omitempty"`
}

type ConnectionDetails struct {
	Provider    string    `json:"provider"`
	Enabled     bool      `json:"enabled"`
	ConnectedAt time.Time `json:"connected_at,omitempty"`
}

// Agent Actions API response types (from Scalar docs)
type AgentConnection struct {
	ID               string                 `json:"id"`
	OrganizationID   string                 `json:"organization_id"`
	Provider         string                 `json:"provider"`
	ProviderMetadata map[string]interface{} `json:"provider_metadata,omitempty"`
	Status           string                 `json:"status"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
}

type ListConnectionsAPIResponse struct {
	Connections []AgentConnection `json:"connections"`
}

type CreateAuthorizationURLRequest struct {
	RedirectURI string `json:"redirect_uri"`
}

type CreateAuthorizationURLResponse struct {
	URL string `json:"url"`
}

func NewConnectionsHandler() (*ConnectionsHandler, error) {
	config, err := getScaleKitConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get ScaleKit config: %w", err)
	}

	return &ConnectionsHandler{
		envURL:       config.EnvURL,
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
	}, nil
}

// getAccessToken gets an M2M access token for API calls
func (h *ConnectionsHandler) getAccessToken(ctx context.Context) (string, error) {
	tokenURL := fmt.Sprintf("%s/oauth/token", h.envURL)

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", h.clientID)
	data.Set("client_secret", h.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed: %s - %s", resp.Status, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

// makeAPIRequest makes an authenticated request to Scalekit Agent Actions API
func (h *ConnectionsHandler) makeAPIRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	token, err := h.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	url := fmt.Sprintf("%s%s", h.envURL, path)

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed [%d]: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// getOrganizationID extracts organization ID from JWT token
func (h *ConnectionsHandler) getOrganizationID(c *gin.Context) (string, error) {
	accessToken, err := c.Cookie("auth_access_token")
	if err != nil {
		return "", fmt.Errorf("no access token found")
	}

	claims, err := decodeJwtPayload(accessToken)
	if err != nil {
		return "", fmt.Errorf("failed to decode token: %w", err)
	}

	orgID, err := getStringClaim(claims, "xoid")
	if err != nil {
		return "", fmt.Errorf("no organization ID in token: %w", err)
	}

	return orgID, nil
}

// RegisterRoutes registers all connection routes
func (h *ConnectionsHandler) RegisterRoutes(r *gin.RouterGroup) {
	connections := r.Group("/connections")
	{
		// Slack routes
		connections.GET("/slack", h.GetSlackStatus)
		connections.POST("/slack", h.ConfigureSlack)
		connections.GET("/slack/callback", h.SlackCallback)

		// GitHub routes
		connections.GET("/github", h.GetGithubStatus)
		connections.POST("/github", h.ConfigureGithub)
		connections.GET("/github/callback", h.GithubCallback)
	}
}

// getConnectionStatus is a helper to check connection status for a provider
func (h *ConnectionsHandler) getConnectionStatus(ctx context.Context, orgID, provider string) (*AgentConnection, error) {
	// List all connections for organization
	respBody, err := h.makeAPIRequest(
		ctx,
		"GET",
		fmt.Sprintf("/api/v1/agent/organizations/%s/connections", orgID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	var apiResp ListConnectionsAPIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Find connection for specified provider
	for i := range apiResp.Connections {
		if apiResp.Connections[i].Provider == provider {
			return &apiResp.Connections[i], nil
		}
	}

	return nil, nil // No connection found
}

// GetSlackStatus retrieves the current Slack connection status
func (h *ConnectionsHandler) GetSlackStatus(c *gin.Context) {
	orgID, err := h.getOrganizationID(c)
	if err != nil {
		log.Printf("Failed to get organization ID: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	slackConn, err := h.getConnectionStatus(c.Request.Context(), orgID, "slack")
	if err != nil {
		log.Printf("Failed to fetch Slack connection: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch connection status"})
		return
	}

	response := ConnectionResponse{
		Success: true,
	}

	if slackConn != nil && slackConn.Status == "connected" {
		createdAt, _ := time.Parse(time.RFC3339, slackConn.CreatedAt)
		response.Connection = &ConnectionDetails{
			Provider:    "slack",
			Enabled:     true,
			ConnectedAt: createdAt,
		}
	} else {
		response.Connection = &ConnectionDetails{
			Provider: "slack",
			Enabled:  false,
		}
	}

	c.JSON(http.StatusOK, response)
}

// ConfigureSlack enables or disables Slack connection
func (h *ConnectionsHandler) ConfigureSlack(c *gin.Context) {
	orgID, err := h.getOrganizationID(c)
	if err != nil {
		log.Printf("Failed to get organization ID: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req ConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Enabled {
		// Get base URL for callback
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}

		baseURL := os.Getenv("PUBLIC_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
		}

		callbackURL := fmt.Sprintf("%s/api/connections/slack/callback", baseURL)

		// Create authorization URL
		authReqBody := CreateAuthorizationURLRequest{
			RedirectURI: callbackURL,
		}

		respBody, err := h.makeAPIRequest(
			c.Request.Context(),
			"POST",
			fmt.Sprintf("/api/v1/agent/organizations/%s/connections/slack/authorize", orgID),
			authReqBody,
		)
		if err != nil {
			log.Printf("Failed to create authorization URL: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create authorization URL: %v", err)})
			return
		}

		var authResp CreateAuthorizationURLResponse
		if err := json.Unmarshal(respBody, &authResp); err != nil {
			log.Printf("Failed to parse authorization response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
			return
		}

		response := ConnectionResponse{
			Success: true,
			Message: "Please complete OAuth authorization",
			Connection: &ConnectionDetails{
				Provider: "slack",
				Enabled:  false,
			},
		}

		c.Header("X-Authorization-URL", authResp.URL)
		c.JSON(http.StatusOK, response)
	} else {
		// Disable connection - find and delete it
		slackConn, err := h.getConnectionStatus(c.Request.Context(), orgID, "slack")
		if err != nil {
			log.Printf("Failed to fetch connections: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch connections"})
			return
		}

		if slackConn != nil {
			_, err := h.makeAPIRequest(
				c.Request.Context(),
				"DELETE",
				fmt.Sprintf("/api/v1/agent/organizations/%s/connections/%s", orgID, slackConn.ID),
				nil,
			)
			if err != nil {
				log.Printf("Failed to delete connection: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete connection"})
				return
			}
		}

		response := ConnectionResponse{
			Success: true,
			Message: "Slack connection disabled successfully",
			Connection: &ConnectionDetails{
				Provider: "slack",
				Enabled:  false,
			},
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetGithubStatus retrieves the current GitHub connection status
func (h *ConnectionsHandler) GetGithubStatus(c *gin.Context) {
	orgID, err := h.getOrganizationID(c)
	if err != nil {
		log.Printf("Failed to get organization ID: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	githubConn, err := h.getConnectionStatus(c.Request.Context(), orgID, "github")
	if err != nil {
		log.Printf("Failed to fetch GitHub connection: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch connection status"})
		return
	}

	response := ConnectionResponse{
		Success: true,
	}

	if githubConn != nil && githubConn.Status == "connected" {
		createdAt, _ := time.Parse(time.RFC3339, githubConn.CreatedAt)
		response.Connection = &ConnectionDetails{
			Provider:    "github",
			Enabled:     true,
			ConnectedAt: createdAt,
		}
	} else {
		response.Connection = &ConnectionDetails{
			Provider: "github",
			Enabled:  false,
		}
	}

	c.JSON(http.StatusOK, response)
}

// ConfigureGithub enables or disables GitHub connection
func (h *ConnectionsHandler) ConfigureGithub(c *gin.Context) {
	orgID, err := h.getOrganizationID(c)
	if err != nil {
		log.Printf("Failed to get organization ID: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req ConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Enabled {
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}

		baseURL := os.Getenv("PUBLIC_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
		}

		callbackURL := fmt.Sprintf("%s/api/connections/github/callback", baseURL)

		authReqBody := CreateAuthorizationURLRequest{
			RedirectURI: callbackURL,
		}

		respBody, err := h.makeAPIRequest(
			c.Request.Context(),
			"POST",
			fmt.Sprintf("/api/v1/agent/organizations/%s/connections/github/authorize", orgID),
			authReqBody,
		)
		if err != nil {
			log.Printf("Failed to create authorization URL: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create authorization URL: %v", err)})
			return
		}

		var authResp CreateAuthorizationURLResponse
		if err := json.Unmarshal(respBody, &authResp); err != nil {
			log.Printf("Failed to parse authorization response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
			return
		}

		response := ConnectionResponse{
			Success: true,
			Message: "Please complete OAuth authorization",
			Connection: &ConnectionDetails{
				Provider: "github",
				Enabled:  false,
			},
		}

		c.Header("X-Authorization-URL", authResp.URL)
		c.JSON(http.StatusOK, response)
	} else {
		githubConn, err := h.getConnectionStatus(c.Request.Context(), orgID, "github")
		if err != nil {
			log.Printf("Failed to fetch connections: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch connections"})
			return
		}

		if githubConn != nil {
			_, err := h.makeAPIRequest(
				c.Request.Context(),
				"DELETE",
				fmt.Sprintf("/api/v1/agent/organizations/%s/connections/%s", orgID, githubConn.ID),
				nil,
			)
			if err != nil {
				log.Printf("Failed to delete connection: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete connection"})
				return
			}
		}

		response := ConnectionResponse{
			Success: true,
			Message: "GitHub connection disabled successfully",
			Connection: &ConnectionDetails{
				Provider: "github",
				Enabled:  false,
			},
		}

		c.JSON(http.StatusOK, response)
	}
}

// SlackCallback handles OAuth callback from Slack
func (h *ConnectionsHandler) SlackCallback(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, `
		<html>
			<head>
				<style>
					body {
						font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
						display: flex;
						align-items: center;
						justify-content: center;
						height: 100vh;
						margin: 0;
						background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
					}
					.container {
						background: white;
						padding: 3rem;
						border-radius: 1rem;
						box-shadow: 0 20px 60px rgba(0,0,0,0.3);
						text-align: center;
					}
					h1 { color: #667eea; margin: 0 0 1rem 0; }
					p { color: #666; margin: 0; }
					.checkmark { font-size: 4rem; margin-bottom: 1rem; }
				</style>
			</head>
			<body>
				<div class="container">
					<div class="checkmark">✅</div>
					<h1>Slack Connected!</h1>
					<p>You can close this window and return to the application.</p>
				</div>
				<script>
					setTimeout(() => window.close(), 2000);
				</script>
			</body>
		</html>
	`)
}

// GithubCallback handles OAuth callback from GitHub
func (h *ConnectionsHandler) GithubCallback(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, `
		<html>
			<head>
				<style>
					body {
						font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
						display: flex;
						align-items: center;
						justify-content: center;
						height: 100vh;
						margin: 0;
						background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
					}
					.container {
						background: white;
						padding: 3rem;
						border-radius: 1rem;
						box-shadow: 0 20px 60px rgba(0,0,0,0.3);
						text-align: center;
					}
					h1 { color: #667eea; margin: 0 0 1rem 0; }
					p { color: #666; margin: 0; }
					.checkmark { font-size: 4rem; margin-bottom: 1rem; }
				</style>
			</head>
			<body>
				<div class="container">
					<div class="checkmark">✅</div>
					<h1>GitHub Connected!</h1>
					<p>You can close this window and return to the application.</p>
				</div>
				<script>
					setTimeout(() => window.close(), 2000);
				</script>
			</body>
		</html>
	`)
}
