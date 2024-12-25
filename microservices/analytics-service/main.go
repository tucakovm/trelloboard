package main

import (
	"analytics-service/config"
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

	// Set up Jaeger exporter for OpenTelemetry
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

	// Set up NATS subscribers for task-related events
	setupTaskEventSubscribers(ctx, natsConn, serviceAnalytics, tp.Tracer("analytics-service"))

	// Set up gRPC server
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(timeoutUnaryInterceptor(5 * time.Second)))
	reflection.Register(grpcServer)

	// Register gRPC service handlers (if necessary for other calls)
	// handlerAnalytics, err := h.NewConnectionHandler(*serviceAnalytics, tp.Tracer("analytics-service"))
	// if err != nil {
	// 	log.Fatalf("Failed to create handler: %v", err)
	// }
	// analytics.RegisterAnalyticsServiceServer(grpcServer, handlerAnalytics)

	// Start gRPC server in a goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal("server error: ", err)
		}
	}()
	log.Println("Analytics Service listening on port:", cfg.Address)

	// Graceful shutdown handling
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGTERM)
	<-stopCh
	grpcServer.Stop()
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

// Setup NATS subscribers for task-related events
func setupTaskEventSubscribers(ctx context.Context, natsConn *nats.Conn, serviceAnalytics service.AnalyticsService, tracer trace.Tracer) {
	// List of subjects and subscription handlers
	subjects := []string{
		"task-status-changed",
		"task-created",
		"task-member-added",
		"task-member-removed",
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

		traceID, spanID := msg.Header.Get("trace_id"), msg.Header.Get("span_id")
		if traceID == "" || spanID == "" {
			log.Println("Missing tracing headers in NATS message")
			return
		}

		// Get parent context from NATS headers
		remoteCtx, err := getNATSParentContext(msg)
		if err != nil {
			log.Fatal(err)
		}

		// Start a new span for the subscriber's work
		_, span := tracer.Start(trace.ContextWithRemoteSpanContext(ctx, remoteCtx), "Subscriber."+subject)
		defer span.End()

		// Handle the different types of messages based on the subject
		handleTaskEvent(ctx, message, serviceAnalytics)
	})
	if err != nil {
		log.Printf("Failed to subscribe to NATS subject %s: %v", subject, err)
	}
}

// Handle task-related events and update analytics accordingly
func handleTaskEvent(ctx context.Context, message map[string]interface{}, serviceAnalytics service.AnalyticsService) {
	projectID, _ := message["projectId"].(string)

	switch message["event"] {
	case "task-created":
		// Task has been created
		taskID, _ := message["taskId"].(string)
		serviceAnalytics.UpdateTaskCount(ctx, projectID, 1) // Increase task count
		log.Printf("Task created: %s for Project: %s", taskID, projectID)

	case "task-status-changed":
		// Task status has changed
		taskID, _ := message["taskId"].(string)
		newStatus, _ := message["newStatus"].(string)
		serviceAnalytics.UpdateTaskStatus(ctx, projectID, taskID, newStatus)
		log.Printf("Task status changed: %s to %s for Project: %s", taskID, newStatus, projectID)

	case "task-member-added":
		// A member has been added to a task
		taskID, _ := message["taskId"].(string)
		memberID, _ := message["memberId"].(string)
		serviceAnalytics.AddMemberToTask(ctx, projectID, taskID, memberID)
		log.Printf("Member added: %s to Task: %s for Project: %s", memberID, taskID, projectID)

	case "task-member-removed":
		// A member has been removed from a task
		taskID, _ := message["taskId"].(string)
		memberID, _ := message["memberId"].(string)
		serviceAnalytics.RemoveMemberFromTask(ctx, projectID, taskID, memberID)
		log.Printf("Member removed: %s from Task: %s for Project: %s", memberID, taskID, projectID)

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

// OpenTelemetry exporter setup
func newExporter(jaegerEndpoint string) (*jaeger.Exporter, error) {
	// Set up Jaeger exporter for OpenTelemetry
	// Example: Use Jaeger collector endpoint to export traces
	// (You can replace this with another exporter if needed)
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaegerEndpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %v", err)
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

// Get NATS context for tracing
func getNATSParentContext(msg *nats.Msg) (context.Context, error) {
	// Assuming you're extracting trace context from NATS message headers
	// Implement trace context extraction here (depending on your tracing setup)
	return context.Background(), nil
}
