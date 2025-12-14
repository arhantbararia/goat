package manager

import (
	"fmt"

	"github.com/arhantbararia/goat/task"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Manager struct {
	Pending       queue.Queue
	TaskDb        map[string][]*task.Task
	EventDb       map[string][]*task.TaskEvent
	Workers       []string
	WorkerTreeMap map[string][]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string
}

func (m *Manager) SelectWorker() {
	fmt.Println("I select good workers")
}

func (m *Manager) UpdateTasks() {
	fmt.Println("I will update Tasks")
}

func (m *Manager) SendWork() {
	fmt.Println("I will send work to workers")
}
