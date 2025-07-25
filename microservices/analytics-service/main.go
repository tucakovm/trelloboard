package main

import (
	"analytics-service/config"
	h "analytics-service/handlers"
	proto "analytics-service/proto/analytics"
	"analytics-service/repositories"
	"analytics-service/services"
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func main() {
	// Initialize config and logger
	cfg := config.GetConfig()
	log.Println("Analytics Service started with config:", cfg.Address)

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

	// Initialize NATS connection
	natsConn := NatsConn()
	defer natsConn.Close()

	// Initialize repository and service
	logger := log.New(os.Stdout, "[analytics-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[analytics-store] ", log.LstdFlags)

	store, err := repositories.NewAnalyticsRepo(storeLogger, tp.Tracer("analytics-service"))
	if err != nil {
		logger.Fatal(err)
	}
	defer store.CloseSession()
	store.CreateTables(ctx)

	// Create service
	serviceAnalytics := service.NewAnalyticsService(*store, tp.Tracer("analytics-service"))

	// Set up NATS subscribers
	setupTaskEventSubscribers(ctx, natsConn, *serviceAnalytics, tp.Tracer("analytics-service"))

	// Set up gRPC server
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", cfg.Address, err)
	}
	defer listener.Close()

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(timeoutUnaryInterceptor(5 * time.Second)))
	reflection.Register(grpcServer)

	// Register gRPC handlers
	handlerAnalytics, err := h.NewAnalyticsHandler(*serviceAnalytics, tp.Tracer("analytics-service"))
	if err != nil {
		log.Fatalf("Failed to create handler: %v", err)
	}
	proto.RegisterAnalyticsServiceServer(grpcServer, handlerAnalytics)

	// Start gRPC server in a goroutine
	go func() {
		log.Println("Analytics Service listening on port:", cfg.Address)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	//test()

	// Graceful shutdown
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGTERM, syscall.SIGINT)
	<-stopCh
	grpcServer.GracefulStop()
	log.Println("Analytics Service stopped")
}

// Timeout for unary RPCs
func timeoutUnaryInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Create a new context with a timeout
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Handle the request
		resp, err := handler(ctx, req)

		// If the context expired, return a timeout error
		if ctx.Err() == context.DeadlineExceeded {
			return nil, status.Error(codes.DeadlineExceeded, "Request timed out")
		}
		return resp, err
	}
}

func getNATSParentSpanContext(msg *nats.Msg) (trace.SpanContext, error) {
	traceID := msg.Header.Get("TRACE_ID")
	spanID := msg.Header.Get("SPAN_ID")

	// Validate the IDs
	tID, err := trace.TraceIDFromHex(traceID)
	if err != nil {
		return trace.SpanContext{}, fmt.Errorf("invalid trace_id: %v", err)
	}
	sID, err := trace.SpanIDFromHex(spanID)
	if err != nil {
		return trace.SpanContext{}, fmt.Errorf("invalid span_id: %v", err)
	}

	// Construct the SpanContext
	return trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tID,
		SpanID:     sID,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	}), nil
}

// Setup NATS subscribers for task-related events
func setupTaskEventSubscribers(ctx context.Context, natsConn *nats.Conn, serviceAnalytics service.AnalyticsService, tracer trace.Tracer) {
	// List of subjects and subscription handlers
	subjects := []string{
		"task-status-changed",
		"create-task",
		"add-to-task",
		"remove-from-task",
		"update-task",
	}

	// Loop over each subject and set up a corresponding subscriber
	for _, subject := range subjects {
		go subscribeToNATS(ctx, natsConn, serviceAnalytics, tracer, subject)
	}
}

// General NATS subscription handler
func subscribeToNATS(ctx context.Context, natsConn *nats.Conn, serviceAnalytics service.AnalyticsService, tracer trace.Tracer, subject string) {
	_, err := natsConn.Subscribe(subject, func(msg *nats.Msg) {
		var message map[string]interface{}
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			return
		}

		traceID, spanID := msg.Header.Get("TRACE_ID"), msg.Header.Get("SPAN_ID")
		if traceID == "" || spanID == "" {
			log.Println("Missing tracing headers in NATS message")
			return
		}

		// Extract SpanContext from NATS headers
		spanCtx, err := getNATSParentSpanContext(msg)
		if err != nil {
			log.Printf("Error extracting SpanContext from headers: %v", err)
			return
		}

		// Start a new span for the subscriber's work
		newCtx, span := tracer.Start(trace.ContextWithRemoteSpanContext(ctx, spanCtx), "Subscriber."+subject)
		defer span.End()

		// Handle the different types of messages based on the subject
		handleTaskEvent(newCtx, message, serviceAnalytics)
	})
	if err != nil {
		log.Printf("Failed to subscribe to NATS subject %s: %v", subject, err)
	}
}

// Handle task-related events and update analytics accordingly
func handleTaskEvent(ctx context.Context, message map[string]interface{}, serviceAnalytics service.AnalyticsService) {
	projectID, _ := message["ProjectId"].(string)

	switch message["event"] {
	case "create-task":
		// Task has been created
		taskID, _ := message["TaskId"].(string)
		serviceAnalytics.UpdateTaskCount(ctx, projectID, taskID, 1) // Increase task count
		log.Printf("Task created: %s for Project: %s", taskID, projectID)

	case "update-task":
		// Task status has changed
		taskID, _ := message["TaskId"].(string)
		newStatus, _ := message["TaskStatus"].(string)
		serviceAnalytics.UpdateTaskStatus(ctx, projectID, taskID, newStatus)
		log.Printf("Task status changed: %s to %s for Project: %s", taskID, newStatus, projectID)

	case "add-to-task":
		// A member has been added to a task
		taskID, _ := message["TaskId"].(string)
		memberID, _ := message["UserId"].(string)
		username, _ := message["UserName"].(string)
		serviceAnalytics.AddMemberToTask(ctx, projectID, taskID, username)
		log.Printf("Member added: %s to Task: %s for Project: %s", memberID, taskID, projectID)

	case "remove-from-task":
		// A member has been removed from a task
		taskID, _ := message["TaskId"].(string)
		username, _ := message["Username"].(string)
		serviceAnalytics.RemoveMemberFromTask(ctx, projectID, taskID, username)
		log.Printf("Member removed: %s from Task: %s for Project: %s", username, taskID, projectID)

	default:
		log.Printf("Unknown event: %v", message["event"])
	}
}

// NATS connection function
func NatsConn() *nats.Conn {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Fatal("NATS_URL environment variable not set")
	}

	opts := []nats.Option{
		nats.Timeout(10 * time.Second), // Set a timeout for connecting
	}

	conn, err := nats.Connect(natsURL, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to NATS at %s: %v", natsURL, err)
	}
	log.Println("Connected to NATS at:", natsURL)
	return conn
}

func newExporter(address string) (*jaeger.Exporter, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(address)))
	if err != nil {
		return nil, err
	}
	return exp, nil
}

// Trace provider setup
func newTraceProvider(exp *jaeger.Exporter) *sdktrace.TracerProvider {
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewSchemaless(semconv.ServiceNameKey.String("analytics-service"))),
	)
}

////// test code bellow
//
//func test() {
//	// Initialize mock repository and tracer (in a real application, use actual implementations)
//	mockRepo := repositories.AnalyticsRepo{}
//	mockTracer := trace.NewNoopTracerProvider().Tracer("")
//
//	// Create the AnalyticsService instance
//	analyticsService := service.NewAnalyticsService(mockRepo, mockTracer)
//
//	// Example data to use in service functions
//	analytic := &proto.Analytic{
//		ProjectId:    "project-1",
//		TotalTasks:   5,
//		StatusCounts: map[string]int32{"in-progress": 3, "completed": 2},
//	}
//
//	// Call Create and check if it was successful
//	if err := createAnalytics(analyticsService, analytic); err != nil {
//		log.Printf("Create failed: %v", err)
//	} else {
//		fmt.Println("Create successful")
//	}
//
//	// Call GetAnalytics and check if it was successful
//	if err := getAnalytics(analyticsService, "project-1"); err != nil {
//		log.Printf("GetAnalytics failed: %v", err)
//	} else {
//		log.Println("GetAnalytics successful")
//	}
//
//	// Call UpdateTaskCount and check if it was successful
//	if err := updateTaskCount(analyticsService, "project-1", "task-2", 5); err != nil {
//		log.Printf("UpdateTaskCount failed: %v", err)
//	} else {
//		log.Println("UpdateTaskCount successful")
//	}
//
//	// Call UpdateTaskStatus and check if it was successful
//	if err := updateTaskStatus(analyticsService, "project-1", "task-1", "completed"); err != nil {
//		log.Printf("UpdateTaskStatus failed: %v", err)
//	} else {
//		log.Println("UpdateTaskStatus successful")
//	}
//
//	// Call AddMemberToTask and check if it was successful
//	if err := addMemberToTask(analyticsService, "project-1", "task-1", "member-1"); err != nil {
//		log.Printf("AddMemberToTask failed: %v", err)
//	} else {
//		log.Println("AddMemberToTask successful")
//	}
//
//	// Call RemoveMemberFromTask and check if it was successful
//	if err := removeMemberFromTask(analyticsService, "project-1", "task-1", "member-1"); err != nil {
//		log.Printf("RemoveMemberFromTask failed: %v", err)
//	} else {
//		log.Println("RemoveMemberFromTask successful")
//	}
//
//	anal := getAnalytics(analyticsService, "project-1")
//	log.Println(anal)
//
//}
//
//// createAnalytics calls the Create method of AnalyticsService
//func createAnalytics(service *service.AnalyticsService, analytic *proto.Analytic) error {
//	err := service.Create(context.Background(), analytic)
//	return err
//}
//
//// getAnalytics calls the GetAnalytics method of AnalyticsService
//func getAnalytics(service *service.AnalyticsService, projectID string) error {
//	_, err := service.GetAnalytics(context.Background(), projectID)
//	return err
//}
//
//// updateTaskCount calls the UpdateTaskCount method of AnalyticsService
//func updateTaskCount(service *service.AnalyticsService, projectID string, taskId string, countDelta int) error {
//	err := service.UpdateTaskCount(context.Background(), projectID, taskId, countDelta)
//	return err
//}
//
//// updateTaskStatus calls the UpdateTaskStatus method of AnalyticsService
//func updateTaskStatus(service *service.AnalyticsService, projectID, taskID, newStatus string) error {
//	err := service.UpdateTaskStatus(context.Background(), projectID, taskID, newStatus)
//	return err
//}
//
//// addMemberToTask calls the AddMemberToTask method of AnalyticsService
//func addMemberToTask(service *service.AnalyticsService, projectID, taskID, memberID string) error {
//	err := service.AddMemberToTask(context.Background(), projectID, taskID, memberID)
//	return err
//}
//
//// removeMemberFromTask calls the RemoveMemberFromTask method of AnalyticsService
//func removeMemberFromTask(service *service.AnalyticsService, projectID, taskID, memberID string) error {
//	err := service.RemoveMemberFromTask(context.Background(), projectID, taskID, memberID)
//	return err
//}
