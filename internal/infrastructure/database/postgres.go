package database

import (
	"database/sql"
	"log"
	"newsletter/config"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// ConnectWithRetry establishes a PostgreSQL connection using the pgx driver.
//
// The function reads the DSN from configuration (DSN env variable) and attempts
// to connect multiple times with a fixed backoff. This is useful in containerized
// or distributed environments where the database may not be immediately available.
//
// On successful connection, a ready-to-use *sql.DB is returned.
// If the database cannot be reached after all retries, the application
// terminates with a fatal log message.
//
// This function belongs to the infrastructure layer and should only be called
// from the application's root.
func InitPostgres() *sql.DB {
	dsn := config.GetEnv("DSN", "")

	for i := 0; i < 10; i++ {
		db, err := sql.Open("pgx", dsn)
		if err == nil && db.Ping() == nil {
			log.Println("Connected to Postgres")
			return db
		}

		log.Println("Postgres not ready, retrying...")
		time.Sleep(2 * time.Second)
	}

	log.Fatal("Could not connect to Postgres")
	return nil
}
