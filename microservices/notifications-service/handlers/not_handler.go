package handlers

import (
	"context"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	proto "not_module/proto/notification"
	"not_module/service"
)

type NotificationHandler struct {
	service service.NotService
	proto.UnimplementedNotificationServiceServer
	Tracer trace.Tracer
}

func NewConnectionHandler(service service.NotService, tracer trace.Tracer) (NotificationHandler, error) {
	return NotificationHandler{
		service: service,
		Tracer:  tracer,
	}, nil
}

func (h NotificationHandler) GetAllNots(ctx context.Context, req *proto.GetAllNotsReq) (res *proto.GetAllNotsRes, err error) {
	ctx, span := h.Tracer.Start(ctx, "h.getAllNots")
	defer span.End()
	id := req.UserId
	nots, err := h.service.GetAllNotUser(ctx, id)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument, "DB exception.")
	}
	response := &proto.GetAllNotsRes{Nots: nots}
	return response, nil
}

func (h NotificationHandler) CreateNot(ctx context.Context, req *proto.CreateNotReq) (res *proto.EmptyResponse, err error) {
	ctx, span := h.Tracer.Start(ctx, "h.createNot")
	defer span.End()

	return
}
