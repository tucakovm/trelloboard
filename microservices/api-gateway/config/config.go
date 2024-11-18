package config

import (
	"fmt"
	"os"
)

type Config struct {
	Address               string
	ProjectServiceAddress string
}

func GetConfig() Config {
	return Config{
		ProjectServiceAddress: fmt.Sprintf(":%s", os.Getenv("PROJECTS_SERVICE_PORT")),
		Address:               fmt.Sprintf(":%s", os.Getenv("GATEWAY_ADDRESS")),
	}
}
