package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *pgxpool.Pool

func Connect() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" || dbURL == "none" {
		log.Println("DATABASE_URL is empty or 'none'")
		DB = nil
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	fmt.Println("✅ Connected to PostgreSQL database successfully!")
	DB = pool
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
		fmt.Println("🔒 Database connection closed.")
	}
}

// OpenGorm opens a gorm.DB from DATABASE_URL env and configures connection pool.
func OpenGorm() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" || dsn == "none" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}

	gormLogger := logger.Default.LogMode(logger.Silent)
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(1 * time.Hour)

	return gdb, nil
}
