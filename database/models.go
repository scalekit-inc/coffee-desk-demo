package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Priority represents task/project priority levels
type Priority string

const (
	PriorityP1 Priority = "P1"
	PriorityP2 Priority = "P2"
	PriorityP3 Priority = "P3"
)

// Status represents task/project status
type Status string

const (
	StatusBacklog    Status = "Backlog"
	StatusTodo       Status = "Todo"
	StatusInProgress Status = "InProgress"
	StatusDone       Status = "Done"
)

// User represents a user in the system (local copy of Scalekit user data)
type User struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ExternalID string         `gorm:"type:varchar(255);not null;unique;index" json:"external_id"` // Scalekit user ID
	Email      string         `gorm:"type:varchar(255);not null;unique" json:"email"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// Organization represents an organization in the system (local copy of Scalekit org data)
type Organization struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ExternalID string         `gorm:"type:varchar(255);not null;unique;index" json:"external_id"` // Scalekit organization ID
	Name       string         `gorm:"type:varchar(255);not null" json:"name"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// Project represents a project in the system
type Project struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizationID string         `gorm:"type:varchar(255);not null;index" json:"organization_id"`
	Name           string         `gorm:"type:varchar(255);not null" json:"name"`
	Description    *string        `gorm:"type:text" json:"description,omitempty"`
	Priority       Priority       `gorm:"type:varchar(10);not null;default:'P3'" json:"priority"`
	Status         Status         `gorm:"type:varchar(20);not null;default:'Backlog'" json:"status"`
	OwnerID        *uuid.UUID     `gorm:"type:uuid;index" json:"owner_id,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Owner *User  `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Tasks []Task `gorm:"foreignKey:ProjectID" json:"tasks,omitempty"`
}

// Task represents a task in the system
type Task struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizationID string         `gorm:"type:varchar(255);not null;index" json:"organization_id"`
	ProjectID      *uuid.UUID     `gorm:"type:uuid;index" json:"project_id,omitempty"`
	Name           string         `gorm:"type:varchar(255);not null" json:"name"`
	Description    *string        `gorm:"type:text" json:"description,omitempty"`
	Priority       Priority       `gorm:"type:varchar(10);not null;default:'P3'" json:"priority"`
	Status         Status         `gorm:"type:varchar(20);not null;default:'Backlog'" json:"status"`
	AssigneeID     *uuid.UUID     `gorm:"type:uuid;index" json:"assignee_id,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Project  *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Assignee *User    `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
}

// TableName returns the table name for Project
func (Project) TableName() string {
	return "projects"
}

// TableName returns the table name for User
func (User) TableName() string {
	return "users"
}

// TableName returns the table name for Organization
func (Organization) TableName() string {
	return "organizations"
}

// TableName returns the table name for Task
func (Task) TableName() string {
	return "tasks"
}

// BeforeCreate hook for User
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook for Organization
func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook for Project
func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook for Task
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
