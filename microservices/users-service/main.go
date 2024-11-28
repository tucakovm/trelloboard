package main

import (
	"context"
	"github.com/redis/go-redis/v9"
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
	)
	projectClient := users.NewProjectServiceClient(projectConn)
	log.Println("ProjectService Gateway registered successfully.")

	timeoutContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	consulAddress := os.Getenv("CONSUL_ADDRESS")
	if consulAddress == "" {
		consulAddress = "localhost:8500" // Default to localhost for fallback
	}

	log.Println("Initializing User Repository...")
	repoUser, err := repositories.NewUserRepo(timeoutContext)
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

	serviceUser, err := services.NewUserService(*repoUser, blacklistRepo)
	if err != nil {
		log.Fatal("Failed to initialize User Service: ", err)
	}
	handlerUser, err := h.NewUserHandler(serviceUser, projectClient)
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
