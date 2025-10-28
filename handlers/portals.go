package handlers

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

// GetPortalLinkHandler generates a portal link for the organization
func GetPortalLinkHandler(c *gin.Context) {
	// Authenticate user and get organization ID
	userInfo, err := extractOrganizationID(c)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	organizationID := userInfo.OrganizationID

	// Get the global ScaleKit client
	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		log.Printf("Failed to get ScaleKit client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Generate the portal link
	link, err := scalekitClient.Organization().GeneratePortalLink(context.Background(), organizationID)
	if err != nil {
		log.Printf("Error generating portal link: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate portal link"})
		return
	}

	// Replace the host with SCALEKIT_BASE_ENVIRONMENT_URL if available
	modifiedLink := link.Location
	baseEnvURL := os.Getenv("SCALEKIT_BASE_ENVIRONMENT_URL")
	if baseEnvURL != "" {
		// Parse the original URL
		parsedURL, err := url.Parse(link.Location)
		if err == nil {
			// Parse the base environment URL
			baseURL, err := url.Parse(baseEnvURL)
			if err == nil {
				// Replace the host and scheme
				parsedURL.Scheme = baseURL.Scheme
				parsedURL.Host = baseURL.Host
				modifiedLink = parsedURL.String()
				log.Printf("Modified portal link from %s to %s", link.Location, modifiedLink)
			}
		}
	}

	// Return the portal link with expiry information
	c.JSON(http.StatusOK, gin.H{
		"link":       modifiedLink,
		"expires_at": link.ExpireTime,
	})
}
