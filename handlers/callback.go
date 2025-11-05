package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"coffee-desk-demo/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	usersv1 "github.com/scalekit-inc/scalekit-sdk-go/v2/pkg/grpc/scalekit/v1/users"
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

	// Check if xoid and xuid exist in the token
	_, hasXoid := claims["xoid"]
	_, hasXuid := claims["xuid"]
	redirectPath := "/onboarding"
	if hasXoid {
		redirectPath = "/dashboard"

		// If xoid is present but xuid is not present, create user in local database
		// and update external_id in ScaleKit
		if !hasXuid {
			// Extract user ID (sub) from token claims
			userID, ok := claims["sub"].(string)
			if ok && userID != "" {
				// Check if user already exists in local database
				var localUser database.User
				err := database.DB.Where("external_id = ?", userID).First(&localUser).Error

				userCreated := false
				if err != nil {
					// User doesn't exist, create it
					// Get user email from userInfo first, or fetch from ScaleKit if not available
					email, _ := userInfo["email"].(string)
					if email == "" {
						// Fetch email from ScaleKit API
						scalekitClient, err := GetScaleKitClient()
						if err == nil {
							userResponse, err := scalekitClient.User().GetUser(context.Background(), userID)
							if err == nil && userResponse.User.Email != "" {
								email = userResponse.User.Email
							}
						}
					}

					// If we still don't have an email, log an error and skip user creation
					if email == "" {
						log.Printf("Unable to get email for user %s from userInfo or ScaleKit API", userID)
						// Skip user creation and ScaleKit update
					} else {
						localUser = database.User{
							ExternalID: userID,
							Email:      email,
						}

						if err := database.DB.Create(&localUser).Error; err != nil {
							log.Printf("Error creating user in local database: %v", err)
							// Continue even if local DB creation fails, but skip ScaleKit update
						} else {
							userCreated = true
						}
					}
				} else {
					userCreated = true
				}

				// Update ScaleKit to set external_id to local database ID
				// This should be done whether the user was just created or already existed
				if userCreated {
					scalekitClient, err := GetScaleKitClient()
					if err == nil {
						localUserIDStr := localUser.ID.String()
						updateUserRequest := &usersv1.UpdateUser{
							ExternalId: &localUserIDStr,
						}

						if _, err := scalekitClient.User().UpdateUser(context.Background(), userID, updateUserRequest); err != nil {
							log.Printf("Error updating external_id in ScaleKit: %v", err)
							// Continue even if ScaleKit update fails
						} else {
							log.Printf("Successfully updated external_id in ScaleKit for user %s", userID)
						}
					}
				}
			}
		}
	}

	// Redirect to the appropriate path
	c.Redirect(http.StatusFound, uiURL+redirectPath)
}
