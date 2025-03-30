package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/altipard/flowcraft/internal/database"
	"github.com/altipard/flowcraft/internal/models"
)

// Engine is the central component for workflow execution
type Engine struct{}

// NewEngine creates a new Engine instance
func NewEngine() *Engine {
	return &Engine{}
}

// ExecuteWorkflow executes a workflow
func (e *Engine) ExecuteWorkflow(executionID uint) error {
	// Load workflow execution
	var execution models.WorkflowExecution
	if err := database.DB.Preload("Workflow").Preload("Workflow.Nodes").Preload("Workflow.Connections").First(&execution, executionID).Error; err != nil {
		return err
	}

	// Update status
	execution.Status = "running"
	execution.StartedAt = time.Now()
	database.DB.Save(&execution)

	// Start execution
	err := e.executeWorkflowInternal(&execution)

	// Completion
	now := time.Now()
	execution.CompletedAt = &now
	if err != nil {
		execution.Status = "failed"
		execution.ErrorMessage = err.Error()
	} else {
		execution.Status = "completed"
	}
	database.DB.Save(&execution)

	return err
}

// executeWorkflowInternal is the internal implementation of workflow execution
func (e *Engine) executeWorkflowInternal(execution *models.WorkflowExecution) error {
	// Workflow data
	workflow := execution.Workflow

	// Start with the start nodes (nodes without incoming connections)
	var startNodes []models.Node
	for _, node := range workflow.Nodes {
		hasIncoming := false
		for _, conn := range workflow.Connections {
			if conn.TargetNodeID == node.ID {
				hasIncoming = true
				break
			}
		}
		if !hasIncoming {
			startNodes = append(startNodes, node)
		}
	}

	if len(startNodes) == 0 {
		return errors.New("workflow has no start nodes")
	}

	// Prepare context for execution
	var inputData map[string]interface{}
	err := json.Unmarshal([]byte(execution.InputData), &inputData)
	if err != nil {
		return fmt.Errorf("failed to parse input data: %v", err)
	}

	context := NewExecutionContext(inputData)

	// Execute start nodes
	for _, node := range startNodes {
		if err := e.executeNode(node.ID, execution.ID, context); err != nil {
			return err
		}
	}

	// Save results to execution
	outputJSON, err := json.Marshal(context.Results)
	if err != nil {
		return fmt.Errorf("failed to marshal output data: %v", err)
	}
	execution.OutputData = string(outputJSON)

	return nil
}

// executeNode executes a single node
func (e *Engine) executeNode(nodeID, executionID uint, context *ExecutionContext) error {
	// Load node and related information
	var node models.Node
	if err := database.DB.First(&node, nodeID).Error; err != nil {
		return err
	}

	// Load node type
	var nodeType models.NodeType
	if err := database.DB.Where("key = ?", node.NodeType).First(&nodeType).Error; err != nil {
		return err
	}

	// Create node execution
	nodeExecution := models.NodeExecution{
		WorkflowExecutionID: executionID,
		NodeID:              nodeID,
		Status:              "running",
	}
	now := time.Now()
	nodeExecution.StartedAt = &now
	database.DB.Create(&nodeExecution)

	// Prepare input data
	inputData := e.prepareNodeInput(node, executionID, context)
	inputJSON, _ := json.Marshal(inputData)
	nodeExecution.InputData = string(inputJSON)
	database.DB.Save(&nodeExecution)

	// Load executor for this node type and execute
	executor, err := LoadExecutor(nodeType.ExecutorClass)
	if err != nil {
		nodeExecution.Status = "failed"
		nodeExecution.ErrorMessage = fmt.Sprintf("failed to load executor: %v", err)
		database.DB.Save(&nodeExecution)
		return err
	}

	// Load node configuration
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(node.Config), &config); err != nil {
		nodeExecution.Status = "failed"
		nodeExecution.ErrorMessage = fmt.Sprintf("failed to parse node config: %v", err)
		database.DB.Save(&nodeExecution)
		return err
	}

	// Execute node
	result, err := executor.Execute(config, inputData)
	if err != nil {
		nodeExecution.Status = "failed"
		nodeExecution.ErrorMessage = fmt.Sprintf("execution failed: %v", err)
		now := time.Now()
		nodeExecution.CompletedAt = &now
		database.DB.Save(&nodeExecution)
		return err
	}

	// Save result
	resultJSON, _ := json.Marshal(result)
	nodeExecution.OutputData = string(resultJSON)
	nodeExecution.Status = "completed"
	now = time.Now()
	nodeExecution.CompletedAt = &now
	database.DB.Save(&nodeExecution)

	// Save result in execution context
	context.Results[nodeID] = result

	// Find and execute subsequent nodes
	var connections []models.Connection
	database.DB.Where("source_node_id = ?", nodeID).Find(&connections)

	for _, conn := range connections {
		targetNodeID := conn.TargetNodeID

		// Check if all incoming connections for the target node are ready
		if e.allInputsReady(targetNodeID, executionID) {
			if err := e.executeNode(targetNodeID, executionID, context); err != nil {
				return err
			}
		}
	}

	return nil
}

// prepareNodeInput prepares the input data for a node
func (e *Engine) prepareNodeInput(node models.Node, executionID uint, context *ExecutionContext) map[string]interface{} {
	// If there are no incoming connections, use the global input
	var connections []models.Connection
	database.DB.Where("target_node_id = ?", node.ID).Find(&connections)

	if len(connections) == 0 {
		return context.Input
	}

	// Otherwise, collect the outputs of the predecessor nodes
	inputs := make(map[string]interface{})

	for _, conn := range connections {
		sourceNodeID := conn.SourceNodeID
		targetHandle := conn.TargetHandle

		if result, ok := context.Results[sourceNodeID]; ok {
			if _, exists := inputs[targetHandle]; !exists {
				inputs[targetHandle] = []interface{}{}
			}

			inputArray, _ := inputs[targetHandle].([]interface{})
			inputs[targetHandle] = append(inputArray, result)
		}
	}

	return inputs
}

// allInputsReady checks if all inputs of a node are ready
func (e *Engine) allInputsReady(nodeID uint, executionID uint) bool {
	var connections []models.Connection
	database.DB.Where("target_node_id = ?", nodeID).Find(&connections)

	for _, conn := range connections {
		var nodeExecution models.NodeExecution
		result := database.DB.Where("workflow_execution_id = ? AND node_id = ? AND status = ?",
			executionID, conn.SourceNodeID, "completed").First(&nodeExecution)

		if result.Error != nil {
			return false
		}
	}

	return true
}

// ExecutionContext holds the state during a workflow execution
type ExecutionContext struct {
	Input   map[string]interface{}
	Results map[uint]interface{}
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(input map[string]interface{}) *ExecutionContext {
	return &ExecutionContext{
		Input:   input,
		Results: make(map[uint]interface{}),
	}
}
