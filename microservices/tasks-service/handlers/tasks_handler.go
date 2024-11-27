package handlers

import (
	"context"
	"log"
	proto "tasks-service/proto/task"
	"tasks-service/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	//"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskHandler struct {
	service        *service.TaskService // Use a pointer here
	projectService proto.ProjectServiceClient
	proto.UnimplementedTaskServiceServer
}

func NewTaskHandler(service *service.TaskService, projectService proto.ProjectServiceClient) *TaskHandler {
	return &TaskHandler{service: service,
		projectService: projectService}
}

func (h *TaskHandler) Delete(ctx context.Context, req *proto.DeleteTaskReq) (*proto.EmptyResponse, error) {
	err := h.service.DeleteTask(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	return nil, nil
}

func (h *TaskHandler) Create(ctx context.Context, req *proto.CreateTaskReq) (*proto.EmptyResponse, error) {
	log.Println(req.Task)
	err := h.service.Create(req.Task)
	if err != nil {
		log.Printf("Error creating project: %v", err)
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	return nil, nil
}

func (h *TaskHandler) GetById(ctx context.Context, req *proto.GetByIdReq) (*proto.TaskResponse, error) {
	task, err := h.service.GetById(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	response := &proto.TaskResponse{Task: task}
	return response, nil
}

func (h *TaskHandler) GetAllByProjectId(ctx context.Context, req *proto.GetAllTasksReq) (*proto.GetAllTasksRes, error) {
	allTasks, err := h.service.GetTasksByProjectId(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch tasks")
	}
	response := &proto.GetAllTasksRes{Tasks: allTasks}
	return response, nil
}

func (h *TaskHandler) AddMemberTask(ctx context.Context, req *proto.AddMemberTaskReq) (*proto.EmptyResponse, error) {
	task, _ := h.service.GetById(req.TaskId)
	userOnProjectReq := &proto.UserOnOneProjectReq{
		UserId:    req.User.Username,
		ProjectId: task.ProjectId,
	}

	projServiceResponse, err := h.projectService.UserOnOneProject(ctx, userOnProjectReq)
	if err != nil {
		return nil, status.Error(codes.Internal, "Error checking project")
	}
	if projServiceResponse.IsOnProj {
		taskId := req.TaskId
		err = h.service.AddMember(taskId, req.User)
		if err != nil {
			log.Printf("Error adding member on project: %v", err)
			return nil, status.Error(codes.InvalidArgument, "Error adding member...")
		}
		return nil, nil

	} else {
		return nil, status.Error(codes.Internal, "User is not assigned to a project.")
	}

}

func (h *TaskHandler) RemoveMemberTask(ctx context.Context, req *proto.RemoveMemberTaskReq) (*proto.EmptyResponse, error) {
	taskId := req.TaskId
	err := h.service.RemoveMember(taskId, req.UserId)
	if err != nil {
		log.Printf("Error creating project: %v", err)
		return nil, status.Error(codes.InvalidArgument, "Error removing member...")
	}
	return nil, nil
}
func (h *TaskHandler) UpdateTask(ctx context.Context, req *proto.UpdateTaskReq) (*proto.EmptyResponse, error) {
	log.Println("Received UpdateTask request for task ID:", req.Id)

	// Validate the task exists
	existingTask, err := h.service.GetById(req.Id)
	if err != nil {
		log.Printf("Error fetching task for update: %v", err)
		return nil, status.Error(codes.NotFound, "Task not found")
	}

	// Update the fields of the task
	updatedTask := existingTask
	updatedTask.Name = req.Name
	updatedTask.Description = req.Description
	updatedTask.Status = req.Status
	updatedTask.Members = req.Members

	// Call the service layer to save changes
	err = h.service.UpdateTask(updatedTask)
	if err != nil {
		log.Printf("Error updating task: %v", err)
		return nil, status.Error(codes.Internal, "Failed to update task")
	}

	log.Println("Task updated successfully:", req.Id)
	return &proto.EmptyResponse{}, nil
}
