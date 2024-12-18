package main

import (
	"api-gateway/config"
	gateway "api-gateway/proto/gateway"
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {

	cfg := config.GetConfig()

	// Create a gRPC Gateway multiplexer
	gwmux := runtime.NewServeMux()

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
	projectClient := gateway.NewProjectServiceClient(projectConn)
	if err := gateway.RegisterProjectServiceHandlerClient(ctx, gwmux, projectClient); err != nil {
		log.Fatalln("Failed to register ProjectService gateway:", err)
	}
	log.Println("ProjectService Gateway registered successfully.")

	// UserService connection
	userConn, err := grpc.DialContext(
		ctx,
		cfg.FullUserServiceAddress(),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial UserService:", err)
	}

	userClient := gateway.NewUsersServiceClient(userConn)

	if err := gateway.RegisterUsersServiceHandlerClient(ctx, gwmux, userClient); err != nil {
		log.Fatalln("Failed to register UserService gateway:", err)
	}
	log.Println("UserService Gateway registered successfully.")

	// TaskService connection
	taskConn, err := grpc.DialContext(
		ctx,
		cfg.FullTaskServiceAddress(),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	taskClient := gateway.NewTaskServiceClient(taskConn)
	if err := gateway.RegisterTaskServiceHandlerClient(ctx, gwmux, taskClient); err != nil {
		log.Fatalln("Failed to register TaskService gateway:", cfg.FullTaskServiceAddress(), err)
	}
	log.Println("TaskService Gateway registered successfully.")

	// NotificationService connection
	notConn, err := grpc.DialContext(
		ctx,
		cfg.FullNotServiceAddress(),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to dial NotificationService: %v", err)
	}
	notClient := gateway.NewNotificationServiceClient(notConn)
	if err := gateway.RegisterNotificationServiceHandlerClient(ctx, gwmux, notClient); err != nil {
		log.Fatalln("Failed to register NotService gateway:", err)
	}
	log.Println("NotService Gateway registered successfully.")
	log.Println(cfg.FullNotServiceAddress())
	if notClient == nil {
		log.Fatal("Notification service client is nil")
	}

	// Start the HTTP server
	gwServer := &http.Server{
		Addr:    cfg.Address,
		Handler: enableCORS(gwmux),
	}

	go func() {
		log.Printf("API Gateway listening on %s\n", cfg.Address)
		if err := gwServer.ListenAndServeTLS("api-gateway/cert2/cert.crt", "api-gateway/cert2/cert.key"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTPS server error: %v\n", err)
			//	if err := gwServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			//	log.Fatalf("HTTP server error: %v\n", err)
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

var rolePermissions = map[string]map[string][]string{
	"User": {
		"GET": {"/api/projects/{username}", "/api/project/{id}", "/api/tasks/{id}", "/api/task/{id}",
			"/api/users/{username}", "/api/notifications/{userId}", "/api/tasks/{taskId}/files/{fileId}", "/api/tasks/{taskId}/files"},
		"POST":   {"/api/tasks/files"},
		"DELETE": {"/api/users/{username}", "/api/tasks/{taskId}/files/{fileId}"},
		"PUT":    {"/api/users/change-password", "/api/tasks/{id}"},
	},
	"Manager": {
		"GET": {"/api/projects/{username}", "/api/project/{id}", "/api/tasks/{id}", "/api/task/{id}",
			"/api/users/{username}", "/api/notifications/{userId}", "/api/tasks/{taskId}/files/{fileId}"},
		"POST": {"/api/project", "/api/task", "/api/tasks/files"},
		"DELETE": {"/api/project/{id}", "/api/task/{id}", "/api/users/{username}", "/api/task/{projectId}/members/{userId}",
			"/api/projects/{projectId}/members/{userId}", "/api/tasks/{taskId}/files/{fileId}"},
		"PUT": {"/api/users/change-password", "/api/task/{id}/members", "/api/projects/{projectId}/members", "/api/tasks/{id}"},
	},
}

var publicRoutes = []string{
	"/api/users/register",
	"/api/users/login",
	"/api/users/verify",
	"/api/users/magic-link",
	"/api/users/recovery",
	"/api/users/recover-password",
	"/api/tasks/{taskId}/files/{fileId}",
}

func matchesRoute(path string, template string) bool {
	router := mux.NewRouter()
	router.Path(template).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	req, _ := http.NewRequest("GET", path, nil)
	match := router.Match(req, &mux.RouteMatch{})
	return match
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Path:", r.URL.Path, "Method:", r.Method)

		if isPublicRoute(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Dodavanje timeout-a iz headera u context
		timeoutHeader := r.Header.Get("Timeout")
		log.Println("Timeout apigateway : " + timeoutHeader)
		if timeoutHeader != "" {
			timeoutMs, err := strconv.Atoi(timeoutHeader)
			if err == nil && timeoutMs > 0 {
				ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutMs)*time.Millisecond)
				defer cancel()
				r = r.WithContext(ctx) // Ažuriraj zahtev sa timeout kontekstom
			}
		}

		// Provera Authorization header-a
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := parseJWT(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		role := claims["user_role"].(string)
		if !isAuthorized(role, r.URL.Path, r.Method) {
			http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
			return
		}

		// Dodavanje roli u gRPC metapodatke
		md := metadata.Pairs("user_role", role)
		ctx := metadata.NewOutgoingContext(r.Context(), md)

		// Prosleđivanje sa ažuriranim kontekstom
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func parseJWT(tokenString string) (jwt.MapClaims, error) {
	secret := []byte("matija_AFK") // Replace with your actual secret
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func isAuthorized(role, path, method string) bool {
	// Proveri da li uloga postoji
	allowedMethods, exists := rolePermissions[role]
	if !exists {
		return false
	}

	// Proveri da li postoje dozvoljene putanje za zadatu metodu
	allowedPaths, methodExists := allowedMethods[method]
	if !methodExists {
		return false
	}

	// Proveri da li putanja odgovara nekoj dozvoljenoj putanji
	for _, allowedPath := range allowedPaths {
		if matchesRoute(path, allowedPath) {
			return true
		}
	}

	return false
}

func isPublicRoute(path string) bool {
	for _, route := range publicRoutes {
		if strings.HasPrefix(path, route) {
			return true
		}
	}
	return false
}

func enableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Timeout")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		// Prolazak kroz authMiddleware sa timeout kontekstom
		authMiddleware(h).ServeHTTP(w, r)
	})
}

//func forwardClaimsToServices(ctx context.Context) context.Context {
//	claims := ctx.Value("claims").(jwt.MapClaims)
//	return context.WithValue(ctx, "role", claims["role"])
//}
