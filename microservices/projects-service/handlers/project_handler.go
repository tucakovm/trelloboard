package handlers

import (
	"context"
	"encoding/json"
	"log"
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
	proto.UnimplementedProjectServiceServer
	natsConn *nats.Conn
	Tracer   trace.Tracer
}

func NewConnectionHandler(service services.ProjectService, taskService proto.TaskServiceClient, natsConn *nats.Conn, Tracer trace.Tracer) (ProjectHandler, error) {
	return ProjectHandler{
		service:     service,
		taskService: taskService,
		natsConn:    natsConn,
		Tracer:      Tracer,
	}, nil
}

func (h ProjectHandler) Create(ctx context.Context, req *proto.CreateProjectReq) (*proto.EmptyResponse, error) {

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
		return nil, status.Error(codes.ResourceExhausted, "Max members must be more than Min member")
	}
	log.Println(req.Project.CompletionDate.AsTime())
	if time.Now().Truncate(24 * time.Hour).After(req.Project.CompletionDate.AsTime()) {
		span.SetStatus(otelCodes.Error, "Invalid completion date")
		return nil, status.Error(codes.ResourceExhausted, "Invalid completion date")
	}
	err := h.service.Create(req.Project, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Printf("Error creating project: %v", err)
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	return nil, nil
}

func (h ProjectHandler) GetAllProjects(ctx context.Context, req *proto.GetAllProjectsReq) (*proto.GetAllProjectsRes, error) {

	if h.Tracer == nil {
		log.Println("Tracer is nil")
		return nil, status.Error(codes.Internal, "tracer is not initialized")
	}
	ctx, span := h.Tracer.Start(ctx, "h.getAllProjects")
	defer span.End()
	allProducts, err := h.service.GetAllProjects(req.Username, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}

	response := &proto.GetAllProjectsRes{Projects: allProducts}
	log.Println(response)

	return response, nil
}

func (h ProjectHandler) Delete(ctx context.Context, req *proto.DeleteProjectReq) (*proto.EmptyResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.deleteProject")
	defer span.End()

	err := h.service.Delete(req.Id, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	return nil, nil
}

func (h ProjectHandler) GetById(ctx context.Context, req *proto.GetByIdReq) (*proto.GetByIdRes, error) {
	log.Printf("Received Project id request: %v", req.Id)
	ctx, span := h.Tracer.Start(ctx, "h.getProjectById")
	defer span.End()
	project, err := h.service.GetById(req.Id, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	response := &proto.GetByIdRes{Project: project}
	return response, nil
}

func (h ProjectHandler) AddMember(ctx context.Context, req *proto.AddMembersRequest) (*proto.EmptyResponse, error) {
	_, span := h.Tracer.Start(ctx, "Publisher.AddMember")
	defer span.End()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	subject := "add-to-project"

	project, _ := h.service.GetById(req.Id, ctx)
	numMembers := len(project.Members)
	if int32(numMembers) >= project.MaxMembers {
		log.Printf("Error adding member, project capacity full")
		return nil, status.Error(codes.ResourceExhausted, "Max members must be less than Min member")
	}
	if req.User.Role == "Manager" {
		log.Printf("Error adding member to project, cannot add a manager")
		return nil, status.Error(codes.Internal, "Error adding member to project, cannot add a manager to a project ")
	}
	err := h.service.AddMember(req.Id, req.User, ctx)
	if err != nil {
		log.Printf("Error adding member on project: %v", err)
		return nil, status.Error(codes.InvalidArgument, "Error adding member...")
	}

	message := map[string]string{
		"UserId":    req.User.Id,
		"ProjectId": req.Id,
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

func (h ProjectHandler) RemoveMember(ctx context.Context, req *proto.RemoveMembersRequest) (*proto.EmptyResponse, error) {
	_, span := h.Tracer.Start(ctx, "Publisher.AddMember")
	defer span.End()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	subject := "removed-from-project"
	log.Printf("Usao u handler od remove membera")
	projectId := req.ProjectId
	err := h.service.RemoveMember(projectId, req.UserId, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Printf("Error creating project: %v", err)
		return nil, status.Error(codes.InvalidArgument, "Error removing member...")
	}

	message := map[string]string{
		"UserId":    req.UserId,
		"ProjectId": req.ProjectId,
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

func (h ProjectHandler) UserOnProject(ctx context.Context, req *proto.UserOnProjectReq) (*proto.UserOnProjectRes, error) {
	time.Sleep(10 * time.Second) // circuit breaker test
	ctx, span := h.Tracer.Start(ctx, "h.userOnProjects")
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
