package tcp

import (
	"context"
	"database/sql"
	"fmt"
)

// ProgressPostgresRepo handles PostgreSQL operations for progress tracking
type ProgressPostgresRepo struct {
	db *sql.DB
}

// NewProgressPostgresRepo creates a new PostgreSQL progress repository
func NewProgressPostgresRepo(db *sql.DB) *ProgressPostgresRepo {
	return &ProgressPostgresRepo{db: db}
}

// SaveProgress = upsert = saves or updates progress in PostgreSQL
func (r *ProgressPostgresRepo) SaveProgress(ctx context.Context, data *ProgressData) error {
	// Upsert query
	query := `
		INSERT INTO user_progress (user_id, manga_id, current_chapter, status, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, manga_id) 
		DO UPDATE SET 
			current_chapter = EXCLUDED.current_chapter,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		data.UserID,
		data.MangaID,
		data.Chapter,
		data.Status,
		data.LastReadAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save progress to postgres: %w", err)
	}

	return nil
}

// BatchInsert inserts multiple progress records in a single transaction
func (r *ProgressPostgresRepo) BatchInsert(ctx context.Context, batch []*ProgressData) error {
	if len(batch) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO user_progress (user_id, manga_id, current_chapter, status, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, manga_id) 
		DO UPDATE SET 
			current_chapter = EXCLUDED.current_chapter,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, data := range batch {
		_, err = stmt.ExecContext(ctx,
			data.UserID,
			data.MangaID,
			data.Chapter,
			data.Status,
			data.LastReadAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert progress: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetProgress retrieves progress from PostgreSQL
func (r *ProgressPostgresRepo) GetProgress(ctx context.Context, userID string, mangaID int64) (*ProgressData, error) {
	query := `
		SELECT user_id, manga_id, current_chapter, status, updated_at 
		FROM user_progress 
		WHERE user_id = $1 AND manga_id = $2
	`

	var data ProgressData
	err := r.db.QueryRowContext(ctx, query, userID, mangaID).Scan(
		&data.UserID,
		&data.MangaID,
		&data.Chapter,
		&data.Status,
		&data.LastReadAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get progress: %w", err)
	}

	return &data, nil
}

// Close closes the database connection
func (r *ProgressPostgresRepo) Close() error {
	return r.db.Close()
}
