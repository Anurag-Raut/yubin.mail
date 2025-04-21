package main

import (
	"os"

	"github.com/joho/godotenv"
)

func main() {
	env := os.Getenv("APP_ENV")
	envFile := "dev.env"
	if env == "prod" {
		envFile = ".env"
	} else {
		envFile = "dev.env"
	}
	godotenv.Load(envFile)
	cfg := NewConfig()

	cfg.SetAddr(os.Getenv("ADDRESS"))
	cfg.SetPort(os.Getenv("PORT"))
	clientServer := NewServer(cfg)
	clientServer.Listen()
}
