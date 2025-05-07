package handlers

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"tasks-service/domain"
	nats_helper "tasks-service/nats_helper"
	"time"

	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"log"
	"strings"
	proto "tasks-service/proto/task"

	grpcstatus "google.golang.org/grpc/status"
	"tasks-service/service"
	//"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskHandler struct {
	service         *service.TaskService // Use a pointer here
	projectService  proto.ProjectServiceClient
	workflowService proto.WorkflowServiceClient
	proto.UnimplementedTaskServiceServer
	natsConn *nats.Conn
	Tracer   trace.Tracer
	//workflowClient WorkflowServiceClient
}

func NewTaskHandler(service *service.TaskService, projectService proto.ProjectServiceClient, workflowService proto.WorkflowServiceClient, natsConn *nats.Conn, tracer trace.Tracer) *TaskHandler {
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
	task, _ := h.service.GetById(req.Id, ctx)
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

	err = h.service.DeleteFromCache(req.Id, task.ProjectId, ctx)
	if err != nil {
		log.Printf("error deleting from cache")
	}
	return nil, nil
}

func (h *TaskHandler) Create(ctx context.Context, req *proto.CreateTaskReq) (*proto.EmptyResponse, error) {
	_, span := h.Tracer.Start(ctx, "h.createTask")
	defer span.End()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	status, _ := domain.ParseTaskStatus2(req.Task.Status)

	task := domain.Task{
		Name:        req.Task.Name,
		Description: req.Task.Description,
		Status:      status,
		ProjectID:   req.Task.ProjectId,
	}

	event := domain.ProjectTaskCreateEvent{
		ProjectID:  req.Task.ProjectId,
		Task:       task,
		OccurredAt: time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return nil, grpcstatus.Error(codes.Internal, "internal error")
	}

	err = h.service.AppendEvent(ctx, req.Task.ProjectId, data, "CreateTask")
	if err != nil {
		log.Printf("Failed to append to event store: %v", err)
		return nil, grpcstatus.Error(codes.Internal, "event store error")
	}

	err = h.natsConn.Publish("task.create.es", data)
	if err != nil {
		log.Printf("Failed to publish to NATS: %v", err)
		return nil, grpcstatus.Error(codes.Internal, "nats error")
	}

	return &proto.EmptyResponse{}, nil
}

func (h *TaskHandler) CreateTask(ctx context.Context) error {
	_, err := h.natsConn.Subscribe("task.create.es", func(msg *nats.Msg) {

		var req domain.ProjectTaskCreateEvent
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			log.Printf("Invalid event format: %v", err)
			return
		}

		_, span := h.Tracer.Start(ctx, "h.createTask.esdb")
		defer span.End()

		headers := nats.Header{}
		headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
		headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

		taskID := primitive.NewObjectID()

		task := &proto.Task{
			Name:        req.Task.Name,
			Description: req.Task.Description,
			ProjectId:   req.ProjectID,
		}
		err := h.service.Create(taskID, task, ctx)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			log.Printf("Error creating task: %v", err)
			return
		}
		err = h.service.PostTaskCacheTTL(taskID, task, ctx)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			log.Printf("Error caching task: %v", err)
			return
		}

		subject := "create-task"

		message := map[string]string{
			"TaskName":  req.Task.Name,
			"ProjectId": req.Task.ProjectID,
		}

		messageData, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling notification message: %v", err)
			return
		}

		msgNot := &nats.Msg{
			Subject: subject,
			Header:  headers,
			Data:    messageData,
		}
		err = h.natsConn.PublishMsg(msgNot)
		if err != nil {
			log.Printf("Error publishing notification: %v", err)
			return
		}

		log.Printf("Notification sent: %s", string(messageData))
	})
	if err != nil {
		return err
	}
	return nil
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

	_, span := h.Tracer.Start(ctx, "h.AddMemberToTask")
	defer span.End()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	task, _ := h.service.GetById(req.TaskId, ctx)
	if task.Status == "Done" {
		log.Printf("cannot add members to a done task")
		return nil, status.Error(codes.NotFound, "cannot add members to a done task")
	}

	userToAdd := domain.User{
		Id:       req.User.Id,
		Username: req.User.Username,
		Role:     req.User.Role,
	}

	event := domain.ProjectTaskAddMemberEvent{
		TaskId:      req.TaskId,
		ProjectID:   task.ProjectId,
		TaskName:    task.Name,
		MemberToAdd: userToAdd,
		OccurredAt:  time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return nil, status.Error(codes.NotFound, "failed to marshal")
	}

	err = h.service.AppendEvent(ctx, task.ProjectId, data, "AddMemberToTask")
	if err != nil {
		log.Printf("Failed to append to event store: %v", err)
		return nil, status.Error(codes.NotFound, "failed to append to event store")
	}

	err = h.natsConn.Publish("task.addMember.es", data)
	if err != nil {
		log.Printf("Failed to publish to NATS: %v", err)
		return nil, status.Error(codes.NotFound, "failed to publish to NATS")
	}

	return &proto.EmptyResponse{}, nil

}

func (h *TaskHandler) AddMemberToTask(ctx context.Context) (*proto.EmptyResponse, error) {
	_, err := h.natsConn.Subscribe("task.addMember.es", func(msg *nats.Msg) {
		log.Printf("log add member to task !!1")
		var req domain.ProjectTaskAddMemberEvent
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			log.Printf("Invalid event format: %v", err)
			return
		}
		_, span := h.Tracer.Start(ctx, "h.AddMemberToTask-esdb")
		defer span.End()

		headers := nats.Header{}
		headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
		headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

		select {
		case <-ctx.Done():
			log.Printf("Handler detected context cancellation or timeout: %v", ctx.Err())
			return
		default:
		}

		task, _ := h.service.GetById(req.TaskId, ctx)
		if task.Status == "Done" {
			log.Printf("cannot add members to a done task")
			return
		}
		//
		//userOnProjectReq := &proto.UserOnOneProjectReq{
		//	UserId:    req.MemberToAdd.Username,
		//	ProjectId: task.ProjectId,
		//}

		// Provera timeout-a pre pozivanja udaljenog servisa
		select {
		case <-ctx.Done():
			log.Printf("Context timeout before calling project service: %v", ctx.Err())
			return
		default:

		}

		//checker := true
		//if h.projectService != nil {
		//	projServiceResponse, err := h.projectService.UserOnOneProject(ctx, userOnProjectReq)
		//	if err != nil {
		//
		//		log.Printf("Error checking project: %v", err)
		//
		//		span.SetStatus(otelCodes.Error, err.Error())
		//		return
		//	}
		//	checker = projServiceResponse.IsOnProj
		//}
		////projServiceResponse, err := h.projectService.UserOnOneProject(ctx, userOnProjectReq)
		////if err != nil {
		////
		////	log.Printf("Error checking project: %v", err)
		////
		////	span.SetStatus(otelCodes.Error, err.Error())
		////	return
		////}
		//
		//if checker {
		//	taskId := req.TaskId

		//time.Sleep(5 * time.Second) // test : Request timeout

		// Provera timeout-a pre dodavanja člana
		select {
		case <-ctx.Done():
			log.Printf("Context timeout before adding member: %v", ctx.Err())
			return
		default:
		}
		subject := "add-to-task"

		memberToAdd := &proto.User{
			Id:       req.MemberToAdd.Id,
			Username: req.MemberToAdd.Username,
			Role:     req.MemberToAdd.Role,
		}

		err := h.service.AddMember(req.TaskId, memberToAdd, ctx)

		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			log.Printf("Error adding member on project: %v", err)
			return
		}
		projectFromTaskReq := &proto.GetByIdReq{
			Id: req.TaskId,
		}
		projectFromTask, _ := h.GetById(ctx, projectFromTaskReq)
		message := map[string]string{
			"UserId":    req.MemberToAdd.Id,
			"TaskId":    req.TaskId,
			"ProjectId": projectFromTask.Task.ProjectId,
		}

		messageData, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling notification message: %v", err)
			return
		}

		msgNot := &nats.Msg{
			Subject: subject,
			Header:  headers,
			Data:    messageData,
		}
		err = h.natsConn.PublishMsg(msgNot)
		if err != nil {
			log.Printf("Error publishing notification: %v", err)
			return
		}

		log.Printf("Notification sent: %s", string(messageData))

		//} else {
		//	return
		//}
	})
	if err != nil {
		return nil, err
	}
	return &proto.EmptyResponse{}, nil
}
func (h *TaskHandler) RemoveMemberTask(ctx context.Context, req *proto.RemoveMemberTaskReq) (*proto.EmptyResponse, error) {
	_, span := h.Tracer.Start(ctx, "h.RemoveMemberFromTask")
	defer span.End()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	task, err := h.service.GetById(req.TaskId, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.NotFound, "project not found")
	}

	event := domain.ProjectTaskRemoveMemberEvent{
		ProjectID:     task.ProjectId,
		TaskName:      task.Name,
		TaskId:        task.Id,
		MemberToAddId: req.UserId,
		OccurredAt:    time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	err = h.service.AppendEvent(ctx, task.ProjectId, data, "RemoveMemberFromProject")
	if err != nil {
		log.Printf("Failed to append to event store: %v", err)
		return nil, status.Error(codes.Internal, "event store error")
	}

	err = h.natsConn.Publish("task.removeMember.es", data)
	if err != nil {
		log.Printf("Failed to publish to NATS: %v", err)
		return nil, status.Error(codes.Internal, "nats error")
	}

	return &proto.EmptyResponse{}, nil
}

func (h *TaskHandler) RemoveMemberFromTask(ctx context.Context) error {
	_, err := h.natsConn.Subscribe("task.removeMember.es", func(msg *nats.Msg) {
		log.Printf("addmember to proj esdb")
		var req domain.ProjectTaskRemoveMemberEvent
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			log.Printf("Invalid event format: %v", err)
			return
		}

		_, span := h.Tracer.Start(ctx, "h.RemoveMemberFromTask.esdb")
		defer span.End()

		headers := nats.Header{}
		headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
		headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

		taskId := req.TaskId
		task, err := h.service.GetById(taskId, ctx)
		if err != nil {
			log.Printf("Error getting task: %v", err)
			return
		}
		log.Println(task.Status)
		if task.Status == "Done" {
			log.Printf("cannot remove member from a done task")
			return
		}
		err = h.service.RemoveMember(taskId, req.MemberToAddId, ctx)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			log.Printf("Error creating project: %v", err)
			return
		}
		projectFromTaskReq := &proto.GetByIdReq{
			Id: taskId,
		}
		projectFromTask, _ := h.GetById(ctx, projectFromTaskReq)
		subject := "remove-from-task"
		message := map[string]string{
			"UserId":    req.MemberToAddId,
			"TaskId":    taskId,
			"ProjectId": projectFromTask.Task.ProjectId,
		}

		messageData, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling notification message: %v", err)
			return
		}

		msgNot := &nats.Msg{
			Subject: subject,
			Header:  headers,
			Data:    messageData,
		}
		err = h.natsConn.PublishMsg(msgNot)
		if err != nil {
			log.Printf("Error publishing notification: %v", err)
			return
		}

		log.Printf("Notification sent: %s", string(messageData))

		return
	})
	if err != nil {
		return err
	}
	return nil
}

func (h *TaskHandler) UpdateTask(ctx context.Context, req *proto.UpdateTaskReq) (*proto.EmptyResponse, error) {
	_, span := h.Tracer.Start(ctx, "h.UpdateTask")
	defer span.End()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	task, err := h.service.GetById(req.Id, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.NotFound, "project not found")
	}

	var mems []domain.User
	for _, m := range req.Members {
		mem := domain.User{
			Id:       m.Id,
			Username: m.Username,
			Role:     m.Role,
		}
		mems = append(mems, mem)
	}

	event := domain.ProjectTaskStatusEvent{
		ProjectID:     task.ProjectId,
		TaskId:        task.Id,
		TaskStatus:    task.Status,
		TaskNewStatus: req.Status,
		Members:       mems,
		OccurredAt:    time.Time{},
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	err = h.service.AppendEvent(ctx, task.ProjectId, data, "StatusUpdate")
	if err != nil {
		log.Printf("Failed to append to event store: %v", err)
		return nil, status.Error(codes.Internal, "event store error")
	}

	err = h.natsConn.Publish("task.update.es", data)
	if err != nil {
		log.Printf("Failed to publish to NATS: %v", err)
		return nil, status.Error(codes.Internal, "nats error")
	}

	return &proto.EmptyResponse{}, nil
}

func (h *TaskHandler) UpdateTaskStatus(ctx context.Context) error {
	_, err := h.natsConn.Subscribe("task.update.es", func(msg *nats.Msg) {

		var req domain.ProjectTaskStatusEvent
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			log.Printf("Invalid event format: %v", err)
			return
		}

		_, span := h.Tracer.Start(ctx, "h.UpdateTask.esdb")
		defer span.End()

		headers := nats.Header{}
		headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
		headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

		log.Println("Received UpdateTask request for task ID:", req.TaskId)

		// Validate the task exists
		existingTask, err := h.service.GetById(req.TaskId, ctx)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			log.Printf("Error fetching task for update: %v", err)
			return
		}

		workflowReq := &proto.IsTaskBlockedReq{TaskID: req.TaskId}
		workflowRes, err := h.workflowService.IsTaskBlocked(ctx, workflowReq)
		if err != nil {
			return
		}
		//log.Printf("response: %v", workflowRes.IsBlocked)

		if workflowRes.IsBlocked {
			return
		}

		updatedTask := existingTask
		updatedTask.Status = req.TaskNewStatus

		err = h.service.UpdateTask(updatedTask, ctx)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			log.Printf("Error updating task: %v", err)
			return
		}

		log.Println("Task updated successfully:", req.TaskId)

		//CQRS workflow----->

		subjectWorkflow := "cqrs-workflow-update"

		messageWorkflow := map[string]interface{}{
			"TaskId":     req.TaskId,
			"TaskStatus": req.TaskNewStatus,
		}

		messageDataWorkflow, err := json.Marshal(messageWorkflow)
		if err != nil {
			log.Printf("Error marshaling cqrs message: %v", err)
			return
		}

		msgWorkflow := &nats.Msg{
			Subject: subjectWorkflow,
			Data:    messageDataWorkflow,
			Header:  headers,
		}
		err = h.natsConn.PublishMsg(msgWorkflow)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			return
		}

		//notification----->

		getProjReq := &proto.GetByIdReq{
			Id: req.TaskId,
		}

		projId, _ := h.GetById(ctx, getProjReq)

		subject := "update-task"
		message := map[string]string{
			"TaskId":     req.TaskId,
			"TaskStatus": req.TaskNewStatus,
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
			return
		}

		msgNot := &nats.Msg{
			Subject: subject,
			Header:  headers,
			Data:    messageData,
		}
		err = h.natsConn.PublishMsg(msgNot)
		if err != nil {
			log.Printf("Error publishing notification: %v", err)
			return
		}

		log.Printf("Notification sent: %s", string(messageData))
		return
	})
	if err != nil {
		return err
	}
	return nil
}
func (h *TaskHandler) UploadFile(ctx context.Context, req *proto.UploadFileRequest) (*proto.EmptyResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.uploadFile")
	defer span.End()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	event := domain.TaskUpdateFileEvent{
		TaskId:      req.TaskId,
		UserId:      req.UserId,
		FileName:    req.FileName,
		FileContent: req.FileContent,
		OccurredAt:  time.Now().UTC(),
	}

	task, err := h.service.GetById(req.TaskId, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.NotFound, "project not found")
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	err = h.service.AppendEvent(ctx, task.ProjectId, data, "UploadFile")
	if err != nil {
		log.Printf("Failed to append to event store: %v", err)
		return nil, status.Error(codes.Internal, "event store error")
	}

	err = h.natsConn.Publish("project.taskUploadFile.es", data)
	if err != nil {
		log.Printf("Failed to publish to NATS: %v", err)
		return nil, status.Error(codes.Internal, "nats error")
	}

	return &proto.EmptyResponse{}, nil

}

func (h *TaskHandler) UploadFileESDB(ctx context.Context) error {
	_, err := h.natsConn.Subscribe("project.taskUploadFile.es", func(msg *nats.Msg) {

		var req domain.TaskUpdateFileEvent
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			log.Printf("Invalid event format: %v", err)
			return
		}

		log.Printf("Received task_id: %s, file_name: %s, file_content length: %d", req.TaskId, req.FileName, len(req.FileContent))

		ctx, span := h.Tracer.Start(ctx, "h.uploadFile.esdb")
		defer span.End()

		headers := nats.Header{}
		headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
		headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

		// Pass the task ID, file name, and raw file content to the service
		err := h.service.UploadFile(ctx, req.TaskId, req.FileName, req.FileContent)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			log.Printf("Failed to upload file")
			return
		}

		log.Printf("File uploaded : %s", req.FileName)

		return
	})
	if err != nil {
		return err
	}
	return nil
}

func (h *TaskHandler) DownloadFile(ctx context.Context, req *proto.DownloadFileRequest) (*proto.FileResponse, error) {
	log.Println("handler Download file")
	ctx, span := h.Tracer.Start(ctx, "h.downloadFile")
	defer span.End()

	fileContent, err := h.service.DownloadFile(ctx, req.TaskId, req.FileId)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "Failed to download file")
	}

	return &proto.FileResponse{
		FileContent: fileContent,
	}, nil
}

func (h *TaskHandler) DeleteFile(ctx context.Context, req *proto.DeleteFileRequest) (*proto.EmptyResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.deleteFile")
	defer span.End()

	err := h.service.DeleteFile(ctx, req.TaskId, req.FileName)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "Failed to delete file")
	}

	log.Printf("File deleted and notification sent: %s")

	return &proto.EmptyResponse{}, nil
}

func (h *TaskHandler) GetAllFiles(ctx context.Context, req *proto.GetTaskFilesRequest) (*proto.GetTaskFilesResponse, error) {
	log.Printf("Received request for task_id: %s", req.TaskId)

	ctx, span := h.Tracer.Start(ctx, "h.getAllFiles")
	defer span.End()

	// Call service layer to fetch the file names
	fileNames, err := h.service.GetAllFiles(ctx, req.TaskId)
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
