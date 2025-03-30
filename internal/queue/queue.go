package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// QueueClient is a client for the message queue
type QueueClient struct {
	redisClient *redis.Client
}

// TaskMessage represents a task in the queue
type TaskMessage struct {
	TaskType string          `json:"task_type"`
	Payload  json.RawMessage `json:"payload"`
}

// NewQueueClient creates a new QueueClient
func NewQueueClient(redisURL string) (*QueueClient, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(options)

	// Test the connection
	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return &QueueClient{
		redisClient: client,
	}, nil
}

// EnqueueTask adds a task to the queue
func (q *QueueClient) EnqueueTask(queueName string, taskType string, payload interface{}) error {
	ctx := context.Background()

	// Serialize payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Create task
	task := TaskMessage{
		TaskType: taskType,
		Payload:  payloadBytes,
	}

	// Serialize task
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %v", err)
	}

	// Add task to queue
	err = q.redisClient.RPush(ctx, queueName, taskBytes).Err()
	if err != nil {
		return fmt.Errorf("failed to push task to queue: %v", err)
	}

	return nil
}

// DequeueTask retrieves a task from the queue
func (q *QueueClient) DequeueTask(queueName string, timeout time.Duration) (*TaskMessage, error) {
	ctx := context.Background()

	// Get task from queue with timeout
	result, err := q.redisClient.BLPop(ctx, timeout, queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No task in queue
		}
		return nil, fmt.Errorf("failed to pop task from queue: %v", err)
	}

	// We receive a slice [queueName, value]
	if len(result) != 2 {
		return nil, fmt.Errorf("unexpected result from BLPOP: %v", result)
	}

	// Deserialize task
	var task TaskMessage
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %v", err)
	}

	return &task, nil
}
