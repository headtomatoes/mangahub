// package database

// import (
// 	"database/sql"
// 	"fmt"
// 	"log/slog" // use slog for structured logging
// 	"mangahub/internal/config"

// 	"github.com/golang-migrate/migrate/v4"
// 	"github.com/golang-migrate/migrate/v4/database/sqlite3"
// 	_ "modernc.org/sqlite"
// )

// func ConnectDB(cfg *config.Config, logger *slog.Logger) (*sql.DB, error) {
// 	db, err := sql.Open("sqlite", cfg.DatabaseURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open database: %w", err)
// 	}

// 	// Verify the connection
// 	if err := db.Ping(); err != nil {
// 		// close the db handle if ping fails to avoid resource leak
// 		db.Close()
// 		return nil, fmt.Errorf("failed to connect to database: %w", err)
// 	}

// 	// Run migrations
// 	if err := runMigrations(db, logger); err != nil {
// 		db.Close()
// 		return nil, fmt.Errorf("failed to run migrations: %w", err)
// 	}

// 	logger.Info("Connected to the database successfully")
// 	return db, nil
// }

// func runMigrations(db *sql.DB, logger *slog.Logger) error {
// 	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
// 	if err != nil {
// 		return fmt.Errorf("failed to create migration driver: %w", err)
// 	}

// 	m, err := migrate.NewWithDatabaseInstance(
// 		"file://database/migrations",
// 		"sqlite3",
// 		driver,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("failed to create migrate instance: %w", err)
// 	}

// 	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
// 		return fmt.Errorf("failed to apply migrations: %w", err)
// 	}

// 	logger.Info("Database migrations applied successfully")
// 	return nil
// }

package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect() {
	dbURL := os.Getenv("DATABASE_URL")
	// if dbURL == "" {
	// 	// Fallback to default local settings
	// 	dbURL = "postgres://mangahub:secret123@localhost:5432/mangahub?sslmode=disable"
	// }

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	if err = pool.Ping(ctx); err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}

	fmt.Println("✅ Connected to PostgreSQL database successfully!")
	DB = pool
}

func Close() {
	if DB != nil {
		DB.Close()
		fmt.Println("🔒 Database connection closed.")
	}
}
