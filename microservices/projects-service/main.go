package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"projects_module/config"
	h "projects_module/handlers"
	proj "projects_module/proto/project"
	"projects_module/repositories"
	"projects_module/services"
	"syscall"
	"time"
)

func main() {
	cfg := config.GetConfig()
	log.Println(cfg.Address)

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatalln(err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(listener)

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

	// Bootstrap gRPC server.
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	// Bootstrap gRPC service server and respond to request.
	proj.RegisterProjectServiceServer(grpcServer, handlerProject)

	//r.HandleFunc("/api/projects/{id}/members", handlerProject.AddMember).Methods(http.MethodPost)
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal("server error: ", err)
		}
	}()

	//r := mux.NewRouter()
	//
	//prjRouter := r.Methods(http.MethodPost).Subrouter()
	//prjRouter.HandleFunc("/api/projects", handlerProject.Create)
	//prjRouter.Use(handlerProject.MiddlewarePatientDeserialization)
	//
	//r.HandleFunc("/api/projects/{username}", handlerProject.GetAll).Methods(http.MethodGet)
	//r.HandleFunc("/api/projects/{id}", handlerProject.Delete).Methods(http.MethodDelete)
	//r.HandleFunc("/api/projects/getById/{id}", handlerProject.GetById).Methods(http.MethodGet)
	//
	//// Define CORS options
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
	//	Addr:    cfg.Address,    // Use the desired port
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
