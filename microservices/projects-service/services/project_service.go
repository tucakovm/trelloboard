package services

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"projects_module/domain"
	proto "projects_module/proto/project"
	"projects_module/repositories"
	"strings"
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
	completionDate := req.CompletionDate.AsTime()

	prj := &domain.Project{
		Name:           req.Name,
		CompletionDate: completionDate.UTC(),
		MinMembers:     req.MinMembers,
		MaxMembers:     req.MaxMembers,
		Manager: domain.User{
			Id:       req.Manager.Id,
			Username: req.Manager.Username,
			Role:     req.Manager.Role,
		}, Members: make([]domain.User, 0),
	}

	log.Printf("SERVICE Received Create Project request: %v", req)
	return s.repo.Create(prj)
}

func (s ProjectService) UserOnProject(username string) (bool, error) {
	return s.repo.DoesManagerExistOnProject(username)
}

func (s ProjectService) GetAllProjects(id string) ([]*proto.Project, error) {
	projects, err := s.repo.GetAllProjects(id)
	if err != nil {
		return nil, status.Error(codes.Internal, "DB exception.")
	}

	var protoProjects []*proto.Project
	for _, dp := range projects {
		var protoMembers []*proto.User
		for _, member := range dp.Members {
			protoMembers = append(protoMembers, &proto.User{
				Id:       member.Id,
				Username: member.Username,
				Role:     member.Role,
			})
		}
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
			Members: protoMembers,
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
	var protoMembers []*proto.User
	for _, member := range prj.Members {
		protoMembers = append(protoMembers, &proto.User{
			Id:       member.Id,
			Username: member.Username,
			Role:     member.Role,
		})
	}

	protoProject := &proto.Project{
		Id:             prj.Id.Hex(),
		Name:           prj.Name,
		CompletionDate: timestamppb.New(prj.CompletionDate),
		MinMembers:     int32(prj.MinMembers),
		MaxMembers:     int32(prj.MaxMembers),
		Manager: &proto.User{
			Id:       prj.Manager.Id,
			Username: prj.Manager.Username,
			Role:     prj.Manager.Role,
		},
		Members: protoMembers,
	}

	return protoProject, nil
}

func (s ProjectService) AddMember(projectId string, protoUser *proto.User) error {
	project, err := s.GetById(projectId)
	if err != nil {
		return status.Error(codes.NotFound, "Project not found")
	}
	log.Println("len project members:", len(project.Members))
	log.Println("max members:", project.MaxMembers)
	log.Println("uslov membera:", len(project.Members) >= int(project.MaxMembers))
	if len(project.Members) >= int(project.MaxMembers) {
		return status.Error(codes.FailedPrecondition, "Maximum number of members reached")
	}
	user := &domain.User{
		Id:       protoUser.Id,
		Username: protoUser.Username,
		Role:     protoUser.Role,
	}
	for _, member := range project.Members {
		log.Println("member.username:", member.Username)
		log.Println("user.username:", user.Username)
		if strings.EqualFold(strings.TrimSpace(member.Username), strings.TrimSpace(user.Username)) {
			return status.Error(codes.AlreadyExists, "Member already part of the project")
		}
	}
	return s.repo.AddMember(projectId, *user)
}
