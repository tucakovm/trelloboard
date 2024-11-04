package service

import (
	"github.com/google/uuid"
	"tasks-service/domain"
)

type TaskService struct {
	repo domain.TasksRepository
}

func NewTaskService(repo domain.TasksRepository) (TaskService, error) {
	return TaskService{
		repo: repo,
	}, nil
}

func (s TaskService) Create(t domain.Task) (domain.Task, error) {
	t.Id = uuid.New()
	task, err := s.repo.Create(t)
	if err != nil {
		return domain.Task{}, err
	}

	return task, nil
}
