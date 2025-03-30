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