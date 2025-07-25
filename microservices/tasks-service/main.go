package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/EventStore/EventStore-Client-Go/esdb"
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

	connString := fmt.Sprintf("esdb://%s:%s@%s:%s?tls=false", cfg.ESDBUser, cfg.ESDBPass, cfg.ESDBHost, cfg.ESDBPort)
	settings, err := esdb.ParseConnectionString(connString)
	if err != nil {
		log.Fatal(err)
	}
	esdbClient, err := esdb.NewClient(settings)
	if err != nil {
		log.Fatal(err)
	}

	// NoSQL: Initialize Product Repository store
	repoTask, err := repository.NewTaskRepo(timeoutContext, esdbClient, cfg.ESDBGroup, storeLogger, tracer)
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

	serviceTask := service.NewTaskService(*repoTask, tracer, repo)
	handleErr(err)

	handlerProject := h.NewTaskHandler(serviceTask, projectClient, workflowClient, natsConn, tracer)
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

	GetTasksForApiComp(ctx, natsConn, *handlerProject, tracer)
	go SubscribeToDeleteTasksSaga(ctx, natsConn, *serviceTask, tracer)

	go func() {
		if _, err := handlerProject.AddMemberToTask(context.Background()); err != nil {
			log.Fatalf("Failed to subscribe: %v", err)
		}
	}()
	go func() {
		if err := handlerProject.RemoveMemberFromTask(context.Background()); err != nil {
			log.Printf("Failed to subscribe: %v", err)
		}
	}()
	go func() {
		if err := handlerProject.UpdateTaskStatus(context.Background()); err != nil {
			log.Printf("Failed to subscribe: %v", err)
		}
	}()
	go func() {
		if err := handlerProject.CreateTask(context.Background()); err != nil {
			log.Printf("Failed to subscribe: %v", err)
		}
	}()
	go func() {
		if err := handlerProject.UploadFileESDB(context.Background()); err != nil {
			log.Printf("Failed to subscribe: %v", err)
		}
	}()

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
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		resp, err := handler(ctx, req)

		if ctx.Err() == context.DeadlineExceeded {
			return nil, status.Error(codes.DeadlineExceeded, "Request timed out")
		}
		return resp, err
	}
}

func GetTasksForApiComp(
	ctx context.Context,
	natsConn *nats.Conn,
	taskHandler h.TaskHandler,
	tracer trace.Tracer,
) {
	subject := "get-tasks-apiComp"

	_, err := natsConn.Subscribe(subject, func(msg *nats.Msg) {
		var message map[string]string
		if err := json.Unmarshal(msg.Data, &message); err != nil {
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
		ctxWithRemote := trace.ContextWithRemoteSpanContext(ctx, remoteCtx)
		_, span := tracer.Start(ctxWithRemote, "Subscriber.GetWorkflow")
		defer span.End()

		projectId, ok := message["ProjectId"]
		if !ok {
			log.Printf("Invalid message format: %v", message)
			return
		}

		protoReg := &tsk.GetAllTasksReq{Id: projectId}

		taskRes, err := taskHandler.GetAllByProjectId(ctx, protoReg)
		if err != nil {
			log.Printf("Error fetching workflow: %v", err)
		}

		messageDataWorkflow, err := json.Marshal(taskRes.Tasks)
		if err != nil {
			log.Printf("Error marshaling task response: %v", err)
			return
		}

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

func SubscribeToDeleteTasksSaga(ctx context.Context, natsConn *nats.Conn, taskService service.TaskService, tracer trace.Tracer) {
	subjectTasks := "delete-tasks-saga"

	_, err := natsConn.Subscribe(subjectTasks, func(msg *nats.Msg) {

		traceID := msg.Header.Get(nats_helper.TRACE_ID)
		spanID := msg.Header.Get(nats_helper.SPAN_ID)
		if traceID == "" || spanID == "" {
			log.Println("Missing tracing headers in NATS message")
		}

		remoteCtx, err := nats_helper.GetNATSParentContext(msg)
		if err != nil {
			log.Fatal(err)
		}
		ctxWithRemote := trace.ContextWithRemoteSpanContext(ctx, remoteCtx)
		_, span := tracer.Start(ctxWithRemote, "Subscriber.DeleteWorkflowSaga")
		defer span.End()

		go handleDeleteTasksMessage(ctxWithRemote, msg, natsConn, taskService, tracer)
	})
	if err != nil {
		log.Printf("Failed to subscribe to subject %s: %v", subjectTasks, err)
	}
}

func handleDeleteTasksMessage(ctx context.Context, msg *nats.Msg, natsConn *nats.Conn, taskService service.TaskService, tracer trace.Tracer) {
	ctx, span := tracer.Start(ctx, "handleDeleteTasksMessage")
	defer span.End()

	projectID := string(msg.Data)
	log.Printf("[Task Saga] Received delete for project: %s", projectID)

	err := taskService.MarkTasksAsDeleting(projectID, ctx)
	if err != nil {
		log.Printf("[Task Saga] Error marking tasks: %v", err)
		return
	}

	// Prepare workflow saga
	replySubjectWorkflow := nats.NewInbox()
	responseChanWorkflow := make(chan *nats.Msg, 1)

	sub, err := natsConn.ChanSubscribe(replySubjectWorkflow, responseChanWorkflow)
	if err != nil {
		log.Printf("[Task Saga] Error subscribing to workflow reply: %v", err)
		return
	}
	defer sub.Unsubscribe()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	msgToSend := &nats.Msg{
		Subject: "delete-workflow-saga",
		Reply:   replySubjectWorkflow,
		Data:    []byte(projectID),
		Header:  headers,
	}
	err = natsConn.PublishMsg(msgToSend)

	select {
	case <-time.After(5 * time.Second):
		log.Println("[Task Saga] Timeout waiting for workflow reply")
		_ = taskService.UnmarkTasksAsDeleting(projectID, ctx)

	case workflowReply := <-responseChanWorkflow:
		if string(workflowReply.Data) != projectID {
			_ = taskService.UnmarkTasksAsDeleting(projectID, ctx)
			return
		}
		log.Println("[Task Saga] Workflow confirmed, deleting tasks...")

		err := taskService.DeleteTasksByProjectId(projectID, ctx)
		if err != nil {
			log.Printf("[Task Saga] Error deleting tasks: %v", err)
			return
		}

		if msg.Reply != "" {
			err = natsConn.Publish(msg.Reply, []byte(projectID))
			if err != nil {
				log.Printf("[Task Saga] Failed to reply to project: %v", err)
			}
		}
	}
}
