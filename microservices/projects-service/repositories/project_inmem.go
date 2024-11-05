package repositories

import (
	"projects_module/domain"
)

type projectsInMemRepository struct {
	projects map[string]domain.Project // Using a map with string keys (IDs)
}

func NewProjectInMem() (domain.ProjectRepository, error) {
	return &projectsInMemRepository{
		projects: make(map[string]domain.Project),
	}, nil
}

func (c projectsInMemRepository) Create(project domain.Project) (domain.Project, error) {
	mapId := project.Id.String()
	c.projects[mapId] = project
	return project, nil
}

func (c projectsInMemRepository) GetAll() ([]domain.Project, error) {
	var projects []domain.Project
	for _, project := range c.projects {
		projects = append(projects, project)
	}
	return projects, nil
}

func (c projectsInMemRepository) Delete(id string) error {
	delete(c.projects, id)
	return nil
}
