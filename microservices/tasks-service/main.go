package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	tsk "tasks-service/proto/task"
	"tasks-service/repository"
	"tasks-service/service"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"tasks-service/config"
	h "tasks-service/handlers"
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
	logger := log.New(os.Stdout, "[task-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[task-store] ", log.LstdFlags)

	// NoSQL: Initialize Product Repository store
	repoTask, err := repository.NewTaskRepo(timeoutContext, storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer repoTask.Disconnect(timeoutContext)
	handleErr(err)

	serviceProject := service.NewTaskService(*repoTask)
	handleErr(err)

	handlerProject := h.NewTaskHandler(serviceProject)
	handleErr(err)

	// Bootstrap gRPC server.
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	// Bootstrap gRPC service server and respond to request.
	tsk.RegisterTaskServiceServer(grpcServer, handlerProject)

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
