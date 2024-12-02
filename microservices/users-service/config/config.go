package config

import (
	"fmt"
	"os"
)

type Config struct {
	SMTPHost              string
	SMTPPort              string
	SMTPUser              string
	UserAddress           string
	UserPort              string
	SMTPPassword          string
	ProjectServicePort    string
	ProjectServiceAddress string
	JaegerEndpoint        string
}

func LoadConfig() (*Config, error) {
	config := &Config{
		SMTPHost:              os.Getenv("SMTP_HOST"),
		SMTPPort:              os.Getenv("SMTP_PORT"),
		UserAddress:           os.Getenv("USER_SERVICE_ADDRESS"),
		SMTPUser:              os.Getenv("SMTP_USER"),
		UserPort:              fmt.Sprintf(":%s", os.Getenv("USER_SERVICE_PORT")),
		SMTPPassword:          os.Getenv("SMTP_PASSWORD"),
		ProjectServicePort:    fmt.Sprintf(":%s", os.Getenv("PROJECTS_SERVICE_PORT")),
		ProjectServiceAddress: os.Getenv("PROJECTS_SERVICE_ADDRESS"),
		JaegerEndpoint:        os.Getenv("JAEGER_ENDPOINT"),
	}
	return config, nil
}

func (cfg Config) FullProjectServiceAddress() string {
	return fmt.Sprintf("%s%s", cfg.ProjectServiceAddress, cfg.ProjectServicePort)
}
