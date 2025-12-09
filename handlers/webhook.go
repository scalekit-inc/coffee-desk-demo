package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	svix "github.com/svix/svix-webhooks/go"

	"coffee-desk-demo/database" // Import your database package
)

// Webhook event structures (keep these)
type ScalekitWebhookEvent struct {
	ID             string          `json:"id"`
	Type           string          `json:"type"`
	SpecVersion    string          `json:"spec_version"`
	OccurredAt     time.Time       `json:"occurred_at"`
	EnvironmentID  string          `json:"environment_id"`
	OrganizationID string          `json:"organization_id"`
	Object         string          `json:"object"`
	Data           json.RawMessage `json:"data"`
}

type DirectoryUser struct {
	ID                string                 `json:"id"`
	OrganizationID    string                 `json:"organization_id"`
	DpID              string                 `json:"dp_id"` // IdP User ID
	Email             string                 `json:"email"`
	Active            bool                   `json:"active"`
	Name              string                 `json:"name"`
	GivenName         string                 `json:"given_name"`
	FamilyName        string                 `json:"family_name"`
	PreferredUsername string                 `json:"preferred_username"`
	PhoneNumber       string                 `json:"phone_number"`
	Roles             []UserRole             `json:"roles"`
	Groups            []UserGroup            `json:"groups"`
	CustomAttributes  map[string]interface{} `json:"custom_attributes"`
	RawAttributes     map[string]interface{} `json:"raw_attributes"`
}

type UserRole struct {
	RoleName string `json:"role_name"`
}

type UserGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DirectoryGroup struct {
	ID             string                 `json:"id"`
	DirectoryID    string                 `json:"directory_id"`
	OrganizationID string                 `json:"organization_id"`
	DisplayName    string                 `json:"display_name"`
	ExternalID     string                 `json:"external_id"`
	RawAttributes  map[string]interface{} `json:"raw_attributes"`
}

// ScalekitWebhookHandler - Main webhook endpoint handler
func ScalekitWebhookHandler(c *gin.Context) {
	// Step 1: Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("❌ Error reading webhook body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading request body"})
		return
	}

	// Step 2: Get webhook secret
	webhookSecret := os.Getenv("SCALEKIT_WEBHOOK_SECRET")

	// Step 3: Verify webhook signature (if secret is configured)
	if webhookSecret != "" {
		log.Printf("🔐 Verifying webhook signature...")

		// Extract Svix headers
		headers := http.Header{}
		headers.Set("svix-id", c.GetHeader("svix-id"))
		headers.Set("svix-timestamp", c.GetHeader("svix-timestamp"))
		headers.Set("svix-signature", c.GetHeader("svix-signature"))

		log.Printf("🔍 Svix headers:")
		log.Printf("  svix-id: %s", c.GetHeader("svix-id"))
		log.Printf("  svix-timestamp: %s", c.GetHeader("svix-timestamp"))
		log.Printf("  svix-signature: %s", c.GetHeader("svix-signature"))

		// Create webhook verifier
		wh, err := svix.NewWebhook(webhookSecret)
		if err != nil {
			log.Printf("❌ Error creating webhook verifier: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Verify signature
		err = wh.Verify(body, headers)
		if err != nil {
			log.Printf("❌ Webhook signature verification failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			return
		}

		log.Printf("✅ Webhook signature verified successfully")
	} else {
		log.Printf("⚠️  Warning: SCALEKIT_WEBHOOK_SECRET not set, skipping verification")
	}

	// Step 4: Parse webhook event
	var event ScalekitWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("❌ Error parsing webhook event: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing event"})
		return
	}

	// Step 5: Log the event
	log.Printf("📬 Received webhook: %s for org: %s", event.Type, event.OrganizationID)

	// Step 6: Process asynchronously
	go processWebhookAsync(event)

	// Step 7: Respond immediately with 200 OK
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// Process webhook events asynchronously
func processWebhookAsync(event ScalekitWebhookEvent) {
	switch event.Type {
	case "organization.directory.user_created":
		handleUserCreated(event)
	case "organization.directory.user_updated":
		handleUserUpdated(event)
	case "organization.directory.user_deleted":
		handleUserDeleted(event)
	case "organization.directory_enabled":
		handleDirectoryEnabled(event)
	case "organization.directory_disabled":
		handleDirectoryDisabled(event)
	default:
		log.Printf("⚠️  Unhandled event type: %s", event.Type)
	}
}

// ============================================
// Event Handlers Using GORM
// ============================================

func handleUserCreated(event ScalekitWebhookEvent) {
	var directoryUser DirectoryUser
	if err := json.Unmarshal(event.Data, &directoryUser); err != nil {
		log.Printf("❌ Error parsing user data: %v", err)
		return
	}

	log.Printf("👤 Creating user: %s (%s) for org: %s",
		directoryUser.Email, directoryUser.Name, directoryUser.OrganizationID)

	// First, ensure the organization exists
	if err := ensureOrganizationExists(directoryUser.OrganizationID); err != nil {
		log.Printf("❌ Error ensuring organization exists: %v", err)
		return
	}

	// Create or update user using GORM
	user := database.User{
		ExternalID: directoryUser.ID, // Scalekit user ID
		Email:      directoryUser.Email,
	}

	// Use FirstOrCreate to handle duplicates gracefully
	result := database.DB.Where(database.User{ExternalID: directoryUser.ID}).
		FirstOrCreate(&user)

	if result.Error != nil {
		log.Printf("❌ Error creating user in database: %v", result.Error)
		return
	}

	if result.RowsAffected > 0 {
		log.Printf("✅ User %s created successfully (ID: %s)", directoryUser.Email, user.ID)
	} else {
		log.Printf("ℹ️  User %s already exists (ID: %s)", directoryUser.Email, user.ID)
	}
}

func handleUserUpdated(event ScalekitWebhookEvent) {
	var directoryUser DirectoryUser
	if err := json.Unmarshal(event.Data, &directoryUser); err != nil {
		log.Printf("❌ Error parsing user data: %v", err)
		return
	}

	log.Printf("🔄 Updating user: %s (%s)", directoryUser.Email, directoryUser.Name)

	// Find user by ExternalID (Scalekit user ID)
	var user database.User
	result := database.DB.Where("external_id = ?", directoryUser.ID).First(&user)

	if result.Error != nil {
		log.Printf("❌ User not found for update: %s", directoryUser.Email)
		// User doesn't exist, create it
		handleUserCreated(event)
		return
	}

	// Update user email if changed
	user.Email = directoryUser.Email

	if err := database.DB.Save(&user).Error; err != nil {
		log.Printf("❌ Error updating user: %v", err)
		return
	}

	log.Printf("✅ User %s updated successfully", directoryUser.Email)
}

func handleUserDeleted(event ScalekitWebhookEvent) {
	var userData struct {
		ID             string `json:"id"`
		OrganizationID string `json:"organization_id"`
		DpID           string `json:"dp_id"`
		Email          string `json:"email"`
	}

	if err := json.Unmarshal(event.Data, &userData); err != nil {
		log.Printf("❌ Error parsing user data: %v", err)
		return
	}

	log.Printf("🗑️  Deleting user: %s", userData.Email)

	// Soft delete user by ExternalID (GORM will set deleted_at)
	result := database.DB.Where("external_id = ?", userData.ID).Delete(&database.User{})

	if result.Error != nil {
		log.Printf("❌ Error deleting user: %v", result.Error)
		return
	}

	if result.RowsAffected > 0 {
		log.Printf("✅ User %s deleted successfully", userData.Email)
	} else {
		log.Printf("⚠️  User %s not found for deletion", userData.Email)
	}
}

func handleDirectoryEnabled(event ScalekitWebhookEvent) {
	log.Printf("✅ Directory sync enabled for org: %s", event.OrganizationID)

	// Ensure organization exists when directory is enabled
	if err := ensureOrganizationExists(event.OrganizationID); err != nil {
		log.Printf("❌ Error creating organization: %v", err)
	}
}

func handleDirectoryDisabled(event ScalekitWebhookEvent) {
	log.Printf("⚠️  Directory sync disabled for org: %s", event.OrganizationID)
}

// ============================================
// Helper Functions
// ============================================

// ensureOrganizationExists creates organization if it doesn't exist
func ensureOrganizationExists(externalOrgID string) error {
	var org database.Organization

	// Check if organization exists
	result := database.DB.Where("external_id = ?", externalOrgID).First(&org)

	if result.Error == nil {
		// Organization already exists
		return nil
	}

	// Organization doesn't exist, create it
	org = database.Organization{
		ExternalID: externalOrgID,
		Name:       "Organization " + externalOrgID, // Default name, can be updated later
	}

	if err := database.DB.Create(&org).Error; err != nil {
		return err
	}

	log.Printf("✅ Organization created: %s (ID: %s)", org.Name, org.ID)
	return nil
}
