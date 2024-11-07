package service

import (
	"github.com/google/uuid"
	"log"
	"tasks-service/domain"
	"tasks-service/repository"
)

type TaskService struct {
	repo *repository.TaskRepo
}

func NewTaskService(repo *repository.TaskRepo) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) CreateTask(task domain.Task) (domain.Task, error) {
	log.Println("pokrenut create task u servisu")
	return s.repo.Create(task)
}

func (s *TaskService) GetAllTasks() ([]domain.Task, error) {
	return s.repo.GetAll()
}

func (s *TaskService) DeleteTask(id uuid.UUID) error {
	return s.repo.Delete(id)
}
