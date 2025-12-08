package task

import (
	"time"

	"github.com/google/uuid"
)

type TaskEvent struct {
	ID        uuid.UUID
	State     State
	TimeStamp time.Time
	Task Task
}



