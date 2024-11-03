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

func NewProjectInMem() (domain.ProjectRepository, error) {
	return &projectsInMemRepository{
		projects: make([]projectsDAO, 0),
	}, nil
}

func (c projectsInMemRepository) Create(project domain.Project) (domain.Project, error) {
	c.projects = append(c.projects, projectsDAO{
		Id:             project.Id,
		Name:           project.Name,
		CompletionDate: project.CompletionDate,
		MinMembers:     project.MinMembers,
		MaxMembers:     project.MaxMembers,
	})
	return project, nil
}
