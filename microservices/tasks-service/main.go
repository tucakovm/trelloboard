package main

import (
	"context"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	h "tasks-service/handlers"
	"tasks-service/repository"
)

func main() {
	// Set up context for MongoDB connection
	ctx := context.Background()

	// Initialize MongoDB repository
	repo, err := repository.NewTaskRepo(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize task repository: %v", err)
	}
	defer repo.Cli.Disconnect(ctx) // Ensure the database connection is closed when done

	// Initialize task service
	//taskService := service.NewTaskService(repo)

	// Initialize task handler
	taskRepo := &repository.TaskRepo{}
	taskHandler := h.NewTaskHandler(taskRepo)

	// Set up router
	r := mux.NewRouter()
	r.HandleFunc("/api/tasks", taskHandler.Create).Methods(http.MethodPost)
	r.HandleFunc("/api/tasks", taskHandler.GetAll).Methods(http.MethodGet)
	r.HandleFunc("/api/tasks", taskHandler.Delete).Methods(http.MethodDelete)

	// Set up CORS
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:4200"}), // Adjust as needed
		handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// Start the server
	srv := &http.Server{
		Handler: corsHandler(r),
		Addr:    ":8002", // Use the desired port
	}

	log.Println("Server is running on port 8002")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	fmt.Println("Welcome to the Tasks Service!")
}

func handleErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
