package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/altipard/flowcraft/internal/database"
	"github.com/altipard/flowcraft/internal/engine"
	"github.com/altipard/flowcraft/internal/queue"
	"github.com/joho/godotenv"
)

// WorkflowExecutionPayload is the payload for workflow execution tasks
type WorkflowExecutionPayload struct {
	ExecutionID uint `json:"execution_id"`
}

func main() {
	// Parse command line flags
	numWorkers := flag.Int("workers", 1, "Number of parallel worker goroutines")
	queueName := flag.String("queue", "workflow_tasks", "Name of the Redis queue to process")
	pollInterval := flag.Duration("poll-interval", 5*time.Second, "How often to poll the queue if empty")
	executionTimeout := flag.Duration("execution-timeout", 30*time.Minute, "Maximum execution time for a workflow")
	flag.Parse()

	log.Printf("Starting worker with configuration: workers=%d, queue=%s, poll-interval=%s, execution-timeout=%s\n", 
		*numWorkers, *queueName, *pollInterval, *executionTimeout)

	// Load environment variables
	godotenv.Load()

	// Initialize database connection
	database.Initialize(os.Getenv("DATABASE_URL"))

	// Initialize queue client
	queueClient, err := queue.NewQueueClient(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize workflow engine
	workflowEngine := engine.NewEngine()

	// Channel for graceful shutdown
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	// Use a WaitGroup to manage worker goroutines
	var wg sync.WaitGroup
	
	// Launch worker goroutines
	for i := 1; i <= *numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Printf("Worker %d started", workerID)
			
			// Create a context with timeout for each workflow execution
			for {
				select {
				case <-stopCh:
					log.Printf("Worker %d received shutdown signal", workerID)
					return
				default:
					// Dequeue task from the queue
					task, err := queueClient.DequeueTask(*queueName, *pollInterval)
					if err != nil {
						log.Printf("Worker %d: Error dequeuing task: %v", workerID, err)
						continue
					}

					// If no task is available, try again
					if task == nil {
						continue
					}

					log.Printf("Worker %d: Processing task: %s", workerID, task.TaskType)

					// Check task type and process accordingly
					switch task.TaskType {
					case "execute_workflow":
						var payload WorkflowExecutionPayload
						if err := json.Unmarshal(task.Payload, &payload); err != nil {
							log.Printf("Worker %d: Error unmarshalling payload: %v", workerID, err)
							continue
						}

						// Execute workflow with timeout
						executionDone := make(chan struct{})
						go func() {
							defer close(executionDone)
							if err := workflowEngine.ExecuteWorkflow(payload.ExecutionID); err != nil {
								log.Printf("Worker %d: Error executing workflow %d: %v", workerID, payload.ExecutionID, err)
							}
						}()

						// Wait for execution to complete or timeout
						select {
						case <-executionDone:
							log.Printf("Worker %d: Workflow %d execution completed", workerID, payload.ExecutionID)
						case <-time.After(*executionTimeout):
							log.Printf("Worker %d: Workflow %d execution timed out after %s", workerID, payload.ExecutionID, *executionTimeout)
							// TODO: Update workflow execution status to failed due to timeout
						}

					default:
						log.Printf("Worker %d: Unknown task type: %s", workerID, task.TaskType)
					}
				}
			}
		}(i)
	}

	// Wait for shutdown signal
	<-stopCh
	log.Println("Shutting down workers gracefully...")
	
	// Use a separate channel to signal forced shutdown after timeout
	forceShutdown := make(chan struct{})
	go func() {
		wg.Wait()
		close(forceShutdown)
	}()

	// Wait for graceful shutdown or force after 10 seconds
	select {
	case <-forceShutdown:
		log.Println("All workers gracefully stopped")
	case <-time.After(10 * time.Second):
		log.Println("Forcing shutdown after timeout")
	}
}
