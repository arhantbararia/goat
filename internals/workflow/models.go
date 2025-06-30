package workflow

import "time"

// NodeType represents the type of a workflow node.
type NodeType string

const (
	NodeTypeTrigger NodeType = "trigger"
	NodeTypeAction  NodeType = "action"
	NodeTypeCondition NodeType = "condition"
)

// Node represents a single node in the workflow DAG.
type Node struct {
	ID       string                 `json:"id"`
	Type     NodeType               `json:"type"`
	Name     string                 `json:"name"`
	Plugin   string                 `json:"plugin"` // e.g., "github.issue.created"
	Inputs   map[string]interface{} `json:"inputs,omitempty"`
	Outputs  map[string]interface{} `json:"outputs,omitempty"`
	Next     []string               `json:"next,omitempty"` // IDs of next nodes
}

// Workflow represents a workflow as a DAG of nodes.
type Workflow struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Nodes       map[string]*Node  `json:"nodes"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Trigger represents a trigger node.
type Trigger struct {
	Node
	// Additional trigger-specific fields can be added here.
}

// Action represents an action node.
type Action struct {
	Node
	// Additional action-specific fields can be added here.
}

// Execution represents a workflow execution instance.
type Execution struct {
	ID         string                 `json:"id"`
	WorkflowID string                 `json:"workflow_id"`
	Status     string                 `json:"status"` // e.g., "pending", "running", "success", "failed"
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt *time.Time             `json:"finished_at,omitempty"`
	Results    map[string]*Result     `json:"results"` // nodeID -> Result
	Context    map[string]interface{} `json:"context,omitempty"`
}

// Result represents the result of a node execution.
type Result struct {
	NodeID    string                 `json:"node_id"`
	Success   bool                   `json:"success"`
	Output    map[string]interface{} `json:"output,omitempty"`
	Error     string                 `json:"error,omitempty"`
	StartedAt time.Time              `json:"started_at"`
	EndedAt   time.Time              `json:"ended_at"`
}
