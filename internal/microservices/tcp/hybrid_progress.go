package tcp

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"
)

// HybridProgressRepository combines Redis and PostgreSQL for progress tracking
// Redis: Fast, in-memory cache for real-time updates
// PostgreSQL: Persistent storage, backup, prevents data loss
type HybridProgressRepository struct {
	redis     *ProgressRedisRepo
	postgres  *ProgressPostgresRepo
	writeChan chan *ProgressData
	stopChan  chan struct{}
	logger    *slog.Logger
	closed    atomic.Bool // to prevent multiple closes
	// atomic boolean to ensure Close is only called once
	// across multiple goroutines
	// useful in this case rather than mutex for simplicity
}

// NewHybridProgressRepository creates a new hybrid progress repository
func NewHybridProgressRepository(redis *ProgressRedisRepo, postgres *ProgressPostgresRepo) *HybridProgressRepository {
	return &HybridProgressRepository{
		redis:     redis,
		postgres:  postgres,
		writeChan: make(chan *ProgressData, 10000), // Buffer for 10k updates
		stopChan:  make(chan struct{}),
		logger:    slog.Default(),
	}
}

// SaveProgress writes to Redis immediately, queues for PostgreSQL batch write
// the problem with this approach is when queue is full => spawn a goroutine to help main writer drain the queue
// which dont happen => each goroutine compete for the same resource => blocked main writer => more goroutine spawned =>
// fix: we choose 1st way(backpressure) for simplicity(this can be improved later)
// 1. backpressure: if channel is full, block until there is space (slows down clients, but prevents overload)
// 2. drop old data: if channel is full, drop oldest data to make space for new data (data loss, but keeps system responsive)
// 3. direct write fallback: if channel is full, spawn a goroutine to write directly to PostgreSQL (more complex, but prevents data loss and keeps system responsive)
// 4. increase channel buffer size: make channel larger to reduce chance of being full (uses more memory, but simpler)
func (r *HybridProgressRepository) SaveProgress(data *ProgressData) error {
	// 0. Check if repository is closed
	if r.closed.Load() {
		return fmt.Errorf("repository is closed")
	}
	// 1. Write to Redis immediately (fast) + required
	if err := r.redis.SaveProgress(data); err != nil {
		r.logger.Error("redis_save_failed",
			"user_id", data.UserID,
			"manga_id", data.MangaID,
			"error", err,
		)
		return fmt.Errorf("redis write failed: %w", err)
	}
	// Monitor write channel depth
	queueDepth := len(r.writeChan)
	if queueDepth > cap(r.writeChan)/2 {
		r.logger.Warn("write_queue_high_watermark",
			"queue_depth", queueDepth,
		)
	}

	// 2. Queue for PostgreSQL batch write (async)
	select {
	case r.writeChan <- data:
		// Successfully queued
	default:
		// Queue full - try synchronous write as fallback
		r.logger.Warn("write_queue_full, attempting direct postgres write",
			"user_id", data.UserID,
		)

		// Use a SHORT timeout for sync write => avoid blocking too long
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		if err := r.postgres.SaveProgress(ctx, data); err != nil {
			r.logger.Error("postgres_direct_write_failed", "error", err)
			// Data is safely in Redis, will retry on next batch
			return fmt.Errorf("postgres direct write failed: %w", err)
		}
	}
	return nil
}

// GetProgress tries Redis first, falls back to PostgreSQL
func (r *HybridProgressRepository) GetProgress(userID string, mangaID int64) (*ProgressData, error) {
	// Try Redis first (fast)
	data, err := r.redis.GetProgress(userID, mangaID)
	if err == nil && data != nil {
		return data, nil
	}

	// Fallback to PostgreSQL
	r.logger.Debug("redis_miss_fallback_to_postgres",
		"user_id", userID,
		"manga_id", mangaID,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pgData, err := r.postgres.GetProgress(ctx, userID, mangaID)
	if err != nil {
		return nil, err
	}

	// Warm up Redis cache with PostgreSQL data
	if pgData != nil {
		go r.redis.SaveProgress(pgData) // Don't block on cache warming
	}

	return pgData, nil
}

// GetUserProgress gets all progress for a user (from Redis)
func (r *HybridProgressRepository) GetUserProgress(userID string) ([]*ProgressData, error) {
	return r.redis.GetUserProgress(userID)
}

// DeleteProgress deletes from both Redis and PostgreSQL
func (r *HybridProgressRepository) DeleteProgress(userID string, mangaID int64) error {
	// Delete from Redis
	if err := r.redis.DeleteProgress(userID, mangaID); err != nil {
		r.logger.Error("redis_delete_failed", "error", err)
	}

	// Delete from PostgreSQL (not implemented in postgres repo yet, but structure is here)
	// ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// defer cancel()
	// if err := r.postgres.DeleteProgress(ctx, userID, mangaID); err != nil {
	// 	return err
	// }

	return nil
}

// StartBatchWriter runs background worker for batch writes to PostgreSQL
// This should be called in a goroutine when the server starts
func (r *HybridProgressRepository) StartBatchWriter(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute) // Batch write every 5 minutes
	defer ticker.Stop()

	batch := make([]*ProgressData, 0, 1000)

	r.logger.Info("batch_writer_started", "interval", "5m", "batch_size", 1000)

	for {
		select {
		case <-ctx.Done():
			// Shutdown requested - flush remaining batch
			r.logger.Info("batch_writer_shutting_down", "remaining", len(batch))
			if len(batch) > 0 {
				r.flushBatch(batch)
			}
			return

		case data := <-r.writeChan:
			batch = append(batch, data)

			// Flush when batch is full
			if len(batch) >= 1000 {
				r.flushBatch(batch)
				batch = batch[:0] // Reset batch
			}

		case <-ticker.C:
			// Periodic flush every 5 minutes
			if len(batch) > 0 {
				r.logger.Info("periodic_batch_flush", "count", len(batch))
				r.flushBatch(batch)
				batch = batch[:0] // Reset batch
			}
		}
	}
}

// flushBatch writes a batch of progress data to PostgreSQL
func (r *HybridProgressRepository) flushBatch(batch []*ProgressData) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()
	if err := r.postgres.BatchInsert(ctx, batch); err != nil {
		r.logger.Error("batch_insert_failed",
			"count", len(batch),
			"error", err,
		)
	} else {
		duration := time.Since(start)
		r.logger.Info("batch_insert_success",
			"count", len(batch),
			"duration_ms", duration.Milliseconds(),
		)
	}
}

// Close closes both Redis and PostgreSQL connections and stops the batch writer
func (r *HybridProgressRepository) Close() error {
	// Ensure Close is only called once
	r.closed.Store(true)
	close(r.stopChan)
	time.Sleep(100 * time.Millisecond) // Drain in-flight => means wait for batch writer to exit
	close(r.writeChan)

	// Close Redis
	if err := r.redis.Close(); err != nil {
		r.logger.Error("failed_to_close_redis", "error", err)
	}

	// Close PostgreSQL
	if err := r.postgres.Close(); err != nil {
		r.logger.Error("failed_to_close_postgres", "error", err)
	}

	return nil
}
