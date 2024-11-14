package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
	h "users_module/handlers"
	"users_module/repositories"
	"users_module/services"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
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

	r := mux.NewRouter()
	r.HandleFunc("/register", handlerUser.RegisterHandler).Methods(http.MethodPost)
	r.HandleFunc("/verify", handlerUser.VerifyHandler).Methods(http.MethodPost)
	r.HandleFunc("/login", handlerUser.LoginUser).Methods(http.MethodPost)
	r.HandleFunc("/user/{username}", handlerUser.GetUserByUsername).Methods(http.MethodGet)
	r.HandleFunc("/user/{username}", handlerUser.DeleteUserByUsername).Methods(http.MethodDelete)
	r.HandleFunc("/user/change-password", handlerUser.ChangePassword).Methods(http.MethodPut)

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:4200"}), // Set the correct origin
		handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "OPTIONS", "PUT"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// Create the HTTP server with CORS handler
	srv := &http.Server{

		Handler: corsHandler(r), // Apply CORS handler to router
		Addr:    ":8003",        // Use the desired port
	}

	// Start the server
	log.Fatal(srv.ListenAndServe())
}

func handleErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
