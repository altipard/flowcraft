package main

import (
	"net/http"
	"os"

	_ "github.com/altipard/flowcraft/docs" // Import Swagger documentation files
	"github.com/altipard/flowcraft/internal/database"
	"github.com/altipard/flowcraft/internal/handlers"
	"github.com/altipard/flowcraft/internal/queue"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title FlowCraft API
// @version 1.0
// @description API Server for FlowCraft - Workflow Management System
// @host localhost:8080
// @BasePath /api
func main() {
	// Load environment variables
	godotenv.Load()

	// Initialize database connection
	database.Initialize(os.Getenv("DATABASE_URL"))

	// Initialize queue client
	queueClient, err := queue.NewQueueClient(os.Getenv("REDIS_URL"))
	if err != nil {
		panic(err)
	}

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.Static("./web/dist"))

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Handlers
	workflowHandler := handlers.NewWorkflowHandler()
	nodeHandler := handlers.NewNodeHandler()
	connectionHandler := handlers.NewConnectionHandler()
	executionHandler := handlers.NewExecutionHandler(queueClient)

	// API routes
	api := e.Group("/api")
	{
		// Workflow routes
		workflows := api.Group("/workflows")
		workflows.GET("", workflowHandler.GetAll)
		workflows.GET("/:id", workflowHandler.GetByID)
		workflows.POST("", workflowHandler.Create)
		workflows.PUT("/:id", workflowHandler.Update)
		workflows.DELETE("/:id", workflowHandler.Delete)
		workflows.POST("/:id/execute", executionHandler.ExecuteWorkflow) // <-- Important: Execution route

		// Node routes
		nodes := api.Group("/nodes")
		nodes.GET("", nodeHandler.GetAll)
		nodes.GET("/:id", nodeHandler.GetByID)
		nodes.POST("", nodeHandler.Create)
		nodes.PUT("/:id", nodeHandler.Update)
		nodes.DELETE("/:id", nodeHandler.Delete)

		// Connection routes
		connections := api.Group("/connections")
		connections.GET("", connectionHandler.GetAll)
		connections.GET("/:id", connectionHandler.GetByID)
		connections.POST("", connectionHandler.Create)
		connections.PUT("/:id", connectionHandler.Update)
		connections.DELETE("/:id", connectionHandler.Delete)

		// Execution routes
		executions := api.Group("/executions")
		executions.GET("/:id/status", executionHandler.GetStatus)
	}

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "FlowCraft API Server is running!")
	})

	e.GET("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "pong"})
	})

	// Start server
	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}
