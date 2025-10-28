package handlers

import (
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/scalekit-inc/scalekit-sdk-go/v2"
)

// AuthorizeHandler initiates the OAuth2 flow and returns the authorization URL
func AuthorizeHandler(c *gin.Context) {
	next := c.Query("next")
	connectionID := c.Query("connection_id")
	provider := c.Query("provider")
	prompt := c.Query("prompt")
	organizationID := c.Query("organization_id")

	// Get the global ScaleKit client
	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	stateObj := map[string]interface{}{
		"next": next,
		"csrf": randomString(12),
	}
	stateBytes, _ := json.Marshal(stateObj)
	state := base64.StdEncoding.EncodeToString(stateBytes)

	host := c.Request.Host
	protocol := "https"
	if strings.Contains(host, "localhost") {
		protocol = "http"
	}
	redirectURL := protocol + "://" + host + "/api/scalekit/callback"

	options := scalekit.AuthorizationUrlOptions{
		State:  state,
		Scopes: []string{"openid", "profile", "email", "offline_access"},
	}
	if connectionID != "" {
		options.ConnectionId = connectionID
	}
	if provider != "" {
		options.Provider = provider
	}
	if prompt != "" {
		options.Prompt = prompt
	}
	if organizationID != "" {
		options.OrganizationId = organizationID
	}
	authURL, err := scalekitClient.GetAuthorizationUrl(redirectURL, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authorization URL"})
		return
	}

	c.Redirect(http.StatusFound, authURL.String())
}

func randomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
