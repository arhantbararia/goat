package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/arhantbararia/goat/task"
	"github.com/go-chi/chi"
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
		msg := fmt.Sprintf("start task request parsing error: %v", err)
		log.Println(msg)
		w.WriteHeader(400)
		e := ErrResponse{
			HTTPStatusCode: 400,
			Message:        msg,
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	a.Worker.AddTask(te.Task)
	log.Printf("Added task : %v \n ", te.Task.ID)
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(te.Task)

}

func (a *API) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "applicaation/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(a.Worker.GetTasks())
}

func (a *API) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")

	if taskID == "" {
		log.Println("No taskID passed in request")
		w.WriteHeader(400)
		e := ErrResponse{
			HTTPStatusCode: 400,
			Message:        "Task Id not in request URL",
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	tID, _ := uuid.Parse(taskID)
	_, ok := a.Worker.Db[tID]
	if !ok {
		log.Println("No tasks with ID : ", tID)
		w.WriteHeader(400)
		e := ErrResponse{
			HTTPStatusCode: 404,
			Message:        "No Tasks found for given ID",
		}
		json.NewEncoder(w).Encode(e)

	}

	taskToStop := a.Worker.Db[tID]
	taskCopy := *taskToStop
	taskCopy.State = task.Completed
	a.Worker.AddTask(taskCopy)

	log.Println("Added Task :", taskToStop.ID)
	log.Println("Stopping Container :", taskToStop.ContainerID)

	w.WriteHeader(204)

}
