package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"projects_module/handlers"
	"projects_module/repositories"
	"projects_module/services"
)

func main() {
	//serverConfig := config.GetConfig()

	repoProject, err := repositories.NewProjectInMem()
	handleErr(err)

	serviceProject, err := services.NewConnectionService(repoProject)
	handleErr(err)

	handlerProject, err := handlers.NewConnectionHandler(serviceProject)
	handleErr(err)

	r := mux.NewRouter()

	r.HandleFunc("/api/project", handlerProject.Create).Methods(http.MethodPost)

	srv := &http.Server{
		Handler: r,
		Addr:    ":8000",
	}
	log.Fatal(srv.ListenAndServe())
}

func handleErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
