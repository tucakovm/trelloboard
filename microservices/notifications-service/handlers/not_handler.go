package handlers

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	proto "not_module/proto/notification"
	"not_module/service"
)

type NotificationHandler struct {
	service service.NotService
	proto.UnimplementedNotificationServiceServer
}

func NewConnectionHandler(service service.NotService) (NotificationHandler, error) {
	return NotificationHandler{
		service: service,
	}, nil
}

func (h NotificationHandler) GetAllNots(ctx context.Context, req *proto.GetAllNotsReq) (res *proto.GetAllNotsRes, err error) {
	id := req.UserId
	nots, err := h.service.GetAllNotUser(id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "DB exception.")
	}
	response := &proto.GetAllNotsRes{Nots: nots}
	return response, nil
}

func (h NotificationHandler) CreateNot(ctx context.Context, req *proto.CreateNotReq) (res *proto.EmptyResponse, err error) {
	return
}
