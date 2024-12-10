package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
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

	repo, err := repository.NewHDFSRepository(storeLogger, "namenode:8020")
	if err != nil {
		log.Fatalf("Failed to initialize HDFS client: %v", err)
	}
	defer repo.Close()
	log.Println("created hdfs repo")

	serviceProject := service.NewTaskService(*repoTask, tracer)
	handleErr(err)

	handlerProject := h.NewTaskHandler(serviceProject, projectClient, natsConn, tracer)
	handleErr(err)

	// Bootstrap gRPC server.
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
	)
	reflection.Register(grpcServer)

	// Bootstrap gRPC service server and respond to request.
	tsk.RegisterTaskServiceServer(grpcServer, handlerProject)

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
