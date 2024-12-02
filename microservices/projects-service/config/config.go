package config

import (
	"fmt"
	"os"
)

type Config struct {
	Address            string
	TaskServicePort    string
	TaskServiceAddress string
	JaegerEndpoint     string
}

func GetConfig() Config {

	return Config{
		Address:            fmt.Sprintf(":%s", os.Getenv("PROJECTS_SERVICE_PORT")),
		TaskServicePort:    fmt.Sprintf(":%s", os.Getenv("TASKS_SERVICE_PORT")),
		TaskServiceAddress: os.Getenv("TASKS_SERVICE_ADDRESS"),
		JaegerEndpoint:     os.Getenv("JAEGER_ENDPOINT"),
	}
}

func (cfg Config) FullTaskServiceAddress() string {
	return fmt.Sprintf("%s%s", cfg.TaskServiceAddress, cfg.TaskServicePort)
}
