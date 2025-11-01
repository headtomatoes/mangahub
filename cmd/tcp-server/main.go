package main

import (
	"fmt"
	"log"
	"log/slog"
	"mangahub/internal/config"
	"mangahub/internal/microservices/tcp"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
	tcpAddr := fmt.Sprintf("localhost:%d", cfg.TCPPort)

	// Parse Redis URL to get address (remove redis:// prefix if present)
	redisAddr := cfg.RedisURL
	redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	redisAddr = strings.TrimPrefix(redisAddr, "rediss://")

	logger.Info("starting_tcp_server",
		"tcp_addr", tcpAddr,
		"redis_addr", redisAddr,
	)

	// Create and start TCP server
	server := tcp.NewServer(tcpAddr, redisAddr)
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
