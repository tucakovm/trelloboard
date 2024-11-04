package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	Address string
}

func GetConfig() Config {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	return Config{
		Address: ":" + os.Getenv("TASKS_SERVICE_PORT"),
	}
}
