package main

import (
	"log"
	"net/http"
	h "projects_module/handlers"
	"projects_module/repositories"
	"projects_module/services"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	//serverConfig := config.GetConfig()

	repoProject, err := repositories.NewProjectInMem()
	handleErr(err)

	serviceProject, err := services.NewConnectionService(repoProject)
	handleErr(err)

	handlerProject, err := h.NewConnectionHandler(serviceProject)
	handleErr(err)

	r := mux.NewRouter()
	r.HandleFunc("/api/projects", handlerProject.Create).Methods(http.MethodPost)

	// Define CORS options
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:4200"}), // Set the correct origin
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// Create the HTTP server with CORS handler
	srv := &http.Server{

		Handler: corsHandler(r), // Apply CORS handler to router
		Addr:    ":8000",        // Use the desired port
	}

	// Start the server
	log.Println("Server is running on port 8000")
	log.Fatal(srv.ListenAndServe())
}

func handleErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
