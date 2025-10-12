package database

import (
	"database/sql"
	"fmt"
	"log/slog"

	// use slog for structured logging
	_ "modernc.org/sqlite"
)

// Config struct holds database configuration details (better to expand in future)
// for now it only contains the connection string
type Config struct {
	ConnectionString string
}

func ConnectDB(cfg Config, logger *slog.Logger) (*sql.DB, error) {
	db, err := sql.Open("sqlite", cfg.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify the connection
	if err := db.Ping(); err != nil {
		// close the db handle if ping fails to avoid resource leak
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Info("Connected to the database successfully")
	return db, nil
}
