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

func (s ProjectService) GetById(id string) (*domain.Project, error) {
	return s.repo.GetById(id)
}

func (s ProjectService) AddMember(projectId string, user domain.User) error {
	return s.repo.AddMember(projectId, user)
}
