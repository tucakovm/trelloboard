package main

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"users_module/config"
	h "users_module/handlers"
	users "users_module/proto/users"
	"users_module/repositories"
	"users_module/services"
)

func main() {

	cfg, _ := config.LoadConfig()

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

	tracer := tp.Tracer("config-service")

	listener, err := net.Listen("tcp", cfg.UserPort)
	if err != nil {
		log.Fatalln("Failed to create listener: ", err)
	}
	defer func(listener net.Listener) {
		log.Println("Closing listener")
		if err := listener.Close(); err != nil {
			log.Fatal("Error closing listener: ", err)
		}
	}(listener)

	// ProjectService connection
	projectConn, err := grpc.DialContext(
		ctx,
		cfg.FullProjectServiceAddress(),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	projectClient := users.NewProjectServiceClient(projectConn)
	log.Println("ProjectService Gateway registered successfully.")

	timeoutContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Initializing User Repository...")
	repoUser, err := repositories.NewUserRepo(timeoutContext, tracer)
	if err != nil {
		log.Fatal("Failed to initialize User Repository: ", err)
	}
	defer repoUser.Disconnect(timeoutContext)
	log.Println("User Repository initialized successfully.")

	serviceUser, err := services.NewUserService(*repoUser, tracer)
	if err != nil {
		log.Fatal("Failed to initialize User Service: ", err)
	}
	handlerUser, err := h.NewUserHandler(serviceUser, projectClient, tracer)
	if err != nil {
		log.Fatal("Failed to initialize User Handler: ", err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	users.RegisterUsersServiceServer(grpcServer, &handlerUser)

	go func() {
		log.Println("Starting gRPC server...")
		if err := grpcServer.Serve(listener); err != nil && err != grpc.ErrServerStopped {
			log.Fatal("gRPC server error: ", err)
		}
	}()

	//r := mux.NewRouter()
	//r.HandleFunc("/register", handlerUser.RegisterHandler).Methods(http.MethodPost)
	//r.HandleFunc("/verify", handlerUser.VerifyHandler).Methods(http.MethodPost)
	//r.HandleFunc("/login", handlerUser.LoginUser).Methods(http.MethodPost)
	//r.HandleFunc("/user/{username}", handlerUser.GetUserByUsername).Methods(http.MethodGet)
	//r.HandleFunc("/user/{username}", handlerUser.DeleteUserByUsername).Methods(http.MethodDelete)
	//r.HandleFunc("/user/change-password", handlerUser.ChangePassword).Methods(http.MethodPut)
	//
	//corsHandler := handlers.CORS(
	//	handlers.AllowedOrigins([]string{"http://localhost:4200"}), // Set the correct origin
	//	handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "OPTIONS", "PUT"}),
	//	handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	//)
	//
	//// Create the HTTP server with CORS handler
	//srv := &http.Server{
	//
	//	Handler: corsHandler(r), // Apply CORS handler to router
	//	Addr:    ":8003",        // Use the desired port
	//}
	//
	//// Start the server
	//log.Fatal(srv.ListenAndServe())
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
		semconv.ServiceNameKey.String("config-service"),
	)

	return sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp),
		sdktrace.WithResource(r),
	)
}
