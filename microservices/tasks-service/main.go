package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"tasks-service/config"
	h "tasks-service/handlers"
	"tasks-service/proto/task"
	"tasks-service/repository"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	cfg := config.GetConfig() // Fetch config
	log.Println("Tasks Service started on:", cfg.Address)

	// Initialize context and logger
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize task repository
	repo, err := repository.NewTaskRepo(timeoutContext)
	if err != nil {
		log.Fatalf("Failed to initialize task repository: %v", err)
	}
	//defer repo.Cli.Disconnect()

	// Initialize task handler
	taskHandler := h.NewTaskHandler(repo)

	// Initialize gRPC server
	grpcServer := grpc.NewServer()

	// Register task service with gRPC
	task.RegisterTaskServiceServer(grpcServer, taskHandler)

	// Enable gRPC reflection
	reflection.Register(grpcServer)

	// Start gRPC server in a goroutine
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatalf("Failed to listen on %v: %v", cfg.Address, err)
	}

	// Start the gRPC server
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Setting up HTTP routes and server
	// Initialize HTTP router and handlers for API Gateway
	// You can reuse the taskHandler or create separate handlers for HTTP if necessary
	httpHandler := h.NewTaskHandler(repo)

	// Set up HTTP router
	httpRouter := mux.NewRouter()
	httpRouter.HandleFunc("/api/tasks", httpHandler.Create).Methods("POST")
	httpRouter.HandleFunc("/api/tasks", httpHandler.GetAll).Methods("GET")
	//httpRouter.HandleFunc("/api/tasks/{id}", httpHandler.GetByID).Methods("GET")
	//httpRouter.HandleFunc("/api/tasks/{id}", httpHandler.Update).Methods("PUT")
	httpRouter.HandleFunc("/api/tasks/{id}", httpHandler.Delete).Methods("DELETE")

	// Set up CORS middleware for the HTTP server
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:4200"}), // Adjust as needed
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// Start the HTTP server for the API Gateway
	httpServer := &http.Server{
		Handler: corsHandler(httpRouter),
		Addr:    ":8002", // Port for the HTTP server
	}

	// Log HTTP server startup
	log.Println("HTTP server started on port 8002")

	// Gracefully shut down servers on SIGTERM or SIGINT
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGTERM, syscall.SIGINT)

	// Block until a signal is received
	<-stopCh

	// Shutdown the servers gracefully
	log.Println("Shutting down servers...")
	grpcServer.Stop()
	httpServer.Close()
	log.Println("Servers gracefully shut down")
}
