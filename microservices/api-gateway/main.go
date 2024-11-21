package main

import (
	"api-gateway/config"
	gateway "api-gateway/proto/gateway"
	"bytes"
	"context"
	"errors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"

	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.GetConfig()
	log.Println("Starting API Gateway...")
	log.Println("Address:", cfg.Address)

	// Create a context for the gateway
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// ProjectService connection
	projectConn, err := grpc.DialContext(
		ctx,
		cfg.FullProjectServiceAddress(),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	conn2, err := grpc.DialContext(
		ctx,
		"users-server:8003",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalln("Failed to dial ProjectService:", err)
	}

	// Create a gRPC Gateway multiplexer
	gwmux := runtime.NewServeMux()

	// Register ProjectService HTTP handlers
	projectClient := gateway.NewProjectServiceClient(projectConn)
	if err := gateway.RegisterProjectServiceHandlerClient(ctx, gwmux, projectClient); err != nil {
		log.Fatalln("Failed to register ProjectService gateway:", err)
	}
	log.Println("ProjectService Gateway registered successfully.")

	// TaskService connection
	taskConn, err := grpc.DialContext(
		ctx,
		cfg.FullTaskServiceAddress(),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	client2 := gateway.NewUsersServiceClient(conn2)
	err = gateway.RegisterUsersServiceHandlerClient(
		context.Background(),
		gwmux,
		client2,
	)
	if err != nil {
		log.Fatalln("Failed to dial TaskService:", err)
	}

	// Register TaskService HTTP handlers
	taskClient := gateway.NewTaskServiceClient(taskConn)
	if err := gateway.RegisterTaskServiceHandlerClient(ctx, gwmux, taskClient); err != nil {
		log.Fatalln("Failed to register TaskService gateway:", cfg.FullTaskServiceAddress(), err)
	}
	log.Println("TaskService Gateway registered successfully.")

	// Start the HTTP server
	gwServer := &http.Server{
		Addr:    cfg.Address,
		Handler: enableCORS(gwmux),
	}

	go func() {
		log.Printf("API Gateway listening on %s\n", cfg.Address)
		if err := gwServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v\n", err)
		}
	}()

	// Graceful shutdown handling
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	<-stopCh

	log.Println("Shutting down API Gateway...")
	if err := gwServer.Close(); err != nil {
		log.Fatalf("Error while stopping server: %v\n", err)
	}
	log.Println("API Gateway stopped.")
}

func enableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func logRequests(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Logovanje osnovnih informacija o HTTP zahtevu
		log.Printf("HTTP Request - Method: %s, URL: %s, Headers: %v", r.Method, r.URL.String(), r.Header)

		// Proveri da li je to POST zahtev na /api/project (pretpostavljam da je ruta za Create)
		if r.Method == http.MethodPost && r.URL.Path == "/api/task" {
			// Možeš koristiti log.Println za logovanje tela zahteva
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("Failed to read request body: %v", err)
			} else {
				log.Printf("Request Body: %s", string(body))
			}
			// Ne zaboravi da vratiš telo nazad, jer ćeš ga inače izgubiti
			r.Body = io.NopCloser(bytes.NewReader(body))
		}

		// Nastavi sa obradom zahteva
		h.ServeHTTP(w, r)
	})
}
