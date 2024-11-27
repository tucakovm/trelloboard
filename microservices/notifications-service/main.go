package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"not_module/config"
	h "not_module/handlers"
	not "not_module/proto/notification"
	"not_module/repositories"
	"not_module/service"
	"os"
	"os/signal"
	"syscall"
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

	log.Println("Not Serices listening on port :" + cfg.Address)

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[notification-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[notification-store] ", log.LstdFlags)

	// NoSQL: Initialize Product Repository store
	store, err := repositories.New(storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.CloseSession()
	store.CreateTables()

	serviceNot := service.NewNotService(*store)

	handlerNots, err := h.NewConnectionHandler(*serviceNot)
	handleErr(err)

	// Bootstrap gRPC server.
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	// Bootstrap gRPC service server and respond to request.
	not.RegisterNotificationServiceServer(grpcServer, handlerNots)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal("server error: ", err)
		}
	}()

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
