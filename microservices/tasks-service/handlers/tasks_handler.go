package handlers

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"strings"
	nats_helper "tasks-service/nats_helper"
	proto "tasks-service/proto/task"

	"tasks-service/service"
	//"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskHandler struct {
	service         *service.TaskService // Use a pointer here
	projectService  proto.ProjectServiceClient
	workflowService proto.WokrflowServiceClient
	proto.UnimplementedTaskServiceServer
	natsConn *nats.Conn
	Tracer   trace.Tracer
	//workflowClient WorkflowServiceClient
}

func NewTaskHandler(service *service.TaskService, projectService proto.ProjectServiceClient, workflowService proto.WokrflowServiceClient, natsConn *nats.Conn, tracer trace.Tracer) *TaskHandler {
	return &TaskHandler{service: service,
		projectService:  projectService,
		natsConn:        natsConn,
		workflowService: workflowService,
		Tracer:          tracer}
}

//func (h *TaskHandler) DoneTasksByProject(ctx context.Context, req *proto.DoneTasksByProjectReq) (*proto.DoneTasksByProjectRes, error) {
//	is, err := h.service.DoneTasksByProject(req.ProjId)
//	if err != nil {
//		return nil, status.Error(codes.InvalidArgument, "bad request ...")
//	}
//	doneTasksByProjectReq := &proto.DoneTasksByProjectRes{
//		IsDone: is,
//	}
//	return doneTasksByProjectReq, nil
//}

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
	log.Printf("sending grpc to workflow TaskID: req.id=%s", req.Id)

	// Provera da li zadatak može biti obrisan
	exists, erre := h.workflowService.TaskExists(ctx, &proto.TaskExistsRequest{TaskId: req.Id})
	log.Printf("exists=%s", exists)

	if erre != nil {
		span.SetStatus(otelCodes.Error, erre.Error())
		log.Printf("Error in workflow Service for task = %s", erre)

		return nil, status.Error(codes.Internal, "failed to check task existence")
	}

	if exists.Exists {
		span.SetStatus(otelCodes.Error, "task is part of a workflow and cannot be deleted")
		log.Printf(" task is part of workflow = %s", erre)

		return nil, status.Error(codes.FailedPrecondition, "task is part of a workflow and cannot be deleted")
	}

	err := h.service.DeleteTask(req.Id, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil,
			status.Error(codes.InvalidArgument, "bad request ...")
	}
	return nil, nil
}

func (h *TaskHandler) Create(ctx context.Context, req *proto.CreateTaskReq) (*proto.EmptyResponse, error) {
	_, span := h.Tracer.Start(ctx, "Publisher.CreateTask")
	defer span.End()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	log.Println(req.Task)
	err := h.service.Create(req.Task, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Printf("Error creating project: %v", err)
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	subject := "create-task"

	message := map[string]string{
		"TaskName":  req.Task.Name,
		"ProjectId": req.Task.ProjectId,
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling notification message: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create notification message...")
	}

	msg := &nats.Msg{
		Subject: subject,
		Header:  headers,
		Data:    messageData,
	}
	err = h.natsConn.PublishMsg(msg)
	if err != nil {
		log.Printf("Error publishing notification: %v", err)
		return nil, status.Error(codes.Internal, "Failed to send notification...")
	}

	log.Printf("Notification sent: %s", string(messageData))

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

	_, span := h.Tracer.Start(ctx, "Publisher.AddMemberToTask")
	defer span.End()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	select {
	case <-ctx.Done():
		log.Printf("Handler detected context cancellation or timeout: %v", ctx.Err())
		return nil, status.Error(codes.DeadlineExceeded, "Request timed out or was canceled")
	default:
	}

	task, _ := h.service.GetById(req.TaskId, ctx)
	if task.Status == "Done" {
		log.Printf("cannot add members to a done task")
		return nil, status.Error(codes.FailedPrecondition, "cannot add members to a done task")
	}
	userOnProjectReq := &proto.UserOnOneProjectReq{
		UserId:    req.User.Username,
		ProjectId: task.ProjectId,
	}

	// Provera timeout-a pre pozivanja udaljenog servisa
	select {
	case <-ctx.Done():
		log.Printf("Context timeout before calling project service: %v", ctx.Err())
		return nil, status.Error(codes.DeadlineExceeded, "Request timed out or was canceled")
	default:

	}
	//time.Sleep(5 * time.Second) // test : Request timeout
	projServiceResponse, err := h.projectService.UserOnOneProject(ctx, userOnProjectReq)
	if err != nil {

		log.Printf("Error checking project: %v", err)

		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "Error checking project")
	}

	if projServiceResponse.IsOnProj {
		taskId := req.TaskId

		// Provera timeout-a pre dodavanja člana
		select {
		case <-ctx.Done():
			log.Printf("Context timeout before adding member: %v", ctx.Err())
			return nil, status.Error(codes.DeadlineExceeded, "Request timed out or was canceled")
		default:
			// Nastavlja sa dodavanjem člana
		}
		subject := "add-to-task"

		err = h.service.AddMember(taskId, req.User, ctx)

		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			log.Printf("Error adding member on project: %v", err)
			return nil, status.Error(codes.InvalidArgument, "Error adding member...")
		}
		projectFromTaskReq := &proto.GetByIdReq{
			Id: taskId,
		}
		projectFromTask, _ := h.GetById(ctx, projectFromTaskReq)
		message := map[string]string{
			"UserId":    req.User.Id,
			"TaskId":    taskId,
			"ProjectId": projectFromTask.Task.ProjectId,
		}

		messageData, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling notification message: %v", err)
			return nil, status.Error(codes.Internal, "Failed to create notification message...")
		}

		msg := &nats.Msg{
			Subject: subject,
			Header:  headers,
			Data:    messageData,
		}
		err = h.natsConn.PublishMsg(msg)
		if err != nil {
			log.Printf("Error publishing notification: %v", err)
			return nil, status.Error(codes.Internal, "Failed to send notification...")
		}

		log.Printf("Notification sent: %s", string(messageData))

		return nil, nil
	} else {
		return nil, status.Error(codes.Internal, "User is not assigned to a project.")
	}
}

func (h *TaskHandler) RemoveMemberTask(ctx context.Context, req *proto.RemoveMemberTaskReq) (*proto.EmptyResponse, error) {
	_, span := h.Tracer.Start(ctx, "Publisher.RemoveMemberFromTask")
	defer span.End()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	taskId := req.TaskId
	task, err := h.service.GetById(taskId, ctx)
	if err != nil {
		log.Printf("Error getting task: %v", err)
		return nil, status.Error(codes.Internal, "Error getting task")
	}
	log.Println(task.Status)
	if task.Status == "Done" {
		log.Printf("cannot remove member from a done task")
		return nil, status.Error(codes.FailedPrecondition, "cannot remove member from a done task")
	}
	err = h.service.RemoveMember(taskId, req.UserId, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Printf("Error creating project: %v", err)
		return nil, status.Error(codes.InvalidArgument, "Error removing member...")
	}
	projectFromTaskReq := &proto.GetByIdReq{
		Id: taskId,
	}
	projectFromTask, _ := h.GetById(ctx, projectFromTaskReq)
	subject := "remove-from-task"
	message := map[string]string{
		"UserId":    req.UserId,
		"TaskId":    taskId,
		"ProjectId": projectFromTask.Task.ProjectId,
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling notification message: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create notification message...")
	}

	msg := &nats.Msg{
		Subject: subject,
		Header:  headers,
		Data:    messageData,
	}
	err = h.natsConn.PublishMsg(msg)
	if err != nil {
		log.Printf("Error publishing notification: %v", err)
		return nil, status.Error(codes.Internal, "Failed to send notification...")
	}

	log.Printf("Notification sent: %s", string(messageData))

	return nil, nil
}
func (h *TaskHandler) UpdateTask(ctx context.Context, req *proto.UpdateTaskReq) (*proto.EmptyResponse, error) {
	_, span := h.Tracer.Start(ctx, "Publisher.UpdateTask")
	defer span.End()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	log.Println("Received UpdateTask request for task ID:", req.Id)

	// Validate the task exists
	existingTask, err := h.service.GetById(req.Id, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Printf("Error fetching task for update: %v", err)
		return nil, status.Error(codes.NotFound, "Task not found")
	}

	// Update the fields of the task
	//Commented out the actual update logic in order to leave only status update
	updatedTask := existingTask
	//pdatedTask.Name = req.Name
	//updatedTask.Description = req.Description
	updatedTask.Status = req.Status
	//updatedTask.Members = req.Members

	// Call the service layer to save changes
	err = h.service.UpdateTask(updatedTask, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Printf("Error updating task: %v", err)
		return nil, status.Error(codes.Internal, "Failed to update task")
	}

	log.Println("Task updated successfully:", req.Id)

	getProjReq := &proto.GetByIdReq{
		Id: req.Id,
	}

	projId, _ := h.GetById(ctx, getProjReq)

	subject := "update-task"
	message := map[string]string{
		"TaskId":     req.Id,
		"TaskStatus": req.Status,
		"ProjectId":  projId.Task.ProjectId,
	}
	if len(req.Members) > 0 {
		var memberIds []string
		for _, member := range req.Members {
			memberIds = append(memberIds, member.Id)
		}
		message["MemberIds"] = strings.Join(memberIds, ",")
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling notification message: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create notification message...")
	}

	msg := &nats.Msg{
		Subject: subject,
		Header:  headers,
		Data:    messageData,
	}
	err = h.natsConn.PublishMsg(msg)
	if err != nil {
		log.Printf("Error publishing notification: %v", err)
		return nil, status.Error(codes.Internal, "Failed to send notification...")
	}

	log.Printf("Notification sent: %s", string(messageData))
	return &proto.EmptyResponse{}, nil
}
