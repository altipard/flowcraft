package handlers

import (
	"net/http"
	"strconv"

	"github.com/altipard/flowcraft/internal/models"
	"github.com/altipard/flowcraft/internal/repository"
	"github.com/labstack/echo/v4"
)

// WorkflowHandler manages the workflow-related API endpoints
type WorkflowHandler struct {
	repo repository.WorkflowRepository
}

// NewWorkflowHandler creates a new WorkflowHandler instance
func NewWorkflowHandler() *WorkflowHandler {
	return &WorkflowHandler{}
}

// GetAll godoc
// @Summary Get all workflows
// @Description Returns a list of all available workflows
// @Tags workflows
// @Accept json
// @Produce json
// @Success 200 {array} models.Workflow
// @Router /workflows [get]
func (h *WorkflowHandler) GetAll(c echo.Context) error {
	workflows, err := h.repo.FindAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, workflows)
}

// GetByID godoc
// @Summary Get workflow by ID
// @Description Returns a specific workflow based on its ID
// @Tags workflows
// @Accept json
// @Produce json
// @Param id path int true "Workflow ID"
// @Success 200 {object} models.Workflow
// @Failure 404 {object} map[string]string
// @Router /workflows/{id} [get]
func (h *WorkflowHandler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	workflow, err := h.repo.FindByID(uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Workflow not found"})
	}

	return c.JSON(http.StatusOK, workflow)
}

// Create godoc
// @Summary Create a new workflow
// @Description Creates a new workflow with the provided data
// @Tags workflows
// @Accept json
// @Produce json
// @Param workflow body models.WorkflowRequest true "Workflow data"
// @Success 201 {object} models.Workflow
// @Failure 400 {object} map[string]string
// @Router /workflows [post]
func (h *WorkflowHandler) Create(c echo.Context) error {
	workflow := new(models.Workflow)
	if err := c.Bind(workflow); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := h.repo.Create(workflow); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, workflow)
}

// Update godoc
// @Summary Update a workflow
// @Description Updates an existing workflow with the provided data
// @Tags workflows
// @Accept json
// @Produce json
// @Param id path int true "Workflow ID"
// @Param workflow body models.WorkflowRequest true "Updated workflow data"
// @Success 200 {object} models.Workflow
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /workflows/{id} [put]
func (h *WorkflowHandler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	workflow, err := h.repo.FindByID(uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Workflow not found"})
	}

	if err := c.Bind(&workflow); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := h.repo.Update(&workflow); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, workflow)
}

// Delete godoc
// @Summary Delete a workflow
// @Description Deletes a workflow based on its ID
// @Tags workflows
// @Accept json
// @Produce json
// @Param id path int true "Workflow ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string
// @Router /workflows/{id} [delete]
func (h *WorkflowHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}
