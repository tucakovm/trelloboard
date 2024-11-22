package handlers

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	proto "projects_module/proto/project"
	"projects_module/services"
)

type ProjectHandler struct {
	service services.ProjectService
	proto.UnimplementedProjectServiceServer
}

func NewConnectionHandler(service services.ProjectService) (ProjectHandler, error) {
	return ProjectHandler{
		service: service,
	}, nil
}

func (h ProjectHandler) Create(ctx context.Context, req *proto.CreateProjectReq) (*proto.EmptyResponse, error) {
	log.Printf("Received Create Project request: %v", req.Project)

	err := h.service.Create(req.Project)
	if err != nil {
		log.Printf("Error creating project: %v", err)
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	return nil, nil
}

func (h ProjectHandler) GetAllProjects(ctx context.Context, req *proto.GetAllProjectsReq) (*proto.GetAllProjectsRes, error) {
	allProducts, err := h.service.GetAllProjects(req.Username)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}

	response := &proto.GetAllProjectsRes{Projects: allProducts}
	log.Println(response)

	return response, nil
}

func (h ProjectHandler) Delete(ctx context.Context, req *proto.DeleteProjectReq) (*proto.EmptyResponse, error) {
	err := h.service.Delete(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	return nil, nil
}

func (h ProjectHandler) GetById(ctx context.Context, req *proto.GetByIdReq) (*proto.GetByIdRes, error) {
	log.Printf("Received Project id request: %v", req.Id)
	project, err := h.service.GetById(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	response := &proto.GetByIdRes{Project: project}
	return response, nil
}

func (h ProjectHandler) AddMember(ctx context.Context, req *proto.AddMembersRequest) (*proto.EmptyResponse, error) {
	log.Printf("PROTOUSER HANDLER: %+v\n", req.User)
	projectId := req.Id

	err := h.service.AddMember(projectId, req.User)
	if err != nil {
		log.Printf("Error creating project: %v", err)
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	return nil, nil
}

func (h ProjectHandler) UserOnProject(ctx context.Context, req *proto.UserOnProjectReq) (*proto.UserOnProjectRes, error) {
	if req.Role == "Manager" {
		res, err := h.service.UserOnProject(req.Username)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "DB exception.")
		}

		return &proto.UserOnProjectRes{OnProject: res}, err
	}
	res, err := h.service.UserOnProjectUser(req.Username)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "DB exception.")
	}

	return &proto.UserOnProjectRes{OnProject: res}, err
}
