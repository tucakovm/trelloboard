package config

import (
	"fmt"
	"os"
)

type Config struct {
	Address string
}

func GetConfig() Config {

	//if err := godotenv.Load(); err != nil {
	//	log.Println("No .env file found")
	//}

	return Config{
		Address: fmt.Sprintf(":%s", os.Getenv("PROJECTS_SERVICE_PORT")),
	}
}
