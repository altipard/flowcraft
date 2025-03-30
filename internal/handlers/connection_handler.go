// internal/handlers/connection_handler.go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/altipard/flowcraft/internal/database"
	"github.com/altipard/flowcraft/internal/models"
	"github.com/labstack/echo/v4"
)

// ConnectionHandler manages the HTTP requests for connections
type ConnectionHandler struct{}

// NewConnectionHandler creates a new ConnectionHandler
func NewConnectionHandler() *ConnectionHandler {
	return &ConnectionHandler{}
}

// GetAll godoc
// @Summary Get all connections
// @Description Returns a list of all connections
// @Tags connections
// @Accept json
// @Produce json
// @Success 200 {array} models.Connection
// @Failure 500 {object} map[string]string
// @Router /connections [get]
func (h *ConnectionHandler) GetAll(c echo.Context) error {
	var connections []models.Connection
	if err := database.DB.Find(&connections).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, connections)
}

// GetByID godoc
// @Summary Get connection by ID
// @Description Returns a specific connection based on its ID
// @Tags connections
// @Accept json
// @Produce json
// @Param id path int true "Connection ID"
// @Success 200 {object} models.Connection
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /connections/{id} [get]
func (h *ConnectionHandler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	var connection models.Connection
	if err := database.DB.First(&connection, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Connection not found"})
	}

	return c.JSON(http.StatusOK, connection)
}

// Create godoc
// @Summary Create a new connection
// @Description Creates a new connection between nodes
// @Tags connections
// @Accept json
// @Produce json
// @Param connection body models.Connection true "Connection data"
// @Success 201 {object} models.Connection
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /connections [post]
func (h *ConnectionHandler) Create(c echo.Context) error {
	connection := new(models.Connection)
	if err := c.Bind(connection); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := database.DB.Create(connection).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, connection)
}

// Update godoc
// @Summary Update a connection
// @Description Updates an existing connection
// @Tags connections
// @Accept json
// @Produce json
// @Param id path int true "Connection ID"
// @Param connection body models.Connection true "Updated connection data"
// @Success 200 {object} models.Connection
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /connections/{id} [put]
func (h *ConnectionHandler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	var connection models.Connection
	if err := database.DB.First(&connection, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Connection not found"})
	}

	if err := c.Bind(&connection); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := database.DB.Save(&connection).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, connection)
}

// Delete godoc
// @Summary Delete a connection
// @Description Deletes a connection based on its ID
// @Tags connections
// @Accept json
// @Produce json
// @Param id path int true "Connection ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /connections/{id} [delete]
func (h *ConnectionHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	if err := database.DB.Delete(&models.Connection{}, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

// GetByWorkflowID godoc
// @Summary Get connections for a workflow
// @Description Returns all connections for a specific workflow
// @Tags connections
// @Accept json
// @Produce json
// @Param workflowId path int true "Workflow ID"
// @Success 200 {array} models.Connection
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /workflows/{workflowId}/connections [get]
func (h *ConnectionHandler) GetByWorkflowID(c echo.Context) error {
	workflowID, err := strconv.Atoi(c.Param("workflowId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid workflow ID"})
	}

	var connections []models.Connection
	if err := database.DB.Where("workflow_id = ?", workflowID).Find(&connections).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, connections)
}
