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
