package repositories

import (
	"github.com/google/uuid"
	"projects_module/domain"
	"time"
)

type projectsInMemRepository struct {
	projects []projectsDAO
}

// trenutno dok nemamo bazu , cuvamo u memoriji.
type projectsDAO struct {
	Id             uuid.UUID
	Name           string
	CompletionDate time.Time
	MinMembers     int32
	MaxMembers     int32
}

func NewConnectionInMem() (domain.ProjectRepository, error) {
	return &projectsInMemRepository{
		projects: make([]projectsDAO, 0),
	}, nil
}

func (c projectsInMemRepository) Create(project domain.Project) (domain.Project, error) {
	//TODO implement me
	panic("implement me")
}
