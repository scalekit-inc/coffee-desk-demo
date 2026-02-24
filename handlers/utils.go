package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// encodeParams encodes a map of parameters into a URL-encoded string
func encodeParams(params map[string]string) string {
	var sb strings.Builder
	first := true
	for k, v := range params {
		if !first {
			sb.WriteString("&")
		}
		first = false
		sb.WriteString(k + "=" + urlEncode(v))
	}
	return sb.String()
}

// urlEncode performs basic URL encoding
func urlEncode(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, " ", "+"), "&", "%26")
}

// decodeJwtPayload decodes the payload (2nd part) of a JWT token and returns the claims map
func decodeJwtPayload(jwt string) (map[string]interface{}, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("error decoding token payload: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("error parsing token claims: %w", err)
	}

	return claims, nil
}

// getStringClaim extracts a string claim from a claims map
func getStringClaim(claims map[string]interface{}, key string) (string, error) {
	value, ok := claims[key].(string)
	if !ok || value == "" {
		return "", fmt.Errorf("claim %s not found or empty", key)
	}
	return value, nil
}

// getUIBaseURL returns the UI base URL from the environment, or falls back to the request host
func getUIBaseURL(c *gin.Context) string {
	uiURL := os.Getenv("UI_URL")
	if uiURL == "" {
		protocol := "https"
		if strings.Contains(c.Request.Host, "localhost") {
			protocol = "http"
		}
		uiURL = protocol + "://" + c.Request.Host
	}
	return uiURL
}

// toExternalWorkspaceID converts an internal organization ID (e.g., org_123) to an external workspace ID (e.g., wspace_123)
func toExternalWorkspaceID(organizationID string) (string, error) {
	parts := strings.Split(organizationID, "_")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid organization ID format: %s", organizationID)
	}
	return "wspace_" + parts[1], nil
}

// GetScaleKitEnvironmentURLHandler returns the SCALEKIT_ENVIRONMENT_URL for frontend use
func GetScaleKitEnvironmentURLHandler(c *gin.Context) {
	envURL := os.Getenv("SCALEKIT_ENVIRONMENT_URL")
	if envURL == "" {
		c.JSON(500, gin.H{"error": "SCALEKIT_ENVIRONMENT_URL not configured"})
		return
	}
	c.JSON(200, gin.H{"url": envURL})
}

// RedirectToPasskeysHandler redirects to the Scalekit passkeys page
func RedirectToPasskeysHandler(c *gin.Context) {
	envURL := os.Getenv("SCALEKIT_ENVIRONMENT_URL")
	if envURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SCALEKIT_ENVIRONMENT_URL not configured"})
		return
	}

	// Construct the passkeys URL and redirect
	passkeysURL := fmt.Sprintf("%s/ui/profile/passkeys", envURL)
	c.Redirect(http.StatusTemporaryRedirect, passkeysURL)
}

// RedirectToSettingsHandler redirects to the Scalekit settings/UI page
func RedirectToSettingsHandler(c *gin.Context) {
	envURL := os.Getenv("SCALEKIT_ENVIRONMENT_URL")
	if envURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SCALEKIT_ENVIRONMENT_URL not configured"})
		return
	}

	// Construct the settings URL and redirect
	settingsURL := fmt.Sprintf("%s/ui", envURL)
	c.Redirect(http.StatusTemporaryRedirect, settingsURL)
}
