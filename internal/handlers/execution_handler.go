package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/altipard/flowcraft/internal/database"
	"github.com/altipard/flowcraft/internal/models"
	"github.com/altipard/flowcraft/internal/queue"
	"github.com/labstack/echo/v4"
)

// ExecutionHandler manages the HTTP requests for workflow executions
type ExecutionHandler struct {
	queueClient *queue.QueueClient
}

// NewExecutionHandler creates a new ExecutionHandler
func NewExecutionHandler(queueClient *queue.QueueClient) *ExecutionHandler {
	return &ExecutionHandler{
		queueClient: queueClient,
	}
}

// ExecuteWorkflow godoc
// @Summary Execute a workflow
// @Description Executes a workflow with the given ID
// @Tags executions
// @Accept json
// @Produce json
// @Param id path int true "Workflow ID"
// @Param inputData body object false "Input data for workflow execution"
// @Success 202 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /workflows/{id}/execute [post]
func (h *ExecutionHandler) ExecuteWorkflow(c echo.Context) error {
	workflowID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid workflow ID"})
	}

	// Check if the workflow exists
	var workflow models.Workflow
	if err := database.DB.First(&workflow, workflowID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Workflow not found"})
	}

	// Input data from request body
	var inputData map[string]interface{}
	if err := c.Bind(&inputData); err != nil {
		// If no input data is provided, use an empty map
		inputData = make(map[string]interface{})
	}

	// Create workflow execution
	execution := models.WorkflowExecution{
		WorkflowID: uint(workflowID),
		Status:     "pending",
		StartedAt:  time.Now(),
	}

	// Save input data as JSON
	inputJSON, _ := json.Marshal(inputData)
	execution.InputData = string(inputJSON)

	if err := database.DB.Create(&execution).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Queue asynchronous execution
	err = h.queueClient.EnqueueTask("workflow_tasks", "execute_workflow", map[string]interface{}{
		"execution_id": execution.ID,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"execution_id": execution.ID,
		"status":       "pending",
	})
}

// GetStatus godoc
// @Summary Get execution status
// @Description Returns the status of a workflow execution
// @Tags executions
// @Accept json
// @Produce json
// @Param id path int true "Execution ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /executions/{id}/status [get]
func (h *ExecutionHandler) GetStatus(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	var execution models.WorkflowExecution
	if err := database.DB.First(&execution, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Execution not found"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":            execution.ID,
		"workflow_id":   execution.WorkflowID,
		"status":        execution.Status,
		"started_at":    execution.StartedAt,
		"completed_at":  execution.CompletedAt,
		"error_message": execution.ErrorMessage,
		"output_data":   execution.OutputData,
	})
}
