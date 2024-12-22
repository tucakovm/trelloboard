package main

import (
	"context"
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

	"github.com/sony/gobreaker"

	"github.com/redis/go-redis/v9"
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

	tracer := tp.Tracer("user-service")

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

	// Set up Redis
	log.Println("Initializing Redis client...")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	defer func() {
		log.Println("Closing Redis client...")
		if err := redisClient.Close(); err != nil {
			log.Fatalf("Failed to close Redis client: %v", err)
		}
	}()
	log.Println("Redis client initialized successfully.")

	// Test Redis connection
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully.")

	// ProjectService connection
	projectConn, err := grpc.DialContext(
		ctx,
		cfg.FullProjectServiceAddress(),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
	)

	projectClient := users.NewProjectServiceClient(projectConn)
	log.Println("ProjectService Gateway registered successfully.")

	timeoutContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	consulAddress := os.Getenv("CONSUL_ADDRESS")

	log.Println("Initializing User Repository...")
	repoUser, err := repositories.NewUserRepo(timeoutContext, tracer)
	if err != nil {
		log.Fatal("Failed to initialize User Repository: ", err)
	}
	defer repoUser.Disconnect(timeoutContext)
	log.Println("User Repository initialized successfully.")

	log.Println("Initializing Blacklist Service...")
	blacklistRepo, err := repositories.NewBlacklistConsul(consulAddress)
	if err != nil {
		log.Fatal("Failed to initialize Blacklist Repository: ", err)
	}
	log.Println("Blacklist Service initialized successfully.")

	userServiceCB := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "User Service Circuit Breaker",
		MaxRequests: 1,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 2
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("[Circuit Breaker] '%s' transitioned from '%s' to '%s'\n", name, from, to)
		},
		IsSuccessful: func(err error) bool {
			if err == nil {
				return true
			}
			grpcErr, ok := status.FromError(err)
			if ok && grpcErr.Code() == codes.DeadlineExceeded {
				return false // Označi kao neuspeh za circuit breaker
			}
			return true // Sve ostale greške tretiraj kao uspeh za circuit breaker
		},
	})

	serviceUser, err := services.NewUserService(*repoUser, blacklistRepo, tracer)

	handlerUser, err := h.NewUserHandler(serviceUser, projectClient, tracer, userServiceCB)
	if err != nil {
		log.Fatal("Failed to initialize User Handler: ", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(timeoutUnaryInterceptor(5 * time.Second)), // Timeout na 5 sekundi
	)
	reflection.Register(grpcServer)
	users.RegisterUsersServiceServer(grpcServer, &handlerUser)

	go func() {
		log.Println("Starting gRPC server...")
		if err := grpcServer.Serve(listener); err != nil && err != grpc.ErrServerStopped {
			log.Fatal("gRPC server error: ", err)
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
		semconv.ServiceNameKey.String("user-service"),
	)

	return sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp),
		sdktrace.WithResource(r),
	)
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
