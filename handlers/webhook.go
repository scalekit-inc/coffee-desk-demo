package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	svix "github.com/svix/svix-webhooks/go"

	"coffee-desk-demo/database"
)

// Webhook event structures
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
	DpID              string                 `json:"dp_id"`
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

// ScalekitWebhookHandler - Main webhook endpoint handler
func ScalekitWebhookHandler(c *gin.Context) {
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("🎯 WEBHOOK HANDLER STARTED")
	log.Printf("   Method: %s", c.Request.Method)
	log.Printf("   Path: %s", c.Request.URL.Path)
	log.Printf("   Full URL: %s", c.Request.URL.String())
	log.Printf("   Remote Address: %s", c.Request.RemoteAddr)
	log.Printf("   Content-Type: %s", c.GetHeader("Content-Type"))
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			log.Printf("🚨 PANIC RECOVERED: %v", r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
	}()

	// Step 1: Read request body
	log.Printf("📖 Step 1: Reading request body...")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("❌ Error reading webhook body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading request body"})
		return
	}
	log.Printf("✅ Body read successfully (%d bytes)", len(body))
	log.Printf("📦 Body content: %s", string(body))

	// Step 2: Get webhook secret
	log.Printf("🔑 Step 2: Getting webhook secret...")
	webhookSecret := os.Getenv("SCALEKIT_WEBHOOK_SECRET")
	if webhookSecret != "" {
		log.Printf("✅ Webhook secret found (length: %d)", len(webhookSecret))
	} else {
		log.Printf("⚠️  No webhook secret configured")
	}

	// Step 3: Verify webhook signature (if secret is configured)
	if webhookSecret != "" {
		log.Printf("🔐 Step 3: Verifying webhook signature...")

		// Extract Svix headers
		headers := http.Header{}
		svixId := c.GetHeader("svix-id")
		svixTimestamp := c.GetHeader("svix-timestamp")
		svixSignature := c.GetHeader("svix-signature")

		headers.Set("svix-id", svixId)
		headers.Set("svix-timestamp", svixTimestamp)
		headers.Set("svix-signature", svixSignature)

		log.Printf("   svix-id: %s", svixId)
		log.Printf("   svix-timestamp: %s", svixTimestamp)
		log.Printf("   svix-signature: %s", svixSignature)

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
		log.Printf("⚠️  Step 3: Skipping signature verification (no secret configured)")
	}

	// Step 4: Parse webhook event
	log.Printf("🔍 Step 4: Parsing webhook event...")
	var event ScalekitWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("❌ Error parsing webhook event: %v", err)
		log.Printf("   Raw body: %s", string(body))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing event"})
		return
	}
	log.Printf("✅ Event parsed successfully")
	log.Printf("   Event ID: %s", event.ID)
	log.Printf("   Event Type: %s", event.Type)
	log.Printf("   Organization ID: %s", event.OrganizationID)
	log.Printf("   Environment ID: %s", event.EnvironmentID)

	// Step 5: Process asynchronously
	log.Printf("🚀 Step 5: Starting async processing...")
	go processWebhookAsync(event)

	// Step 6: Respond immediately
	log.Printf("✅ Step 6: Sending 200 OK response")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	c.JSON(http.StatusOK, gin.H{"status": "received", "event_id": event.ID})
}

// Process webhook events asynchronously
func processWebhookAsync(event ScalekitWebhookEvent) {
	log.Printf("⚙️  ASYNC PROCESSING STARTED for event: %s", event.Type)

	defer func() {
		if r := recover(); r != nil {
			log.Printf("🚨 PANIC in async processing: %v", r)
		}
	}()

	switch event.Type {
	case "organization.directory.user_created":
		log.Printf("📌 Routing to handleUserCreated")
		handleUserCreated(event)
	case "organization.directory.user_updated":
		log.Printf("📌 Routing to handleUserUpdated")
		handleUserUpdated(event)
	case "organization.directory.user_deleted":
		log.Printf("📌 Routing to handleUserDeleted")
		handleUserDeleted(event)
	case "organization.directory_enabled":
		log.Printf("📌 Routing to handleDirectoryEnabled")
		handleDirectoryEnabled(event)
	case "organization.directory_disabled":
		log.Printf("📌 Routing to handleDirectoryDisabled")
		handleDirectoryDisabled(event)
	default:
		log.Printf("⚠️  Unhandled event type: %s", event.Type)
	}

	log.Printf("✅ ASYNC PROCESSING COMPLETED for event: %s", event.Type)
}

// ============================================
// Event Handlers
// ============================================

func handleUserCreated(event ScalekitWebhookEvent) {
	log.Printf("👤 ═══ handleUserCreated START ═══")

	var directoryUser DirectoryUser
	if err := json.Unmarshal(event.Data, &directoryUser); err != nil {
		log.Printf("❌ Error parsing user data: %v", err)
		log.Printf("   Raw data: %s", string(event.Data))
		return
	}

	log.Printf("   User Email: %s", directoryUser.Email)
	log.Printf("   User Name: %s", directoryUser.Name)
	log.Printf("   User ID: %s", directoryUser.ID)
	log.Printf("   Organization ID: %s", directoryUser.OrganizationID)

	// Ensure organization exists
	log.Printf("🏢 Ensuring organization exists...")
	if err := ensureOrganizationExists(directoryUser.OrganizationID); err != nil {
		log.Printf("❌ Error ensuring organization exists: %v", err)
		return
	}

	// Check if user exists by ExternalID OR Email
	log.Printf("💾 Checking if user exists...")
	var existingUser database.User
	result := database.DB.Where("external_id = ? OR email = ?", directoryUser.ID, directoryUser.Email).First(&existingUser)

	if result.Error == nil {
		// User exists - update it
		log.Printf("ℹ️  User already exists (ID: %s), updating...", existingUser.ID)

		existingUser.ExternalID = directoryUser.ID
		existingUser.Email = directoryUser.Email

		if err := database.DB.Save(&existingUser).Error; err != nil {
			log.Printf("❌ Error updating existing user: %v", err)
			return
		}

		log.Printf("✅ User updated successfully (ID: %s)", existingUser.ID)
	} else {
		// User doesn't exist - create it
		log.Printf("📝 Creating new user...")

		newUser := database.User{
			ExternalID: directoryUser.ID,
			Email:      directoryUser.Email,
		}

		if err := database.DB.Create(&newUser).Error; err != nil {
			log.Printf("❌ Error creating user: %v", err)
			return
		}

		log.Printf("✅ User created successfully (ID: %s)", newUser.ID)
	}

	log.Printf("👤 ═══ handleUserCreated END ═══")
}

func handleUserUpdated(event ScalekitWebhookEvent) {
	log.Printf("🔄 ═══ handleUserUpdated START ═══")

	var directoryUser DirectoryUser
	if err := json.Unmarshal(event.Data, &directoryUser); err != nil {
		log.Printf("❌ Error parsing user data: %v", err)
		return
	}

	log.Printf("   User Email: %s", directoryUser.Email)
	log.Printf("   User ID: %s", directoryUser.ID)

	var user database.User
	result := database.DB.Where("external_id = ?", directoryUser.ID).First(&user)

	if result.Error != nil {
		log.Printf("⚠️  User not found, creating instead...")
		handleUserCreated(event)
		return
	}

	user.Email = directoryUser.Email

	if err := database.DB.Save(&user).Error; err != nil {
		log.Printf("❌ Error updating user: %v", err)
		return
	}

	log.Printf("✅ User updated successfully")
	log.Printf("🔄 ═══ handleUserUpdated END ═══")
}

func handleUserDeleted(event ScalekitWebhookEvent) {
	log.Printf("🗑️  ═══ handleUserDeleted START ═══")

	var userData struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}

	if err := json.Unmarshal(event.Data, &userData); err != nil {
		log.Printf("❌ Error parsing user data: %v", err)
		return
	}

	log.Printf("   Deleting user: %s (ID: %s)", userData.Email, userData.ID)

	result := database.DB.Where("external_id = ?", userData.ID).Delete(&database.User{})

	if result.Error != nil {
		log.Printf("❌ Error deleting user: %v", result.Error)
		return
	}

	if result.RowsAffected > 0 {
		log.Printf("✅ User deleted successfully")
	} else {
		log.Printf("⚠️  User not found for deletion")
	}

	log.Printf("🗑️  ═══ handleUserDeleted END ═══")
}

func handleDirectoryEnabled(event ScalekitWebhookEvent) {
	log.Printf("🟢 Directory sync enabled for org: %s", event.OrganizationID)

	if err := ensureOrganizationExists(event.OrganizationID); err != nil {
		log.Printf("❌ Error creating organization: %v", err)
	}
}

func handleDirectoryDisabled(event ScalekitWebhookEvent) {
	log.Printf("🔴 Directory sync disabled for org: %s", event.OrganizationID)
}

// ============================================
// Helper Functions
// ============================================

func ensureOrganizationExists(externalOrgID string) error {
	log.Printf("   🔍 Checking if organization exists: %s", externalOrgID)

	var org database.Organization
	result := database.DB.Where("external_id = ?", externalOrgID).First(&org)

	if result.Error == nil {
		log.Printf("   ✅ Organization already exists (ID: %s)", org.ID)
		return nil
	}

	log.Printf("   📝 Creating new organization...")
	org = database.Organization{
		ExternalID: externalOrgID,
		Name:       fmt.Sprintf("Organization %s", externalOrgID),
	}

	if err := database.DB.Create(&org).Error; err != nil {
		log.Printf("   ❌ Failed to create organization: %v", err)
		return err
	}

	log.Printf("   ✅ Organization created successfully (ID: %s)", org.ID)
	return nil
}
