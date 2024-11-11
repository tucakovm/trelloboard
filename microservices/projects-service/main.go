package main

import (
	"context"
	"log"
	"net/http"
	"os"
	h "projects_module/handlers"
	"projects_module/repositories"
	"projects_module/services"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	//cfg := config.GetConfig()

	// Initialize context
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[project-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[project-store] ", log.LstdFlags)

	// NoSQL: Initialize Product Repository store
	repoProject, err := repositories.New(timeoutContext, storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer repoProject.Disconnect(timeoutContext)
	handleErr(err)

	serviceProject, err := services.NewProjectService(*repoProject)
	handleErr(err)

	handlerProject, err := h.NewConnectionHandler(serviceProject)
	handleErr(err)

	r := mux.NewRouter()

	prjRouter := r.Methods(http.MethodPost).Subrouter()
	prjRouter.HandleFunc("/api/projects", handlerProject.Create)
	prjRouter.Use(handlerProject.MiddlewareProjectDeserialization)

	r.HandleFunc("/api/projects", handlerProject.GetAll).Methods(http.MethodGet)
	r.HandleFunc("/api/projects/{id}", handlerProject.Delete).Methods(http.MethodDelete)

	// Define CORS options
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:4200"}), // Set the correct origin
		handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "OPTIONS", "PUT"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// Create the HTTP server with CORS handler
	srv := &http.Server{

		Handler: corsHandler(r), // Apply CORS handler to router
		Addr:    ":8001",        // Use the desired port
	}

	// Start the server
	log.Fatal(srv.ListenAndServe())
}

func handleErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
