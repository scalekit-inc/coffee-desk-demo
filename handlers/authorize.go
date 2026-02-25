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

const (
	idpInitiatedLoginParam = "idp_initiated_login"
	scalekitCallbackPath   = "/api/scalekit/callback"
)

// getCallbackRedirectURL returns the full URL for the ScaleKit OAuth callback using the request host.
func getCallbackRedirectURL(c *gin.Context) string {
	host := c.Request.Host
	protocol := "https"
	if strings.Contains(host, "localhost") {
		protocol = "http"
	}
	return protocol + "://" + host + scalekitCallbackPath
}

// IdpInitiatedLoginHandler handles IDP-initiated SSO: validates the JWT via ScaleKit SDK,
// then builds authorization URL options from the claims and redirects to ScaleKit using GetAuthorizationUrl.
func IdpInitiatedLoginHandler(c *gin.Context) {
	rawToken := c.Query(idpInitiatedLoginParam)
	if rawToken == "" {
		// return
	}

	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	claims := &scalekit.IdpInitiatedLoginClaims{}
	if rawToken != "" {
		var err error
		claims, err = scalekitClient.GetIdpInitiatedLoginClaims(c.Request.Context(), rawToken)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid idp_initiated_login token: " + err.Error()})
			return
		}
	}

	next := c.Query("next")
	if claims.RelayState != nil && *claims.RelayState != "" {
		next = *claims.RelayState
	}
	stateObj := map[string]interface{}{
		"next": next,
		"csrf": randomString(12),
	}
	stateBytes, _ := json.Marshal(stateObj)
	state := base64.StdEncoding.EncodeToString(stateBytes)

	redirectURL := getCallbackRedirectURL(c)

	options := scalekit.AuthorizationUrlOptions{
		State:  state,
		Scopes: []string{"openid", "profile", "email", "offline_access"},
	}
	if claims.OrganizationID != "" {
		options.OrganizationId = claims.OrganizationID
	}
	if claims.ConnectionID != "" {
		options.ConnectionId = claims.ConnectionID
	}
	if claims.LoginHint != "" {
		options.LoginHint = claims.LoginHint
	}
	authURL, err := scalekitClient.GetAuthorizationUrl(redirectURL, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authorization URL"})
		return
	}

	c.Redirect(http.StatusFound, authURL.String())
}

// AuthorizeHandler initiates the OAuth2 flow and returns the authorization URL
func AuthorizeHandler(c *gin.Context) {
	next := c.Query("next")
	connectionID := c.Query("connection_id")
	provider := c.Query("provider")
	prompt := c.Query("prompt")
	organizationID := c.Query("organization_id")
	loginHint := c.Query("login_hint")

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

	redirectURL := getCallbackRedirectURL(c)

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
	if loginHint != "" {
		options.LoginHint = loginHint
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
