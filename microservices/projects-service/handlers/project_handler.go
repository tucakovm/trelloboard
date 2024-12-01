package handlers

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	proto "projects_module/proto/project"
	"projects_module/services"
)

type ProjectHandler struct {
	service     services.ProjectService
	taskService proto.TaskServiceClient
	proto.UnimplementedProjectServiceServer
	Tracer trace.Tracer
}

func NewConnectionHandler(service services.ProjectService, taskService proto.TaskServiceClient, tracer trace.Tracer) (ProjectHandler, error) {
	return ProjectHandler{
		service:     service,
		taskService: taskService,
		Tracer:      tracer,
	}, nil
}

func (h ProjectHandler) Create(ctx context.Context, req *proto.CreateProjectReq) (*proto.EmptyResponse, error) {
	log.Printf("Received Create Project request: %v", req.Project)
	ctx, span := h.Tracer.Start(ctx, "h.createProject")
	defer span.End()
	err := h.service.Create(req.Project, ctx)
	if err != nil {
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
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	response := &proto.GetByIdRes{Project: project}
	return response, nil
}

func (h ProjectHandler) AddMember(ctx context.Context, req *proto.AddMembersRequest) (*proto.EmptyResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.addMemberToProject")
	defer span.End()
	projId := req.Id
	reqForTaskClient := &proto.DoneTasksByProjectReq{
		ProjId: projId,
	}
	is, _ := h.taskService.DoneTasksByProject(ctx, reqForTaskClient)
	if is.IsDone {
		err := h.service.AddMember(projId, req.User, ctx)
		if err != nil {
			log.Printf("Error adding member on project: %v", err)
			return nil, status.Error(codes.InvalidArgument, "Error adding member...")
		}
		return nil, nil
	}
	return nil, status.Error(codes.Aborted, "Project has done.")
}
func (h ProjectHandler) RemoveMember(ctx context.Context, req *proto.RemoveMembersRequest) (*proto.EmptyResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.removeMemberFromProject")
	defer span.End()
	log.Printf("Usao u handler od remove membera")
	projectId := req.ProjectId
	err := h.service.RemoveMember(projectId, req.UserId, ctx)
	if err != nil {
		log.Printf("Error creating project: %v", err)
		return nil, status.Error(codes.InvalidArgument, "Error removing member...")
	}
	return nil, nil
}

func (h ProjectHandler) UserOnProject(ctx context.Context, req *proto.UserOnProjectReq) (*proto.UserOnProjectRes, error) {
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
		return nil, status.Error(codes.InvalidArgument, "DB exception.")
	}

	return &proto.UserOnProjectRes{OnProject: res}, err
}

func (h ProjectHandler) UserOnOneProject(ctx context.Context, req *proto.UserOnOneProjectReq) (*proto.UserOnProjectRes, error) {
	ctx, span := h.Tracer.Start(ctx, "h.userOnAProject")
	defer span.End()

	res, err := h.service.UserOnOneProject(req.ProjectId, req.UserId, ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "DB exception.")
	}
	log.Println("LOG BOOL :")
	log.Println(res)
	return &proto.UserOnProjectRes{OnProject: res}, err
}
