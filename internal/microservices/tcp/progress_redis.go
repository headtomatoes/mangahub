package tcp

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type ProgressData struct {
	UserID  string `json:"user_id"`
	MangaID string `json:"manga_id"`
	Chapter int    `json:"chapter"`
	//Page       int       `json:"page"` // current page not used but can be useful for tracking
	LastReadAt time.Time `json:"last_read_at"`
	Status     string    `json:"status"` // "reading", "completed", "on_hold"
}

type ProgressRepository struct {
	client *redis.Client   // Redis client instance
	ctx    context.Context // Context for managing request lifecycle
}

// constructor for ProgressRepository
func NewProgressRepository(redisAddr string) (*ProgressRepository, error) {
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

	return &ProgressRepository{
		client: rdb,
		ctx:    context.Background(),
	}, nil
}

// Save progress (upsert)
func (r *ProgressRepository) SaveProgress(data *ProgressData) error {
	key := fmt.Sprintf("progress:user:%s:manga:%s", data.UserID, data.MangaID)

	// Convert struct to a map[string]interface{} for HSET
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
func (r *ProgressRepository) GetProgress(userID, mangaID string) (*ProgressData, error) {
	key := fmt.Sprintf("progress:user:%s:manga:%s", userID, mangaID)

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
		MangaID: fields["manga_id"],
		Status:  fields["status"],
	}

	// Parse chapter as int
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
func (r *ProgressRepository) GetUserProgress(userID string) ([]*ProgressData, error) {
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
			data := &ProgressData{
				UserID:  fields["user_id"],
				MangaID: fields["manga_id"],
				Status:  fields["status"],
			}
			// Parse chapter as int
			if ch, ok := fields["chapter"]; ok {
				fmt.Sscanf(ch, "%d", &data.Chapter)
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
func (r *ProgressRepository) DeleteProgress(userID, mangaID string) error {
	key := fmt.Sprintf("progress:user:%s:manga:%s", userID, mangaID)
	return r.client.Del(r.ctx, key).Err()
}

// // Atomic increment (for chapter/page tracking) - example for incrementing chapter
// func (r *ProgressRepository) IncrementPage(userID, mangaID string) error {
// 	key := fmt.Sprintf("progress:user:%s:manga:%s", userID, mangaID)
// 	field := "page"
// 	return r.client.HIncrBy(r.ctx, key, field, 1).Err()
// }

func (r *ProgressRepository) Close() error {
	return r.client.Close()
}
