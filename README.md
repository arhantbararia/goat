# GOAT 🐐
*Go Orchestration & Automation Toolkit*

## Introduction

**GOAT** is a lightweight, distributed task orchestration engine written in Go. Inspired by platforms like Kubernetes, it provides a simple yet powerful foundation for scheduling and running containerized tasks across a cluster of worker nodes. It's designed for developers and DevOps engineers who need a simple orchestrator for managing distributed workloads without the complexity of a full-blown container platform.

---

## Core Concepts & Architecture

GOAT operates on a manager-worker architecture:
-   **Manager**: The central control plane. It exposes a REST API for users to submit tasks. The manager maintains the state of all tasks, selects appropriate workers, and dispatches the work.
-   **Worker**: A node responsible for executing tasks. It receives instructions from the manager, runs the specified Docker container, and reports the task's status back to the manager.

---

## Features

-   **Distributed Task Execution**: Run tasks as Docker containers across multiple worker nodes.
-   **Manager-Worker Architecture**: A central manager orchestrates tasks on a cluster of workers.
-   **State Management**: Tracks the lifecycle of each task (e.g., `Scheduled`, `Running`, `Completed`, `Failed`).
-   **REST API**: Simple HTTP-based API to submit, view, and stop tasks.
-   **Round-Robin Scheduling**: Basic load distribution by assigning tasks to workers in a round-robin fashion.
-   **Built in Go**: A single, statically-linked binary for both manager and worker components, ensuring easy deployment.

---

## Getting Started

The `main.go` file provides a simple demonstration that runs one manager and one worker on your local machine.

```bash
# Clone the repository
git clone https://github.com/your-org/goat.git
cd goat

# Tidy dependencies
go mod tidy

# Run the demo (starts one manager and one worker)
go run main.go
```

---
## Project Structure

-   `manager/` – Contains the logic for the central manager node, including task scheduling and worker communication.
-   `worker/` – Contains the logic for worker nodes, including task execution via Docker.
-   `task/` – Defines the core `Task` data structures and Docker interaction logic.
-   `main.go` – The main application entrypoint for running a demo cluster.
---


## API Usage

You can interact with the manager's API to control tasks.

**Submit a New Task:**
```http
POST /tasks
"Task": {
        "Name": "my-new-task",
        "Image": "strm/helloworld-http"
    }
}
```

**List All Tasks:**
```http
GET /tasks
```

**Stop a Task:**
```http
DELETE /tasks/{taskID}
```

---

## Roadmap

-   [ ] **Persistence**: Add a database layer (e.g., SQLite, Postgres) to persist task state.
-   [ ] **Advanced Scheduling**: Implement resource-aware scheduling instead of simple round-robin.
-   [ ] **Fault Tolerance**: Improve handling of worker failures and enable task retries.
-   [ ] **Worker Discovery**: Implement a mechanism for workers to dynamically register with the manager.
-   [ ] **CLI Tool**: Develop a command-line interface for interacting with the manager.
-   [ ] **Improved API**: Enhance the API with more endpoints and better error handling.

---

## Contributing

Contributions are welcome! To get started:

1.  Fork the repo and create your branch
2.  Submit pull requests for features, bugfixes, or documentation improvements.
3.  See `CONTRIBUTING.md` for more detailed guidelines.

---

## License

MIT License

---

## Credits

Inspired by container orchestration platforms like Kubernetes and HashiCorp Nomad.

```
goat
├─ .air.toml
├─ cmd
│  └─ api
├─ docker-compose.yml
├─ go.mod
├─ go.sum
├─ internal
│  ├─ database
│  │  ├─ database.go
│  │  ├─ database_test.go
│  │  ├─ mongodb.go
│  │  ├─ mysql.go
│  │  ├─ postgres.go
│  │  └─ sqlite.go
│  ├─ models
│  │  ├─ engine.go
│  │  ├─ example_workflow.json
│  │  ├─ executor.go
│  │  ├─ node.go
│  │  ├─ registry.go
│  │  ├─ runner.go
│  │  ├─ send_message.go
│  │  └─ workflow.go
│  └─ server
│     ├─ routes.go
│     ├─ routes_test.go
│     ├─ server.go
│     └─ urls.go
├─ Makefile
├─ pkg
├─ README.md
└─ scripts

```