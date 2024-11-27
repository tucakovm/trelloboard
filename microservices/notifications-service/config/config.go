package config

import (
	"fmt"
	"os"
)

type Config struct {
	Address string
}

func GetConfig() Config {

	return Config{
		Address: fmt.Sprintf(":%s", os.Getenv("NOTIFICATIONS_SERVICE_PORT")),
	}
}
