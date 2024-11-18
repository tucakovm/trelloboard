package main

import (
	"api-gateway/config"
	gateway "api-gateway/proto/gateway"
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := config.GetConfig()
	log.Println(cfg.Address)
	log.Println(cfg.ProjectServiceAddress)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		"projects-server:8001",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	client := gateway.NewProjectServiceClient(conn)
	err = gateway.RegisterProjectServiceHandlerClient(
		context.Background(),
		gwmux,
		client,
	)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}
	log.Println("gRPC Gateway registered successfully.")
	log.Printf("ProjectServiceAddress: %s", cfg.ProjectServiceAddress)

	gwServer := &http.Server{
		Addr:    cfg.Address,
		Handler: logRequests(enableCORS(gwmux)),
	}

	go func() {
		if err := gwServer.ListenAndServe(); err != nil {
			log.Fatal("server error: ", err)
		}
	}()

	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, syscall.SIGTERM)

	<-stopCh

	if err = gwServer.Close(); err != nil {
		log.Fatalln("error while stopping server: ", err)
	}
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
		if r.Method == http.MethodPost && r.URL.Path == "/api/project" {
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
