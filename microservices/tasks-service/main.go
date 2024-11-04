package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"log"
	"net/http"
	h "tasks-service/handlers"
	"tasks-service/repository"
	"tasks-service/service"

	"github.com/gorilla/mux"
)

func main() {

	repoTask, err := repository.NewTaskInMem()
	handleErr(err)

	serviceTask, err := service.NewTaskService(repoTask)
	handleErr(err)

	handlerTask, err := h.NewConnectionHandler(serviceTask)
	handleErr(err)

	r := mux.NewRouter()
	r.HandleFunc("/api/tasks", handlerTask.Create).Methods(http.MethodPost)

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:4200"}), // Set the correct origin
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	srv := &http.Server{

		Handler: corsHandler(r),
		Addr:    ":8001",
	}

	log.Println("Server is running on port 8001")
	log.Fatal(srv.ListenAndServe())

	fmt.Println("Welcome to the Tasks Service!")
}

func handleErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
