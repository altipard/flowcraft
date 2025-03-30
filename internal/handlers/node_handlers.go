// internal/handlers/node_handler.go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/altipard/flowcraft/internal/database"
	"github.com/altipard/flowcraft/internal/models"
	"github.com/labstack/echo/v4"
)

// NodeHandler manages the HTTP requests for nodes
type NodeHandler struct{}

// NewNodeHandler creates a new NodeHandler
func NewNodeHandler() *NodeHandler {
	return &NodeHandler{}
}

// GetAll godoc
// @Summary Get all nodes
// @Description Returns a list of all nodes
// @Tags nodes
// @Accept json
// @Produce json
// @Success 200 {array} models.Node
// @Failure 500 {object} map[string]string
// @Router /nodes [get]
func (h *NodeHandler) GetAll(c echo.Context) error {
	var nodes []models.Node
	if err := database.DB.Find(&nodes).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, nodes)
}

// GetByID godoc
// @Summary Get node by ID
// @Description Returns a specific node based on its ID
// @Tags nodes
// @Accept json
// @Produce json
// @Param id path int true "Node ID"
// @Success 200 {object} models.Node
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /nodes/{id} [get]
func (h *NodeHandler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	var node models.Node
	if err := database.DB.First(&node, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Node not found"})
	}

	return c.JSON(http.StatusOK, node)
}

// Create godoc
// @Summary Create a new node
// @Description Creates a new node in a workflow
// @Tags nodes
// @Accept json
// @Produce json
// @Param node body models.Node true "Node data"
// @Success 201 {object} models.Node
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nodes [post]
func (h *NodeHandler) Create(c echo.Context) error {
	node := new(models.Node)
	if err := c.Bind(node); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if node.Config == "" {
		node.Config = "{}"
	}

	if err := database.DB.Create(node).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, node)
}

// Update godoc
// @Summary Update a node
// @Description Updates an existing node
// @Tags nodes
// @Accept json
// @Produce json
// @Param id path int true "Node ID"
// @Param node body models.Node true "Updated node data"
// @Success 200 {object} models.Node
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nodes/{id} [put]
func (h *NodeHandler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	var node models.Node
	if err := database.DB.First(&node, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Node not found"})
	}

	if err := c.Bind(&node); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := database.DB.Save(&node).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, node)
}

// Delete godoc
// @Summary Delete a node
// @Description Deletes a node based on its ID
// @Tags nodes
// @Accept json
// @Produce json
// @Param id path int true "Node ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nodes/{id} [delete]
func (h *NodeHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	if err := database.DB.Delete(&models.Node{}, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

// GetByWorkflowID godoc
// @Summary Get nodes for a workflow
// @Description Returns all nodes for a specific workflow
// @Tags nodes
// @Accept json
// @Produce json
// @Param workflowId path int true "Workflow ID"
// @Success 200 {array} models.Node
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /workflows/{workflowId}/nodes [get]
func (h *NodeHandler) GetByWorkflowID(c echo.Context) error {
	workflowID, err := strconv.Atoi(c.Param("workflowId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid workflow ID"})
	}

	var nodes []models.Node
	if err := database.DB.Where("workflow_id = ?", workflowID).Find(&nodes).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, nodes)
}
