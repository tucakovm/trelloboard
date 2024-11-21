package handlers

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net/http"
	"projects_module/domain"
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

func (h ProjectHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Get the projectId from URL parameter
	vars := mux.Vars(r)
	projectId := vars["id"]
	if projectId == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	// Parse user from request body
	var user domain.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid user data", http.StatusBadRequest)
		return
	}

	// Call the service to add the member to the project
	err = h.service.AddMember(projectId, user)
	if err != nil {
		http.Error(w, "Error adding member to project", http.StatusInternalServerError)
		return
	}

	// Respond with success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Member added successfully"}`))
}