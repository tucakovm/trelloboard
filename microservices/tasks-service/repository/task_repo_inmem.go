package repository

import (
	"github.com/google/uuid"
	"tasks-service/domain"
)

type tasksInMemRepository struct {
	tasks []tasksDAO
}

func (t tasksInMemRepository) Create(task domain.Task) (domain.Task, error) {
	//TODO implement me
	panic("implement me")
}

func (t tasksInMemRepository) CreateInMem(task domain.Task) (domain.Task, error) {
	if task.Status != domain.Pending {
		task.Status = domain.Pending
	}
	t.tasks = append(t.tasks, tasksDAO{
		//Id:          task.Id,
		Name:        task.Name,
		Description: task.Description,
		Status:      task.Status,
	})
	return task, nil

}

type tasksDAO struct {
	Id          uuid.UUID
	Name        string
	Description string
	Status      domain.Status
}

func NewTaskInMem() (domain.TasksRepository, error) {
	return &tasksInMemRepository{
		tasks: make([]tasksDAO, 0),
	}, nil
}
