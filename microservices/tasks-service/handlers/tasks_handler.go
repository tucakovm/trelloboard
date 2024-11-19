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
	service *service.TaskService // Use a pointer here
	proto.UnimplementedTaskServiceServer
}

func NewTaskHandler(service *service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
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
