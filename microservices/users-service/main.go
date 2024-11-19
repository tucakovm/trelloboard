package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"time"
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
		log.Fatalln(err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(listener)

	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[user-api] ", log.LstdFlags)
	//storeLogger := log.New(os.Stdout, "[user-store] ", log.LstdFlags)

	// NoSQL: Initialize Product Repository store
	repoUser, err := repositories.NewUserRepo(timeoutContext)
	if err != nil {
		logger.Fatal(err)
	}
	defer repoUser.Disconnect(timeoutContext)
	handleErr(err)

	serviceUser, err := services.NewUserService(*repoUser)
	handleErr(err)

	handlerUser, err := h.NewUserHandler(serviceUser)
	handleErr(err)

	// Bootstrap gRPC server.
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	// Bootstrap gRPC service server and respond to request.
	users.RegisterUsersServiceServer(grpcServer, &handlerUser)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal("server error: ", err)
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
}

func handleErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
