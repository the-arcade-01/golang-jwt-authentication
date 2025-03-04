package config

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/utils"
)

// newDBClient initializes a new database client using the configuration
// specified in the environment variables. It sets the connection
// maximum lifetime, maximum open connections, and maximum idle connections
// based on the environment variables. It also pings the database to ensure
// the connection is established successfully.
//
// Returns a pointer to the sql.DB instance and an error if any occurs
// during the process.
func newDBClient() (*sql.DB, error) {
	db, err := sql.Open(Envs.DB_DRIVER, Envs.DB_URL)
	if err != nil {
		utils.Log.Error("error establishing db conn", "error", err)
		return nil, err
	}

	db.SetConnMaxLifetime(time.Duration(Envs.DB_MAX_CONN_TIME_SEC) * time.Second)
	db.SetMaxOpenConns(Envs.DB_MAX_OPEN_CONN)
	db.SetMaxIdleConns(Envs.DB_MAX_IDLE_CONN)

	if err := db.Ping(); err != nil {
		utils.Log.Error("error pinging db", "error", err)
		return nil, err
	}

	utils.Log.Info("DB connection established")
	return db, nil
}
