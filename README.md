# FlowCraft

FlowCraft is a powerful workflow automation engine built in Go. It allows you to create, manage, and execute complex workflows with different node types, enabling automation of various tasks and processes.

## Overview

FlowCraft provides a flexible API for defining and executing workflows. It is designed to be:

- **Modular**: Create workflows with reusable, connectable nodes
- **Extensible**: Add new node types to support different actions
- **Asynchronous**: Execute workflows in the background while monitoring their status
- **API-first**: Control everything through a RESTful API with Swagger documentation

## Features

- Create and manage workflows, nodes, and connections
- Execute workflows asynchronously with a queue-based worker system
- Support for multiple node types (HTTP requests, filters, transformations)
- Track the status and results of workflow executions
- API documentation with Swagger

## Installation

### Prerequisites

- Go 1.21 or later
- PostgreSQL database
- Redis (for the task queue)

### Setup

#### Option 1: Manual Setup

1. Clone the repository:

```bash
git clone https://github.com/yourusername/flowcraft.git
cd flowcraft
```

2. Install dependencies:

```bash
go mod tidy
```

3. Set up environment variables (create a `.env` file):

```
PORT=8080
DATABASE_URL=postgres://username:password@localhost:5432/flowcraft
REDIS_URL=redis://localhost:6379
```

4. Run the server:

```bash
go run cmd/server/main.go
```

#### Option 2: Docker Compose

The easiest way to get started is with Docker Compose, which will start all the required components:

1. Clone the repository:

```bash
git clone https://github.com/yourusername/flowcraft.git
cd flowcraft
```

2. Start the stack:

```bash
docker-compose up -d
```

This will start:
- PostgreSQL database
- Redis queue
- FlowCraft API server (accessible at http://localhost:8080)
- FlowCraft worker (with 3 worker processes)

To view logs:
```bash
docker-compose logs -f
```

To stop all services:
```bash
docker-compose down
```

### Worker Setup

FlowCraft uses a worker system to process and execute workflows asynchronously. The worker component pulls tasks from the Redis queue and executes the workflow steps.

1. Run the worker (in a separate terminal):

```bash
go run cmd/worker/main.go
```

You can run multiple worker instances for higher throughput:

```bash
# Start 3 worker instances
go run cmd/worker/main.go --workers=3
```

The worker uses the same environment variables as the server, so make sure your `.env` file is properly configured.

#### Worker Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `--workers` | 1 | Number of parallel worker goroutines |
| `--queue` | workflow_tasks | Name of the Redis queue to process |
| `--poll-interval` | 5s | How often to poll the queue if empty |
| `--execution-timeout` | 30m | Maximum execution time for a workflow |

## API Documentation

FlowCraft comes with built-in Swagger documentation.

### Generating Swagger Documentation

1. Install swag:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

2. Generate documentation:

```bash
swag init -g cmd/server/main.go
```

3. Access the documentation:

Once the server is running, navigate to `/swagger/index.html` in your browser to access the interactive API documentation.

## Node Executors

FlowCraft comes with several built-in node executors that perform different types of operations. Each node type has specific configuration options and input/output handling.

### HTTP Request Executor

The HTTP Request executor makes HTTP calls to external APIs or services.

**Purpose**: Fetch data from external APIs, send data to external systems, or trigger external processes.

**Configuration Options**:

| Option | Type | Description |
|--------|------|-------------|
| `url` | string | The URL to make the request to (required) |
| `method` | string | HTTP method (GET, POST, PUT, DELETE) |
| `headers` | object | HTTP headers to include with the request |
| `json_data` | object | JSON payload for POST/PUT requests |

**Example Configuration**:

```json
{
  "url": "https://api.example.com/data",
  "method": "POST",
  "headers": {
    "Authorization": "Bearer token123",
    "Content-Type": "application/json"
  },
  "json_data": {
    "name": "Test Item",
    "value": 42
  }
}
```

**Template Support**: The URL can include template placeholders using the format `{{key}}` which will be replaced with values from the input data.

**Example with Template**:

```json
{
  "url": "https://api.example.com/users/{{user_id}}",
  "method": "GET"
}
```

**Output**: Returns an object with `status_code` and `data` properties.

### Filter Executor

The Filter executor filters data based on specified conditions.

**Purpose**: Remove unwanted items from a dataset, select only the items that match certain criteria.

**Configuration Options**:

| Option | Type | Description |
|--------|------|-------------|
| `field` | string | The field path to check (supports dot notation for nested fields) |
| `operator` | string | Comparison operator: equals, not_equals, contains, greater_than, less_than |
| `value` | any | The value to compare against |

**Example Configuration**:

```json
{
  "field": "status",
  "operator": "equals",
  "value": "active"
}
```

**Example with Nested Fields**:

```json
{
  "field": "user.profile.age",
  "operator": "greater_than",
  "value": 18
}
```

**Input**: Array of objects to filter

**Output**: Filtered array containing only items that match the condition

### Transform Executor

The Transform executor maps data from one structure to another.

**Purpose**: Reshape data, select specific fields, rename fields, or create new computed fields.

**Configuration Options**:

| Option | Type | Description |
|--------|------|-------------|
| `mapping` | object | The template that defines how to transform input data |

**Example Configuration**:

```json
{
  "mapping": {
    "id": "{{id}}",
    "fullName": "{{firstName}} {{lastName}}",
    "contact": {
      "email": "{{email}}",
      "phone": "{{phone}}"
    },
    "isActive": true
  }
}
```

**Template Format**: Use `{{fieldPath}}` to reference fields from the input data, including nested paths with dot notation.

**Input**: Array of objects to transform

**Output**: Array of transformed objects according to the mapping template

## Extending FlowCraft with Custom Executors

FlowCraft supports extending the system with custom executors using Go plugins. This allows you to add custom functionality without modifying the core codebase.

### Creating a Custom Executor Plugin

Here's a step-by-step guide to create a custom executor plugin:

#### 1. Create a New Go Project for Your Plugin

Create a directory for your plugin:

```bash
mkdir -p plugins/math-executor
cd plugins/math-executor
```

Initialize a new Go module:

```bash
go mod init math-executor
```

#### 2. Implement the NodeExecutor Interface

Create a main.go file with your custom executor implementation:

```go
package main

import (
	"fmt"
	"strconv"
)

// MathExecutor performs basic math operations on input numbers
type MathExecutor struct{}

// NewExecutor is the exported function that FlowCraft will call to create your executor
// This function name is required and must return a NodeExecutor interface
func NewExecutor() interface{} {
	return &MathExecutor{}
}

// Execute implements the NodeExecutor interface
func (e *MathExecutor) Execute(config map[string]interface{}, input map[string]interface{}) (interface{}, error) {
	// Get operation from config
	operation, ok := config["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation is required in config")
	}

	// Get operands (can be from input or config)
	var value1, value2 float64
	var err error

	// First operand can come from config or input
	if val1Config, exists := config["value1"]; exists {
		// If value1 is directly in config
		if strVal, ok := val1Config.(string); ok {
			value1, err = strconv.ParseFloat(strVal, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid value1: %v", err)
			}
		} else if numVal, ok := val1Config.(float64); ok {
			value1 = numVal
		} else {
			return nil, fmt.Errorf("value1 must be a number")
		}
	} else if inputVal, exists := input["value1"]; exists {
		// If value1 is in input
		if numVal, ok := inputVal.(float64); ok {
			value1 = numVal
		} else {
			return nil, fmt.Errorf("input value1 must be a number")
		}
	} else {
		return nil, fmt.Errorf("value1 is required in config or input")
	}

	// Second operand, similar to first
	if val2Config, exists := config["value2"]; exists {
		if strVal, ok := val2Config.(string); ok {
			value2, err = strconv.ParseFloat(strVal, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid value2: %v", err)
			}
		} else if numVal, ok := val2Config.(float64); ok {
			value2 = numVal
		} else {
			return nil, fmt.Errorf("value2 must be a number")
		}
	} else if inputVal, exists := input["value2"]; exists {
		if numVal, ok := inputVal.(float64); ok {
			value2 = numVal
		} else {
			return nil, fmt.Errorf("input value2 must be a number")
		}
	} else {
		return nil, fmt.Errorf("value2 is required in config or input")
	}

	// Perform the operation
	var result float64
	switch operation {
	case "add":
		result = value1 + value2
	case "subtract":
		result = value1 - value2
	case "multiply":
		result = value1 * value2
	case "divide":
		if value2 == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = value1 / value2
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}

	// Return the result
	return map[string]interface{}{
		"result": result,
	}, nil
}
```

#### 3. Build the Plugin

Compile your plugin as a shared object (.so) file:

```bash
go build -buildmode=plugin -o math-executor.so .
```

#### 4. Move the Plugin to an Accessible Location

Create a plugins directory in your FlowCraft project (if it doesn't exist) and copy your plugin:

```bash
mkdir -p /path/to/flowcraft/plugins
cp math-executor.so /path/to/flowcraft/plugins/
```

### Using Your Custom Executor in FlowCraft

#### 1. Create a Node with Your Custom Executor

When creating a node, specify the node type using the `plugin:` prefix followed by the path to your plugin:

```bash
curl -X POST http://localhost:8080/api/nodes \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_id": 1,
    "node_type": "plugin:/path/to/flowcraft/plugins/math-executor.so",
    "position_x": 400,
    "position_y": 200,
    "name": "Calculate Sum",
    "config": "{\"operation\":\"add\",\"value1\":10,\"value2\":5}"
  }'
```

#### 2. Connect Your Node in a Workflow

Connect it like any other node, and the system will dynamically load and use your custom executor when the workflow runs.

### Example Workflow Using Custom Math Executor

Here's a complete example of a workflow that uses the custom math executor:

1. Create a workflow:

```bash
curl -X POST http://localhost:8080/api/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Math Workflow",
    "description": "Perform math operations"
  }'
```

2. Add an HTTP request node to get some data:

```bash
curl -X POST http://localhost:8080/api/nodes \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_id": 1,
    "node_type": "httpRequest",
    "position_x": 100,
    "position_y": 100,
    "name": "Get Numbers",
    "config": "{\"url\":\"https://api.example.com/numbers\",\"method\":\"GET\"}"
  }'
```

3. Add your custom math node:

```bash
curl -X POST http://localhost:8080/api/nodes \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_id": 1,
    "node_type": "plugin:/path/to/flowcraft/plugins/math-executor.so",
    "position_x": 300,
    "position_y": 100,
    "name": "Calculate",
    "config": "{\"operation\":\"multiply\"}"
  }'
```

4. Connect the nodes:

```bash
curl -X POST http://localhost:8080/api/connections \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_id": 1,
    "source_node_id": 1,
    "target_node_id": 2,
    "source_handle": "output",
    "target_handle": "input"
  }'
```

5. Run the workflow:

```bash
curl -X POST http://localhost:8080/api/workflows/1/execute
```

### Notes on Plugin Development

1. **Compatibility**: Your plugin must be compiled with the same version of Go as FlowCraft.

2. **Interface Compliance**: Your plugin must export a `NewExecutor()` function that returns an object implementing the `NodeExecutor` interface.

3. **Error Handling**: Proper error handling in your plugin is essential, as errors will be propagated to the workflow execution.

4. **Deployment**: When running in Docker, you need to mount your plugins directory into the container.

5. **Security**: Be cautious when loading plugins, as they run with the same privileges as the main application.

## Example: Creating a Simple Workflow

Here's an example of how to create a basic workflow that fetches data from an API and filters the results:

### 1. Create a Workflow

```bash
curl -X POST http://localhost:8080/api/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Fetch and Filter Data",
    "description": "Get data from an API and filter results"
  }'
```

Response:
```json
{
  "id": 1,
  "name": "Fetch and Filter Data",
  "description": "Get data from an API and filter results",
  "created_at": "2023-04-25T15:30:45Z",
  "updated_at": "2023-04-25T15:30:45Z"
}
```

### 2. Add HTTP Request Node

```bash
curl -X POST http://localhost:8080/api/nodes \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_id": 1,
    "node_type": "httpRequest",
    "position_x": 100,
    "position_y": 100,
    "name": "Fetch Data",
    "config": "{\"url\":\"https://api.example.com/data\",\"method\":\"GET\"}"
  }'
```

Response:
```json
{
  "id": 1,
  "workflow_id": 1,
  "node_type": "httpRequest",
  "position_x": 100,
  "position_y": 100,
  "name": "Fetch Data",
  "config": "{\"url\":\"https://api.example.com/data\",\"method\":\"GET\"}"
}
```

### 3. Add Filter Node

```bash
curl -X POST http://localhost:8080/api/nodes \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_id": 1,
    "node_type": "filter",
    "position_x": 300,
    "position_y": 100,
    "name": "Filter Results",
    "config": "{\"field\":\"status\",\"operator\":\"equals\",\"value\":\"active\"}"
  }'
```

Response:
```json
{
  "id": 2,
  "workflow_id": 1,
  "node_type": "filter",
  "position_x": 300,
  "position_y": 100,
  "name": "Filter Results",
  "config": "{\"field\":\"status\",\"operator\":\"equals\",\"value\":\"active\"}"
}
```

### 4. Connect the Nodes

```bash
curl -X POST http://localhost:8080/api/connections \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_id": 1,
    "source_node_id": 1,
    "target_node_id": 2,
    "source_handle": "output",
    "target_handle": "input"
  }'
```

Response:
```json
{
  "id": 1,
  "workflow_id": 1,
  "source_node_id": 1,
  "target_node_id": 2,
  "source_handle": "output",
  "target_handle": "input"
}
```

### 5. Execute the Workflow

```bash
curl -X POST http://localhost:8080/api/workflows/1/execute \
  -H "Content-Type: application/json" \
  -d '{}'
```

Response:
```json
{
  "execution_id": 1,
  "status": "pending"
}
```

### 6. Check Execution Status

```bash
curl -X GET http://localhost:8080/api/executions/1/status
```

Response:
```json
{
  "id": 1,
  "workflow_id": 1,
  "status": "completed",
  "started_at": "2023-04-25T15:35:10Z",
  "completed_at": "2023-04-25T15:35:12Z",
  "output_data": "{\"2\":[{\"id\":1,\"name\":\"Example\",\"status\":\"active\"}]}"
}
```

## System Architecture

FlowCraft follows a modular architecture:

- **API Server**: Handles HTTP requests and manages resources
- **Engine**: Core component that executes workflows
- **Worker**: Processes tasks from the queue and runs workflows
- **Queue**: Manages asynchronous execution of workflows
- **Database**: Stores workflows, nodes, connections, and execution history
- **Executors**: Implementations for different node types

![FlowCraft Architecture](docs/architecture.png)

### Workflow Execution Flow

1. Client sends a request to execute a workflow
2. API server validates the request and creates an execution record
3. Server enqueues the execution task in Redis
4. Worker picks up the task from the queue
5. Engine executes the workflow nodes in the correct order
6. Execution results are stored in the database
7. Client can query the execution status and results

## Roadmap

### Planned Features

- Implement more node types (database operations, email, messaging, etc.)
- Develop a frontend UI for visual workflow editing
- Add trigger functionality (webhooks, schedules, events)
- Improve error handling and retry mechanisms
- Add comprehensive logging and monitoring
- Support for conditional branches and loops
- User authentication and permissions
- Version control for workflows

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 