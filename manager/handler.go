package manager

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/arhantbararia/goat/task"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ErrResponse struct {
	HTTPStatusCode int
	Message        string
}

func (a *API) StartTaskHandler(w http.ResponseWriter, r *http.Request) {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	te := task.TaskEvent{}
	err := d.Decode(&te)
	if err != nil {
		msg := fmt.Sprintf("Error serializing body: %v ", err)
		log.Printf(msg)
		e := ErrResponse{
			HTTPStatusCode: 400,
			Message:        msg,
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	a.Manager.AddTask(te)
	log.Println("Added Task: ", te.Task.ID)
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(te.Task)

}

func (a *API) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(a.Manager.GetTasks())
}

func (a *API) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	if taskID == "" {
		log.Println("no tasks id in request")
		w.WriteHeader(400)
	}

	tId, _ := uuid.Parse(taskID)
	taskToStop, ok := a.Manager.TaskDb[tId]
	if !ok {
		log.Println("No task with task id: ", tId)
		w.WriteHeader(400)
	}

	te := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Completed,
		TimeStamp: time.Now(),
	}

	taskCopy := *taskToStop
	taskCopy.State = task.Completed
	te.Task = taskCopy
	a.Manager.AddTask(te)

	log.Printf("Added task event %v to stop %v \n", te.ID, taskToStop.ID)
	w.WriteHeader(204)

}
