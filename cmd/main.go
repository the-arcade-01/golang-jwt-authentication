package main

import (
	"net/http"

	"github.com/joho/godotenv"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/api"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/config"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		config.Log.Error("error loading env file", "err", err)
		panic(err)
	}
	_, err = config.ParseEnvs()
	if err != nil {
		panic(err)
	}
}

func main() {
	server := api.NewServer()
	config.Log.Info("server running on port:8080")
	err := http.ListenAndServe(":8080", server.Router)
	if err != nil {
		panic(err)
	}
}
