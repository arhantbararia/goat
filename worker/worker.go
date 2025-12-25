package worker

import (
	"fmt"
	"log"
	"time"

	"github.com/arhantbararia/goat/task"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

var WORKER_SLEEP_TIME = 15

type Worker struct {
	Name      string
	Queue     queue.Queue
	Db        map[uuid.UUID]*task.Task
	TaskCount int
	Stats     *Stats
}

func (w *Worker) runTask() task.DockerResult {
	t := w.Queue.Dequeue()
	if t == nil {
		log.Println("No task in the queue")
		return task.DockerResult{Error: nil}

	}

	taskQueued := t.(task.Task) //proper type conversion after queue retrieval

	taskPersisted := w.Db[taskQueued.ID]
	if taskPersisted == nil {
		//task appeared first time
		taskPersisted = &taskQueued
		w.Db[taskQueued.ID] = taskPersisted
	}

	var result task.DockerResult

	if task.ValidaStateTransition(taskPersisted.State, taskQueued.State) {

		switch taskQueued.State {
		case task.Scheduled:
			result = w.StartTask(taskQueued)
		case task.Completed:
			result = w.StopTask(taskQueued)
		default:
			result.Error = fmt.Errorf("this is unexpected")
		}

	} else {
		err := fmt.Errorf("invalid Transition from %v --> %v ", taskPersisted.State, taskQueued.State)
		result.Error = err

	}

	return result

}

func (w *Worker) AddTask(t task.Task) {
	w.Queue.Enqueue(t)
}

func (w *Worker) StartTask(t task.Task) task.DockerResult {

	t.StartTime = time.Now().UTC()

	config := task.NewConfig(&t)
	dock := task.NewDocker(config)

	result := dock.Run()

	if result.Error != nil {
		log.Printf("Err running task %v: %v\n", t.ID, result.Error)
		t.State = task.Failed
		w.Db[t.ID] = &t
		return result
	}

	t.ContainerID = result.ContainerId
	t.State = task.Running
	w.Db[t.ID] = &t

	return result

}

func (w *Worker) StopTask(t task.Task) task.DockerResult {
	config := task.NewConfig(&t)
	dock := task.NewDocker(config)

	result := dock.Stop(t.ContainerID)

	if result.Error != nil {
		log.Printf("error stopping container: %v , %v \n", t.ContainerID, result.Error)
		return result
	}

	t.FinishTime = time.Now().UTC()
	t.State = task.Completed
	w.Db[t.ID] = &t

	log.Printf("stopped and removed container %v for task %v \n", t.ContainerID, t.ID)

	return result

}

func (w *Worker) GetTasks() []task.Task {
	//returns all tasks
	tasks := []task.Task{}

	for _, value := range w.Db {
		tasks = append(tasks, *value)

	}

	return tasks

}

func (w *Worker) CollectStats() {
	for {
		log.Println("Collecting state")
		w.Stats = GetStats()
		w.Stats.TaskCount = w.TaskCount
		time.Sleep(15 * time.Second)
	}
}

func (w *Worker) RunTasks() {
	fmt.Println("Running Task collection Loop")
	for {

		fmt.Println("Queued Tasks: ", w.Queue.Len())
		if w.Queue.Len() != 0 {
			result := w.runTask()
			if result.Error != nil {
				log.Println("Error running task- ", result.Error)
			} else {
				log.Println("No tasks to process currently")
			}

		} else {
			log.Printf("Sleeping for %v seconds", WORKER_SLEEP_TIME)
			time.Sleep(time.Duration(WORKER_SLEEP_TIME) * time.Second)

		}
	}
}
