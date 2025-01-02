package config

import (
	"fmt"
	"os"
)

type Config struct {
	Address                 string
	ProjectServicePort      string
	ProjectServiceAddress   string
	TaskServiceAddress      string
	TaskServicePort         string
	UserServiceAddress      string
	UserServicePort         string
	NotServiceAddress       string
	NotServicePort          string
	AnalyticsServiceAddress string
	AnalyticsServicePort    string
}

func GetConfig() Config {
	return Config{
		ProjectServicePort:      fmt.Sprintf(":%s", os.Getenv("PROJECTS_SERVICE_PORT")),
		ProjectServiceAddress:   os.Getenv("PROJECTS_SERVICE_ADDRESS"),
		Address:                 fmt.Sprintf(":%s", os.Getenv("GATEWAY_PORT")),
		TaskServicePort:         fmt.Sprintf(":%s", os.Getenv("TASKS_SERVICE_PORT")),
		TaskServiceAddress:      os.Getenv("TASKS_SERVICE_ADDRESS"),
		UserServicePort:         fmt.Sprintf(":%s", os.Getenv("USER_SERVICE_PORT")),
		UserServiceAddress:      os.Getenv("USER_SERVICE_ADDRESS"),
		NotServicePort:          fmt.Sprintf(":%s", os.Getenv("NOTIFICATIONS_SERVICE_PORT")),
		NotServiceAddress:       os.Getenv("NOTIFICATIONS_SERVICE_ADDRESS"),
		AnalyticsServicePort:    fmt.Sprintf(":%s", os.Getenv("ANALYTICS_SERVICE_PORT")),
		AnalyticsServiceAddress: os.Getenv("ANALYTICS_SERVICE_ADDRESS"),
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
func (cfg Config) FullAnalServiceAddress() string {
	return fmt.Sprintf("%s%s", cfg.AnalyticsServiceAddress, cfg.AnalyticsServicePort)
}

type ErrResp struct {
	URL        string
	Method     string
	StatusCode int
}

func (e ErrResp) Error() string {
	return fmt.Sprintf("error [status code %d] for request: HTTP %s %s", e.StatusCode, e.Method, e.URL)
}
