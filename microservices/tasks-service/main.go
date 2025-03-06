package main

import (
	"context"
	"encoding/json"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	nats_helper "tasks-service/nats_helper"
	tsk "tasks-service/proto/task"
	"tasks-service/repository"
	"tasks-service/service"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"tasks-service/config"
	h "tasks-service/handlers"
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

	tracer := tp.Tracer("task-service")

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

	// WorkflowService connection
	workflowConn, err := grpc.DialContext(
		ctx,
		cfg.FullWorkflowServiceAddress(),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	workflowClient := tsk.NewWorkflowServiceClient(workflowConn)
	log.Println("WorkflowService Gateway registered successfully.")

	// ProjectService connection
	projectConn, err := grpc.DialContext(
		ctx,
		cfg.FullProjectServiceAddress(),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
	)
	projectClient := tsk.NewProjectServiceClient(projectConn)
	log.Println("ProjectService Gateway registered successfully.")

	//Nats Conn
	natsConn := NatsConn()
	defer natsConn.Close()

	// Initialize context
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[task-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[task-store] ", log.LstdFlags)

	// NoSQL: Initialize Product Repository store
	repoTask, err := repository.NewTaskRepo(timeoutContext, storeLogger, tracer)
	if err != nil {
		logger.Fatal(err)
	}
	defer repoTask.Disconnect(timeoutContext)
	handleErr(err)

	serviceProject := service.NewTaskService(*repoTask, tracer)
	handleErr(err)

	handlerProject := h.NewTaskHandler(serviceProject, projectClient, workflowClient, natsConn, tracer)
	handleErr(err)

	// Bootstrap gRPC server.
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
	)
	reflection.Register(grpcServer)

	// Bootstrap gRPC service server and respond to request.
	tsk.RegisterTaskServiceServer(grpcServer, handlerProject)

	GetTasksForApiComp(ctx, natsConn, *handlerProject, tracer)

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
		semconv.ServiceNameKey.String("task-service"),
	)

	return sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp),
		sdktrace.WithResource(r),
	)
}

func GetTasksForApiComp(ctx context.Context, natsConn *nats.Conn, taskHandler h.TaskHandler, tracer trace.Tracer) {
	subject := "get-tasks-apiComp"

	_, err := natsConn.Subscribe(subject, func(msg *nats.Msg) {
		// Parsiraj primljenu poruku
		var message map[string]string
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			return
		}

		// Dohvati tracing podatke iz headera
		traceID := msg.Header.Get(nats_helper.TRACE_ID)
		spanID := msg.Header.Get(nats_helper.SPAN_ID)
		if traceID == "" || spanID == "" {
			log.Println("Missing tracing headers in NATS message")
			return
		}

		// Kreiraj kontekst za tracing koristeći roditeljski span
		remoteCtx, err := nats_helper.GetNATSParentContext(msg)
		if err != nil {
			log.Fatal(err)
		}
		ctxWithRemote := trace.ContextWithRemoteSpanContext(ctx, remoteCtx)
		_, span := tracer.Start(ctxWithRemote, "Subscriber.GetWorkflow")
		defer span.End()

		// Proveri prisustvo ProjectId-a
		projectId, ok := message["ProjectId"]
		if !ok {
			log.Printf("Invalid message format: %v", message)
			return
		}

		protoReg := &tsk.GetAllTasksReq{Id: projectId}

		// Pozovi workflow servis
		taskRes, err := taskHandler.GetAllByProjectId(ctx, protoReg)
		if err != nil {
			log.Printf("Error fetching workflow: %v", err)
		}

		// Serijalizuj rezultat
		messageDataWorkflow, err := json.Marshal(taskRes.Tasks)
		if err != nil {
			log.Printf("Error marshaling task response: %v", err)
			return
		}

		// Odgovori na zahtev koristeći reply subject iz primljene poruke
		if msg.Reply != "" {
			if err := natsConn.Publish(msg.Reply, messageDataWorkflow); err != nil {
				log.Printf("Error publishing task response: %v", err)
			}
		} else {
			log.Println("No reply subject provided in the request")
		}
	})
	if err != nil {
		log.Printf("Error subscribing to subject %s: %v", subject, err)
	}
}
