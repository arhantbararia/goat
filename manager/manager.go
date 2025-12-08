package manager

import (
	task "github.com/arhantbararia/goat/Task"
	"github.com/docker/docker/libcontainerd/queue"
	"github.com/google/uuid"
)

type Manager struct {
	Pending       queue.Queue
	TaskDb        map[string][]*task.Task
	EventDb       map[string][]*task.TaskEvent
	Workers       []string
	WorkerTreeMap map[string][]uuid.NewUUID
	TaskWorkerMap map[uuid.UUID]string
}

