package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func goDotEnvVariable(key string) string {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Could not load .env file")
	}

	return os.Getenv(key)
}
