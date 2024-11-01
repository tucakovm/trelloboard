package services

import "projects_module/domain"

type ProjectService struct {
	conns domain.ProjectRepository
}

func NewConnectionService(conns domain.ProjectRepository) (ProjectService, error) {
	return ProjectService{
		conns: conns,
	}, nil
}
