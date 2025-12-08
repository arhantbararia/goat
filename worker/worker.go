package worker

import (
	"fmt"

	"github.com/arhantbararia/goat/task"
	"github.com/docker/docker/libcontainerd/queue"
	"github.com/google/uuid"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	Db        map[uuid.UUID]*task.Task
	TaskCount int
}

func (w *Worker) CollectState() {
	fmt.Println("I will collect statistics")

}

func (w *Worker) RunTask() {
	fmt.Println("I will Run Task")

}

func (w *Worker) StartTask() {
	fmt.Println("I will start a task")
}

func (w *Worker) StopTask() {
	fmt.Println("I will stop a task")

}
