package task

import (
	"context"
	"io"
	"log"
	"math"
	"os"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/uuid"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

type State int

const (
	Pending State = iota
	Scheduled
	Running
	Completed
	Failed
)

var stateTransitionMap = map[State][]State{
	Pending:   []State{Scheduled},
	Scheduled: []State{Scheduled, Running, Failed},
	Running:   []State{Running, Completed, Failed},
	Completed: []State{},
	Failed:    []State{},
}

func Contains(states []State, state State) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}

func ValidaStateTransition(src State, dst State) bool {
	return Contains(stateTransitionMap[src], dst)
}

type Task struct {
	ID            uuid.UUID
	ContainerID   string
	Name          string
	State         State
	Image         string
	Memory        int //required memory
	Disk          int //required disk space
	ExposedPorts  network.PortSet
	PortBindings  map[string]string
	RestartPolicy string
	StartTime     time.Time
	FinishTime    time.Time
}

type Config struct {
	Name          string
	ContainerID   string
	AttachStdin   bool
	AttachStdout  bool
	AttachStderr  bool
	ExposedPorts  network.PortSet
	Cmd           []string
	Image         string
	Cpu           float64
	Memory        int64
	Disk          int64
	Env           []string
	RestartPolicy string
}

func NewConfig(task *Task) Config {
	return Config{
		Name:          task.Name,
		ContainerID:   task.ContainerID,
		ExposedPorts:  task.ExposedPorts,
		Image:         task.Image,
		Memory:        int64(task.Memory),
		Disk:          int64(task.Disk),
		RestartPolicy: task.RestartPolicy,
	}
}

type Docker struct {
	Client client.Client
	Config Config
}

type DockerResult struct {
	Error       error
	Action      string
	ContainerId string
	Result      string
}

func NewDocker(conf Config) *Docker {
	new_client, err := client.New()
	if err != nil || new_client == nil {
		log.Println("Docker daemon unreachable")
	}

	return &Docker{
		Client: *new_client,
		Config: conf,
	}
}

func (d *Docker) Run() DockerResult {
	ctx := context.Background()

	reader, err := d.Client.ImagePull(
		ctx,
		d.Config.Image,
		client.ImagePullOptions{},
	)
	if err != nil {
		log.Printf("Error Pulling image: %s: %v\n", d.Config.Image, err)
	}

	io.Copy(os.Stdout, reader)

	rp := container.RestartPolicy{
		Name: container.RestartPolicyMode(d.Config.RestartPolicy),
	}

	r := container.Resources{
		Memory:   d.Config.Memory,
		NanoCPUs: int64(d.Config.Cpu * math.Pow(10, 9)),
	}

	cc := container.Config{
		Image:        d.Config.Image,
		Tty:          false,
		Env:          d.Config.Env,
		ExposedPorts: d.Config.ExposedPorts,
	}

	hc := container.HostConfig{
		RestartPolicy:   rp,
		Resources:       r,
		PublishAllPorts: true,
	}

	// &cc,
	// 	&hc,
	// 	nil,
	// 	nil,
	// 	d.Config.Name

	resp, err := d.Client.ContainerCreate(
		ctx,
		client.ContainerCreateOptions{
			Config:           &cc,
			HostConfig:       &hc,
			NetworkingConfig: nil,
			Platform:         nil,
			Name:             d.Config.Name,
		},
	)

	if err != nil {
		log.Printf("error creating the container using image: %s, %v \n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	_, err = d.Client.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{})
	if err != nil {
		log.Printf("error starting the container. ID: %s, %v \n", resp.ID, err)
		return DockerResult{Error: err}
	}

	d.Config.ContainerID = resp.ID

	out, err := d.Client.ContainerLogs(
		ctx,
		resp.ID,
		client.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true},
	)
	if err != nil {
		log.Printf("Error getting logs for container %s: %v\n", resp.ID, err)
		return DockerResult{Error: err}
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	return DockerResult{
		ContainerId: resp.ID,
		Action:      "start",
		Result:      "success",
	}

}

func (d *Docker) Stop(id string) DockerResult {
	log.Printf("Attempting to stop container. ID: %v \n", id)
	ctx := context.Background()
	_, err := d.Client.ContainerStop(ctx, id, client.ContainerStopOptions{})
	if err != nil {
		log.Printf("Error stopping container %s: %v\n", id, err)
		return DockerResult{Error: err}
	}

	_, err = d.Client.ContainerRemove(ctx, id, client.ContainerRemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		Force:         false,
	})

	if err != nil {
		log.Printf("Error removing container %s: %v\n", id, err)
		return DockerResult{Error: err}
	}

	return DockerResult{
		Action: "stop",
		Result: "success",
		Error:  nil,
	}

}
