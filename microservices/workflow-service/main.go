package main

import (
	"context"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	h "workflow-service/handlers"
	workflow "workflow-service/proto/workflows"
	"workflow-service/repository"
	"workflow-service/services"
)

func main() {
	// Ensure required environment variables are set
	checkEnvVars()

	// Initialize context and cancel function
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize tracing
	exp, err := newExporter(os.Getenv("JAEGER_ENDPOINT"))
	if err != nil {
		log.Fatalf("failed to initialize exporter: %v", err)
	}
	tp := newTraceProvider(exp)
	defer func() { _ = tp.Shutdown(ctx) }()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Create gRPC listener
	listener, err := net.Listen("tcp", os.Getenv("WORKFLOW_ENDPOINT"))
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Set up Redis client
	log.Println("Initializing Redis client...")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("PASSWORD"),
		DB:       0,
	})
	defer redisClient.Close()

	// Test Redis connection
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully.")

	// Initialize Workflow Repository
	log.Println("Initializing Workflow Repository...")
	repoWorkflow, err := repository.NewWorkflowRepository(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Workflow Repository: %v", err)
	}
	defer repoWorkflow.Driver.Close(ctx)

	// Initialize service
	serviceWorkflow := services.NewWorkflowService(*repoWorkflow)
	handlerWorkflow := h.NewWorkflowHandler(serviceWorkflow)

	// Initialize and start gRPC server
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	workflow.RegisterWorkflowServiceServer(grpcServer, handlerWorkflow)

	// Run gRPC server in a separate goroutine
	go func() {
		log.Println("Starting gRPC server...")
		if err := grpcServer.Serve(listener); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// Wait for termination signal
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGTERM)
	<-stopCh

	// Gracefully stop the server
	log.Println("Shutting down gRPC server gracefully...")
	grpcServer.GracefulStop()
}

// handleErr is simplified to log errors and exit.
func handleErr(err error) {
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// checkEnvVars checks that the necessary environment variables are set.
func checkEnvVars() {
	requiredVars := []string{
		"JAEGER_ENDPOINT", "WORKFLOW_ENDPOINT", "REDIS_ADDRESS", "PASSWORD",
	}

	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			log.Fatalf("Environment variable %s is not set", v)
		}
	}
}

// newExporter creates a Jaeger exporter based on the address.
func newExporter(address string) (sdktrace.SpanExporter, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(address)))
	if err != nil {
		return nil, err
	}
	return exp, nil
}

// newTraceProvider creates a new trace provider with resource attributes.
func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	r := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("workflow-service"),
	)

	return sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp),
		sdktrace.WithResource(r),
	)
}
