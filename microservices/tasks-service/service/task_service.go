package service

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"strings"
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
		Members:     make([]domain.User, 0),
	}
	log.Println(newTask)
	return s.repo.Create(*newTask)
}

func (s *TaskService) DeleteTask(id string) error {
	return s.repo.Delete(id)
}

func (s *TaskService) GetById(id string) (*proto.Task, error) {
	task, err := s.repo.GetById(id)
	if err != nil {
		return nil, status.Error(codes.Internal, "DB exception.")
	}
	var protoMembers []*proto.User
	for _, member := range task.Members {
		protoMembers = append(protoMembers, &proto.User{
			Id:       member.Id,
			Username: member.Username,
			Role:     member.Role,
		})
	}
	protoTask := &proto.Task{
		Id:          task.Id.Hex(),
		Name:        task.Name,
		Description: task.Description,
		Status:      task.Status.String(),
		ProjectId:   task.ProjectID,
		Members:     protoMembers,
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
		var protoMembers []*proto.User
		for _, member := range dp.Members {
			protoMembers = append(protoMembers, &proto.User{
				Id:       member.Id,
				Username: member.Username,
				Role:     member.Role,
			})
		}
		protoTasks = append(protoTasks, &proto.Task{
			Id:          dp.Id.Hex(),
			Name:        dp.Name,
			Description: dp.Description,
			Status:      dp.Status.String(),
			ProjectId:   dp.ProjectID,
			Members:     protoMembers,
		})
	}
	log.Println(protoTasks)
	return protoTasks, err
}

func (t *TaskService) AddMember(projectId string, protoUser *proto.User) error {
	task, err := t.repo.GetById(projectId)
	if err != nil {
		return status.Error(codes.NotFound, "Project not found")
	}

	user := &domain.User{
		Id:       protoUser.Id,
		Username: protoUser.Username,
		Role:     protoUser.Role,
	}
	log.Println("TASK SERVICE gettask repo->: ", task)
	for _, member := range task.Members {
		if strings.EqualFold(strings.TrimSpace(member.Username), strings.TrimSpace(user.Username)) {
			return status.Error(codes.AlreadyExists, "Member already part of the task")
		}
	}
	return t.repo.AddMember(projectId, *user)
}

func (t *TaskService) RemoveMember(projectId string, userId string) error {
	return t.repo.RemoveMember(projectId, userId)
}
func (s *TaskService) UpdateTask(taskReq *proto.Task) error {
	// Fetch the existing task
	existingTask, err := s.repo.GetById(taskReq.Id)
	if err != nil {
		return status.Error(codes.NotFound, "Task not found")
	}

	existingTask.Name = taskReq.Name
	existingTask.Description = taskReq.Description

	statusEnum, err := domain.ParseTaskStatus2(taskReq.Status)
	if err != nil {
		return status.Error(codes.InvalidArgument, "Invalid task status")
	}

	existingTask.Status = statusEnum

	// Update members if provided
	if len(taskReq.Members) > 0 {
		var updatedMembers []domain.User
		for _, member := range taskReq.Members {
			updatedMembers = append(updatedMembers, domain.User{
				Id:       member.Id,
				Username: member.Username,
				Role:     member.Role,
			})
		}
		existingTask.Members = updatedMembers
	}

	// Call the repository to persist the changes
	err = s.repo.Update(*existingTask)
	if err != nil {
		return status.Error(codes.Internal, "Failed to update task in the database")
	}

	return nil
}
