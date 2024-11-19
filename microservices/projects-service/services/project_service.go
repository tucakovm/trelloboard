package services

import (
	"errors"
	"log"
	"projects_module/domain"
	proto "projects_module/proto/project"
	"projects_module/repositories"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProjectService struct {
	repo repositories.ProjectRepo
}

func NewProjectService(repo repositories.ProjectRepo) (ProjectService, error) {
	return ProjectService{
		repo: repo,
	}, nil
}

func (s ProjectService) Create(req *proto.Project) error {
	if req == nil {
		log.Println("Received nil project.")
		return errors.New("received nil project")
	}

	log.Printf("SERVICE Received Create Project request: %v", req)
	return s.repo.Create(req)
}

func (s ProjectService) GetAllProjects(id string) ([]*proto.Project, error) {
	projects, err := s.repo.GetAllProjects(id)
	if err != nil {
		return nil, status.Error(codes.Internal, "DB exception.")
	}

	var protoProjects []*proto.Project
	for _, dp := range projects {
		protoProjects = append(protoProjects, &proto.Project{
			Id:             dp.Id.Hex(),
			Name:           dp.Name,
			CompletionDate: timestamppb.New(dp.CompletionDate),
			MinMembers:     int32(dp.MinMembers),
			MaxMembers:     int32(dp.MaxMembers),
			Manager: &proto.User{
				Id:       dp.Manager.Id,
				Username: dp.Manager.Username,
				Role:     dp.Manager.Role,
			},
		})
	}
	return protoProjects, nil
}

func (s ProjectService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s ProjectService) GetById(id string) (*proto.Project, error) {
	prj, err := s.repo.GetById(id)
	if err != nil {
		return nil, status.Error(codes.Internal, "DB exception.")
	}
	protoProject := &proto.Project{
		Id:             prj.Id.String(),
		Name:           prj.Name,
		CompletionDate: timestamppb.New(prj.CompletionDate),
		MinMembers:     int32(prj.MinMembers),
		MaxMembers:     int32(prj.MaxMembers),
		Manager: &proto.User{
			Id:       prj.Manager.Id,
			Username: prj.Manager.Username,
			Role:     prj.Manager.Role,
		},
	}

	return protoProject, nil
}

func (s ProjectService) AddMember(projectId string, user domain.User) error {
	return s.repo.AddMember(projectId, user)
}
