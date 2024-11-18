package handlers

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	proto "projects_module/proto/project"
	"projects_module/services"
)

type KeyProduct struct{}

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

	err := h.service.Create(req.Project) // Prosleđivanje samo req.Project
	if err != nil {
		log.Printf("Error creating project: %v", err)
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	return nil, nil
}

func (h ProjectHandler) GetAllProjects(ctx context.Context, req *proto.GetAllProjectsReq) (*proto.GetAllProjectsRes, error) {
	allProducts, err := h.service.GetAllProjects(req.Username)
	log.Println("project handler getAll")
	log.Println(allProducts)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}

	response := &proto.GetAllProjectsRes{Projects: allProducts}
	//data, err := p.Marshal(response)
	//if err != nil {
	//	log.Println("Error serializing response:", err)
	//} else {
	//	log.Println("Serialized response:", data)
	//	// Try unmarshaling back to verify the integrity of data
	//	var deserializedResponse proto.GetAllProjectsRes
	//	err := p.Unmarshal(data, &deserializedResponse)
	//	if err != nil {
	//		log.Println("Error unmarshaling response:", err)
	//	} else {
	//		log.Println("Deserialized response:", deserializedResponse)
	//	}
	//}

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
