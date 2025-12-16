package main

import (
	"fmt"
	"os"
	"time"

	"github.com/arhantbararia/goat/task"
	"github.com/moby/moby/client"
)

func main() {

	fmt.Printf("create a test container\n")
	dockerTask, createResult := createContainer()
	if createResult.Error != nil {
		fmt.Printf("%v", createResult.Error)
		os.Exit(1)
	}

	time.Sleep(time.Second * 5)
	fmt.Printf("stopping container %s\n", createResult.ContainerId)
	_ = stopContainer(dockerTask, createResult.ContainerId)

}

func createContainer() (*task.Docker, *task.DockerResult) {
	c := task.Config{
		Name:  "test-container-1",
		Image: "postgres:13",
		Env: []string{
			"POSTGRES_USER=test",
			"POSTGRES_PASSWORD=secret",
		},
	}

	dc, _ := client.New(client.FromEnv)
	d := task.Docker{
		Client: *dc,
		Config: c,
	}

	result := d.Run()
	if result.Error != nil {
		fmt.Printf("%v\n", result.Error)
		return nil, nil
	}

	fmt.Printf("Container %s is running with config %v\n", result.ContainerId, c)
	return &d, &result

}

func stopContainer(d *task.Docker, id string) *task.DockerResult {
	result := d.Stop(id)
	if result.Error != nil {
		fmt.Printf("%v\n", result.Error)
		return nil
	}

	fmt.Printf(
		"Container %s has been stopped and removed\n", result.ContainerId)
	return &result
}

// t := task.Task{
// 	ID:     uuid.New(),
// 	Name:   "Task-1",
// 	State:  task.Pending,
// 	Image:  "image-1",
// 	Memory: 1024,
// 	Disk:   1,
// }

// te := task.TaskEvent{
// 	ID:        uuid.New(),
// 	State:     task.Pending,
// 	TimeStamp: time.Now(),
// 	Task:      t,
// }

// fmt.Println(t)

// fmt.Println(te)

// w := worker.Worker{
// 	Name:  "worker-1",
// 	Queue: *queue.New(),
// 	Db:    make(map[uuid.UUID]*task.Task),
// }
// fmt.Println(w)
// w.CollectState()
// w.RunTask()
// w.StartTask()
// w.StopTask()

// m := manager.Manager{
// 	Pending: *queue.New(),
// 	TaskDb:  make(map[string][]*task.Task),
// 	EventDb: make(map[string][]*task.TaskEvent),
// 	Workers: []string{w.Name},
// }

// fmt.Println(m)
// m.SelectWorker()
// m.UpdateTasks()
// m.SendWork()
// n := node.Node{
// 	Name:   "Node-1",
// 	Ip:     "192.168.1.1",
// 	Cores:  4,
// 	Memory: 1024,
// 	Disk:   25,
// 	Role:   "worker",
// }
// fmt.Printf("node: %v\n", n)
