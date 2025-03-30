package models

import (
	"time"

	"gorm.io/gorm"
)

// Workflow represents an automation workflow
type Workflow struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	CreatedBy    uint           `json:"created_by"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	WorkflowData string         `json:"workflow_data" gorm:"type:jsonb;default:'{}'"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Nodes       []Node       `json:"nodes" gorm:"foreignKey:WorkflowID"`
	Connections []Connection `json:"connections" gorm:"foreignKey:WorkflowID"`
}

// Node represents a single step in the workflow
type Node struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	WorkflowID uint    `json:"workflow_id"`
	NodeType   string  `json:"node_type"`
	PositionX  float64 `json:"position_x"`
	PositionY  float64 `json:"position_y"`
	Name       string  `json:"name"`
	Config     string  `json:"config" gorm:"type:jsonb"`
}

// Connection represents a connection between two nodes
type Connection struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	WorkflowID   uint   `json:"workflow_id"`
	SourceNodeID uint   `json:"source_node_id"`
	TargetNodeID uint   `json:"target_node_id"`
	SourceHandle string `json:"source_handle" gorm:"default:'output'"`
	TargetHandle string `json:"target_handle" gorm:"default:'input'"`
}

// WorkflowRequest represents the input data for workflow creation/update
type WorkflowRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// Point represents an x,y coordinate for a node
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
