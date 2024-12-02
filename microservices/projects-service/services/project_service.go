package services

import (
	"context"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	repo   repositories.ProjectRepo
	Tracer trace.Tracer
}

func NewProjectService(repo repositories.ProjectRepo, tracer trace.Tracer) (ProjectService, error) {
	return ProjectService{
		repo:   repo,
		Tracer: tracer,
	}, nil
}

func (s ProjectService) Create(req *proto.Project, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "s.createProject")
	defer span.End()
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
	return s.repo.Create(prj, ctx)
}

func (s ProjectService) UserOnProject(username string, ctx context.Context) (bool, error) {
	ctx, span := s.Tracer.Start(ctx, "s.userOnProject")
	defer span.End()
	return s.repo.DoesManagerExistOnProject(username, ctx)
}

func (s ProjectService) UserOnProjectUser(username string, ctx context.Context) (bool, error) {
	ctx, span := s.Tracer.Start(ctx, "s.userOnProject")
	defer span.End()
	return s.repo.DoesUserExistOnProject(username, ctx)
}

func (s ProjectService) UserOnOneProject(prjId string, userId string, ctx context.Context) (bool, error) {
	ctx, span := s.Tracer.Start(ctx, "s.userOnAProject")
	defer span.End()
	return s.repo.DoesMemberExistOnProject(prjId, userId, ctx)
}

func (s ProjectService) GetAllProjects(id string, ctx context.Context) ([]*proto.Project, error) {
	if s.Tracer == nil {
		log.Println("Tracer is nil")
		return nil, status.Error(codes.Internal, "tracer is not initialized")
	}
	ctx, span := s.Tracer.Start(ctx, "s.getAllProjects")
	defer span.End()
	projects, err := s.repo.GetAllProjects(id, ctx)
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

func (s ProjectService) Delete(id string, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "s.Delete")
	defer span.End()
	return s.repo.Delete(id, ctx)
}

func (s ProjectService) GetById(id string, ctx context.Context) (*proto.Project, error) {
	ctx, span := s.Tracer.Start(ctx, "s.getById")
	defer span.End()
	prj, err := s.repo.GetById(id, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
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

func (s ProjectService) AddMember(projectId string, protoUser *proto.User, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "s.addMemberToProject")
	defer span.End()
	project, err := s.GetById(projectId, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
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
	return s.repo.AddMember(projectId, *user, ctx)
}

func (s ProjectService) RemoveMember(projectId string, userId string, ctx context.Context) error {
	return s.repo.RemoveMember(projectId, userId, ctx)
}
