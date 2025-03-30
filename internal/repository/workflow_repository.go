package repository

import (
	"github.com/altipard/flowcraft/internal/database"
	"github.com/altipard/flowcraft/internal/models"
)

// WorkflowRepository contains all database operations for workflows
type WorkflowRepository struct{}

// FindAll returns all workflows
func (r *WorkflowRepository) FindAll() ([]models.Workflow, error) {
    var workflows []models.Workflow
    result := database.DB.Find(&workflows)
    return workflows, result.Error
}

// FindByID returns a workflow by its ID
func (r *WorkflowRepository) FindByID(id uint) (models.Workflow, error) {
    var workflow models.Workflow
    result := database.DB.Preload("Nodes").Preload("Connections").First(&workflow, id)
    return workflow, result.Error
}

// Create creates a new workflow
func (r *WorkflowRepository) Create(workflow *models.Workflow) error {
    return database.DB.Create(workflow).Error
}

// Update updates an existing workflow
func (r *WorkflowRepository) Update(workflow *models.Workflow) error {
    return database.DB.Save(workflow).Error
}

// Delete deletes a workflow
func (r *WorkflowRepository) Delete(id uint) error {
    return database.DB.Delete(&models.Workflow{}, id).Error
}
