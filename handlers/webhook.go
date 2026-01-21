package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	usersv1 "github.com/scalekit-inc/scalekit-sdk-go/v2/pkg/grpc/scalekit/v1/users"
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
	log.Printf("   Remote Address: %s", c.Request.RemoteAddr)
	log.Printf("   User-Agent: %s", c.GetHeader("User-Agent"))

	// 🔍 LOG KEY HEADERS
	log.Println("\n📋 WEBHOOK HEADERS:")
	webhookHeaders := []string{
		"Webhook-Id", "Webhook-Timestamp", "Webhook-Signature",
		"Svix-Id", "Svix-Timestamp", "Svix-Signature",
		"Content-Type", "User-Agent",
	}
	for _, header := range webhookHeaders {
		if value := c.GetHeader(header); value != "" {
			log.Printf("   %s: %s", header, truncate(value, 60))
		}
	}
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	defer func() {
		if r := recover(); r != nil {
			log.Printf("🚨 PANIC RECOVERED: %v", r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
	}()

	// Step 1: Read body
	log.Printf("\n📖 Step 1: Reading request body...")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("❌ Error reading webhook body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading request body"})
		return
	}
	log.Printf("✅ Body read successfully (%d bytes)", len(body))

	// Step 2: Get webhook secret
	log.Printf("\n🔑 Step 2: Checking webhook secret...")
	webhookSecret := os.Getenv("SCALEKIT_WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Printf("⚠️  WARNING: No webhook secret configured")
		log.Printf("   Set SCALEKIT_WEBHOOK_SECRET in your environment")
	} else {
		log.Printf("✅ Webhook secret configured")
	}

	// Step 3: Verify signature
	if webhookSecret != "" {
		log.Printf("\n🔐 Step 3: Verifying webhook signature...")

		wh, err := svix.NewWebhook(webhookSecret)
		if err != nil {
			log.Printf("❌ Error creating webhook verifier: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		err = wh.Verify(body, c.Request.Header)
		if err != nil {
			log.Printf("❌ Signature verification FAILED: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid webhook signature",
			})
			return
		}

		log.Printf("✅ Webhook signature verified!")
	}

	// Step 4: Parse event
	log.Printf("\n🔍 Step 4: Parsing webhook event...")
	var event ScalekitWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("❌ Failed to parse webhook event: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event format"})
		return
	}

	log.Printf("✅ Event parsed successfully!")
	log.Printf("   📌 Event ID: %s", event.ID)
	log.Printf("   📌 Event Type: %s", event.Type)
	log.Printf("   📌 Organization: %s", event.OrganizationID)

	// Step 5: Process async
	log.Printf("\n🚀 Step 5: Processing event asynchronously...")
	go processWebhookAsync(event)

	// Step 6: Respond immediately
	log.Printf("✅ Step 6: Sending 200 OK")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	c.JSON(http.StatusOK, gin.H{
		"status":   "received",
		"event_id": event.ID,
	})
}

// extractSvixHeaders extracts Svix webhook headers (supports both formats)
func extractSvixHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)

	// Svix supports TWO header formats:
	// 1. Standard format: Webhook-Id, Webhook-Timestamp, Webhook-Signature
	// 2. Legacy format: Svix-Id, Svix-Timestamp, Svix-Signature

	// Try standard format first (Webhook-*)
	if id := c.GetHeader("Webhook-Id"); id != "" {
		headers["svix-id"] = id
		log.Printf("   Found Webhook-Id: %s", id)
	}
	if timestamp := c.GetHeader("Webhook-Timestamp"); timestamp != "" {
		headers["svix-timestamp"] = timestamp
		log.Printf("   Found Webhook-Timestamp: %s", timestamp)
	}
	if signature := c.GetHeader("Webhook-Signature"); signature != "" {
		headers["svix-signature"] = signature
		log.Printf("   Found Webhook-Signature: %s", truncate(signature, 50))
	}

	// Try legacy format (Svix-*) if standard format not found
	if headers["svix-id"] == "" {
		if id := c.GetHeader("Svix-Id"); id != "" {
			headers["svix-id"] = id
			log.Printf("   Found Svix-Id: %s", id)
		}
	}
	if headers["svix-timestamp"] == "" {
		if timestamp := c.GetHeader("Svix-Timestamp"); timestamp != "" {
			headers["svix-timestamp"] = timestamp
			log.Printf("   Found Svix-Timestamp: %s", timestamp)
		}
	}
	if headers["svix-signature"] == "" {
		if signature := c.GetHeader("Svix-Signature"); signature != "" {
			headers["svix-signature"] = signature
			log.Printf("   Found Svix-Signature: %s", truncate(signature, 50))
		}
	}

	return headers
}

// stringOrEmpty returns the string or "(empty)" for logging
func stringOrEmpty(s string) string {
	if s == "" {
		return "(empty)"
	}
	if len(s) > 50 {
		return s[:50] + "..."
	}
	return s
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Process webhook events asynchronously
func processWebhookAsync(event ScalekitWebhookEvent) {
	startTime := time.Now()
	log.Printf("\n⚙️  ═══════════════════════════════════════════════════")
	log.Printf("⚙️  ASYNC PROCESSING STARTED")
	log.Printf("   Event Type: %s", event.Type)
	log.Printf("   Event ID: %s", event.ID)
	log.Printf("   Organization: %s", event.OrganizationID)
	log.Printf("⚙️  ═══════════════════════════════════════════════════\n")

	defer func() {
		if r := recover(); r != nil {
			log.Printf("\n🚨 ═══════════════════════════════════════════════════")
			log.Printf("🚨 PANIC RECOVERED in async processing!")
			log.Printf("   Event Type: %s", event.Type)
			log.Printf("   Event ID: %s", event.ID)
			log.Printf("   Error: %v", r)
			log.Printf("🚨 ═══════════════════════════════════════════════════\n")
		}

		duration := time.Since(startTime)
		log.Printf("\n✅ ═══════════════════════════════════════════════════")
		log.Printf("✅ ASYNC PROCESSING COMPLETED")
		log.Printf("   Event Type: %s", event.Type)
		log.Printf("   Duration: %v", duration)
		log.Printf("✅ ═══════════════════════════════════════════════════\n")
	}()

	// Route to appropriate handler
	switch event.Type {
	case "organization.directory.user_created":
		log.Printf("📌 Routing to: handleUserCreated")
		handleUserCreated(event)

	case "organization.directory.user_updated":
		log.Printf("📌 Routing to: handleUserUpdated")
		handleUserUpdated(event)

	case "organization.directory.user_deleted":
		log.Printf("📌 Routing to: handleUserDeleted")
		handleUserDeleted(event)

	case "organization.directory_enabled":
		log.Printf("📌 Routing to: handleDirectoryEnabled")
		handleDirectoryEnabled(event)

	case "organization.directory_disabled":
		log.Printf("📌 Routing to: handleDirectoryDisabled")
		handleDirectoryDisabled(event)

	default:
		log.Printf("⚠️  UNHANDLED EVENT TYPE: %s", event.Type)
		log.Printf("   Available handlers:")
		log.Printf("   - organization.directory.user_created")
		log.Printf("   - organization.directory.user_updated")
		log.Printf("   - organization.directory.user_deleted")
		log.Printf("   - organization.directory_enabled")
		log.Printf("   - organization.directory_disabled")
	}
}

// ============================================
// In-memory processed event tracker (replace with DB for persistence)
var processedEvents = struct {
	sync.Mutex
	events map[string]struct{}
}{events: make(map[string]struct{})}

// Checks if event has been processed already
func isEventProcessed(eventID string) bool {
	processedEvents.Lock()
	defer processedEvents.Unlock()
	_, ok := processedEvents.events[eventID]
	return ok
}

func markEventProcessed(eventID string) {
	processedEvents.Lock()
	defer processedEvents.Unlock()
	processedEvents.events[eventID] = struct{}{}
}

// ============================================
// Event Handlers
// ============================================

func handleUserCreated(event ScalekitWebhookEvent) {
	log.Printf("👤 ═══ handleUserCreated START ═══")

	var directoryUser DirectoryUser
	if err := json.Unmarshal(event.Data, &directoryUser); err != nil {
		log.Printf("❌ Error parsing user data: %v", err)
		return
	}

	log.Printf("   User Email: %s", directoryUser.Email)
	log.Printf("   User ID: %s", directoryUser.ID)
	log.Printf("   Organization ID: %s", directoryUser.OrganizationID)

	// Ensure organization exists locally
	if err := ensureOrganizationExists(directoryUser.OrganizationID); err != nil {
		log.Printf("❌ Error ensuring organization exists: %v", err)
		return
	}

	// Get ScaleKit client
	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		log.Printf("❌ Error getting ScaleKit client: %v", err)
		return
	}

	// Create user in ScaleKit
	createReq := &usersv1.CreateUser{
		Email:      directoryUser.Email,
		ExternalId: &directoryUser.ID,
	}

	_, err = scalekitClient.User().CreateUserAndMembership(
		context.Background(),
		directoryUser.OrganizationID,
		createReq,
		true, // send invitation email
	)
	if err != nil {
		log.Printf("❌ Error creating user in ScaleKit: %v", err)
		// continue, we can still sync locally
	}

	// Insert user into local DB if not exists
	var existingUser database.User
	result := database.DB.Where("external_id = ? OR email = ?", directoryUser.ID, directoryUser.Email).First(&existingUser)

	if result.Error == nil {
		log.Printf("ℹ️  User already exists locally (ID: %s)", existingUser.ID)
	} else {
		newUser := database.User{
			ExternalID: directoryUser.ID,
			Email:      directoryUser.Email,
		}
		if err := database.DB.Create(&newUser).Error; err != nil {
			log.Printf("❌ Error creating user locally: %v", err)
		} else {
			log.Printf("✅ User created locally (ID: %s)", newUser.ID)
		}
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

	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		log.Printf("❌ Error getting ScaleKit client: %v", err)
		return
	}

	// Update user in ScaleKit
	updateReq := &usersv1.UpdateUser{
		UserProfile: &usersv1.UpdateUserProfile{
			FirstName: &directoryUser.GivenName,
			LastName:  &directoryUser.FamilyName,
			Name:      &directoryUser.Name,
		},
	}

	_, err = scalekitClient.User().UpdateUser(
		context.Background(),
		directoryUser.ID,
		updateReq,
	)
	if err != nil {
		log.Printf("❌ Error updating user in ScaleKit: %v", err)
	}

	// Update local DB
	var user database.User
	result := database.DB.Where("external_id = ?", directoryUser.ID).First(&user)
	if result.Error != nil {
		log.Printf("⚠️  User not found locally, creating instead")
		handleUserCreated(event)
		return
	}

	user.Email = directoryUser.Email
	if err := database.DB.Save(&user).Error; err != nil {
		log.Printf("❌ Error updating user locally: %v", err)
	}

	log.Printf("🔄 ═══ handleUserUpdated END ═══")
}

func handleUserDeleted(event ScalekitWebhookEvent) {
	log.Printf("🗑️  ═══ handleUserDeleted START ═══")

	var userData struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(event.Data, &userData); err != nil {
		log.Printf("❌ Error parsing user data: %v", err)
		return
	}

	scalekitClient, err := GetScaleKitClient()
	if err != nil {
		log.Printf("❌ Error getting ScaleKit client: %v", err)
		return
	}

	if err := scalekitClient.User().DeleteUser(context.Background(), userData.ID); err != nil {
		log.Printf("❌ Error deleting user in ScaleKit: %v", err)
	}

	// Remove from local DB
	if err := database.DB.Where("external_id = ?", userData.ID).Delete(&database.User{}).Error; err != nil {
		log.Printf("❌ Error deleting user locally: %v", err)
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

// Add this route for testing
func TestWebhookHeaders(c *gin.Context) {
	log.Println("\n🔬 WEBHOOK HEADER DIAGNOSTIC")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	log.Println("📋 All Headers Received:")
	headerCount := 0
	for name, values := range c.Request.Header {
		for _, value := range values {
			headerCount++
			log.Printf("   [%d] %s: %s", headerCount, name, value)
		}
	}

	log.Println("\n🔍 Svix Header Check:")
	svixHeaders := []string{"svix-id", "svix-timestamp", "svix-signature"}
	for _, header := range svixHeaders {
		value := c.GetHeader(header)
		if value != "" {
			log.Printf("   ✅ %s: %s", header, value)
		} else {
			log.Printf("   ❌ %s: MISSING", header)
		}
	}

	body, _ := io.ReadAll(c.Request.Body)
	log.Printf("\n📦 Body (%d bytes): %s", len(body), string(body))

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	c.JSON(200, gin.H{
		"headers_received": headerCount,
		"has_svix_headers": c.GetHeader("svix-id") != "",
		"body_size":        len(body),
	})
}
