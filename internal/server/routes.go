package server

import (
	"goat/internal/models"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health())

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

// POST /api/workflows
func (s *Server) handleCreateWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	var wf models.Workflow
	if err := json.Unmarshal(body, &wf); err != nil {
		http.Error(w, "Malformed JSON", http.StatusBadRequest)
		return
	}

	// Generate UUID for workflow
	wf.ID = uuid.NewString()
	now := time.Now().UTC()
	wf.CreatedAt = now
	wf.UpdatedAt = now

	// Validation: must have at least one node
	if len(wf.Nodes) == 0 {
		http.Error(w, "Workflow must have at least one node", http.StatusBadRequest)
		return
	}
	// Validation: all next node IDs must exist
	for _, node := range wf.Nodes {
		for _, nextID := range node.Next {
			if _, ok := wf.Nodes[nextID]; !ok {
				http.Error(w, "Node "+node.ID+" references missing next node "+nextID, http.StatusBadRequest)
				return
			}
		}
	}

	// Uniqueness: check if workflow with same name exists
	existing, err := s.db.Read(r.Context(), "workflows", map[string]interface{}{"name": wf.Name})
	if err == nil && len(existing) > 0 {
		http.Error(w, "Workflow name must be unique", http.StatusConflict)
		return
	}

	// Insert workflow
	data, err := json.Marshal(wf)
	if err != nil {
		http.Error(w, "Failed to encode workflow", http.StatusInternalServerError)
		return
	}
	_, err = s.db.Create(r.Context(), "workflows", map[string]interface{}{
		"id":          wf.ID,
		"name":        wf.Name,
		"description": wf.Description,
		"nodes":       string(data), // store as JSON string
		"created_at":  wf.CreatedAt,
		"updated_at":  wf.UpdatedAt,
	})
	if err != nil {
		http.Error(w, "Failed to create workflow: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(wf)
}

// GET /api/workflows/{id}
func (s *Server) handleReadWorkflow(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/workflows/")
	if id == "" {
		http.Error(w, "Missing workflow id", http.StatusBadRequest)
		return
	}
	rows, err := s.db.Read(r.Context(), "workflows", map[string]interface{}{"id": id})
	if err != nil || len(rows) == 0 {
		http.Error(w, "Workflow not found", http.StatusNotFound)
		return
	}
	var wf models.Workflow
	if err := json.Unmarshal([]byte(rows[0]["nodes"].(string)), &wf); err != nil {
		http.Error(w, "Corrupt workflow data", http.StatusInternalServerError)
		return
	}
	// Fill top-level fields
	wf.ID = rows[0]["id"].(string)
	wf.Name = rows[0]["name"].(string)
	wf.Description, _ = rows[0]["description"].(string)
	wf.CreatedAt, _ = rows[0]["created_at"].(time.Time)
	wf.UpdatedAt, _ = rows[0]["updated_at"].(time.Time)
	json.NewEncoder(w).Encode(wf)
}

// DELETE /api/workflows/{id}
func (s *Server) handleDeleteWorkflow(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/workflows/")
	if id == "" {
		http.Error(w, "Missing workflow id", http.StatusBadRequest)
		return
	}
	count, err := s.db.Delete(r.Context(), "workflows", map[string]interface{}{"id": id})
	if err != nil {
		http.Error(w, "Failed to delete workflow", http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.Error(w, "Workflow not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/workflows
func (s *Server) handleListWorkflows(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Read(r.Context(), "workflows", map[string]interface{}{})
	if err != nil {
		http.Error(w, "Failed to list workflows", http.StatusInternalServerError)
		return
	}
	var workflows []models.Workflow
	for _, row := range rows {
		var wf models.Workflow
		if err := json.Unmarshal([]byte(row["nodes"].(string)), &wf); err != nil {
			continue
		}
		wf.ID = row["id"].(string)
		wf.Name = row["name"].(string)
		wf.Description, _ = row["description"].(string)
		wf.CreatedAt, _ = row["created_at"].(time.Time)
		wf.UpdatedAt, _ = row["updated_at"].(time.Time)
		workflows = append(workflows, wf)
	}
	json.NewEncoder(w).Encode(workflows)
}
