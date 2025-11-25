package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"mangahub/internal/config"
	"mangahub/internal/microservices/tcp"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Configuration
	// Load config (fallback to env/default)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Build TCP address from config
	// Use 0.0.0.0 to accept connections from outside the container
	tcpAddr := fmt.Sprintf("0.0.0.0:%d", cfg.TCPPort)

	// Parse Redis URL to get address (remove redis:// prefix if present)
	redisAddr := cfg.RedisURL
	redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	redisAddr = strings.TrimPrefix(redisAddr, "rediss://")

	logger.Info("starting_tcp_server",
		"tcp_addr", tcpAddr,
		"redis_addr", redisAddr,
	)

	// Connect to PostgreSQL for hybrid storage
	dbURL := os.Getenv("DATABASE_URL")
	var server *tcp.TCPServer

	if dbURL != "" && dbURL != "none" {
		// Create hybrid server with PostgreSQL + Redis
		logger.Info("initializing_hybrid_storage", "database_url", "***")

		db, err := sql.Open("pgx", dbURL) // Use pgx driver
		if err != nil {
			log.Fatalf("Failed to open database connection: %v", err)
		}
		defer db.Close()

		// Test database connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := db.PingContext(ctx); err != nil {
			cancel()
			log.Fatalf("Failed to ping database: %v", err)
		}
		cancel()

		// Configure connection pool
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(1 * time.Hour)

		logger.Info("database_connected", "max_open_conns", 25)

		// Create server with hybrid storage and authentication
		server = tcp.NewServerWithHybridStorage(tcpAddr, redisAddr, db, cfg.JWTSecret)
	} else {
		// Fallback to Redis-only mode
		logger.Warn("database_url_not_set", "storage_mode", "redis_only")
		server = tcp.NewServer(tcpAddr, redisAddr)
	}

	if server == nil {
		log.Fatal("Failed to create TCP server")
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		logger.Info("received_shutdown_signal")
		server.Stop()
		logger.Info("server_stopped_gracefully")
	case err := <-errChan:
		logger.Error("server_error", "error", err.Error())
		os.Exit(1)
	}
}
