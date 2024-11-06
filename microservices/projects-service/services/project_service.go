package services

import (
	"github.com/google/uuid"
	"projects_module/domain"
	"projects_module/repositories"
)

type ProjectService struct {
	repo repositories.ProjectRepo
}

func NewConnectionService(repo repositories.ProjectRepo) (ProjectService, error) {
	return ProjectService{
		repo: repo,
	}, nil
}

func (s ProjectService) Create(p *domain.Project) error {
	p.Id = uuid.New()
	s.repo.Create(p)

	return nil
}

func (s ProjectService) GetAll() (domain.Projects, error) {
	return s.repo.GetAll()
}

func (s ProjectService) Delete(id string) error {
	return s.repo.Delete(id)
}
