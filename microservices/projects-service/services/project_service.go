package services

import (
	"github.com/google/uuid"
	"projects_module/domain"
)

type ProjectService struct {
	repo domain.ProjectRepository
}

func NewConnectionService(repo domain.ProjectRepository) (ProjectService, error) {
	return ProjectService{
		repo: repo,
	}, nil
}

func (s ProjectService) Create(p domain.Project) (domain.Project, error) {
	p.Id = uuid.New()
	project, err := s.repo.Create(p)
	if err != nil {
		return domain.Project{}, err
	}

	return project, nil
}

func (s ProjectService) GetAll() ([]domain.Project, error) {
	return s.repo.GetAll()
}

func (s ProjectService) Delete(id string) error {
	return s.repo.Delete(id)
}
