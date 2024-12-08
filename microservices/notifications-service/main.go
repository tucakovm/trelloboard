package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"not_module/config"
	h "not_module/handlers"
	nats_helper "not_module/nats_helper"
	not "not_module/proto/notification"
	"not_module/repositories"
	"not_module/service"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	cfg := config.GetConfig()
	log.Println(cfg.Address)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exp, err := newExporter(cfg.JaegerEndpoint)
	if err != nil {
		log.Fatalf("failed to initialize exporter: %v", err)
	}

	tp := newTraceProvider(exp)
	defer func() { _ = tp.Shutdown(ctx) }()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tracer := tp.Tracer("notifications-service")

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatalln(err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(listener)

	log.Println("Not Serices listening on port :" + cfg.Address)

	//Nats Conn
	natsConn := NatsConn()
	defer natsConn.Close()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[notification-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[notification-store] ", log.LstdFlags)

	// NoSQL: Initialize Product Repository store
	store, err := repositories.New(storeLogger, tracer)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.CloseSession()
	store.CreateTables(ctx)

	serviceNot := service.NewNotService(*store, tracer)

	handlerNots, err := h.NewConnectionHandler(*serviceNot, tracer)
	handleErr(err)

	//NATS subs
	NatsRemoveFromProject(ctx, natsConn, *serviceNot, tracer)
	NatsAddToProject(ctx, natsConn, *serviceNot, tracer)
	NatsAddToTask(ctx, natsConn, *serviceNot, tracer)
	NatsRemoveFromTask(ctx, natsConn, *serviceNot, tracer)
	NatsUpdateTask(ctx, natsConn, *serviceNot, tracer)
	NatsCreateTask(ctx, natsConn, *serviceNot, tracer)

	// Bootstrap gRPC server.
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	// Bootstrap gRPC service server and respond to request.
	not.RegisterNotificationServiceServer(grpcServer, handlerNots)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal("server error: ", err)
		}
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGTERM)

	<-stopCh

	grpcServer.Stop()
}

func handleErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func NatsConn() *nats.Conn {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Fatal("NATS_URL environment variable not set")
	}

	opts := []nats.Option{
		nats.Timeout(10 * time.Second), // Postavi timeout za povezivanje
	}

	conn, err := nats.Connect(natsURL, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to NATS at %s: %v", natsURL, err)
	}
	log.Println("Connected to NATS at:", natsURL)
	return conn
}

func NatsRemoveFromProject(ctx context.Context, natsConn *nats.Conn, notService service.NotService, tracer trace.Tracer) {
	subjectRFP := "removed-from-project"

	_, _ = natsConn.Subscribe(subjectRFP, func(msg *nats.Msg) {

		var message map[string]string
		err := json.Unmarshal(msg.Data, &message)
		if err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			return
		}

		traceID := msg.Header.Get(nats_helper.TRACE_ID)
		spanID := msg.Header.Get(nats_helper.SPAN_ID)

		if traceID == "" || spanID == "" {
			log.Println("Missing tracing headers in NATS message")
			return
		}

		remoteCtx, err := nats_helper.GetNATSParentContext(msg)
		if err != nil {
			log.Fatal(err)
		}
		_, span := tracer.Start(trace.ContextWithRemoteSpanContext(ctx, remoteCtx), "Subscriber.RemoveFromProject")
		defer span.End()

		userId, userOk := message["UserId"]
		projectId, projectOk := message["ProjectId"]

		if !userOk || !projectOk {
			log.Printf("Invalid message format: %v", message)
			return
		}
		notificationMessage := fmt.Sprintf("You have been removed from project with id :%s", projectId)
		notificationMessageProj := fmt.Sprintf("Member with id :%s has been removed from the project with id :%s", userId, projectId)

		notification := &not.Notification{
			UserId:    userId,
			CreatedAt: timestamppb.New(time.Now()),
			Message:   notificationMessage,
			Status:    "unread",
		}

		notificationProject := &not.Notification{
			UserId:    projectId,
			CreatedAt: timestamppb.New(time.Now()),
			Message:   notificationMessageProj,
			Status:    "unread",
		}

		err = notService.Create(ctx, notification)
		if err != nil {
			log.Printf("Error saving notification: %v", err)
		} else {
			log.Printf("Notification saved: %s", notificationMessage)
		}
		err = notService.Create(ctx, notificationProject)
		if err != nil {
			log.Printf("Error saving notification: %v", err)
		} else {
			log.Printf("Notification saved: %s", notificationMessage)
		}
	})
}

func NatsAddToProject(ctx context.Context, natsConn *nats.Conn, notService service.NotService, tracer trace.Tracer) {
	subjectATP := "add-to-project"

	_, _ = natsConn.Subscribe(subjectATP, func(msg *nats.Msg) {
		var message map[string]string
		err := json.Unmarshal(msg.Data, &message)
		if err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			return
		}

		traceID := msg.Header.Get(nats_helper.TRACE_ID)
		spanID := msg.Header.Get(nats_helper.SPAN_ID)

		if traceID == "" || spanID == "" {
			log.Println("Missing tracing headers in NATS message")
			return
		}

		remoteCtx, err := nats_helper.GetNATSParentContext(msg)
		if err != nil {
			log.Fatal(err)
		}
		_, span := tracer.Start(trace.ContextWithRemoteSpanContext(ctx, remoteCtx), "Subscriber.AddToProject")
		defer span.End()

		userId, userOk := message["UserId"]
		projectId, projectOk := message["ProjectId"]

		if !userOk || !projectOk {
			log.Printf("Invalid message format: %v", message)
			return
		}
		notificationMessage := fmt.Sprintf("You have been added to the project with id :%s", projectId)
		notificationMessageProj := fmt.Sprintf("Member with id :%s has been added to the project with id :%s", userId, projectId)

		notification := &not.Notification{
			UserId:    userId,
			CreatedAt: timestamppb.New(time.Now()),
			Message:   notificationMessage,
			Status:    "unread",
		}

		notificationProject := &not.Notification{
			UserId:    projectId,
			CreatedAt: timestamppb.New(time.Now()),
			Message:   notificationMessageProj,
			Status:    "unread",
		}

		err = notService.Create(ctx, notification)
		if err != nil {
			log.Printf("Error saving notification: %v", err)
		} else {
			log.Printf("Notification saved: %s", notificationMessage)
		}
		err = notService.Create(ctx, notificationProject)
		if err != nil {
			log.Printf("Error saving notification: %v", err)
		} else {
			log.Printf("Notification saved: %s", notificationMessage)
		}
	})
}

func NatsAddToTask(ctx context.Context, natsConn *nats.Conn, notService service.NotService, tracer trace.Tracer) {
	subjectATP := "add-to-task"

	_, _ = natsConn.Subscribe(subjectATP, func(msg *nats.Msg) {
		var message map[string]string
		err := json.Unmarshal(msg.Data, &message)
		if err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			return
		}

		traceID := msg.Header.Get(nats_helper.TRACE_ID)
		spanID := msg.Header.Get(nats_helper.SPAN_ID)

		if traceID == "" || spanID == "" {
			log.Println("Missing tracing headers in NATS message")
			return
		}

		remoteCtx, err := nats_helper.GetNATSParentContext(msg)
		if err != nil {
			log.Fatal(err)
		}
		_, span := tracer.Start(trace.ContextWithRemoteSpanContext(ctx, remoteCtx), "Subscriber.AddToTask")
		defer span.End()

		userId, userOk := message["UserId"]
		taskId, taskOk := message["TaskId"]
		projectId, projectOk := message["ProjectId"]

		if !userOk || !taskOk || !projectOk {
			log.Printf("Invalid message format: %v", message)
			return
		}
		notificationMessage := fmt.Sprintf("You have been added to the task with id :%s", taskId)
		notificationMessageProj := fmt.Sprintf("Member with id :%s has been added to the task with id :%s", userId, taskId)

		notification := &not.Notification{
			UserId:    userId,
			CreatedAt: timestamppb.New(time.Now()),
			Message:   notificationMessage,
			Status:    "unread",
		}

		notificationProject := &not.Notification{
			UserId:    projectId,
			CreatedAt: timestamppb.New(time.Now()),
			Message:   notificationMessageProj,
			Status:    "unread",
		}

		err = notService.Create(ctx, notification)
		if err != nil {
			log.Printf("Error saving notification: %v", err)
		} else {
			log.Printf("Notification saved: %s", notificationMessage)
		}
		err = notService.Create(ctx, notificationProject)
		if err != nil {
			log.Printf("Error saving notification: %v", err)
		} else {
			log.Printf("Notification saved: %s", notificationMessage)
		}
	})
}

func NatsRemoveFromTask(ctx context.Context, natsConn *nats.Conn, notService service.NotService, tracer trace.Tracer) {
	subjectRTT := "remove-from-task"

	_, _ = natsConn.Subscribe(subjectRTT, func(msg *nats.Msg) {
		var message map[string]string
		err := json.Unmarshal(msg.Data, &message)
		if err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			return
		}

		traceID := msg.Header.Get(nats_helper.TRACE_ID)
		spanID := msg.Header.Get(nats_helper.SPAN_ID)

		if traceID == "" || spanID == "" {
			log.Println("Missing tracing headers in NATS message")
			return
		}

		remoteCtx, err := nats_helper.GetNATSParentContext(msg)
		if err != nil {
			log.Fatal(err)
		}
		_, span := tracer.Start(trace.ContextWithRemoteSpanContext(ctx, remoteCtx), "Subscriber.RemoveFromTask")
		defer span.End()

		userId, userOk := message["UserId"]
		taskId, taskOk := message["TaskId"]
		projectId, projectOk := message["ProjectId"]

		if !userOk || !taskOk || !projectOk {
			log.Printf("Invalid message format: %v", message)
			return
		}
		notificationMessage := fmt.Sprintf("You have been removed from the task with id :%s", taskId)
		notificationMessageProj := fmt.Sprintf("Member with id :%s has been removed from the task with id :%s", userId, taskId)

		notification := &not.Notification{
			UserId:    userId,
			CreatedAt: timestamppb.New(time.Now()),
			Message:   notificationMessage,
			Status:    "unread",
		}

		notificationProject := &not.Notification{
			UserId:    projectId,
			CreatedAt: timestamppb.New(time.Now()),
			Message:   notificationMessageProj,
			Status:    "unread",
		}

		err = notService.Create(ctx, notification)
		if err != nil {
			log.Printf("Error saving notification: %v", err)
		} else {
			log.Printf("Notification saved: %s", notificationMessage)
		}
		err = notService.Create(ctx, notificationProject)
		if err != nil {
			log.Printf("Error saving notification: %v", err)
		} else {
			log.Printf("Notification saved: %s", notificationMessage)
		}
	})
}

func NatsUpdateTask(ctx context.Context, natsConn *nats.Conn, notService service.NotService, tracer trace.Tracer) {
	subjectRTT := "update-task"

	_, _ = natsConn.Subscribe(subjectRTT, func(msg *nats.Msg) {
		var message map[string]string
		err := json.Unmarshal(msg.Data, &message)
		if err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			return
		}

		traceID := msg.Header.Get(nats_helper.TRACE_ID)
		spanID := msg.Header.Get(nats_helper.SPAN_ID)

		if traceID == "" || spanID == "" {
			log.Println("Missing tracing headers in NATS message")
			return
		}

		remoteCtx, err := nats_helper.GetNATSParentContext(msg)
		if err != nil {
			log.Fatal(err)
		}
		_, span := tracer.Start(trace.ContextWithRemoteSpanContext(ctx, remoteCtx), "Subscriber.UpdateTask")
		defer span.End()

		taskId, taskOk := message["TaskId"]
		memberIds, membersOk := message["MemberIds"]
		taskStatus, taskStatusOk := message["TaskStatus"]
		projectId, projectIdOk := message["ProjectId"]

		if !taskOk || !membersOk || !taskStatusOk || !projectIdOk {
			log.Printf("Invalid message format: %v", message)
			return
		}

		memberIdList := strings.Split(memberIds, ",")

		notificationMessage := fmt.Sprintf("The status of the task with ID: %s has been updated to: %s.", taskId, taskStatus)

		for _, userId := range memberIdList {
			notification := &not.Notification{
				UserId:    userId,
				CreatedAt: timestamppb.New(time.Now()),
				Message:   notificationMessage,
				Status:    "unread",
			}

			err = notService.Create(ctx, notification)
			if err != nil {
				log.Printf("Error saving notification for user %s: %v", userId, err)
			} else {
				log.Printf("Notification saved for user %s: %s", userId, notificationMessage)
			}
		}
		notificationMessageProjectNot := fmt.Sprintf("The status of the task with ID: %s , on project with ID: %s ,has been updated to: %s.", taskId, projectId, taskStatus)

		notificationProject := &not.Notification{
			UserId:    projectId,
			CreatedAt: timestamppb.New(time.Now()),
			Message:   notificationMessageProjectNot,
			Status:    "unread",
		}

		err = notService.Create(ctx, notificationProject)
		if err != nil {
			log.Printf("Error saving notification for project %s: %v", projectId, err)
		} else {
			log.Printf("Notification saved for project %s: %s", projectId, notificationMessageProjectNot)
		}

	})
}

func NatsCreateTask(ctx context.Context, natsConn *nats.Conn, notService service.NotService, tracer trace.Tracer) {
	subjectCT := "create-task"

	_, _ = natsConn.Subscribe(subjectCT, func(msg *nats.Msg) {
		var message map[string]string
		err := json.Unmarshal(msg.Data, &message)
		if err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			return
		}

		traceID := msg.Header.Get(nats_helper.TRACE_ID)
		spanID := msg.Header.Get(nats_helper.SPAN_ID)

		if traceID == "" || spanID == "" {
			log.Println("Missing tracing headers in NATS message")
			return
		}

		remoteCtx, err := nats_helper.GetNATSParentContext(msg)
		if err != nil {
			log.Fatal(err)
		}
		_, span := tracer.Start(trace.ContextWithRemoteSpanContext(ctx, remoteCtx), "Subscriber.CreateTask")
		defer span.End()

		taskName, taskNameOk := message["TaskName"]
		projectId, projectIdOk := message["ProjectId"]

		if !taskNameOk || !projectIdOk {
			log.Printf("Invalid message format: %v", message)
			return
		}
		notificationMessage := fmt.Sprintf("Task %s has been added.", taskName)

		notification := &not.Notification{
			UserId:    projectId,
			CreatedAt: timestamppb.New(time.Now()),
			Message:   notificationMessage,
			Status:    "unread",
		}

		err = notService.Create(ctx, notification)
		if err != nil {
			log.Printf("Error saving notification: %v", err)
		} else {
			log.Printf("Notification saved: %s", notificationMessage)
		}
	})
}
func newExporter(address string) (sdktrace.SpanExporter, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(address)))
	if err != nil {
		return nil, err
	}
	return exp, nil
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	r := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("notifications-service"),
	)

	return sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp),
		sdktrace.WithResource(r),
	)
}
