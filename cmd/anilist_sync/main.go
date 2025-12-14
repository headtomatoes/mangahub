package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "strconv"
    "syscall"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "mangahub/internal/ingestion/anilist"
)

func main() {
    log.Println("=== AniList Sync Service ===")

    // Load configuration
    config := anilist.SyncConfig{
        DatabaseURL:      getEnv("DATABASE_URL", "postgres://mangahub:mangahub_secret@localhost:5432/mangahub?sslmode=disable"),
        UDPServerURL:     getEnv("UDP_SERVER_URL", "http://localhost:8085"),
        InitialSyncLimit: getEnvInt("ANILIST_SYNC_INITIAL_COUNT", 150),
        WorkerCount:      getEnvInt("ANILIST_SYNC_WORKERS", 10),
        RateConcurrency:  getEnvInt("ANILIST_RATE_CONCURRENCY", 5),
    }

    // Connect to database
    db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    sqlDB, err := db.DB()
    if err != nil {
        log.Fatalf("Failed to get database instance: %v", err)
    }
    defer sqlDB.Close()

    log.Println("✅ Connected to database")

    // Create sync service
    syncService := anilist.NewSyncService(config, db)

    // Setup context with cancellation
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle shutdown signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("\n🛑 Shutdown signal received, stopping sync service...")
        cancel()
    }()

    // Run initial sync
    log.Println("Starting initial sync...")
    if err := syncService.RunInitialSync(ctx); err != nil {
        if err == context.Canceled {
            log.Println("Initial sync cancelled")
        } else {
            log.Printf("❌ Initial sync failed: %v", err)
        }
    }

    // Start pollers
    syncService.StartPollers(ctx)

    // Keep running until interrupted
    log.Println("✅ AniList sync service running. Press Ctrl+C to stop.")
    <-ctx.Done()

    log.Println("👋 AniList sync service stopped")
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}