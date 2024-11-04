package domain

import (
	"github.com/google/uuid"
	"time"
)

type Project struct {
	Id             uuid.UUID
	Name           string
	CompletionDate time.Time
	MinMembers     int32
	MaxMembers     int32
	//Manager        User
	//Members        []User
}

type ProjectRepository interface {
	Create(project Project) (Project, error)
}
