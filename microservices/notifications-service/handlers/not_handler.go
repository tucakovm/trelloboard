package handlers

import (
	"context"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
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
	notsCache, err := h.service.GetAllNotUserCache(ctx, id)
	if err != nil {
		nots, err := h.service.GetAllNotUser(ctx, id)
		if err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, status.Error(codes.InvalidArgument, "DB exception.")
		}
		err = h.service.PostAllNotsCache(id, nots, ctx)
		if err != nil {
			return nil, err
		}
		response := &proto.GetAllNotsRes{Nots: nots}
		return response, nil
	}
	log.Println("response from cache")
	response := &proto.GetAllNotsRes{Nots: notsCache}
	return response, nil
}
