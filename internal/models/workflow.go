package models

import "time"

// Workflow represents a workflow as a DAG of nodes.
type Workflow struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Nodes       map[string]*Node `json:"nodes"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}
