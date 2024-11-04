package domain

import (
	"github.com/google/uuid"
	"time"
)

t+

type ProjectRepository interface {
	Create(project Project) (Project, error)
}
