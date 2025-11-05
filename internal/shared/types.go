package shared

import "time"

// shared types across the application
// 1st: progress data structure for manga reading progress TCP
// 2nd: auth claims structure for JWT authentication in HTTP API
// 3rd: add more shared types as needed

type Progress struct {
	UserID     string    `json:"user_id" db:"user_id"`           // user identifier(UUID)
	MangaID    int64     `json:"manga_id" db:"manga_id"`         // manga identifier (int)
	Chapter    int       `json:"chapter" db:"chapter"`           // current chapter
	Status     string    `json:"status" db:"status"`             // reading status (e.g., "reading", "completed")
	LastReadAt time.Time `json:"last_read_at" db:"last_read_at"` // timestamp of last read action
}

type AuthClaims struct {
	UserID   string `json:"user_id" db:"user_id"`    // user identifier(UUID)
	UserName string `json:"username" db:"user_name"` // username
}
