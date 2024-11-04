package repositories

import (
	"projects_module/domain"
)

type projectsInMemRepository struct {
	projects []domain.Project
}

func NewProjectInMem() (domain.ProjectRepository, error) {
	return &projectsInMemRepository{
		projects: make([]domain.Project, 0),
	}, nil
}

func (c projectsInMemRepository) Create(project domain.Project) (domain.Project, error) {
	c.projects = append(c.projects, project)
	return project, nil
}
