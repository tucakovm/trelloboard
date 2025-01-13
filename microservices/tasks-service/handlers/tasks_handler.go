package handlers

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	nats_helper "tasks-service/nats_helper"

	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"log"
	"strings"
	proto "tasks-service/proto/task"
	"tasks-service/service"
	//"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskHandler struct {
	service        *service.TaskService // Use a pointer here
	projectService proto.ProjectServiceClient
	proto.UnimplementedTaskServiceServer
	natsConn *nats.Conn
	Tracer   trace.Tracer
}

func NewTaskHandler(service *service.TaskService, projectService proto.ProjectServiceClient, natsConn *nats.Conn, tracer trace.Tracer) *TaskHandler {
	return &TaskHandler{service: service,
		projectService: projectService,
		natsConn:       natsConn,
		Tracer:         tracer}
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
	task, _ := h.service.GetById(req.Id, ctx)
	err := h.service.DeleteTask(req.Id, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil,
			status.Error(codes.InvalidArgument, "bad request ...")
	}

	err = h.service.DeleteFromCache(req.Id, task.ProjectId, ctx)
	if err != nil {
		log.Printf("error deleting from cache")
	}
	return nil, nil
}

func (h *TaskHandler) Create(ctx context.Context, req *proto.CreateTaskReq) (*proto.EmptyResponse, error) {
	_, span := h.Tracer.Start(ctx, "Publisher.CreateTask")
	defer span.End()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	taskID := primitive.NewObjectID()

	log.Println(req.Task)
	err := h.service.Create(taskID, req.Task, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Printf("Error creating task: %v", err)
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	err = h.service.PostTaskCacheTTL(taskID, req.Task, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Printf("Error caching task: %v", err)
		return nil, status.Error(codes.InvalidArgument, "error caching task")
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
	task, err := h.service.GetByIdCache(req.Id, ctx)
	if err != nil {
		task, err := h.service.GetById(req.Id, ctx)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, status.Error(codes.InvalidArgument, "bad request ...")
		}
		err = h.service.PostTaskCache(task, ctx)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "post task cache error")
		}
		response := &proto.TaskResponse{Task: task}
		return response, nil
	}
	log.Println("response from cache:")
	response := &proto.TaskResponse{Task: task}
	return response, nil
}

func (h *TaskHandler) GetAllByProjectId(ctx context.Context, req *proto.GetAllTasksReq) (*proto.GetAllTasksRes, error) {
	ctx, span := h.Tracer.Start(ctx, "h.GetAllByProjectId")
	defer span.End()

	allTasks, err := h.service.GetAllTasksCache(req.Id, ctx)
	if err != nil {
		allTasks, err := h.service.GetTasksByProjectId(req.Id, ctx)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, status.Errorf(codes.Internal, "Failed to fetch tasks")
		}
		err = h.service.PostAllTasksCache(req.Id, allTasks, ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to fetch tasks from cache")
		}
		response := &proto.GetAllTasksRes{Tasks: allTasks}
		return response, nil
	}
	log.Println("response from cache:")
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

	projServiceResponse, err := h.projectService.UserOnOneProject(ctx, userOnProjectReq)
	if err != nil {

		log.Printf("Error checking project: %v", err)

		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "Error checking project")
	}

	if projServiceResponse.IsOnProj {
		taskId := req.TaskId

		//time.Sleep(5 * time.Second) // test : Request timeout

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

	//notification----->

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

func (h *TaskHandler) UploadFile(ctx context.Context, req *proto.UploadFileRequest) (*proto.EmptyResponse, error) {
	log.Printf("Received task_id: %s, file_name: %s, file_content length: %d", req.TaskId, req.FileName, len(req.FileContent))

	_, span := h.Tracer.Start(ctx, "Publisher.UploadFile")
	defer span.End()

	log.Printf("handler upload file")
	log.Println("task name", req.FileName)
	log.Println("end on file name")

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	// Pass the task ID, file name, and raw file content to the service
	err := h.service.UploadFile(req.TaskId, string(req.FileContent), []byte(req.FileName))
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "Failed to upload file")
	}

	log.Printf("File uploaded : %s")

	return &proto.EmptyResponse{}, nil
}

func (h *TaskHandler) DownloadFile(ctx context.Context, req *proto.DownloadFileRequest) (*proto.DownloadFileResponse, error) {
	_, span := h.Tracer.Start(ctx, "Publisher.DownloadFile")
	defer span.End()

	fileContent, err := h.service.DownloadFile(req.TaskId, req.FileName)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "Failed to download file")
	}

	return &proto.DownloadFileResponse{
		FileContent: fileContent,
	}, nil
}

func (h *TaskHandler) DeleteFile(ctx context.Context, req *proto.DeleteFileRequest) (*proto.EmptyResponse, error) {
	_, span := h.Tracer.Start(ctx, "Publisher.DeleteFile")
	defer span.End()

	err := h.service.DeleteFile(req.TaskId, req.FileName)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "Failed to delete file")
	}

	log.Printf("File deleted and notification sent: %s")

	return &proto.EmptyResponse{}, nil
}

func (h *TaskHandler) GetTaskFiles(ctx context.Context, req *proto.GetTaskFilesRequest) (*proto.GetTaskFilesResponse, error) {
	log.Printf("Received request for task_id: %s", req.TaskId)

	_, span := h.Tracer.Start(ctx, "Handler.GetTaskFiles")
	defer span.End()

	// Call service layer to fetch the file names
	fileNames, err := h.service.GetTaskFiles(req.TaskId)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "Failed to fetch files")
	}

	log.Printf("Successfully fetched file names for task_id: %s", req.TaskId)

	// Return the response with the file names
	return &proto.GetTaskFilesResponse{
		FileNames: fileNames,
	}, nil
}
