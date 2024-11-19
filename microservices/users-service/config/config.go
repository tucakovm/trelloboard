package config

import (
	"fmt"
	"os"
)

type Config struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	UserAdress   string
	UserPort     string
	SMTPPassword string
}

func LoadConfig() (*Config, error) {
	config := &Config{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		UserAdress:   os.Getenv("USER_SERVICE_ADDRESS"),
		SMTPUser:     os.Getenv("SMTP_USER"),
		UserPort:     fmt.Sprintf(":%s", os.Getenv("USER_SERVICE_PORT")),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
	}
	return config, nil
}
