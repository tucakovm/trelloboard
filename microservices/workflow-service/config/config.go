package config

import (
	"fmt"
	"os"
)

type Config struct {
	Address               string
	UserServicePort       string
	UserServiceAddress    string
	ProjectServicePort    string
	ProjectServiceAddress string
	JaegerEndpoint        string
	WorkflowServicePort   string
	TaskServiceAddress    string
	TaskServicePort       string
}

func LoadConfig() Config {

	return Config{
		Address:               fmt.Sprintf(":%s", os.Getenv("TASKS_SERVICE_PORT")),
		UserServicePort:       fmt.Sprintf(":%s", os.Getenv("USER_SERVICE_PORT")),
		WorkflowServicePort:   fmt.Sprintf(":%s", os.Getenv("WORKFLOW_SERVICE_PORT")),
		UserServiceAddress:    os.Getenv("USER_SERVICE_ADDRESS"),
		ProjectServicePort:    fmt.Sprintf(":%s", os.Getenv("PROJECTS_SERVICE_PORT")),
		ProjectServiceAddress: os.Getenv("PROJECTS_SERVICE_ADDRESS"),
		JaegerEndpoint:        os.Getenv("JAEGER_ENDPOINT"),
		TaskServiceAddress:    os.Getenv("TASKS_SERVICE_ADDRESS"),
		TaskServicePort:       fmt.Sprintf(":%s", os.Getenv("TASKS_SERVICE_PORT")),
	}
}

func (cfg Config) FullUserServiceAddress() string {
	return fmt.Sprintf("%s%s", cfg.UserServiceAddress, cfg.UserServicePort)
}

func (cfg Config) FullProjectServiceAddress() string {
	return fmt.Sprintf("%s%s", cfg.ProjectServiceAddress, cfg.ProjectServicePort)
}

func (cfg Config) FullTaskServiceAddress() string {
	return fmt.Sprintf("%s%s", cfg.TaskServiceAddress, cfg.TaskServicePort)
}
