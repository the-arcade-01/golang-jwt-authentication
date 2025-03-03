package config

import (
	"fmt"
	"os"
	"sync"
)

var envOnce sync.Once
var Envs *AppEnvs

type AppEnvs struct {
	ENV                  string
	DB_DRIVER            string
	DB_URL               string
	DB_MAX_CONN_TIME_SEC int
	DB_MAX_OPEN_CONN     int
	DB_MAX_IDLE_CONN     int
}

func ParseEnvs() (*AppEnvs, error) {
	var err error
	envOnce.Do(func() {
		Envs = &AppEnvs{
			ENV: os.Getenv("ENV"),
		}
		if Envs.ENV == "" {
			Log.Error("invalid variables parsed, please check .env file")
			err = fmt.Errorf("invalid variables parsed")
			return
		}
	})

	if err != nil {
		return nil, err
	}

	Log.Info("envs loaded successfully", "ENV", Envs.ENV)
	if Envs.ENV == "development" {
		Log.Info("envs values", "envs", Envs)
	}

	return Envs, nil
}
