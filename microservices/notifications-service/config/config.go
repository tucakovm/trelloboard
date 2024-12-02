package config

import (
	"fmt"
	"os"
)

type Config struct {
	Address        string
	JaegerEndpoint string
}

func GetConfig() Config {
	return Config{
		Address:        fmt.Sprintf(":%s", os.Getenv("NOTIFICATIONS_SERVICE_PORT")),
		JaegerEndpoint: os.Getenv("JAEGER_ENDPOINT"),
	}
}
