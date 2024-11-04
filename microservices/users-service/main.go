package main

import (
	"log"
	"net/http"
	"users_module/handlers"
)

func registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method == http.MethodPost {
		w.Write([]byte("User registered successfully"))
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/verify", handlers.VerifyHandler)
	http.ListenAndServe(":8080", nil)
	log.Println("Server running on http://localhost:8080")
	//log.Fatal(http.ListenAndServe(":8080", nil))

}
