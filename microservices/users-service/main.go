package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"users_module/config"
	h "users_module/handlers"
	users "users_module/proto/users"
	"users_module/repositories"
	"users_module/services"
)

func main() {

	cfg, _ := config.LoadConfig()
	log.Println(cfg.UserPort)

	listener, err := net.Listen("tcp", ":8003")
	if err != nil {
		log.Fatalln("Failed to create listener: ", err)
	}
	defer func(listener net.Listener) {
		log.Println("Closing listener")
		if err := listener.Close(); err != nil {
			log.Fatal("Error closing listener: ", err)
		}
	}(listener)

	timeoutContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Initializing User Repository...")
	repoUser, err := repositories.NewUserRepo(timeoutContext)
	if err != nil {
		log.Fatal("Failed to initialize User Repository: ", err)
	}
	defer repoUser.Disconnect(timeoutContext)
	log.Println("User Repository initialized successfully.")

	serviceUser, err := services.NewUserService(*repoUser)
	if err != nil {
		log.Fatal("Failed to initialize User Service: ", err)
	}
	handlerUser, err := h.NewUserHandler(serviceUser)
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
