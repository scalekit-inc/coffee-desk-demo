package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// CallbackHandler handles the OAuth2 callback, exchanges code for tokens, and fetches user info
func CallbackHandler(c *gin.Context) {
	code := c.Query("code")
	errorMsg := c.Query("error")
	errorDesc := c.Query("error_description")

	if errorMsg != "" {
		log.Printf("Error in callback: %s - %s", errorMsg, errorDesc)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorDesc})
		return
	}

	envURL := os.Getenv("SCALEKIT_ENVIRONMENT_URL")
	clientID := os.Getenv("SCALEKIT_CLIENT_ID")
	clientSecret := os.Getenv("SCALEKIT_CLIENT_SECRET")

	redirectURL := getUIBaseURL(c) + "/api/scalekit/callback"

	log.Printf("Using redirect URL: %s", redirectURL)

	// Exchange code for token
	tokenURL := envURL + "/oauth/token"
	tokenResp, err := http.Post(
		tokenURL,
		"application/x-www-form-urlencoded",
		strings.NewReader(encodeParams(map[string]string{
			"code":          code,
			"redirect_uri":  redirectURL,
			"client_id":     clientID,
			"client_secret": clientSecret,
			"grant_type":    "authorization_code",
		})),
	)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token"})
		return
	}
	defer tokenResp.Body.Close()
	if tokenResp.StatusCode != 200 {
		body, readErr := io.ReadAll(tokenResp.Body)
		bodyStr := string(body)
		if readErr != nil {
			bodyStr = "Failed to read response body: " + readErr.Error()
		}
		log.Printf("Token exchange failed with status %d. URL: %s, Response body: %s", tokenResp.StatusCode, tokenURL, bodyStr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token"})
		return
	}

	var tokenData map[string]interface{}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
		log.Printf("Error decoding token response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode token response"})
		return
	}

	accessToken, _ := tokenData["access_token"].(string)
	idToken, _ := tokenData["id_token"].(string)
	refreshToken, _ := tokenData["refresh_token"].(string)

	log.Printf("Successfully obtained tokens")

	// Fetch user info
	userReq, _ := http.NewRequest("GET", envURL+"/userinfo", nil)
	userReq.Header.Set("Authorization", "Bearer "+accessToken)
	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil || userResp.StatusCode != 200 {
		log.Printf("Error fetching user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer userResp.Body.Close()

	userBody, _ := io.ReadAll(userResp.Body)
	var userInfo map[string]interface{}
	if err := json.Unmarshal(userBody, &userInfo); err != nil {
		log.Printf("Error decoding user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}

	log.Printf("Successfully fetched user info")

	// TODO: Handle the user info and tokens as needed
	// For demo: set cookies and redirect
	// In production, to set secure, httpOnly, and other cookie options
	c.SetCookie("auth_access_token", accessToken, 86400, "/", "", false, true)
	c.SetCookie("auth_refresh_token", refreshToken, 2592000, "/", "", false, true)
	c.SetCookie("id_token", idToken, 86400, "/", "", false, false)

	// Get UI URL from environment variable, fallback to default if not set
	uiURL := os.Getenv("UI_URL")
	if uiURL == "" {
		// Use the request origin as fallback for production
		protocol := "https"
		if strings.Contains(c.Request.Host, "localhost") {
			protocol = "http"
		}
		uiURL = protocol + "://" + c.Request.Host
	}

	// Decode the access token to check for xoid
	token, _, err := jwt.NewParser().ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		log.Printf("Error parsing access token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse access token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("Error getting token claims")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token claims"})
		return
	}

	// Check if xoid exists in the token
	_, hasXoid := claims["xoid"]
	redirectPath := "/onboarding"
	if hasXoid {
		redirectPath = "/dashboard"
	}

	// Redirect to the appropriate path
	c.Redirect(http.StatusFound, uiURL+redirectPath)
}
