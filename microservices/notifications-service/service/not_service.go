package service

import (
	"context"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"not_module/domain"
	proto "not_module/proto/notification"
	"not_module/repositories"
)

type NotService struct {
	repo   repositories.NotRepo
	Tracer trace.Tracer
}

func NewNotService(repo repositories.NotRepo, tracer trace.Tracer) *NotService {
	return &NotService{
		repo:   repo,
		Tracer: tracer,
	}
}

func (s *NotService) Create(ctx context.Context, req *proto.Notification) error {
	ctx, span := s.Tracer.Start(ctx, "s.createNot")
	defer span.End()
	newNot := &domain.Notification{
		UserId:    req.UserId,
		CreatedAt: req.CreatedAt.AsTime(),
		Message:   req.Message,
		Status:    req.Status,
	}
	log.Println(newNot)
	return s.repo.InsertNotByUser(ctx, newNot)
}

func (s *NotService) GetAllNotUser(ctx context.Context, id string) ([]*proto.Notification, error) {
	ctx, span := s.Tracer.Start(ctx, "s.getAllNotsUser")
	defer span.End()
	nots, err := s.repo.GetNotsByUser(ctx, id)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "s:DB exception.")
	}
	var protoNots []*proto.Notification
	for _, not := range nots {
		protoNots = append(protoNots, &proto.Notification{
			UserId:    not.UserId,
			CreatedAt: timestamppb.New(not.CreatedAt),
			NotId:     not.NotificationId.String(),
			Message:   not.Message,
			Status:    not.Status,
		})
	}
	return protoNots, nil
}
