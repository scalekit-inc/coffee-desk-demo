package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"coffee-desk-demo/database"

	"github.com/gin-gonic/gin"
	usersv1 "github.com/scalekit-inc/scalekit-sdk-go/v2/pkg/grpc/scalekit/v1/users"
)

// OnboardingHandler handles the onboarding process for organizations
func OnboardingHandler(c *gin.Context) {
	// Get the ID token from the cookie
	idToken, err := c.Cookie("id_token")
	if err != nil {
		log.Printf("No ID token found in cookies: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Parse the request body
	var requestBody struct {
		WorkspaceName string `json:"workspaceName" binding:"required"`
		FirstName     string `json:"firstName"`
		LastName      string `json:"lastName"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		log.Printf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Decode the JWT token to extract the organization ID (oid) and user ID (sub)
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		log.Printf("Invalid ID token format")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid ID token"})
		return
	}

	// Decode the payload (second part)
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

	// Extract the organization ID (oid) from token claims
	organizationID, ok := claims["oid"].(string)
	if !ok {
		log.Printf("No organization ID (oid) found in token claims")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims - no organization ID"})
		return
	}

	// Extract the user ID (sub) from token claims
	userID, ok := claims["sub"].(string)
	if !ok {
		log.Printf("No user ID (sub) found in token claims")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims - no user ID"})
		return
	}

	// Get the global ScaleKit client
	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		log.Printf("Failed to get ScaleKit client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Step 1: Update user profile if first name or last name is provided
	if requestBody.FirstName != "" || requestBody.LastName != "" {
		updateUserRequest := &usersv1.UpdateUser{
			UserProfile: &usersv1.UpdateUserProfile{},
		}

		// Only set fields if they are provided
		if requestBody.FirstName != "" {
			updateUserRequest.UserProfile.FirstName = &requestBody.FirstName
		}
		if requestBody.LastName != "" {
			updateUserRequest.UserProfile.LastName = &requestBody.LastName
		}

		_, err = scalekitClient.User().UpdateUser(context.Background(), userID, updateUserRequest)
		if err != nil {
			log.Printf("Error updating user profile: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
			return
		}

	}

	// Step 2: Create the organization in the local database
	localOrg := database.Organization{
		ExternalID: organizationID,
		Name:       requestBody.WorkspaceName,
	}

	if err := database.DB.Create(&localOrg).Error; err != nil {
		log.Printf("Error creating organization in local database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create organization"})
		return
	}

	// Step 3: Link the local organization to the ScaleKit user via metadata
	orgIDStr := localOrg.ID.String()
	metadataUpdate := &usersv1.UpdateUser{
		Metadata: map[string]string{
			"coffee_org_id": orgIDStr,
		},
	}

	if _, err := scalekitClient.User().UpdateUser(context.Background(), userID, metadataUpdate); err != nil {
		log.Printf("Error updating user metadata with coffee_org_id: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link organization to user"})
		return
	}

	// Create user in local database (we need to get user email from Scalekit)
	// For now, we'll create with empty email and update it later when we have it
	localUser := database.User{
		ExternalID: userID,
		Email:      "", // We'll update this when we get user info from Scalekit
	}

	if err := database.DB.Create(&localUser).Error; err != nil {
		log.Printf("Error creating user in local database: %v", err)
		// Continue even if local DB creation fails
	} else {
		// Update ScaleKit to set external_id to local database ID
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

	// Try to get user email from Scalekit and update local record
	userResponse, err := scalekitClient.User().GetUser(context.Background(), userID)
	if err == nil && userResponse.User.Email != "" {
		if err := database.DB.Model(&database.User{}).Where("external_id = ?", userID).Update("email", userResponse.User.Email).Error; err != nil {
			log.Printf("Error updating user email in local database: %v", err)
			// Continue even if email update fails
		}
	}

	// Return the updated organization details
	c.JSON(http.StatusOK, gin.H{
		"message": "Onboarding completed successfully",
		"organization": gin.H{
			"id":           localOrg.ID.String(),
			"display_name": localOrg.Name,
			"external_id":  localOrg.ExternalID,
		},
		"user_updated": requestBody.FirstName != "" || requestBody.LastName != "",
	})
}
