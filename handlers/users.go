package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"coffee-desk-demo/database"

	"github.com/gin-gonic/gin"
	commons "github.com/scalekit-inc/scalekit-sdk-go/v2/pkg/grpc/scalekit/v1/commons"
	usersv1 "github.com/scalekit-inc/scalekit-sdk-go/v2/pkg/grpc/scalekit/v1/users"
)

// ScaleKitConfig holds the configuration for ScaleKit client
type ScaleKitConfig struct {
	EnvURL       string
	ClientID     string
	ClientSecret string
}

// getScaleKitConfig retrieves and validates ScaleKit configuration from environment variables
func getScaleKitConfig() (*ScaleKitConfig, error) {
	envURL := os.Getenv("SCALEKIT_ENVIRONMENT_URL")
	clientID := os.Getenv("SCALEKIT_CLIENT_ID")
	clientSecret := os.Getenv("SCALEKIT_CLIENT_SECRET")

	if envURL == "" || clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("missing ScaleKit environment variables")
	}

	return &ScaleKitConfig{
		EnvURL:       envURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}

// UserInfo represents the user information from the token
type UserInfo struct {
	OrganizationID string
}

// extractOrganizationID extracts the organization ID from the access token
func extractOrganizationID(c *gin.Context) (*UserInfo, error) {
	// Get the access token from the cookie
	accessToken, err := c.Cookie("auth_access_token")
	if err != nil {
		return nil, fmt.Errorf("no access token found in cookies: %w", err)
	}

	// Decode JWT and extract claims
	claims, err := decodeJwtPayload(accessToken)
	if err != nil {
		return nil, err
	}

	// Extract the organization ID (oid) from token claims
	organizationID, err := getStringClaim(claims, "oid")
	if err != nil {
		return nil, fmt.Errorf("no organization ID (oid) found in token claims")
	}

	return &UserInfo{
		OrganizationID: organizationID,
	}, nil
}

// handleAuthError handles authentication errors consistently
func handleAuthError(c *gin.Context, err error) {
	log.Printf("Authentication error: %v", err)
	c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
}

// handleInternalError handles internal server errors consistently
func handleInternalError(c *gin.Context, err error) {
	log.Printf("Internal server error: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}

// GetWorkspaceMembersHandler fetches all members of the workspace using the organization ID from the user's token
func GetWorkspaceMembersHandler(c *gin.Context) {
	// Authenticate user and get organization ID
	userInfo, err := extractOrganizationID(c)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	// Get the global ScaleKit client
	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		handleInternalError(c, err)
		return
	}

	// Use the SDK to list organization users
	membersResponse, err := scalekitClient.User().ListOrganizationUsers(context.Background(), userInfo.OrganizationID, nil)
	if err != nil {
		log.Printf("Error fetching workspace members using SDK: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workspace members"})
		return
	}

	// Transform only the membership status to ensure it's a string
	transformedUsers := make([]map[string]interface{}, len(membersResponse.Users))
	for i, user := range membersResponse.Users {
		userMap := map[string]interface{}{
			"id":             user.Id,
			"email":          user.Email,
			"create_time":    user.CreateTime,
			"update_time":    user.UpdateTime,
			"environment_id": user.EnvironmentId,
			"external_id":    user.ExternalId,
			"metadata":       user.Metadata,
			"user_profile":   user.UserProfile,
		}

		// Transform memberships to ensure status is a string
		if user.Memberships != nil {
			memberships := make([]map[string]interface{}, len(user.Memberships))
			for j, membership := range user.Memberships {
				membershipMap := map[string]interface{}{
					"organization_id":   membership.OrganizationId,
					"membership_status": membership.MembershipStatus.String(), // Convert enum to string
					"name":              membership.Name,
					"roles":             membership.Roles,
				}

				// Include ExpiresAt field if membership status is PENDING_INVITE
				if membership.MembershipStatus.String() == "PENDING_INVITE" && membership.ExpiresAt != nil {
					// Convert protobuf timestamp to RFC 3339 string format
					membershipMap["expires_at"] = membership.ExpiresAt.AsTime().Format(time.RFC3339)
				}

				memberships[j] = membershipMap
			}
			userMap["memberships"] = memberships
		} else {
			userMap["memberships"] = []map[string]interface{}{}
		}

		transformedUsers[i] = userMap
	}

	// Return the response with transformed users
	c.JSON(http.StatusOK, gin.H{
		"next_page_token": membersResponse.NextPageToken,
		"prev_page_token": membersResponse.PrevPageToken,
		"total_size":      membersResponse.TotalSize,
		"users":           transformedUsers,
	})
}

// CreateWorkspaceMemberHandler creates a new user and adds them to the workspace
func CreateWorkspaceMemberHandler(c *gin.Context) {
	// Authenticate user and get organization ID
	userInfo, err := extractOrganizationID(c)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	// Parse the request body
	var requestBody struct {
		Email               string  `json:"email" binding:"required,email"`
		FirstName           *string `json:"firstName,omitempty"`
		LastName            *string `json:"lastName,omitempty"`
		SendInvitationEmail *bool   `json:"sendInvitationEmail,omitempty"`
		Role                string  `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		log.Printf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get the global ScaleKit client
	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		handleInternalError(c, err)
		return
	}

	// Create the user with membership
	user := &usersv1.CreateUser{
		Email:       requestBody.Email,
		UserProfile: &usersv1.CreateUserProfile{},
		Membership: &usersv1.CreateMembership{
			Roles: []*commons.Role{
				{
					Name: requestBody.Role,
				},
			},
		},
	}

	// Only set fields if they are provided
	if requestBody.FirstName != nil {
		user.UserProfile.FirstName = *requestBody.FirstName
	}
	if requestBody.LastName != nil {
		user.UserProfile.LastName = *requestBody.LastName
	}

	// Default to true if not provided
	sendInvitationEmail := true
	if requestBody.SendInvitationEmail != nil {
		sendInvitationEmail = *requestBody.SendInvitationEmail
	}

	// Create user in Scalekit first
	response, err := scalekitClient.User().CreateUserAndMembership(context.Background(), userInfo.OrganizationID, user, sendInvitationEmail)
	if err != nil {
		log.Printf("Error creating user and membership: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Insert user into local database
	localUser := database.User{
		ExternalID: response.User.Id,
		Email:      response.User.Email,
	}

	if err := database.DB.Create(&localUser).Error; err != nil {
		log.Printf("Error creating user in local database: %v", err)
		// Don't fail the request, just log the error

	}

	c.JSON(http.StatusOK, gin.H{"user_id": response.User.Id})
}

// DeleteWorkspaceMemberHandler deletes a user from the workspace
func DeleteWorkspaceMemberHandler(c *gin.Context) {
	if _, err := extractOrganizationID(c); err != nil {
		handleAuthError(c, err)
		return
	}

	memberID := c.Param("member_id")
	if memberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "member_id is required"})
		return
	}

	// Get the global ScaleKit client
	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		handleInternalError(c, err)
		return
	}

	// Delete the user from Scalekit
	if err := scalekitClient.User().DeleteUser(context.Background(), memberID); err != nil {
		log.Printf("Error deleting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	// Delete user from local database
	if err := database.DB.Where("external_id = ?", memberID).Delete(&database.User{}).Error; err != nil {
		log.Printf("Error deleting user from local database: %v", err)
		// Don't fail the request, just log the error
		// The user was deleted from Scalekit successfully
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ResendInviteHandler resends an invitation email for a pending member in the organization
func ResendInviteHandler(c *gin.Context) {
	// Ensure user is authenticated (we don't strictly require org from token to match request now)
	if _, err := extractOrganizationID(c); err != nil {
		handleAuthError(c, err)
		return
	}

	// Parse the request body
	var req struct {
		OrganizationID string `json:"organization_id" binding:"required"`
		UserID         string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid resend invite request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization_id and user_id are required"})
		return
	}

	// Get ScaleKit client
	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		handleInternalError(c, err)
		return
	}

	// Call SDK to resend invite
	if _, err := scalekitClient.User().ResendInvite(context.Background(), req.OrganizationID, req.UserID); err != nil {
		log.Printf("Error resending invite (org=%s user=%s): %v", req.OrganizationID, req.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resend invite"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invite resent successfully"})
}

// UpdateUserProfileHandler handles updating user profile information
func UpdateUserProfileHandler(c *gin.Context) {
	// Get the access token from the cookie
	accessToken, err := c.Cookie("auth_access_token")
	if err != nil {
		log.Printf("No access token found in cookies: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Parse the request body
	var requestBody struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Name      string `json:"name"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		log.Printf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Decode JWT and extract claims
	claims, err := decodeJwtPayload(accessToken)
	if err != nil {
		log.Printf("Error decoding token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Extract the user ID (sub) from token claims
	userID, err := getStringClaim(claims, "sub")
	if err != nil {
		log.Printf("No user ID (sub) found in token claims")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims - no user ID"})
		return
	}

	// Get the global ScaleKit client
	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		handleInternalError(c, err)
		return
	}

	// Create the update user request
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
	if requestBody.Name != "" {
		updateUserRequest.UserProfile.Name = &requestBody.Name
	}

	// Update the user using the SDK
	response, err := scalekitClient.User().UpdateUser(context.Background(), userID, updateUserRequest)
	if err != nil {
		log.Printf("Error updating user profile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}

	// Update user in local database (only email if it changed)
	if response.User.Email != "" {
		if err := database.DB.Model(&database.User{}).Where("external_id = ?", userID).Update("email", response.User.Email).Error; err != nil {
			log.Printf("Error updating user email in local database: %v", err)
			// Don't fail the request, just log the error
		}
	}

	log.Printf("Updated user profile for user %s", userID)

	// Return the updated user details
	c.JSON(http.StatusOK, gin.H{
		"message": "User profile updated successfully",
		"user": gin.H{
			"id":         response.User.Id,
			"email":      response.User.Email,
			"first_name": response.User.UserProfile.FirstName,
			"last_name":  response.User.UserProfile.LastName,
			"name":       response.User.UserProfile.Name,
		},
	})
}
