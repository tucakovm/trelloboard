package services

import (
	"projects_module/domain"
	"projects_module/repositories"
)

type ProjectService struct {
	repo repositories.ProjectRepo
}

func NewProjectService(repo repositories.ProjectRepo) (ProjectService, error) {
	return ProjectService{
		repo: repo,
	}, nil
}

func (s ProjectService) Create(p *domain.Project) error {
	return s.repo.Create(p)
}

func (s ProjectService) GetAll(id string) (domain.Projects, error) {
	return s.repo.GetAll(id)
}

func (s ProjectService) Delete(id string) error {
	return s.repo.Delete(id)
}
