package main

import (
	"net/http"

	"github.com/joho/godotenv"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/api"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/config"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/utils"
)

// init loads environment variables from a .env file and parses them.
func init() {
	err := godotenv.Load()
	if err != nil {
		utils.Log.Error("error loading env file", "err", err)
		panic(err)
	}
	_, err = config.ParseEnvs()
	if err != nil {
		panic(err)
	}
}

// main initializes the server and starts listening on port 8080.
func main() {
	server := api.NewServer()
	utils.Log.Info("server running on port:8080")
	err := http.ListenAndServe(":8080", server.Router)
	if err != nil {
		panic(err)
	}
}
