package handlers

import (
	"coffee-desk-demo/database"
	"net/http"
	"strconv"

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
	Name        string     `json:"name" binding:"required"`
	Description *string    `json:"description,omitempty"`
	Priority    string     `json:"priority" binding:"required,oneof=P1 P2 P3"`
	Status      string     `json:"status" binding:"required,oneof=Backlog Todo InProgress Done"`
	OwnerID     *uuid.UUID `json:"owner_id,omitempty"`
}

// UpdateProjectRequest represents the request body for updating a project
type UpdateProjectRequest struct {
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	Priority    *string    `json:"priority,omitempty" binding:"omitempty,oneof=P1 P2 P3"`
	Status      *string    `json:"status,omitempty" binding:"omitempty,oneof=Backlog Todo InProgress Done"`
	OwnerID     *uuid.UUID `json:"owner_id,omitempty"`
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

	// Create project
	project := database.Project{
		OrganizationID: userInfo.OrganizationID,
		Name:           req.Name,
		Description:    req.Description,
		Priority:       database.Priority(req.Priority),
		Status:         database.Status(req.Status),
		OwnerID:        req.OwnerID,
	}

	if err := database.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
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
		updates["owner_id"] = *req.OwnerID
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
