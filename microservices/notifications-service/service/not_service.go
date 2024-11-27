package service

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"not_module/domain"
	proto "not_module/proto/notification"
	"not_module/repositories"
)

type NotService struct {
	repo repositories.NotRepo
}

func NewNotService(repo repositories.NotRepo) *NotService {
	return &NotService{repo: repo}
}

func (s *NotService) Create(req *proto.Notification) error {
	newNot := &domain.Notification{
		UserId:    req.UserId,
		CreatedAt: req.CreatedAt.AsTime(),
		Message:   req.Message,
		Status:    req.Status,
	}
	log.Println(newNot)
	return s.repo.InsertNotByUser(newNot)
}

func (s *NotService) GetAllNotUser(id string) ([]*proto.Notification, error) {
	nots, err := s.repo.GetNotsByUser(id)
	if err != nil {
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
