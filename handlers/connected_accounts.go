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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	connectorSlack         = "slack"
	connectorGitHubActions = "github-actions"
	connectorGmail         = "gmail"
	statusActive           = "ACTIVE"
	apiTimeout             = 30 * time.Second
	tokenTimeout           = 10 * time.Second
)

type ConnectedAccountsHandler struct {
	envURL       string
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

// NewConnectedAccountsHandler creates a new ConnectedAccountsHandler instance
func NewConnectedAccountsHandler() (*ConnectedAccountsHandler, error) {
	config, err := getScaleKitConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get ScaleKit config: %w", err)
	}

	return &ConnectedAccountsHandler{
		envURL:       config.EnvURL,
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		httpClient:   &http.Client{Timeout: apiTimeout},
	}, nil
}

// validateConnector validates and normalizes the connector parameter
func validateConnector(connector string) (string, error) {
	connector = strings.ToLower(connector)
	if connector != connectorSlack && connector != connectorGitHubActions && connector != connectorGmail {
		return "", fmt.Errorf("connector must be either '%s', '%s', or '%s'", connectorSlack, connectorGitHubActions, connectorGmail)
	}
	return connector, nil
}

// makeAuthenticatedRequest makes an authenticated HTTP request to Scalekit API
func (h *ConnectedAccountsHandler) makeAuthenticatedRequest(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	clientToken, err := h.getClientAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client access token: %w", err)
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewBuffer(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", clientToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Scalekit API: %w", err)
	}

	return resp, nil
}

// readAndParseResponse reads the response body and parses it as JSON
func readAndParseResponse(resp *http.Response) ([]byte, map[string]interface{}, error) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	var parsedResp map[string]interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &parsedResp); err != nil {
			// If it's not JSON, that's okay - return the raw body
			return respBody, nil, nil
		}
	}

	return respBody, parsedResp, nil
}

// handleAPIError handles non-2xx HTTP responses
func handleAPIError(c *gin.Context, statusCode int, respBody []byte, parsedResp map[string]interface{}) {
	if parsedResp != nil {
		c.JSON(statusCode, parsedResp)
	} else {
		c.JSON(statusCode, gin.H{
			"error": string(respBody),
		})
	}
}

// extractConnectionData extracts connection details from Scalekit API response
func extractConnectionData(scalekitResp map[string]interface{}, connector string) gin.H {
	var status string
	var connectedAt string
	var isEnabled bool

	// Check for connected_account in response
	if ca, ok := scalekitResp["connected_account"].(map[string]interface{}); ok {
		if s, ok := ca["status"].(string); ok {
			status = s
			isEnabled = (status == statusActive)
		} else {
			isEnabled = false
		}

		// Extract timestamp (prefer updated_at, fallback to last_used_at or created_at)
		if ua, ok := ca["updated_at"].(string); ok {
			connectedAt = ua
		} else if lua, ok := ca["last_used_at"].(string); ok {
			connectedAt = lua
		} else if cat, ok := ca["created_at"].(string); ok {
			connectedAt = cat
		}
	} else if connectionData, ok := scalekitResp["connection"].(map[string]interface{}); ok {
		// Fallback: check for "connection" key (older format)
		if s, ok := connectionData["status"].(string); ok {
			status = s
			isEnabled = (status == statusActive)
		} else {
			isEnabled = false
		}
		if cat, ok := connectionData["connected_at"].(string); ok {
			connectedAt = cat
		} else if cat, ok := connectionData["created_at"].(string); ok {
			connectedAt = cat
		}
	} else {
		// No connection data found, default to disabled
		isEnabled = false
	}

	connection := gin.H{
		"provider": connector,
		"enabled":  isEnabled,
	}
	if connectedAt != "" {
		connection["connected_at"] = connectedAt
	}

	return connection
}

// getClientAccessToken gets an M2M access token for API calls
func (h *ConnectedAccountsHandler) getClientAccessToken(ctx context.Context) (string, error) {
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

	client := &http.Client{Timeout: tokenTimeout}
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

// getUserID extracts user ID from access token
func (h *ConnectedAccountsHandler) getUserID(c *gin.Context) (string, error) {
	accessToken, err := c.Cookie("auth_access_token")
	if err != nil {
		return "", fmt.Errorf("no access token found")
	}

	claims, err := decodeJwtPayload(accessToken)
	if err != nil {
		return "", fmt.Errorf("failed to decode token: %w", err)
	}

	userID, err := getStringClaim(claims, "sub")
	if err != nil {
		return "", fmt.Errorf("no user ID (sub) in token: %w", err)
	}

	return userID, nil
}

// GetConnectedAccountStatusHandler handles GET /api/connections?connector=slack|github_actions
func (h *ConnectedAccountsHandler) GetConnectedAccountStatusHandler(c *gin.Context) {
	connector := c.Query("connector")
	if connector == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "connector query parameter is required",
		})
		return
	}

	connector, err := validateConnector(connector)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.getUserID(c)
	if err != nil {
		log.Printf("Failed to get user ID: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Failed to authenticate user",
		})
		return
	}

	ctx := c.Request.Context()
	apiURL := fmt.Sprintf("%s/api/v1/connected_accounts/auth?connector=%s&identifier=%s",
		h.envURL, connector, url.QueryEscape(userID))

	resp, err := h.makeAuthenticatedRequest(ctx, "GET", apiURL, nil)
	if err != nil {
		log.Printf("Failed to make authenticated request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	respBody, parsedResp, err := readAndParseResponse(resp)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read response",
		})
		return
	}

	// Handle 404 as "not connected" - this is a valid state
	if resp.StatusCode == http.StatusNotFound {
		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"connection": nil, // null in JSON - frontend will interpret as enabled=false
		})
		return
	}

	// Handle non-2xx responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		handleAPIError(c, resp.StatusCode, respBody, parsedResp)
		return
	}

	// Parse and transform response
	if parsedResp == nil {
		log.Printf("Failed to parse Scalekit response")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse response",
		})
		return
	}

	connection := extractConnectionData(parsedResp, connector)
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"connection": connection,
	})
}

// ConnectedAccountRequest represents the request body for enabling/disabling connected accounts
type ConnectedAccountRequest struct {
	Enabled bool `json:"enabled"`
}

// ConfigureConnectedAccountHandler handles POST /api/connections?connector=slack|github
func (h *ConnectedAccountsHandler) ConfigureConnectedAccountHandler(c *gin.Context) {
	connector := c.Query("connector")
	if connector == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "connector query parameter is required",
		})
		return
	}

	connector, err := validateConnector(connector)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req ConnectedAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body. Expected: { enabled: boolean }",
		})
		return
	}

	userID, err := h.getUserID(c)
	if err != nil {
		log.Printf("Failed to get user ID: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Failed to authenticate user",
		})
		return
	}

	ctx := c.Request.Context()

	if req.Enabled {
		h.enableConnection(c, ctx, connector, userID)
	} else {
		h.disableConnection(c, ctx, connector, userID)
	}
}

// enableConnection enables a connected account by getting a magic link
func (h *ConnectedAccountsHandler) enableConnection(c *gin.Context, ctx context.Context, connector, userID string) {
	apiURL := fmt.Sprintf("%s/api/v1/connected_accounts/magic_link", h.envURL)

	requestBody := map[string]interface{}{
		"connector":  connector,
		"identifier": userID,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Failed to marshal request body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create request",
		})
		return
	}

	resp, err := h.makeAuthenticatedRequest(ctx, "POST", apiURL, jsonData)
	if err != nil {
		log.Printf("Failed to make authenticated request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	respBody, parsedResp, err := readAndParseResponse(resp)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read response",
		})
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		handleAPIError(c, resp.StatusCode, respBody, parsedResp)
		return
	}

	// Forward Scalekit response directly to UI (contains "link" and "expiry")
	if parsedResp != nil {
		c.JSON(http.StatusOK, parsedResp)
	} else {
		c.JSON(http.StatusOK, gin.H{"error": "Invalid response format"})
	}
}

// disableConnection disables a connected account by deleting it
func (h *ConnectedAccountsHandler) disableConnection(c *gin.Context, ctx context.Context, connector, userID string) {
	apiURL := fmt.Sprintf("%s/api/v1/connected_accounts:delete", h.envURL)

	requestBody := map[string]interface{}{
		"connector":  connector,
		"identifier": userID,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Failed to marshal request body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create request",
		})
		return
	}

	resp, err := h.makeAuthenticatedRequest(ctx, "POST", apiURL, jsonData)
	if err != nil {
		log.Printf("Failed to make authenticated request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	respBody, parsedResp, err := readAndParseResponse(resp)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read response",
		})
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		handleAPIError(c, resp.StatusCode, respBody, parsedResp)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("%s connection disabled successfully", connector),
	})
}

// RegisterRoutes registers the connected accounts routes
func (h *ConnectedAccountsHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/connections", h.GetConnectedAccountStatusHandler)
	r.POST("/connections", h.ConfigureConnectedAccountHandler)
}
