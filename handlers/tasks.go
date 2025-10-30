package handlers

import (
	"coffee-desk-demo/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
	Priority    string  `json:"priority" binding:"required,oneof=P1 P2 P3"`
	Status      string  `json:"status" binding:"required,oneof=Backlog Todo InProgress Done"`
	ProjectID   string  `json:"project_id,omitempty"`
	AssigneeID  string  `json:"assignee_id,omitempty"`
}

// UpdateTaskRequest represents the request body for updating a task
type UpdateTaskRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Priority    *string `json:"priority,omitempty" binding:"omitempty,oneof=P1 P2 P3"`
	Status      *string `json:"status,omitempty" binding:"omitempty,oneof=Backlog Todo InProgress Done"`
	ProjectID   *string `json:"project_id,omitempty"`
	AssigneeID  *string `json:"assignee_id,omitempty"`
}

// GetTasksHandler handles fetching all tasks for an organization
func GetTasksHandler(c *gin.Context) {
	// Authenticate user and get organization ID
	userInfo, err := extractOrganizationID(c)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	// Parse query parameters
	page := 1
	limit := 20
	projectID := c.Query("project_id")
	status := c.Query("status")
	assigneeID := c.Query("assignee_id")

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

	// Query tasks
	var tasks []database.Task
	var total int64

	query := database.DB.Where("organization_id = ?", userInfo.OrganizationID)

	// Apply filters
	if projectID != "" {
		if projectUUID, err := uuid.Parse(projectID); err == nil {
			query = query.Where("project_id = ?", projectUUID)
		}
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if assigneeID != "" {
		query = query.Where("assignee_id = ?", assigneeID)
	}

	// Count total
	if err := query.Model(&database.Task{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count tasks"})
		return
	}

	// Fetch tasks with pagination, project info, and assignee info
	if err := query.Preload("Project").Preload("Assignee").
		Offset(offset).Limit(limit).Order("created_at DESC").Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// CreateTaskHandler handles creating a new task
func CreateTaskHandler(c *gin.Context) {
	// Authenticate user and get organization ID
	userInfo, err := extractOrganizationID(c)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	// Parse request body
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Parse project_id if provided (can be UUID or project name)
	var projectID *uuid.UUID
	if req.ProjectID != "" {
		// Try to parse as UUID first
		if parsedID, err := uuid.Parse(req.ProjectID); err == nil {
			projectID = &parsedID
		} else {
			// If not a UUID, treat as project name and look up project
			var project database.Project
			if err := database.DB.Where("name = ? AND organization_id = ?", req.ProjectID, userInfo.OrganizationID).
				First(&project).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Project not found with name: " + req.ProjectID})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to look up project"})
				return
			}
			projectID = &project.ID
		}
	}

	// Parse assignee_id if provided (can be UUID or email)
	var assigneeID *uuid.UUID
	if req.AssigneeID != "" {
		// Try to parse as UUID first
		if parsedID, err := uuid.Parse(req.AssigneeID); err == nil {
			assigneeID = &parsedID
		} else {
			// If not a UUID, treat as email and look up user
			var user database.User
			if err := database.DB.Where("email = ?", req.AssigneeID).First(&user).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusBadRequest, gin.H{"error": "User not found with email: " + req.AssigneeID})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to look up user"})
				return
			}
			assigneeID = &user.ID
		}
	}

	// Create task
	task := database.Task{
		OrganizationID: userInfo.OrganizationID,
		ProjectID:      projectID,
		Name:           req.Name,
		Description:    req.Description,
		Priority:       database.Priority(req.Priority),
		Status:         database.Status(req.Status),
		AssigneeID:     assigneeID,
	}

	if err := database.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	// Load project info for response
	if err := database.DB.Preload("Project").First(&task, task.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch created task"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Task created successfully",
		"task":    task,
	})
}

// GetTaskHandler handles fetching a single task by ID
func GetTaskHandler(c *gin.Context) {
	// Authenticate user and get organization ID
	userInfo, err := extractOrganizationID(c)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	// Get task ID from URL parameter
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Query task with project and assignee info
	var task database.Task
	if err := database.DB.Where("id = ? AND organization_id = ?", taskID, userInfo.OrganizationID).
		Preload("Project").Preload("Assignee").
		First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// UpdateTaskHandler handles updating a task
func UpdateTaskHandler(c *gin.Context) {
	// Authenticate user and get organization ID
	userInfo, err := extractOrganizationID(c)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	// Get task ID from URL parameter
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Parse request body
	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Check if task exists
	var task database.Task
	if err := database.DB.Where("id = ? AND organization_id = ?", taskID, userInfo.OrganizationID).
		First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
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
	if req.ProjectID != nil {
		if *req.ProjectID == "" {
			// Set to nil if empty string
			updates["project_id"] = nil
		} else {
			// Try to parse as UUID first
			if parsedID, err := uuid.Parse(*req.ProjectID); err == nil {
				updates["project_id"] = parsedID
			} else {
				// If not a UUID, treat as project name and look up project
				var project database.Project
				if err := database.DB.Where("name = ? AND organization_id = ?", *req.ProjectID, userInfo.OrganizationID).
					First(&project).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						c.JSON(http.StatusBadRequest, gin.H{"error": "Project not found with name: " + *req.ProjectID})
						return
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to look up project"})
					return
				}
				updates["project_id"] = project.ID
			}
		}
	}
	if req.AssigneeID != nil {
		if *req.AssigneeID == "" {
			// Set to nil if empty string
			updates["assignee_id"] = nil
		} else {
			// Try to parse as UUID first
			if parsedID, err := uuid.Parse(*req.AssigneeID); err == nil {
				updates["assignee_id"] = parsedID
			} else {
				// If not a UUID, treat as email and look up user
				var user database.User
				if err := database.DB.Where("email = ?", *req.AssigneeID).First(&user).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						c.JSON(http.StatusBadRequest, gin.H{"error": "User not found with email: " + *req.AssigneeID})
						return
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to look up user"})
					return
				}
				updates["assignee_id"] = user.ID
			}
		}
	}

	// Apply updates
	if err := database.DB.Model(&task).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	// Fetch updated task with project info
	if err := database.DB.Preload("Project").Where("id = ?", taskID).First(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task updated successfully",
		"task":    task,
	})
}

// DeleteTaskHandler handles deleting a task
func DeleteTaskHandler(c *gin.Context) {
	// Authenticate user and get organization ID
	userInfo, err := extractOrganizationID(c)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	// Get task ID from URL parameter
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Check if task exists
	var task database.Task
	if err := database.DB.Where("id = ? AND organization_id = ?", taskID, userInfo.OrganizationID).
		First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
		return
	}

	// Delete task (soft delete)
	if err := database.DB.Delete(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}
