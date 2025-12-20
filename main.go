package main

import (
	"fmt"
	"log"
	"time"

	"github.com/arhantbararia/goat/task"
	"github.com/arhantbararia/goat/worker"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

var SLEEP_TIME = 7

func main() {

	host := "127.0.0.1"
	port := 8000

	fmt.Println("Starting Goat worker")
	w := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}

	//runtime loop
	go runTasks(&w)

	api := worker.API{
		Address: host,
		Port:    port,
		Worker:  &w,
	}

	api.Start()

}

func runTasks(w *worker.Worker) {
	fmt.Println("Running Task collection Loop")
	for {

		fmt.Println("Queued Tasks: ", w.Queue.Len())
		if w.Queue.Len() != 0 {
			result := w.RunTask()
			if result.Error != nil {
				log.Println("Error running task- ", result.Error)
			} else {
				log.Println("No tasks to process currently")
			}

		} else {
			log.Printf("Sleeping for %v seconds", SLEEP_TIME)
			time.Sleep(time.Duration(SLEEP_TIME) * time.Second)

		}
	}
}

//// //////////////////////// using task structs raw
// func main() {

// 	fmt.Printf("create a test container\n")
// 	dockerTask, createResult := createContainer()
// 	if createResult.Error != nil {
// 		fmt.Printf("%v", createResult.Error)
// 		os.Exit(1)
// 	}

// 	time.Sleep(time.Second * 5)
// 	fmt.Printf("stopping container %s\n", createResult.ContainerId)
// 	_ = stopContainer(dockerTask, createResult.ContainerId)

// }

// func createContainer() (*task.Docker, *task.DockerResult) {
// 	c := task.Config{
// 		Name:  "test-container-1",
// 		Image: "postgres:13",
// 		Env: []string{
// 			"POSTGRES_USER=test",
// 			"POSTGRES_PASSWORD=secret",
// 		},
// 	}

// 	dc, _ := client.New(client.FromEnv)
// 	d := task.Docker{
// 		Client: *dc,
// 		Config: c,
// 	}

// 	result := d.Run()
// 	if result.Error != nil {
// 		fmt.Printf("%v\n", result.Error)
// 		return nil, nil
// 	}

// 	fmt.Printf("Container %s is running with config %v\n", result.ContainerId, c)
// 	return &d, &result

// }

// func stopContainer(d *task.Docker, id string) *task.DockerResult {
// 	result := d.Stop(id)
// 	if result.Error != nil {
// 		fmt.Printf("%v\n", result.Error)
// 		return nil
// 	}

// 	fmt.Printf(
// 		"Container %s has been stopped and removed\n", result.ContainerId)
// 	return &result
// }

// //////////////// Using work structs RAW /////////////

// func main() {
// 	db := make(map[uuid.UUID]*task.Task)
// 	w := worker.Worker{
// 		Queue: *queue.New(),
// 		Db:    db,
// 	}
// 	PORT, _ := network.PortFrom(80, "tcp")

// 	ports := network.PortSet{PORT: struct{}{}}

// 	t := task.Task{
// 		ID:           uuid.New(),
// 		Name:         "test-http-container-1",
// 		State:        task.Scheduled,
// 		Image:        "strm/helloworld-http",
// 		ExposedPorts: ports,
// 	}

// 	// First time the worker will see the task
// 	fmt.Println("Starting Task")
// 	w.AddTask(t)
// 	result := w.RunTask()
// 	if result.Error != nil {
// 		panic(result.Error)
// 	}

// 	t.ContainerID = result.ContainerId
// 	fmt.Printf("Task %s us running in container %s \n", t.Name, t.ContainerID)

// 	fmt.Println("Starting Up time rest")
// 	time.Sleep(time.Second * 10)

// 	// Test the running HTTP server
// 	fmt.Println("Testing HTTP server...")
// 	resp, err := http.Get("http://localhost:80") // Assuming the container exposes port 80
// 	if err != nil {
// 		log.Printf("Error sending HTTP request: %v", err)
// 	} else {
// 		defer resp.Body.Close()
// 		body, err := ioutil.ReadAll(resp.Body)
// 		if err != nil {
// 			log.Printf("Error reading HTTP response body: %v", err)
// 		}
// 		fmt.Printf("HTTP Response Status: %d\n", resp.StatusCode)
// 		fmt.Printf("HTTP Response Body: %s\n", body)
// 		if resp.StatusCode != http.StatusOK {
// 			log.Fatalf("HTTP server returned non-200 status code: %d", resp.StatusCode)
// 		}
// 	}

// 	//sleep for 10 seconds
// 	fmt.Println("Sleepy time")
// 	time.Sleep(time.Second * 30)

// 	fmt.Println("Stopping task")
// 	t.State = task.Completed
// 	w.AddTask(t)
// 	result = w.RunTask()
// 	if result.Error != nil {
// 		panic(result.Error)
// 	}

// }
