package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/the-arcade-01/golang-jwt-authentication/internal/utils"
)

var (
	envOnce sync.Once
	Envs    *AppEnvs
)

// AppEnvs holds the environment variables for the application.
type AppEnvs struct {
	ENV                       string
	WEB_URL                   string
	JWT_SECRET_KEY            string
	DB_DRIVER                 string
	DB_URL                    string
	DB_MAX_IDLE_CONN          int
	DB_MAX_OPEN_CONN          int
	DB_MAX_CONN_TIME_SEC      int
	HTTP_COOKIE_HTTPONLY      bool
	HTTP_COOKIE_SECURE        bool
	HTTP_ACCESS_TOKEN_EXPIRE  int
	HTTP_REFRESH_TOKEN_EXPIRE int
}

// ParseEnvs parses the environment variables and stores them in the AppEnvs struct.
// It ensures that the environment variables are parsed only once using sync.Once.
// If any required environment variable is missing or invalid, it returns an error.
// Wherever you want to use environment variables, first parse them using this function
// and then use the Envs variable declared about in vars.
func ParseEnvs() (*AppEnvs, error) {
	var err error
	envOnce.Do(func() {
		Envs = &AppEnvs{
			ENV:            os.Getenv("ENV"),
			WEB_URL:        os.Getenv("WEB_URL"),
			JWT_SECRET_KEY: os.Getenv("JWT_SECRET_KEY"),
			DB_DRIVER:      os.Getenv("DB_DRIVER"),
			DB_URL:         os.Getenv("DB_URL"),
		}

		if Envs.ENV == "" || Envs.WEB_URL == "" || Envs.JWT_SECRET_KEY == "" || Envs.DB_DRIVER == "" || Envs.DB_URL == "" {
			err = fmt.Errorf("invalid env variables in .env file, please check")
			return
		}

		Envs.DB_MAX_IDLE_CONN, err = stringToInt(os.Getenv("DB_MAX_IDLE_CONN"))
		if err != nil || Envs.DB_MAX_IDLE_CONN <= 0 {
			err = fmt.Errorf("invalid DB_MAX_IDLE_CONN value")
			return
		}

		Envs.DB_MAX_OPEN_CONN, err = stringToInt(os.Getenv("DB_MAX_OPEN_CONN"))
		if err != nil || Envs.DB_MAX_OPEN_CONN <= 0 {
			err = fmt.Errorf("invalid DB_MAX_OPEN_CONN value")
			return
		}

		Envs.DB_MAX_CONN_TIME_SEC, err = stringToInt(os.Getenv("DB_MAX_CONN_TIME_SEC"))
		if err != nil || Envs.DB_MAX_CONN_TIME_SEC <= 0 {
			err = fmt.Errorf("invalid DB_MAX_CONN_TIME_SEC value")
			return
		}

		httpCookieHttpOnly, parseErr := strconv.ParseBool(os.Getenv("HTTP_COOKIE_HTTPONLY"))
		if parseErr != nil {
			utils.Log.Error("Error parsing HTTP_COOKIE_HTTPONLY", "err", parseErr)
			err = parseErr
			return
		}
		Envs.HTTP_COOKIE_HTTPONLY = httpCookieHttpOnly

		httpCookieSecure, parseErr := strconv.ParseBool(os.Getenv("HTTP_COOKIE_SECURE"))
		if parseErr != nil {
			utils.Log.Error("Error parsing HTTP_COOKIE_SECURE", "err", parseErr)
			err = parseErr
			return
		}
		Envs.HTTP_COOKIE_SECURE = httpCookieSecure

		httpRefreshTokenExpire, parseErr := strconv.Atoi(os.Getenv("HTTP_REFRESH_TOKEN_EXPIRE"))
		if parseErr != nil {
			utils.Log.Error("Error parsing HTTP_REFRESH_TOKEN_EXPIRE", "err", parseErr)
			err = parseErr
			return
		}
		Envs.HTTP_REFRESH_TOKEN_EXPIRE = httpRefreshTokenExpire

		httpAccessTokenExpire, parseErr := strconv.Atoi(os.Getenv("HTTP_ACCESS_TOKEN_EXPIRE"))
		if parseErr != nil {
			utils.Log.Error("Error parsing HTTP_ACCESS_TOKEN_EXPIRE", "err", parseErr)
			err = parseErr
			return
		}
		Envs.HTTP_ACCESS_TOKEN_EXPIRE = httpAccessTokenExpire
	})
	if err != nil {
		return nil, err
	}
	if Envs.ENV == "development" {
		utils.Log.Info("env variables", "env", Envs)
	}
	utils.Log.Info("envs parsed successfully")

	return Envs, nil
}

// stringToInt converts a string to an integer.
// It returns an error if the string cannot be converted to an integer.
func stringToInt(s string) (int, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return -1, err
	}
	return val, nil
}
