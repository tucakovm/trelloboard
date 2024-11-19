package config

import (
	"fmt"
	"os"
)

type Config struct {
	Address               string
	ProjectServicePort    string
	ProjectServiceAddress string
	TaskServiceAddress    string
	TaskServicePort       string
}

func GetConfig() Config {
	return Config{
		ProjectServicePort:    fmt.Sprintf(":%s", os.Getenv("PROJECTS_SERVICE_PORT")),
		ProjectServiceAddress: os.Getenv("PROJECTS_SERVICE_ADDRESS"),
		Address:               fmt.Sprintf(":%s", os.Getenv("GATEWAY_ADDRESS")),
		TaskServicePort:       fmt.Sprintf(":%s", os.Getenv("TASKS_SERVICE_PORT")),
		TaskServiceAddress:    os.Getenv("TASKS_SERVICE_ADDRESS"),
	}
}

func (cfg Config) FullProjectServiceAddress() string {
	return fmt.Sprintf("%s%s", cfg.ProjectServiceAddress, cfg.ProjectServicePort)
}

func (cfg Config) FullTaskServiceAddress() string {
	return fmt.Sprintf("%s%s", cfg.TaskServiceAddress, cfg.TaskServicePort)
}
