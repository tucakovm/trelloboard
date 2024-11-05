package main

import (
	"github.com/google/uuid"
	"log"
	"net/http"
	"projects_module/config"
	"projects_module/domain"
	h "projects_module/handlers"
	"projects_module/repositories"
	"projects_module/services"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	cfg := config.GetConfig()

	repoProject, err := repositories.NewProjectInMem()
	handleErr(err)

	prj1 := domain.Project{
		Id:             uuid.New(),
		Name:           "prj1",
		CompletionDate: time.Time{},
		MinMembers:     2,
		MaxMembers:     3,
	}

	prj2 := domain.Project{
		Id:             uuid.New(),
		Name:           "prj2",
		CompletionDate: time.Time{},
		MinMembers:     2,
		MaxMembers:     3,
	}

	repoProject.Create(prj1)
	repoProject.Create(prj2)

	serviceProject, err := services.NewConnectionService(repoProject)
	handleErr(err)

	handlerProject, err := h.NewConnectionHandler(serviceProject)
	handleErr(err)

	r := mux.NewRouter()
	r.HandleFunc("/api/projects", handlerProject.Create).Methods(http.MethodPost)
	r.HandleFunc("/api/projects", handlerProject.GetAll).Methods(http.MethodGet)
	r.HandleFunc("/api/projects/{id}", handlerProject.Delete).Methods(http.MethodDelete)

	// Define CORS options
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:4200"}), // Set the correct origin
		handlers.AllowedMethods([]string{"GET", "POST", "DELETE"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// Create the HTTP server with CORS handler
	srv := &http.Server{

		Handler: corsHandler(r), // Apply CORS handler to router
		Addr:    cfg.Address,    // Use the desired port
	}

	// Start the server
	log.Fatal(srv.ListenAndServe())
}

func handleErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
