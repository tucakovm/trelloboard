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
	service     services.ProjectService
	taskService proto.TaskServiceClient
	proto.UnimplementedProjectServiceServer
}

func NewConnectionHandler(service services.ProjectService, taskService proto.TaskServiceClient) (ProjectHandler, error) {
	return ProjectHandler{
		service:     service,
		taskService: taskService,
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
	projId := req.Id
	reqForTaskClient := &proto.DoneTasksByProjectReq{
		ProjId: projId,
	}
	is, _ := h.taskService.DoneTasksByProject(ctx, reqForTaskClient)
	if is.IsDone {
		err := h.service.AddMember(projId, req.User)
		if err != nil {
			log.Printf("Error adding member on project: %v", err)
			return nil, status.Error(codes.InvalidArgument, "Error adding member...")
		}
		return nil, nil
	}
	return nil, status.Error(codes.Aborted, "Project has done.")
}
func (h ProjectHandler) RemoveMember(ctx context.Context, req *proto.RemoveMembersRequest) (*proto.EmptyResponse, error) {
	log.Printf("Usao u handler od remove membera")
	projectId := req.ProjectId
	err := h.service.RemoveMember(projectId, req.UserId)
	if err != nil {
		log.Printf("Error creating project: %v", err)
		return nil, status.Error(codes.InvalidArgument, "Error removing member...")
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

func (h ProjectHandler) UserOnOneProject(ctx context.Context, req *proto.UserOnOneProjectReq) (*proto.UserOnProjectRes, error) {

	res, err := h.service.UserOnOneProject(req.ProjectId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "DB exception.")
	}
	log.Println("LOG BOOL :")
	log.Println(res)
	return &proto.UserOnProjectRes{OnProject: res}, err
}
