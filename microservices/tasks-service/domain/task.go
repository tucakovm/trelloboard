package domain

import (
	"github.com/google/uuid"
)

type Task struct {
	Id          uuid.UUID
	Name        string
	Description string
	Status      Status
}

type TasksRepository interface {
	Create(task Task) (Task, error)
}
