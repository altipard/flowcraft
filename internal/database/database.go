package database

import (
	"log"

	"github.com/altipard/flowcraft/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Initialize establishes the connection to the database and performs migrations
func Initialize(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migration for models
	err = DB.AutoMigrate(
		&models.Workflow{},
		&models.Node{},
		&models.Connection{},
		&models.WorkflowExecution{},
		&models.NodeExecution{},
		&models.NodeType{},
		&models.Trigger{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Register default node types
	registerDefaultNodeTypes()
}

// Registers the default node types in the database if they don't exist yet
func registerDefaultNodeTypes() {
	nodeTypes := []models.NodeType{
		{
			Key:           "httpRequest",
			Name:          "HTTP Request",
			Description:   "Executes HTTP requests",
			Icon:          "globe",
			Category:      "API",
			ConfigSchema:  `{"properties":{"url":{"type":"string"},"method":{"type":"string","enum":["GET","POST","PUT","DELETE"]},"headers":{"type":"object"},"json_data":{"type":"object"}}}`,
			InputSchema:   `{}`,
			OutputSchema:  `{}`,
			ExecutorClass: "httpRequest",
		},
		{
			Key:           "filter",
			Name:          "Filter",
			Description:   "Filters data based on conditions",
			Icon:          "filter",
			Category:      "Data Processing",
			ConfigSchema:  `{"properties":{"field":{"type":"string"},"operator":{"type":"string","enum":["equals","not_equals","contains","greater_than","less_than"]},"value":{"type":"string"}}}`,
			InputSchema:   `{}`,
			OutputSchema:  `{}`,
			ExecutorClass: "filter",
		},
		{
			Key:           "transform",
			Name:          "Transform",
			Description:   "Transforms data based on a mapping",
			Icon:          "rotate",
			Category:      "Data Processing",
			ConfigSchema:  `{"properties":{"mapping":{"type":"object"}}}`,
			InputSchema:   `{}`,
			OutputSchema:  `{}`,
			ExecutorClass: "transform",
		},
	}

	// Register node types in the database if they don't exist yet
	for _, nodeType := range nodeTypes {
		var count int64
		DB.Model(&models.NodeType{}).Where("key = ?", nodeType.Key).Count(&count)
		if count == 0 {
			log.Printf("Registering node type: %s", nodeType.Key)
			if err := DB.Create(&nodeType).Error; err != nil {
				log.Printf("Warning: Failed to register node type %s: %v", nodeType.Key, err)
			}
		}
	}
}
