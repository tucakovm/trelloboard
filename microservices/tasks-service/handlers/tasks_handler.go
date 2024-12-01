package handlers

import (
	"context"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	Tracer trace.Tracer
}

func NewTaskHandler(service *service.TaskService, projectService proto.ProjectServiceClient, tracer trace.Tracer) *TaskHandler {
	return &TaskHandler{service: service,
		projectService: projectService,
		Tracer:         tracer}
}

func (h *TaskHandler) DoneTasksByProject(ctx context.Context, req *proto.DoneTasksByProjectReq) (*proto.DoneTasksByProjectRes, error) {
	ctx, span := h.Tracer.Start(ctx, "h.doneTaskByProject")
	defer span.End()

	is, err := h.service.DoneTasksByProject(req.ProjId, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument,
			"bad request ...")
	}
	doneTasksByProjectReq := &proto.DoneTasksByProjectRes{
		IsDone: is,
	}
	return doneTasksByProjectReq, nil
}

func (h *TaskHandler) Delete(ctx context.Context, req *proto.DeleteTaskReq) (*proto.EmptyResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.deleteTask")
	defer span.End()
	err := h.service.DeleteTask(req.Id, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil,
			status.Error(codes.InvalidArgument, "bad request ...")
	}
	return nil, nil
}

func (h *TaskHandler) Create(ctx context.Context, req *proto.CreateTaskReq) (*proto.EmptyResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.create")
	defer span.End()
	log.Println(req.Task)
	err := h.service.Create(req.Task, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Printf("Error creating project: %v", err)
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	return nil, nil
}

func (h *TaskHandler) GetById(ctx context.Context, req *proto.GetByIdReq) (*proto.TaskResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.getById")
	defer span.End()
	task, err := h.service.GetById(req.Id, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	response := &proto.TaskResponse{Task: task}
	return response, nil
}

func (h *TaskHandler) GetAllByProjectId(ctx context.Context, req *proto.GetAllTasksReq) (*proto.GetAllTasksRes, error) {
	ctx, span := h.Tracer.Start(ctx, "h.GetAllByProjectId")
	defer span.End()
	allTasks, err := h.service.GetTasksByProjectId(req.Id, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Errorf(codes.Internal, "Failed to fetch tasks")
	}
	response := &proto.GetAllTasksRes{Tasks: allTasks}
	return response, nil
}

func (h *TaskHandler) AddMemberTask(ctx context.Context, req *proto.AddMemberTaskReq) (*proto.EmptyResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.AddMemberTask")
	defer span.End()
	task, _ := h.service.GetById(req.TaskId, ctx)
	userOnProjectReq := &proto.UserOnOneProjectReq{
		UserId:    req.User.Username,
		ProjectId: task.ProjectId,
	}

	projServiceResponse, err := h.projectService.UserOnOneProject(ctx, userOnProjectReq)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "Error checking project")
	}
	if projServiceResponse.IsOnProj {
		taskId := req.TaskId
		err = h.service.AddMember(taskId, req.User, ctx)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			log.Printf("Error adding member on project: %v", err)
			return nil, status.Error(codes.InvalidArgument, "Error adding member...")
		}
		return nil, nil

	} else {
		return nil, status.Error(codes.Internal, "User is not assigned to a project.")
	}

}

func (h *TaskHandler) RemoveMemberTask(ctx context.Context, req *proto.RemoveMemberTaskReq) (*proto.EmptyResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.removeMemberTask")
	defer span.End()
	taskId := req.TaskId
	err := h.service.RemoveMember(taskId, req.UserId, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Printf("Error creating project: %v", err)
		return nil, status.Error(codes.InvalidArgument, "Error removing member...")
	}
	return nil, nil
}
func (h *TaskHandler) UpdateTask(ctx context.Context, req *proto.UpdateTaskReq) (*proto.EmptyResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.updateTask")
	defer span.End()
	log.Println("Received UpdateTask request for task ID:", req.Id)

	// Validate the task exists
	existingTask, err := h.service.GetById(req.Id, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
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
	err = h.service.UpdateTask(updatedTask, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Printf("Error updating task: %v", err)
		return nil, status.Error(codes.Internal, "Failed to update task")
	}

	log.Println("Task updated successfully:", req.Id)
	return &proto.EmptyResponse{}, nil
}
