package models

import (
	"time"

	"gorm.io/gorm"
)

// WorkflowExecution repräsentiert eine einzelne Ausführung eines Workflows
type WorkflowExecution struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	WorkflowID   uint           `json:"workflow_id"`
	Status       string         `json:"status" gorm:"default:'pending'"` // pending, running, completed, failed
	StartedAt    time.Time      `json:"started_at"`
	CompletedAt  *time.Time     `json:"completed_at"`
	InputData    string         `json:"input_data" gorm:"type:jsonb;default:'{}'"`
	OutputData   string         `json:"output_data" gorm:"type:jsonb;default:'{}'"`
	ErrorMessage string         `json:"error_message"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Beziehungen
	Workflow       Workflow        `json:"-" gorm:"foreignKey:WorkflowID"`
	NodeExecutions []NodeExecution `json:"node_executions" gorm:"foreignKey:WorkflowExecutionID"`
}

// NodeExecution repräsentiert eine einzelne Node-Ausführung innerhalb einer Workflow-Ausführung
type NodeExecution struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	WorkflowExecutionID uint       `json:"workflow_execution_id"`
	NodeID              uint       `json:"node_id"`
	Status              string     `json:"status" gorm:"default:'pending'"` // pending, running, completed, failed, skipped
	StartedAt           *time.Time `json:"started_at"`
	CompletedAt         *time.Time `json:"completed_at"`
	InputData           string     `json:"input_data" gorm:"type:jsonb;default:'{}'"`
	OutputData          string     `json:"output_data" gorm:"type:jsonb;default:'{}'"`
	ErrorMessage        string     `json:"error_message"`

	// Beziehungen
	WorkflowExecution WorkflowExecution `json:"-" gorm:"foreignKey:WorkflowExecutionID"`
	Node              Node              `json:"-" gorm:"foreignKey:NodeID"`
}

// NodeType repräsentiert einen verfügbaren Node-Typ mit seinen Eigenschaften
type NodeType struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	Key           string `json:"key" gorm:"uniqueIndex"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Icon          string `json:"icon"`
	Category      string `json:"category" gorm:"default:'Uncategorized'"`
	ConfigSchema  string `json:"config_schema" gorm:"type:jsonb"`
	InputSchema   string `json:"input_schema" gorm:"type:jsonb"`
	OutputSchema  string `json:"output_schema" gorm:"type:jsonb"`
	ExecutorClass string `json:"executor_class"`
}

// Trigger repräsentiert einen Auslöser für einen Workflow
type Trigger struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	WorkflowID     uint   `json:"workflow_id"`
	Name           string `json:"name"`
	TriggerType    string `json:"trigger_type"` // webhook, schedule, event
	Config         string `json:"config" gorm:"type:jsonb"`
	WebhookPath    string `json:"webhook_path" gorm:"uniqueIndex"`
	CronExpression string `json:"cron_expression"`
	IsActive       bool   `json:"is_active" gorm:"default:true"`

	// Beziehungen
	Workflow Workflow `json:"-" gorm:"foreignKey:WorkflowID"`
}
