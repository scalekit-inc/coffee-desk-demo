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
	organizationsv1 "github.com/scalekit-inc/scalekit-sdk-go/v2/pkg/grpc/scalekit/v1/organizations"
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

	// Generate the external ID in the format wspace_id_<org_id_suffix>
	// If org_id is org_59615193906282635, external_id will be wspace_id_59615193906282635
	externalID, err := toExternalWorkspaceID(organizationID)
	if err != nil {
		log.Printf("Invalid organization ID format: %s", organizationID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID format"})
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

	// Step 2: Update the organization
	updateOrg := &organizationsv1.UpdateOrganization{
		DisplayName: &requestBody.WorkspaceName,
		ExternalId:  &externalID,
	}

	// Update the organization using the SDK
	response, err := scalekitClient.Organization().UpdateOrganization(context.Background(), organizationID, updateOrg)
	if err != nil {
		log.Printf("Error updating organization: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update organization"})
		return
	}

	// Create organization in local database
	localOrg := database.Organization{
		ExternalID: organizationID,
		Name:       response.Organization.DisplayName,
	}

	if err := database.DB.Create(&localOrg).Error; err != nil {
		log.Printf("Error creating organization in local database: %v", err)
		// Don't fail the request, just log the error
		// The organization was updated in Scalekit successfully
	}

	// Create user in local database (we need to get user email from Scalekit)
	// For now, we'll create with empty email and update it later when we have it
	localUser := database.User{
		ExternalID: userID,
		Email:      "", // We'll update this when we get user info from Scalekit
	}

	if err := database.DB.Create(&localUser).Error; err != nil {
		log.Printf("Error creating user in local database: %v", err)

	}

	// Try to get user email from Scalekit and update local record
	userResponse, err := scalekitClient.User().GetUser(context.Background(), userID)
	if err == nil && userResponse.User.Email != "" {
		if err := database.DB.Model(&database.User{}).Where("external_id = ?", userID).Update("email", userResponse.User.Email).Error; err != nil {
			log.Printf("Error updating user email in local database: %v", err)

		}
	}

	// Return the updated organization details
	c.JSON(http.StatusOK, gin.H{
		"message": "Onboarding completed successfully",
		"organization": gin.H{
			"id":           response.Organization.Id,
			"display_name": response.Organization.DisplayName,
			"external_id":  response.Organization.ExternalId,
		},
		"user_updated": requestBody.FirstName != "" || requestBody.LastName != "",
	})
}
