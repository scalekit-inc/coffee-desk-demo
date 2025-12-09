package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/scalekit-inc/scalekit-sdk-go/v2"
)

// SessionHandler verifies the user's session and returns their information
func SessionHandler(c *gin.Context) {
	// Get the access token and refresh token from cookies
	accessToken, err := c.Cookie("auth_access_token")
	if err != nil {
		log.Printf("No access token found in cookies: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	refreshToken, err := c.Cookie("auth_refresh_token")
	if err != nil {
		log.Printf("No refresh token found in cookies: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Get the global ScaleKit client
	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		log.Printf("Failed to get ScaleKit client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Validate the access token using the SDK
	log.Printf("Validating access token...")

	isValid, err := scalekitClient.ValidateAccessToken(accessToken)
	if err != nil {
		log.Printf("Error validating access token: %v", err)
		// If validation fails (e.g., token expired), treat it as invalid and try to refresh
		isValid = false
	}

	// If access token is not valid, try to refresh it
	if !isValid {
		log.Printf("Access token is invalid, attempting to refresh")

		// Call RefreshAccessToken SDK method
		refreshResponse, err := scalekitClient.RefreshAccessToken(refreshToken)
		if err != nil {
			log.Printf("Error refreshing token: %v", err)
			// If refresh fails, call LogoutHandler
			LogoutHandler(c)
			return
		}

		// Update cookies with new tokens
		c.SetCookie("auth_access_token", refreshResponse.AccessToken, 86400, "/", "", false, true)
		c.SetCookie("auth_refresh_token", refreshResponse.RefreshToken, 2592000, "/", "", false, true)

		// Update access token for further use
		accessToken = refreshResponse.AccessToken
	}

	// Extract user ID and organization ID from access token
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		log.Printf("Invalid access token format")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
		return
	}

	// Decode the payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		log.Printf("Error decoding token payload: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode token"})
		return
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		log.Printf("Error parsing token claims: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token claims"})
		return
	}

	userId, ok := claims["sub"].(string)
	if !ok {
		log.Printf("No sub claim found in token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	// Get user information using the SDK
	userResponse, err := scalekitClient.User().GetUser(context.Background(), userId)
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	// Create simplified user object with required fields
	user := gin.H{
		"id":          userResponse.User.Id,
		"email":       userResponse.User.Email,
		"first_name":  userResponse.User.UserProfile.FirstName,
		"last_name":   userResponse.User.UserProfile.LastName,
		"name":        userResponse.User.UserProfile.Name,
		"roles":       []string{"admin"},
		"is_admin":    true,
		"permissions": []string{"*"},
	}

	// Create workspaces list from memberships
	workspaces := make([]gin.H, 0)
	for idx, membership := range userResponse.User.Memberships {
		workspace := gin.H{
			"id":           membership.OrganizationId,
			"display_name": membership.Name,
			"is_current":   idx == 0,
		}
		workspaces = append(workspaces, workspace)
	}

	// Return the simplified response
	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user":          user,
		"workspaces":    workspaces,
	})
}

// LogoutHandler clears the session cookies and redirects to ScaleKit logout URL
func LogoutHandler(c *gin.Context) {
	// Get the ID token from cookies
	idToken, err := c.Cookie("id_token")
	if err != nil {
		log.Printf("No ID token found in cookies: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Get the global ScaleKit client
	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		log.Printf("Failed to get ScaleKit client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Generate logout URL with options
	options := scalekit.LogoutUrlOptions{
		IdTokenHint:           idToken,
		PostLogoutRedirectUri: getUIBaseURL(c),
	}

	logoutURL, err := scalekitClient.GetLogoutUrl(options)
	if err != nil {
		log.Printf("Error generating logout URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate logout URL"})
		return
	}

	c.SetCookie("auth_access_token", "", -1, "/", "", false, true)
	c.SetCookie("auth_refresh_token", "", -1, "/", "", false, true)
	c.SetCookie("id_token", "", -1, "/", "", false, false)

	// Redirect to ScaleKit logout URL
	c.Redirect(http.StatusFound, logoutURL.String())
}
