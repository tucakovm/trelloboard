package main

import (
	"context"
	"log"
	"net/http"
	"users_module/handlers"
	"users_module/repositories"
)

func main() {
	ctx := context.Background()

	// Initialize the UserRepo (MongoDB client)
	_, err := repositories.NewUserRepo(ctx)
	if err != nil {
		log.Fatalf("Could not create UserRepo: %v", err)
	}

	// Pass TaskRepo to handlers
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		handlers.RegisterHandler(w, r)
	})
	http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		handlers.VerifyHandler(w, r)
	})

	// Start server
	log.Println("Server running on http://localhost:8003")
	log.Fatal(http.ListenAndServe(":8003", nil))
}
