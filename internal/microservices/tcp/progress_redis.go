package tcp

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type ProgressData struct {
	UserID     string    `json:"user_id"`
	MangaID    int64     `json:"manga_id"` // Changed from string to int64 for consistency
	Chapter    int64     `json:"chapter"`  // Changed from int to int64
	LastReadAt time.Time `json:"last_read_at"`
	Status     string    `json:"status"` // "reading", "completed", "on_hold"
}

type ProgressRedisRepo struct {
	client *redis.Client   // Redis client instance
	ctx    context.Context // Context for managing request lifecycle
}

// constructor for ProgressRedisRepo
func NewProgressRedisRepo(redisAddr string) (*ProgressRedisRepo, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     "",
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &ProgressRedisRepo{
		client: rdb,
		ctx:    context.Background(),
	}, nil
}

// Save progress (upsert)
func (r *ProgressRedisRepo) SaveProgress(data *ProgressData) error {
	if r == nil || r.client == nil {
		// No-op for testing/mock mode - return success
		return nil
	}
	key := fmt.Sprintf("progress:user:%s:manga:%d", data.UserID, data.MangaID)

	// Convert struct to a map[string]any for HSET
	fields := map[string]any{
		"user_id":  data.UserID,
		"manga_id": data.MangaID,
		"chapter":  data.Chapter,
		//"page":         data.Page,
		"last_read_at": data.LastReadAt.Format(time.RFC3339Nano),
		"status":       data.Status,
	}

	// Use HSET to set all fields in the hash
	if err := r.client.HSet(r.ctx, key, fields).Err(); err != nil {
		return err
	}

	// Set the expiration on the whole key
	return r.client.Expire(r.ctx, key, 90*24*time.Hour).Err()
}

// Get progress
func (r *ProgressRedisRepo) GetProgress(userID string, mangaID int64) (*ProgressData, error) {
	if r == nil || r.client == nil {
		// No-op for testing/mock mode - return not found
		return nil, nil
	}
	key := fmt.Sprintf("progress:user:%s:manga:%d", userID, mangaID)

	// Use HGetAll to retrieve all fields from hash
	fields, err := r.client.HGetAll(r.ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, nil // Not found
	}

	// Parse fields into struct
	data := &ProgressData{
		UserID:  fields["user_id"],
		MangaID: mangaID,
		Status:  fields["status"],
	}

	// Parse chapter as int64
	if ch, ok := fields["chapter"]; ok {
		fmt.Sscanf(ch, "%d", &data.Chapter)
	}

	// Parse timestamp
	if ts, ok := fields["last_read_at"]; ok {
		data.LastReadAt, _ = time.Parse(time.RFC3339Nano, ts)
	}

	return data, nil
}

// Get all manga progress for a user
func (r *ProgressRedisRepo) GetUserProgress(userID string) ([]*ProgressData, error) {
	if r == nil || r.client == nil {
		// No-op for testing/mock mode - return empty list
		return []*ProgressData{}, nil
	}
	pattern := fmt.Sprintf("progress:user:%s:manga:*", userID)
	var results []*ProgressData
	var cursor uint64

	for {
		// SCAN returns keys in batches without blocking
		keys, nextCursor, err := r.client.Scan(r.ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			// Use HGetAll here if using hashes
			fields, err := r.client.HGetAll(r.ctx, key).Result()
			if err != nil {
				continue
			}
			// Parse fields into ProgressData...
			mangaID, err := strconv.ParseInt(fields["manga_id"], 10, 64)
			if err != nil {
				fmt.Errorf("invalid manga_id in redis for key %s: %v", key, err)
				continue
			}
			data := &ProgressData{
				UserID:  fields["user_id"],
				MangaID: mangaID,
				Status:  fields["status"],
			}
			// Parse chapter as int64
			if ch, ok := fields["chapter"]; ok {
				data.Chapter, _ = strconv.ParseInt(ch, 10, 64)
			}
			// Parse timestamp
			if ts, ok := fields["last_read_at"]; ok {
				data.LastReadAt, _ = time.Parse(time.RFC3339Nano, ts)
			}
			results = append(results, data)
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return results, nil
}

// Delete progress
func (r *ProgressRedisRepo) DeleteProgress(userID string, mangaID int64) error {
	if r == nil || r.client == nil {
		// No-op for testing/mock mode
		return nil
	}
	key := fmt.Sprintf("progress:user:%s:manga:%d", userID, mangaID)
	return r.client.Del(r.ctx, key).Err()
}

// // Atomic increment (for chapter/page tracking) - example for incrementing chapter
// func (r *ProgressRedisRepo) IncrementPage(userID, mangaID string) error {
// 	key := fmt.Sprintf("progress:user:%s:manga:%s", userID, mangaID)
// 	field := "page"
// 	return r.client.HIncrBy(r.ctx, key, field, 1).Err()
// }

func (r *ProgressRedisRepo) Close() error {
	if r == nil || r.client == nil {
		// No-op for testing/mock mode
		return nil
	}
	return r.client.Close()
}
