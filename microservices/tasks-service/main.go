package main

import (
	"context"
	"encoding/base64"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	tsk "tasks-service/proto/task"
	"tasks-service/repository"
	"tasks-service/service"
	"tasks-service/utils"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

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

	repo, err := repository.NewHDFSRepository(storeLogger, cfg.NamenodeUrl, tracer)
	if err != nil {
		log.Fatalf("Failed to initialize HDFS client: %v", err)
	}
	defer repo.Close()
	log.Println("created hdfs repo")
	//checkHDFSConnection(repo)

	serviceProject := service.NewTaskService(*repoTask, tracer, repo)
	handleErr(err)

	handlerProject := h.NewTaskHandler(serviceProject, projectClient, natsConn, tracer)
	handleErr(err)

	// Bootstrap gRPC server.
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(timeoutUnaryInterceptor(5 * time.Second)), // Timeout na 5 sekundi
	)
	reflection.Register(grpcServer)

	// Bootstrap gRPC service server and respond to request.
	tsk.RegisterTaskServiceServer(grpcServer, handlerProject)
	if repo.Client == nil {
		log.Println("main.go repo.client nil")
	}
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

func checkHDFSConnection(ctx context.Context, repo *repository.HDFSRepository) {
	// Create a test directory in HDFS
	err := repo.Client.MkdirAll("/tasks/test-dir", 0755)
	if err != nil {
		log.Fatalf("Error creating test directory in HDFS: %v", err)
	} else {
		log.Println("HDFS connection successful: Test directory created.")
	}

	// Generate a unique file name for testing
	// The repo doesn't allow files with the same name
	name := utils.GenerateCode()

	// Create test file content and encode it to Base64
	testFileContent := "This is a test file uploaded on startup."
	encodedContent := base64.StdEncoding.EncodeToString([]byte(testFileContent)) // Base64 encode content

	// Upload the test file to HDFS
	err = repo.UploadFile(ctx, "test-task-id", name, encodedContent)
	if err != nil {
		log.Fatalf("Error uploading test file to HDFS: %v", err)
	} else {
		log.Println("Test file uploaded successfully to HDFS.")
	}

	// Attempt to download the test file from HDFS
	file, err := repo.DownloadFile(ctx, "test-task-id", name)
	if err != nil {
		log.Println("Error downloading test file from HDFS:", err)
	} else {
		// Decode Base64 content after downloading
		decodedContent, decodeErr := base64.StdEncoding.DecodeString(string(file))
		if decodeErr != nil {
			log.Println("Error decoding downloaded file content:", decodeErr)
		} else {
			log.Printf("Downloaded file content: %s\n", string(decodedContent))
		}
	}

	// Delete the test file from HDFS
	err = repo.DeleteFile(ctx, "test-task-id", name)
	if err != nil {
		log.Println("Error deleting test file from HDFS:", err)
	} else {
		log.Println("Deleted test file from HDFS successfully.")
	}
}

func timeoutUnaryInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Kreiraj novi kontekst sa timeout-om
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Obradi zahtev sa novim kontekstom
		resp, err := handler(ctx, req)

		// Ako je kontekst istekao, vrati odgovarajući status
		if ctx.Err() == context.DeadlineExceeded {
			return nil, status.Error(codes.DeadlineExceeded, "Request timed out")
		}
		return resp, err
	}
}
