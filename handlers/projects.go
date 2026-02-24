package handlers

import (
	"context"
	"coffee-desk-demo/database"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// convertScalekitUserIDToLocal converts Scalekit user ID to local user UUID
func convertScalekitUserIDToLocal(scalekitUserID string) (*uuid.UUID, error) {
	if scalekitUserID == "" {
		return nil, nil
	}

	var user database.User
	if err := database.DB.Where("external_id = ?", scalekitUserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // User not found in local DB
		}
		return nil, err
	}

	return &user.ID, nil
}

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
	Priority    string  `json:"priority" binding:"required,oneof=P1 P2 P3"`
	Status      string  `json:"status" binding:"required,oneof=Backlog Todo InProgress Done"`
	OwnerID     string  `json:"owner_id,omitempty"`
}

// UpdateProjectRequest represents the request body for updating a project
type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Priority    *string `json:"priority,omitempty" binding:"omitempty,oneof=P1 P2 P3"`
	Status      *string `json:"status,omitempty" binding:"omitempty,oneof=Backlog Todo InProgress Done"`
	OwnerID     *string `json:"owner_id,omitempty"`
}

// GetProjectsHandler handles fetching all projects for an organization
func GetProjectsHandler(c *gin.Context) {
	// Authenticate user and get organization ID
	userInfo, err := extractOrganizationID(c)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	// Parse query parameters
	page := 1
	limit := 20
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	// Query projects
	var projects []database.Project
	var total int64

	query := database.DB.Where("organization_id = ?", userInfo.OrganizationID)

	// Count total
	if err := query.Model(&database.Project{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count projects"})
		return
	}

	// Fetch projects with pagination and owner information
	if err := query.Preload("Owner").Offset(offset).Limit(limit).Order("created_at DESC").Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// CreateProjectHandler handles creating a new project
func CreateProjectHandler(c *gin.Context) {
	// Authenticate user and get organization ID
	userInfo, err := extractOrganizationID(c)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	// Parse request body
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Parse owner_id if provided (can be UUID or email)
	var ownerID *uuid.UUID
	if req.OwnerID != "" {
		// Try to parse as UUID first
		if parsedID, err := uuid.Parse(req.OwnerID); err == nil {
			ownerID = &parsedID
		} else {
			// If not a UUID, treat as email and look up user
			var user database.User
			if err := database.DB.Where("email = ?", req.OwnerID).First(&user).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusBadRequest, gin.H{"error": "User not found with email: " + req.OwnerID})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to look up user"})
				return
			}
			ownerID = &user.ID
		}
	}

	// Create project
	project := database.Project{
		OrganizationID: userInfo.OrganizationID,
		Name:           req.Name,
		Description:    req.Description,
		Priority:       database.Priority(req.Priority),
		Status:         database.Status(req.Status),
		OwnerID:        ownerID,
	}

	if err := database.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	// Load the owner for Slack notification
	if ownerID != nil {
		if err := database.DB.Preload("Owner").Where("id = ?", project.ID).First(&project).Error; err == nil {
			// Extract user ID from token before goroutine (context may not be valid in goroutine)
			userID, err := getUserIDFromToken(c)
			if err == nil {
				// Try to send Slack notification if Slack is enabled (async)
				go sendProjectCreationSlackMessage(userID, project)
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Project created successfully",
		"project": project,
	})
}

// GetProjectHandler handles fetching a single project by ID
func GetProjectHandler(c *gin.Context) {
	// Authenticate user and get organization ID
	userInfo, err := extractOrganizationID(c)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	// Get project ID from URL parameter
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Query project with tasks and owner
	var project database.Project
	if err := database.DB.Where("id = ? AND organization_id = ?", projectID, userInfo.OrganizationID).
		Preload("Owner").
		Preload("Tasks", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Assignee").Order("created_at DESC")
		}).
		First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"project": project})
}

// UpdateProjectHandler handles updating a project
func UpdateProjectHandler(c *gin.Context) {
	// Authenticate user and get organization ID
	userInfo, err := extractOrganizationID(c)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	// Get project ID from URL parameter
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Parse request body
	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Check if project exists
	var project database.Project
	if err := database.DB.Where("id = ? AND organization_id = ?", projectID, userInfo.OrganizationID).
		First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project"})
		return
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Priority != nil {
		updates["priority"] = database.Priority(*req.Priority)
	}
	if req.Status != nil {
		updates["status"] = database.Status(*req.Status)
	}
	if req.OwnerID != nil {
		if *req.OwnerID == "" {
			// Set to nil if empty string
			updates["owner_id"] = nil
		} else {
			// Try to parse as UUID first
			if parsedID, err := uuid.Parse(*req.OwnerID); err == nil {
				updates["owner_id"] = parsedID
			} else {
				// If not a UUID, treat as email and look up user
				var user database.User
				if err := database.DB.Where("email = ?", *req.OwnerID).First(&user).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						c.JSON(http.StatusBadRequest, gin.H{"error": "User not found with email: " + *req.OwnerID})
						return
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to look up user"})
					return
				}
				updates["owner_id"] = user.ID
			}
		}
	}

	// Apply updates
	if err := database.DB.Model(&project).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	// Fetch updated project
	if err := database.DB.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project updated successfully",
		"project": project,
	})
}

// DeleteProjectHandler handles deleting a project
func DeleteProjectHandler(c *gin.Context) {
	// Authenticate user and get organization ID
	userInfo, err := extractOrganizationID(c)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	// Get project ID from URL parameter
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Check if project exists
	var project database.Project
	if err := database.DB.Where("id = ? AND organization_id = ?", projectID, userInfo.OrganizationID).
		First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project"})
		return
	}

	// Delete project (soft delete)
	if err := database.DB.Delete(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

// getUserIDFromToken extracts user ID from the JWT token
func getUserIDFromToken(c *gin.Context) (string, error) {
	accessToken, err := c.Cookie("auth_access_token")
	if err != nil {
		return "", fmt.Errorf("no access token found in cookies: %w", err)
	}

	claims, err := decodeJwtPayload(accessToken)
	if err != nil {
		return "", fmt.Errorf("failed to decode token: %w", err)
	}

	userID, err := getStringClaim(claims, "sub")
	if err != nil {
		return "", fmt.Errorf("no user ID (sub) in token: %w", err)
	}

	return userID, nil
}

// isSlackEnabled checks if Slack integration is enabled for a user
func isSlackEnabled(ctx context.Context, userID string) (bool, error) {
	config, err := getScaleKitConfig()
	if err != nil {
		return false, fmt.Errorf("failed to get ScaleKit config: %w", err)
	}

	handler := &ConnectedAccountsHandler{
		envURL:       config.EnvURL,
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}

	apiURL := fmt.Sprintf("%s/api/v1/connected_accounts/auth?connector=slack&identifier=%s",
		config.EnvURL, url.QueryEscape(userID))

	resp, err := handler.makeAuthenticatedRequest(ctx, "GET", apiURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to check Slack status: %w", err)
	}
	defer resp.Body.Close()

	// 404 means not connected
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	// Check for other errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Parse response
	respBody, parsedResp, err := readAndParseResponse(resp)
	if err != nil {
		return false, fmt.Errorf("failed to read response: %w", err)
	}

	if parsedResp == nil {
		return false, fmt.Errorf("failed to parse response: %s", string(respBody))
	}

	// Check if connected_account exists and status is ACTIVE
	if ca, ok := parsedResp["connected_account"].(map[string]interface{}); ok {
		if status, ok := ca["status"].(string); ok {
			return status == "ACTIVE", nil
		}
	}

	return false, nil
}

// executeSlackSendMessage executes the Slack send message tool via Scalekit API
func executeSlackSendMessage(ctx context.Context, userID, channel, text string) error {
	config, err := getScaleKitConfig()
	if err != nil {
		return fmt.Errorf("failed to get ScaleKit config: %w", err)
	}

	handler := &ConnectedAccountsHandler{
		envURL:       config.EnvURL,
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}

	// Prepare tool parameters for slack_send_message
	// Required: channel (channel ID, channel name like #general, or user ID for DM) and text
	toolParams := map[string]interface{}{
		"channel": channel,
		"text":    text,
	}

	requestBody := map[string]interface{}{
		"tool_name": "slack_send_message",
		"identifier": userID,
		"connector": "slack",
		"params":    toolParams,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	apiURL := fmt.Sprintf("%s/api/v1/execute_tool", config.EnvURL)
	resp, err := handler.makeAuthenticatedRequest(ctx, "POST", apiURL, jsonData)
	if err != nil {
		return fmt.Errorf("failed to execute tool: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// sendProjectCreationSlackMessage sends a Slack notification to the project assignee
func sendProjectCreationSlackMessage(userID string, project database.Project) {
	// Check if Slack is enabled for the user creating the project
	ctx := context.Background()
	slackEnabled, err := isSlackEnabled(ctx, userID)
	if err != nil {
		log.Printf("Failed to check Slack status: %v", err)
		return
	}

	if !slackEnabled {
		log.Printf("Slack not enabled for user %s, skipping Slack notification", userID)
		return
	}

	// Check if project has an assignee (owner)
	if project.Owner == nil || project.Owner.Email == "" {
		log.Printf("Project has no assignee, skipping Slack notification")
		return
	}

	// Prepare Slack message content
	// Extract assignee name from email (part before @) for a friendly mention
	assigneeName := project.Owner.Email
	if atIndex := strings.Index(project.Owner.Email, "@"); atIndex > 0 {
		assigneeName = project.Owner.Email[:atIndex]
		// Replace dots/underscores with spaces
		assigneeName = strings.ReplaceAll(assigneeName, ".", " ")
		assigneeName = strings.ReplaceAll(assigneeName, "_", " ")
		// Capitalize first letter of each word
		words := strings.Fields(assigneeName)
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
			}
		}
		assigneeName = strings.Join(words, " ")
	}

	message := "📋 *New Project Assigned*\n\n"
	message += fmt.Sprintf("Hey *%s* (%s)! You have been assigned to a new project:\n\n", assigneeName, project.Owner.Email)
	message += fmt.Sprintf("*Project Name:* %s\n", project.Name)
	if project.Description != nil && *project.Description != "" {
		message += fmt.Sprintf("*Description:* %s\n", *project.Description)
	}
	message += fmt.Sprintf("*Priority:* %s\n", project.Priority)
	message += fmt.Sprintf("*Status:* %s\n\n", project.Status)
	message += "Please review and start working on this project."

	// Send to the Coffee Desk team channel
	channel := "#all-coffeedesk"

	// Send Slack message
	if err := executeSlackSendMessage(ctx, userID, channel, message); err != nil {
		log.Printf("Failed to send project creation Slack message: %v", err)
		return
	}

	log.Printf("Successfully sent project creation Slack message to %s for project %s", channel, project.Name)
}
