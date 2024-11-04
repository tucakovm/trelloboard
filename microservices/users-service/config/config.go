package config

import (
	"fmt"
	"os"
)

type Config struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
}

func LoadConfig() (*Config, error) {
	config := &Config{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUser:     os.Getenv("SMTP_USER"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
	}

	if config.SMTPHost == "" || config.SMTPPort == "" || config.SMTPUser == "" || config.SMTPPassword == "" {
		return nil, fmt.Errorf("missing required SMTP configuration environment variables")
	}

	return config, nil
}
