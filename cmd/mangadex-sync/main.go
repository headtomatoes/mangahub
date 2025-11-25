package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"mangahub/database"
	"mangahub/internal/ingestion/mangadex"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("===========================================")
	log.Println("   MangaDex Sync Service Starting...")
	log.Println("===========================================")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("[Warning] .env file not found, using system environment variables")
	}

	// Initialize database connection
	db, err := database.OpenGorm()
	if err != nil {
		log.Fatalf("[Fatal] Failed to connect to database: %v", err)
	}

	log.Println("[Database] ✅ Connected successfully")

	// Get configuration from environment
	config := mangadex.SyncConfig{
		APIKey:           getEnv("MANGADX_API_KEY", ""),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		UDPServerURL:     getEnv("UDP_SERVER_URL", "http://udp-server:8085"),
		InitialSyncLimit: getEnvInt("MANGA_SYNC_INITIAL_COUNT", 150),
		WorkerCount:      getEnvInt("MANGA_SYNC_WORKERS", 10),
		RateConcurrency:  getEnvInt("MANGA_SYNC_RATE_CONCURRENCY", 5),
	}

	log.Println("[Config] Loaded configuration:")
	log.Printf("  - API Key: %s", maskAPIKey(config.APIKey))
	log.Printf("  - UDP Server: %s", config.UDPServerURL)
	log.Printf("  - Initial Sync Limit: %d", config.InitialSyncLimit)
	log.Printf("  - Worker Count: %d", config.WorkerCount)
	log.Printf("  - Rate Concurrency: %d", config.RateConcurrency)

	// Create sync service
	syncService := mangadex.NewSyncService(config, db)
	log.Println("[SyncService] ✅ Service initialized")

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n[Shutdown] Received shutdown signal, gracefully stopping...")
		cancel()
	}()

	// Check if initial sync should run
	shouldRunInitialSync := getEnvBool("MANGA_SYNC_INITIAL", true)

	if shouldRunInitialSync {
		log.Println("===========================================")
		log.Println("   Running Initial Sync")
		log.Println("===========================================")

		if err := syncService.RunInitialSync(ctx); err != nil {
			log.Printf("[InitialSync] Error: %v", err)
		} else {
			log.Println("[InitialSync] ✅ Initial sync completed successfully!")
		}
	} else {
		log.Println("[InitialSync] Skipped (MANGA_SYNC_INITIAL=false)")
	}

	// Start scheduled pollers
	log.Println("===========================================")
	log.Println("   Starting Scheduled Pollers")
	log.Println("===========================================")
	log.Println("  • New Manga Poll: Every 24 hours")
	log.Println("  • Chapter Updates: Every 48 hours (2 days)")
	log.Println("===========================================")

	syncService.StartPollers(ctx)

	log.Println("[Service] ✅ MangaDex Sync Service is running!")
	log.Println("[Service] Press Ctrl+C to stop")

	// Wait for shutdown signal
	<-ctx.Done()

	log.Println("[Shutdown] Service stopped gracefully")
	log.Println("===========================================")
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if value == "true" || value == "1" || value == "yes" {
			return true
		}
		if value == "false" || value == "0" || value == "no" {
			return false
		}
	}
	return defaultValue
}

func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "(not set)"
	}
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}
