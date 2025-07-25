package handlers

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"projects_module/domain"
	nats_helper "projects_module/nats_helper"
	proto "projects_module/proto/project"
	"projects_module/services"
	"time"

	"github.com/nats-io/nats.go"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProjectHandler struct {
	service     services.ProjectService
	taskService proto.TaskServiceClient
	userService proto.UserServiceClient
	proto.UnimplementedProjectServiceServer
	natsConn *nats.Conn
	Tracer   trace.Tracer
}

func NewConnectionHandler(service services.ProjectService, taskService proto.TaskServiceClient, userService proto.UserServiceClient, natsConn *nats.Conn, Tracer trace.Tracer) (ProjectHandler, error) {
	return ProjectHandler{
		service:     service,
		taskService: taskService,
		natsConn:    natsConn,
		Tracer:      Tracer,
		userService: userService,
	}, nil
}

func (h ProjectHandler) Create(ctx context.Context, req *proto.CreateProjectReq) (*proto.EmptyResponse, error) {
	err := h.CreateProject(ctx)
	if err != nil {
		return nil, err
	}

	ctx, span := h.Tracer.Start(ctx, "h.Create")
	defer span.End()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	projectID := primitive.NewObjectID()

	event := domain.ProjectEvent{
		ProjectID:       projectID.Hex(),
		Name:            req.Project.Name,
		CompletionDate:  req.Project.CompletionDate.AsTime(),
		MinMembers:      req.Project.MinMembers,
		MaxMembers:      req.Project.MaxMembers,
		ManagerID:       req.Project.Manager.Id,
		ManagerUsername: req.Project.Manager.Username,
		ManagerRole:     req.Project.Manager.Role,
		OccurredAt:      time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	err = h.service.AppendEvent(ctx, projectID.Hex(), data, "ProjectCreated")
	if err != nil {
		log.Printf("Failed to append to event store: %v", err)
		return nil, status.Error(codes.Internal, "event store error")
	}

	err = h.natsConn.Publish("project.created.es", data)
	if err != nil {
		log.Printf("Failed to publish to NATS: %v", err)
		return nil, status.Error(codes.Internal, "nats error")
	}

	return &proto.EmptyResponse{}, nil
}

func (h ProjectHandler) CreateProject(ctx context.Context) error {
	_, err := h.natsConn.Subscribe("project.created.es", func(msg *nats.Msg) {
		log.Printf("create project esdb sub")
		var event domain.ProjectEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Invalid event format: %v", err)
			return
		}

		req := &proto.CreateProjectReq{
			Project: &proto.Project{
				Name:           event.Name,
				CompletionDate: timestamppb.New(event.CompletionDate),
				MinMembers:     event.MinMembers,
				MaxMembers:     event.MaxMembers,
				Manager: &proto.User{
					Id:       event.ManagerID,
					Username: event.ManagerUsername,
					Role:     event.ManagerRole,
				},
			},
		}

		// TESTIRANJE ZA GLOBALNI TIMEOUT
		//time.Sleep(6 * time.Second)
		//if ctx.Err() == context.DeadlineExceeded {
		//	return nil, status.Error(codes.DeadlineExceeded, "Project creation timed out")
		//}

		log.Printf("Received Create Project request: %v", req.Project)

		ctx, span := h.Tracer.Start(ctx, "h.createProject")
		defer span.End()

		if req.Project.MaxMembers <= req.Project.MinMembers {
			span.SetStatus(otelCodes.Error, "Max members must be more than Min member")
			return
		}
		if time.Now().Truncate(24 * time.Hour).After(req.Project.CompletionDate.AsTime()) {
			span.SetStatus(otelCodes.Error, "Invalid completion date")
			return
		}

		projectId := primitive.NewObjectID()
		err := h.service.Create(projectId, req.Project, ctx)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			log.Printf("Error creating project: %v", err)
			return
		}

		_, err = h.service.GetAllProjectsCache(req.Project.Manager.Username, ctx)
		if err == nil {
			err = h.service.PostProjectCacheTTL(projectId, req.Project, ctx)
			if err != nil {
				return
			}
		}

	})
	if err != nil {
		return err
	}
	return nil
}

func (h ProjectHandler) GetAllProjects(ctx context.Context, req *proto.GetAllProjectsReq) (*proto.GetAllProjectsRes, error) {

	if h.Tracer == nil {
		log.Println("Tracer is nil")
		return nil, status.Error(codes.Internal, "tracer is not initialized")
	}
	ctx, span := h.Tracer.Start(ctx, "h.getAllProjects")
	defer span.End()

	username := req.Username

	allProductsCache, err := h.service.GetAllProjectsCache(username, ctx)

	if err != nil {
		allProducts, err := h.service.GetAllProjects(username, ctx)

		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, status.Error(codes.InvalidArgument, "bad request ...")
		}

		err = h.service.PostAllProjectsCache(username, allProducts, ctx)
		if err != nil {
			return nil, status.Error(codes.Internal, "error caching projects")
		}

		response := &proto.GetAllProjectsRes{Projects: allProducts}

		log.Println(response)

		return response, nil
	} else {

		response := &proto.GetAllProjectsRes{Projects: allProductsCache}
		log.Println("response from cache:")
		log.Println(response)

		return response, nil
	}

}

func (h ProjectHandler) Delete(ctx context.Context, req *proto.DeleteProjectReq) (*proto.EmptyResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.deleteProject-esdb")
	defer span.End()

	err := h.DeleteProject(ctx)
	if err != nil {
		return nil, err
	}

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	project, err := h.service.GetById(req.Id, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.NotFound, "project not found")
	}

	event := domain.ProjectEvent{
		ProjectID:       project.Id,
		Name:            project.Name,
		CompletionDate:  project.CompletionDate.AsTime(),
		MinMembers:      project.MinMembers,
		MaxMembers:      project.MaxMembers,
		ManagerID:       project.Manager.Id,
		ManagerUsername: project.Manager.Username,
		ManagerRole:     project.Manager.Role,
		OccurredAt:      time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	err = h.service.AppendEvent(ctx, project.Id, data, "ProjectDeleted")
	if err != nil {
		log.Printf("Failed to append to event store: %v", err)
		return nil, status.Error(codes.Internal, "event store error")
	}

	err = h.natsConn.Publish("project.deleted.es", data)
	if err != nil {
		log.Printf("Failed to publish to NATS: %v", err)
		return nil, status.Error(codes.Internal, "nats error")
	}

	return &proto.EmptyResponse{}, nil
}

func (h ProjectHandler) DeleteProject(ctx context.Context) error {
	_, err := h.natsConn.Subscribe("project.deleted.es", func(msg *nats.Msg) {
		log.Printf("delete project esdb sub")
		var req domain.ProjectEvent
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			log.Printf("Invalid event format: %v", err)
			return
		}

		ctx, span := h.Tracer.Start(ctx, "h.deleteProject-saga")
		defer span.End()

		headers := nats.Header{}
		headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
		headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

		subjectTasks := "delete-tasks-saga"
		replySubjectTask := "delete-tasks-saga-reply"

		username, err := h.service.GetById(req.ProjectID, ctx)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			return
		}
		err = h.service.MarkAsDeleting(req.ProjectID, ctx)
		if err != nil {
			return
		}

		responseChan := make(chan *nats.Msg, 1)

		_, err = h.natsConn.ChanSubscribe(replySubjectTask, responseChan)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			return
		}

		if err != nil {
			log.Printf("[ERROR] Failed to marshal message: %v", err)
			return
		}

		log.Println("Publishing delete-tasks-saga message...")
		msgNot := &nats.Msg{
			Subject: subjectTasks,
			Reply:   replySubjectTask,
			Data:    []byte(req.ProjectID),
			Header:  headers,
		}
		err = h.natsConn.PublishMsg(msgNot)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			return
		}

		select {
		case msg := <-responseChan:
			log.Printf("msg projectServer data : %s\n", string(msg.Data))
			_, err = h.service.Delete(req.ProjectID, ctx)
			if err != nil {
				span.SetStatus(otelCodes.Error, err.Error())
				return
			}

			err = h.service.DeleteFromCache(req.ProjectID, username.Manager.Username, ctx)
			if err != nil {
				log.Printf("Error deleting from cache: %v", err)
			}

			return

		case <-time.After(5 * time.Second):
			log.Println("Timeout waiting for task deletion response, aborting project deletion")
			err = h.service.UnMarkAsDeleting(req.ProjectID, ctx)
			if err != nil {
				return
			}
			return
		}
	})
	if err != nil {
		return err
	}
	return nil
}

func (h ProjectHandler) GetById(ctx context.Context, req *proto.GetByIdReq) (*proto.GetByIdRes, error) {
	log.Printf("Received Project id request: %v", req.Id)
	ctx, span := h.Tracer.Start(ctx, "h.getProjectById")
	defer span.End()

	projectCache, err := h.service.GetByIdCache(req.Id, ctx)
	if err != nil {
		project, err := h.service.GetById(req.Id, ctx)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, status.Error(codes.InvalidArgument, "bad request ...")
		}
		err = h.service.PostProjectCache(project, ctx)
		if err != nil {
			return nil, status.Error(codes.Internal, "error caching project")
		}

		response := &proto.GetByIdRes{Project: project}
		return response, nil
	}
	log.Println("response from cache:")
	response := &proto.GetByIdRes{Project: projectCache}
	return response, nil
}

func (h ProjectHandler) AddMember(ctx context.Context, req *proto.AddMembersRequest) (*proto.EmptyResponse, error) {
	//err := h.AddMemberToProject(ctx)
	//if err != nil {
	//	return nil, err
	//}

	_, span := h.Tracer.Start(ctx, "h.AddMember")
	defer span.End()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	project, err := h.service.GetById(req.Id, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.NotFound, "project not found")
	}

	memberToAdd := domain.User{
		Id:       req.User.Id,
		Username: req.User.Username,
		Role:     req.User.Role,
	}

	event := domain.ProjectAddMemberEvent{
		ProjectID:       project.Id,
		Name:            project.Name,
		CompletionDate:  project.CompletionDate.AsTime(),
		MinMembers:      project.MinMembers,
		MaxMembers:      project.MaxMembers,
		ManagerID:       project.Manager.Id,
		ManagerUsername: project.Manager.Username,
		ManagerRole:     project.Manager.Role,
		MemberToAdd:     memberToAdd,
		OccurredAt:      time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	err = h.service.AppendEvent(ctx, project.Id, data, "AddMemberToProject")
	if err != nil {
		log.Printf("Failed to append to event store: %v", err)
		return nil, status.Error(codes.Internal, "event store error")
	}

	err = h.natsConn.Publish("project.addMember.es", data)
	if err != nil {
		log.Printf("Failed to publish to NATS: %v", err)
		return nil, status.Error(codes.Internal, "nats error")
	}

	return &proto.EmptyResponse{}, nil
}

func (h ProjectHandler) AddMemberToProject(ctx context.Context) error {
	_, err := h.natsConn.Subscribe("project.addMember.es", func(msg *nats.Msg) {
		log.Printf("addmember to proj esdb")
		var req domain.ProjectAddMemberEvent
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			log.Printf("Invalid event format: %v", err)
			return
		}

		_, span := h.Tracer.Start(ctx, "h.AddMember.esdb")
		defer span.End()

		headers := nats.Header{}
		headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
		headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

		subject := "add-to-project"

		project, _ := h.service.GetById(req.ProjectID, ctx)
		numMembers := len(project.Members)
		if int32(numMembers) >= project.MaxMembers {
			log.Printf("Error adding member, project capacity full")
			return
		}
		if req.MemberToAdd.Role == "Manager" {
			log.Printf("Error adding member to project, cannot add a manager")
			return
		}
		protoUser := &proto.User{
			Id:       req.MemberToAdd.Id,
			Username: req.MemberToAdd.Username,
			Role:     req.MemberToAdd.Role,
		}
		err := h.service.AddMember(req.ProjectID, protoUser, ctx)
		if err != nil {
			log.Printf("Error adding member on project: %v", err)
			return
		}

		message := map[string]string{
			"UserId":    req.MemberToAdd.Id,
			"ProjectId": req.ProjectID,
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

func (h ProjectHandler) RemoveMember(ctx context.Context, req *proto.RemoveMembersRequest) (*proto.EmptyResponse, error) {
	//err := h.RemoveMemberFromProject(ctx)
	//if err != nil {
	//	return nil, err
	//}

	_, span := h.Tracer.Start(ctx, "p.AddMember")
	defer span.End()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	project, err := h.service.GetById(req.ProjectId, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.NotFound, "project not found")
	}

	event := domain.ProjectRemoveMemberEvent{
		ProjectID:        project.Id,
		Name:             project.Name,
		CompletionDate:   project.CompletionDate.AsTime(),
		MinMembers:       project.MinMembers,
		MaxMembers:       project.MaxMembers,
		ManagerID:        project.Manager.Id,
		ManagerUsername:  project.Manager.Username,
		ManagerRole:      project.Manager.Role,
		MemberToRemoveId: req.UserId,
		OccurredAt:       time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	err = h.service.AppendEvent(ctx, project.Id, data, "RemoveMemberFromProject")
	if err != nil {
		log.Printf("Failed to append to event store: %v", err)
		return nil, status.Error(codes.Internal, "event store error")
	}

	err = h.natsConn.Publish("project.removeMember.es", data)
	if err != nil {
		log.Printf("Failed to publish to NATS: %v", err)
		return nil, status.Error(codes.Internal, "nats error")
	}

	return &proto.EmptyResponse{}, nil

}

func (h ProjectHandler) RemoveMemberFromProject(ctx context.Context) error {
	_, err := h.natsConn.Subscribe("project.removeMember.es", func(msg *nats.Msg) {

		var req domain.ProjectRemoveMemberEvent
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			log.Printf("Invalid event format: %v", err)
			return
		}

		_, span := h.Tracer.Start(ctx, "p.AddMember.esdb")
		defer span.End()

		headers := nats.Header{}
		headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
		headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

		subject := "removed-from-project"

		projectId := req.ProjectID
		err := h.service.RemoveMember(projectId, req.MemberToRemoveId, ctx)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			log.Printf("Error RemoveMember project: %v", err)
			return
		}

		message := map[string]string{
			"UserId":    req.MemberToRemoveId,
			"ProjectId": req.ProjectID,
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

func (h ProjectHandler) UserOnProject(ctx context.Context, req *proto.UserOnProjectReq) (*proto.UserOnProjectRes, error) {

	count := req.Count
	log.Printf("UserOnProject called with count: %d", count)

	//api gateway test
	//time.Sleep(10 * time.Second)

	retryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ctx, span := h.Tracer.Start(retryCtx, "h.userOnProjects")
	defer span.End()

	if req.Role == "Manager" {
		res, err := h.service.UserOnProject(req.Username, ctx)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "DB exception.")
		}

		return &proto.UserOnProjectRes{OnProject: res}, err
	}

	res, err := h.service.UserOnProjectUser(req.Username, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument, "DB exception.")
	}

	return &proto.UserOnProjectRes{OnProject: res}, err
}

func (h ProjectHandler) UserOnOneProject(ctx context.Context, req *proto.UserOnOneProjectReq) (*proto.UserOnProjectRes, error) {
	ctx, span := h.Tracer.Start(ctx, "h.userOnAProject")
	defer span.End()

	res, err := h.service.UserOnOneProject(req.ProjectId, req.UserId, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument, "DB exception.")
	}
	log.Println("LOG BOOL :")
	log.Println(res)
	return &proto.UserOnProjectRes{OnProject: res}, err
}
