package main

import (
	"context"
	"fmt"
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
	"workflow-service/config"
	h "workflow-service/handlers"
	"workflow-service/models"
	workflow "workflow-service/proto/workflows"
	"workflow-service/repository"
	"workflow-service/services"
)

func main() {
	// Ensure required environment variables are set
	checkEnvVars()
	cfg := config.LoadConfig()

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
	listener, err := net.Listen("tcp", cfg.WorkflowServiceport)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Initialize Workflow Repository
	log.Println("Initializing Workflow Repository...")
	repoWorkflow, err := repository.NewWorkflowRepository(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Workflow Repository: %v", err)
	}
	log.Println("Initialized repo")
	//defer repoWorkflow.Driver.Close(ctx)

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
	/*
	   /*
	   	exists, err := repoWorkflow.CheckWorkflowsExist(ctx)
	   	if err != nil {
	   		log.Fatalf("Error checking if workflows exist: %v", err)
	   	}

	   	if exists {
	   		log.Println("Workflows already exist in the database. Skipping workflow generation.")
	   	} else {
	   		// Add test workflows if no workflows exist
	   		log.Println("No workflows found in the database. Generating test workflows...")*/
	generateTestWorkflows(ctx, repoWorkflow)
	//}

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
		"JAEGER_ENDPOINT", "WORKFLOW_ENDPOINT",
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

func generateTestWorkflows(ctx context.Context, repo *repository.WorkflowRepository) {
	testProjects := []struct {
		ProjectID   string
		ProjectName string
	}{
		{"1", "Project One"},
		{"2", "Project Two"},
		{"3", "Project Three"},
	}

	for _, project := range testProjects {
		// Create the workflow
		err := repo.CreateWorkflow(ctx, models.Workflow{
			ProjectID:   project.ProjectID,
			ProjectName: project.ProjectName,
		})
		if err != nil {
			log.Printf("Failed to create workflow for project %s: %v", project.ProjectID, err)
			continue
		}

		// Add main task
		mainTask := models.TaskNode{
			TaskID:   fmt.Sprintf("main-task-%s", project.ProjectID),
			TaskName: "Main Task",
		}
		err = repo.AddTask(ctx, project.ProjectID, mainTask)
		if err != nil {
			log.Printf("Failed to add main task for project %s: %v", project.ProjectID, err)
			continue
		}

		// Add dependent task
		dependentTask := models.TaskNode{
			TaskID:       fmt.Sprintf("dependent-task-%s", project.ProjectID),
			TaskName:     "Dependent Task",
			Dependencies: []string{mainTask.TaskID},
		}
		err = repo.AddTask(ctx, project.ProjectID, dependentTask)
		if err != nil {
			log.Printf("Failed to add dependent task for project %s: %v", project.ProjectID, err)
			continue
		}

		log.Printf("Successfully created test workflow for project %s", project.ProjectID)
	}
}

func fetchWorkflows(repo *repository.WorkflowRepository) {
	workflow, _ := repo.GetWorkflow(context.Background(), "1")
	log.Println(workflow)
}
