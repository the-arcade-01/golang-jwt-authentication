package config

import (
	"database/sql"
	"sync"

	"github.com/go-chi/jwtauth/v5"
)

var (
	once      sync.Once
	appConfig *AppConfig
)

type AppConfig struct {
	DB      *sql.DB
	JWTAuth *jwtauth.JWTAuth
}

// NewAppConfig initializes and returns a singleton instance of AppConfig.
// It ensures that the configuration is loaded only once using sync.Once.
// This function sets up the JWT authentication client and the database client.
// If there is an error initializing the database client, it will panic.
// Wherever you need any config variables, use this function call directly as it's a singleton.
func NewAppConfig() *AppConfig {
	once.Do(func() {
		appConfig = &AppConfig{
			JWTAuth: newJWTAuthClient(),
		}
		db, err := newDBClient()
		if err != nil {
			panic(err)
		}
		appConfig.DB = db
	})

	return appConfig
}
