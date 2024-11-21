package service

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"tasks-service/domain"
	proto "tasks-service/proto/task"
	"tasks-service/repository"
)

type TaskService struct {
	repo repository.TaskRepo
}

func NewTaskService(repo repository.TaskRepo) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) Create(taskReq *proto.Task) error {
	newTask := &domain.Task{
		Name:        taskReq.Name,
		Description: taskReq.Description,
		Status:      0,
		ProjectID:   taskReq.ProjectId,
	}
	log.Println(newTask)
	return s.repo.Create(*newTask)
}

func (s *TaskService) GetAllTasks() ([]domain.Task, error) {
	return s.repo.GetAll()
}

func (s *TaskService) DeleteTask(id string) error {
	return s.repo.Delete(id)
}

func (s *TaskService) GetById(id string) (*proto.Task, error) {
	task, err := s.repo.GetById(id)
	if err != nil {
		return nil, status.Error(codes.Internal, "DB exception.")
	}
	protoTask := &proto.Task{
		Id:          task.Id.Hex(),
		Name:        task.Name,
		Description: task.Description,
		Status:      task.Status.String(),
		ProjectId:   task.ProjectID,
	}

	return protoTask, nil
}

func (s *TaskService) DeleteTasksByProjectId(id string) error {
	return s.repo.DeleteAllByProjectID(id)
}

func (s *TaskService) GetTasksByProjectId(id string) ([]*proto.Task, error) {
	tasks, err := s.repo.GetAllByProjectID(id)
	if err != nil {
		return nil, status.Error(codes.Internal, "DB exception.")
	}
	log.Println("SErvice tasks")
	log.Println(tasks)
	var protoTasks []*proto.Task
	for _, dp := range tasks {
		protoTasks = append(protoTasks, &proto.Task{
			Id:          dp.Id.Hex(),
			Name:        dp.Name,
			Description: dp.Description,
			Status:      dp.Status.String(),
			ProjectId:   dp.ProjectID,
		})
	}
	log.Println(protoTasks)
	return protoTasks, err
}
