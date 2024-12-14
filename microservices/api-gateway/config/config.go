package config

import (
	"fmt"
	"os"
)

type Config struct {
	Address                string
	ProjectServicePort     string
	ProjectServiceAddress  string
	TaskServiceAddress     string
	TaskServicePort        string
	UserServiceAddress     string
	UserServicePort        string
	NotServiceAddress      string
	NotServicePort         string
	WorkflowServiceAddress string
	WorkflowServicePort    string
}

func GetConfig() Config {
	return Config{
		ProjectServicePort:     fmt.Sprintf(":%s", os.Getenv("PROJECTS_SERVICE_PORT")),
		ProjectServiceAddress:  os.Getenv("PROJECTS_SERVICE_ADDRESS"),
		Address:                fmt.Sprintf(":%s", os.Getenv("GATEWAY_PORT")),
		TaskServicePort:        fmt.Sprintf(":%s", os.Getenv("TASKS_SERVICE_PORT")),
		TaskServiceAddress:     os.Getenv("TASKS_SERVICE_ADDRESS"),
		UserServicePort:        fmt.Sprintf(":%s", os.Getenv("USER_SERVICE_PORT")),
		UserServiceAddress:     os.Getenv("USER_SERVICE_ADDRESS"),
		NotServicePort:         fmt.Sprintf(":%s", os.Getenv("NOTIFICATIONS_SERVICE_PORT")),
		NotServiceAddress:      os.Getenv("NOTIFICATIONS_SERVICE_ADDRESS"),
		WorkflowServicePort:    fmt.Sprintf(":%s", os.Getenv("WORKFLOW_SERVICE_PORT")),
		WorkflowServiceAddress: os.Getenv("WORKFLOW_SERVICE_ADDRESS"),
	}
}

func (cfg Config) FullProjectServiceAddress() string {
	return fmt.Sprintf("%s%s", cfg.ProjectServiceAddress, cfg.ProjectServicePort)
}

func (cfg Config) FullTaskServiceAddress() string {
	return fmt.Sprintf("%s%s", cfg.TaskServiceAddress, cfg.TaskServicePort)
}

func (cfg Config) FullUserServiceAddress() string {
	return fmt.Sprintf("%s%s", cfg.UserServiceAddress, cfg.UserServicePort)
}

func (cfg Config) FullNotServiceAddress() string {
	return fmt.Sprintf("%s%s", cfg.NotServiceAddress, cfg.NotServicePort)
}
func (cfg Config) FullWorkflowServiceAddress() string {
	return fmt.Sprintf("%s%s", cfg.WorkflowServiceAddress, cfg.WorkflowServicePort)
}
