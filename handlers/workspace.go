package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"

	"coffee-desk-demo/database"

	"github.com/gin-gonic/gin"
	"github.com/scalekit-inc/scalekit-sdk-go/v2"
	commons "github.com/scalekit-inc/scalekit-sdk-go/v2/pkg/grpc/scalekit/v1/commons"
	organizationsv1 "github.com/scalekit-inc/scalekit-sdk-go/v2/pkg/grpc/scalekit/v1/organizations"
	usersv1 "github.com/scalekit-inc/scalekit-sdk-go/v2/pkg/grpc/scalekit/v1/users"
)

// UpdateWorkspaceHandler handles updating workspace/organization details
func UpdateWorkspaceHandler(c *gin.Context) {
	// Authenticate user and get organization ID
	userInfo, err := extractOrganizationID(c)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	// Parse the request body
	var requestBody struct {
		DisplayName string `json:"display_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		log.Printf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate display name
	if requestBody.DisplayName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Display name cannot be empty"})
		return
	}

	// Get the global ScaleKit client
	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		handleInternalError(c, err)
		return
	}

	// Create the update organization request
	updateOrg := &organizationsv1.UpdateOrganization{
		DisplayName: &requestBody.DisplayName,
	}

	// Update the organization using the SDK
	response, err := scalekitClient.Organization().UpdateOrganization(context.Background(), userInfo.OrganizationID, updateOrg)
	if err != nil {
		log.Printf("Error updating organization: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update organization"})
		return
	}

	// Update organization in local database
	if err := database.DB.Model(&database.Organization{}).Where("external_id = ?", userInfo.OrganizationID).Update("name", response.Organization.DisplayName).Error; err != nil {
		log.Printf("Error updating organization in local database: %v", err)
		// Don't fail the request, just log the error
	}

	// Return the updated organization details
	c.JSON(http.StatusOK, gin.H{
		"message": "Workspace updated successfully",
		"workspace": gin.H{
			"id":           response.Organization.Id,
			"display_name": response.Organization.DisplayName,
			"external_id":  response.Organization.ExternalId,
		},
	})
}

// CreateWorkspaceHandler handles creating a new workspace/organization
func CreateWorkspaceHandler(c *gin.Context) {
	// Get the access token from the cookie
	accessToken, err := c.Cookie("auth_access_token")
	if err != nil {
		log.Printf("No access token found in cookies: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Parse the request body
	var requestBody struct {
		WorkspaceName string `json:"workspaceName" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		log.Printf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate workspace name
	if requestBody.WorkspaceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Workspace name cannot be empty"})
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

	// Step 1: Create the organization using the correct method signature
	createOrgResponse, err := scalekitClient.Organization().CreateOrganization(
		context.Background(),
		requestBody.WorkspaceName,
		scalekit.CreateOrganizationOptions{
			// We'll set the external ID after creation
		},
	)
	if err != nil {
		log.Printf("Error creating organization: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create organization"})
		return
	}

	organizationID := createOrgResponse.Organization.Id

	// Step 2: Create membership for the current user
	membership := &usersv1.CreateMembership{
		Roles: []*commons.Role{
			{
				Name: "admin", // Make the creator an admin
			},
		},
	}

	membershipResponse, err := scalekitClient.User().CreateMembership(
		context.Background(),
		organizationID, // organization ID
		userID,         // user ID
		membership,
		false, // sendInvitationEmail
	)
	if err != nil {
		log.Printf("Error creating membership: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create membership"})
		return
	}

	// Update organization with external ID
	// Generate the external ID in the format wspace_id_<org_id_suffix>
	externalID, err := toExternalWorkspaceID(organizationID)
	if err != nil {
		log.Printf("Invalid organization ID format: %s", organizationID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid organization ID format"})
		return
	}

	updateOrgRequest := &organizationsv1.UpdateOrganization{
		ExternalId: &externalID,
	}

	updateOrgResponse, err := scalekitClient.Organization().UpdateOrganization(context.Background(), organizationID, updateOrgRequest)
	if err != nil {
		log.Printf("Error updating organization with external ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update organization"})
		return
	}

	log.Printf("Updated organization with external ID: %s", externalID)

	// Insert organization into local database
	localOrg := database.Organization{
		ExternalID: organizationID,
		Name:       updateOrgResponse.Organization.DisplayName,
	}

	if err := database.DB.Create(&localOrg).Error; err != nil {
		log.Printf("Error creating organization in local database: %v", err)
		// Don't fail the request, just log the error
		// The organization was created in Scalekit successfully
	}

	// Return the created organization details
	c.JSON(http.StatusOK, gin.H{
		"message": "Workspace created successfully",
		"workspace": gin.H{
			"id":           updateOrgResponse.Organization.Id,
			"display_name": updateOrgResponse.Organization.DisplayName,
			"external_id":  updateOrgResponse.Organization.ExternalId,
		},
		"membership": gin.H{
			"user_id": membershipResponse.User.Id,
			"email":   membershipResponse.User.Email,
		},
	})
}

// CreateDomainsHandler handles creating allowed email domains for the workspace
func CreateDomainsHandler(c *gin.Context) {
	// Authenticate user and get organization ID
	userInfo, err := extractOrganizationID(c)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	// Parse the request body
	var requestBody struct {
		Domain string `json:"domain" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		log.Printf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate domain
	if requestBody.Domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain is required"})
		return
	}

	// Get the global ScaleKit client
	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		handleInternalError(c, err)
		return
	}

	// Clean and validate domain
	cleanDomain := strings.TrimSpace(requestBody.Domain)
	if cleanDomain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain cannot be empty"})
		return
	}

	// Remove @ symbol if present
	cleanDomain = strings.TrimPrefix(cleanDomain, "@")

	// Create domain using ScaleKit SDK with ALLOWED_EMAIL_DOMAIN type
	response, err := scalekitClient.Domain().CreateDomain(
		context.Background(),
		userInfo.OrganizationID,
		cleanDomain,
		&scalekit.CreateDomainOptions{
			DomainType: scalekit.DomainTypeAllowedEmail,
		},
	)

	if err != nil {
		log.Printf("Error creating domain %s: %v", cleanDomain, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create domain"})
		return
	}

	// Return the created domain details
	c.JSON(http.StatusOK, gin.H{
		"message": "Domain created successfully",
		"domain": gin.H{
			"id":   response.Domain.Id,
			"name": response.Domain.Domain,
			"type": response.Domain.DomainType.String(),
		},
	})
}

// GetDomainsHandler handles fetching all domains for the workspace
func GetDomainsHandler(c *gin.Context) {
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

	// List domains using the SDK
	response, err := scalekitClient.Domain().ListDomains(
		context.Background(),
		userInfo.OrganizationID,
	)
	if err != nil {
		log.Printf("Error fetching domains: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch domains"})
		return
	}

	// Transform domains to a simpler format, filtering only ALLOWED_EMAIL_DOMAIN types
	// This ensures we only show domains that are relevant for email domain restrictions
	var domains []gin.H
	for _, domain := range response.Domains {
		// Only include domains of type ALLOWED_EMAIL_DOMAIN
		// Filter out ORGANIZATION_DOMAIN and DOMAIN_TYPE_UNSPECIFIED
		if domain.DomainType.String() == "ALLOWED_EMAIL_DOMAIN" {
			domains = append(domains, gin.H{
				"id":         domain.Id,
				"name":       domain.Domain,
				"type":       domain.DomainType.String(),
				"created_at": domain.CreateTime,
			})
		}
	}

	// Return the domains list
	c.JSON(http.StatusOK, gin.H{
		"domains": domains,
	})
}

// DeleteDomainHandler handles deleting a domain from the workspace
func DeleteDomainHandler(c *gin.Context) {
	// Authenticate user and get organization ID
	userInfo, err := extractOrganizationID(c)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	// Get domain ID from URL parameter
	domainID := c.Param("domain_id")
	if domainID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain ID is required"})
		return
	}

	// Get the global ScaleKit client
	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		handleInternalError(c, err)
		return
	}

	// Delete domain using the SDK
	log.Printf("Attempting to delete domain %s from organization %s", domainID, userInfo.OrganizationID)

	// Correct parameter order: ctx, id, organizationId
	err = scalekitClient.Domain().DeleteDomain(
		context.Background(),
		domainID,
		userInfo.OrganizationID,
	)
	if err != nil {
		log.Printf("Error deleting domain %s: %v", domainID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete domain"})
		return
	}

	log.Printf("Successfully deleted domain %s", domainID)

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Domain deleted successfully",
	})
}
