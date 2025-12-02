package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	// "time"

	"github.com/google/uuid"

	"mangahub/internal/microservices/http-api/models"
	"mangahub/internal/microservices/http-api/repository"
)

type MangaService interface {
	GetAll(ctx context.Context, page, pageSize int) ([]models.Manga, int64, error)
	GetByID(ctx context.Context, id int64) (*models.Manga, error)
	Create(ctx context.Context, m *models.Manga) error
	Update(ctx context.Context, id int64, m *models.Manga) error
	Delete(ctx context.Context, id int64) error

	// new search method
	SearchByTitle(ctx context.Context, title string) ([]models.Manga, error)

	// manga <-> genres (for handler endpoints)
	ReplaceGenresForManga(ctx context.Context, mangaID int64, genreIDs []int64) error
}

type mangaService struct {
	repo *repository.MangaRepo
}

func NewMangaService(r *repository.MangaRepo) MangaService {
	return &mangaService{repo: r}
}

func (s *mangaService) GetAll(ctx context.Context, page, pageSize int) ([]models.Manga, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.GetAll(ctx, page, pageSize)
}

func (s *mangaService) GetByID(ctx context.Context, id int64) (*models.Manga, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *mangaService) Create(ctx context.Context, m *models.Manga) error {
	// basic validation
	if strings.TrimSpace(m.Title) == "" {
		return errors.New("title is required")
	}

	// ensure slug exists, generate from title if missing
	if m.Slug == nil || strings.TrimSpace(*m.Slug) == "" {
		slug := generateSlug(m.Title)
		// add short uuid suffix to avoid collisions
		slug = fmt.Sprintf("%s-%s", slug, uuid.New().String()[:8])
		m.Slug = &slug
	}

	// business rules can go here (e.g. normalize fields)
	if m.Author != nil {
		a := strings.TrimSpace(*m.Author)
		m.Author = &a
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return err
	}

	// notify UDP server (best-effort, non-blocking)
	go notifyNewManga(m.ID, m.Title)
	return nil
}

func (s *mangaService) Update(ctx context.Context, id int64, m *models.Manga) error {
	// ensure exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// merge minimal validation/business logic
	if strings.TrimSpace(m.Title) == "" && (m.Slug == nil || *m.Slug == "") {
		// allow updates that don't change title/slug but require at least one non-empty field
		// We won't reject if other fields present
	}

	// Track which fields are being updated with old and new values
	type fieldChange struct {
		Field    string      `json:"field"`
		OldValue interface{} `json:"old_value,omitempty"`
		NewValue interface{} `json:"new_value"`
	}
	var changes []string
	var detailedChanges []fieldChange

	// Apply fields that are non-nil / non-zero in m to existing
	if m.Slug != nil && (existing.Slug == nil || *m.Slug != *existing.Slug) {
		oldVal := ""
		if existing.Slug != nil {
			oldVal = *existing.Slug
		}
		detailedChanges = append(detailedChanges, fieldChange{
			Field:    "slug",
			OldValue: oldVal,
			NewValue: *m.Slug,
		})
		existing.Slug = m.Slug
		changes = append(changes, "slug")
	}
	if strings.TrimSpace(m.Title) != "" && m.Title != existing.Title {
		detailedChanges = append(detailedChanges, fieldChange{
			Field:    "title",
			OldValue: existing.Title,
			NewValue: m.Title,
		})
		existing.Title = m.Title
		changes = append(changes, "title")
	}
	if m.Author != nil {
		if existing.Author == nil || *m.Author != *existing.Author {
			oldVal := ""
			if existing.Author != nil {
				oldVal = *existing.Author
			}
			detailedChanges = append(detailedChanges, fieldChange{
				Field:    "author",
				OldValue: oldVal,
				NewValue: *m.Author,
			})
			existing.Author = m.Author
			changes = append(changes, "author")
		}
	}
	if m.Status != nil {
		if existing.Status == nil || *m.Status != *existing.Status {
			oldVal := ""
			if existing.Status != nil {
				oldVal = *existing.Status
			}
			detailedChanges = append(detailedChanges, fieldChange{
				Field:    "status",
				OldValue: oldVal,
				NewValue: *m.Status,
			})
			existing.Status = m.Status
			changes = append(changes, "status")
		}
	}
	if m.TotalChapters != nil {
		if existing.TotalChapters == nil || *m.TotalChapters != *existing.TotalChapters {
			oldVal := 0
			if existing.TotalChapters != nil {
				oldVal = *existing.TotalChapters
			}
			detailedChanges = append(detailedChanges, fieldChange{
				Field:    "total_chapters",
				OldValue: oldVal,
				NewValue: *m.TotalChapters,
			})
			existing.TotalChapters = m.TotalChapters
			changes = append(changes, "total chapters")
		}
	}
	if m.Description != nil {
		if existing.Description == nil || *m.Description != *existing.Description {
			oldVal := ""
			if existing.Description != nil {
				oldVal = *existing.Description
			}
			detailedChanges = append(detailedChanges, fieldChange{
				Field:    "description",
				OldValue: oldVal,
				NewValue: *m.Description,
			})
			existing.Description = m.Description
			changes = append(changes, "description")
		}
	}
	if m.CoverURL != nil {
		if existing.CoverURL == nil || *m.CoverURL != *existing.CoverURL {
			oldVal := ""
			if existing.CoverURL != nil {
				oldVal = *existing.CoverURL
			}
			detailedChanges = append(detailedChanges, fieldChange{
				Field:    "cover_url",
				OldValue: oldVal,
				NewValue: *m.CoverURL,
			})
			existing.CoverURL = m.CoverURL
			changes = append(changes, "cover image")
		}
	}

	// update updated_at business rule could be here

	if err := s.repo.Update(ctx, id, existing); err != nil {
		return err
	}

	// fire a best-effort notification about the update with specific changes
	if len(changes) > 0 {
		// Convert to []interface{} for JSON marshaling
		detailedChangesInterface := make([]interface{}, len(detailedChanges))
		for i, v := range detailedChanges {
			detailedChangesInterface[i] = v
		}
		go notifyMangaUpdateDetailed(id, existing.Title, changes, detailedChangesInterface)
	}
	return nil
}

// notifyNewManga posts to the UDP service HTTP trigger. Non-blocking caller should
// call this in a goroutine.
func notifyNewManga(mangaID int64, title string) {
	url := os.Getenv("UDP_TRIGGER_URL")
	if url == "" {
		url = "http://udp-server:8085/notify/new-manga"
	}
	payload := map[string]interface{}{"manga_id": mangaID, "title": title}
	b, _ := json.Marshal(payload)
	_, _ = http.Post(url, "application/json", bytes.NewReader(b))
}

func notifyMangaUpdate(mangaID int64, title string, changes []string) {
	url := os.Getenv("UDP_TRIGGER_URL")
	if url == "" {
		// call the dedicated manga-update trigger in the UDP server
		url = "http://udp-server:8085/notify/manga-update"
	}
	payload := map[string]interface{}{
		"manga_id": mangaID,
		"title":    title,
		"changes":  changes,
	}
	b, _ := json.Marshal(payload)
	_, _ = http.Post(url, "application/json", bytes.NewReader(b))
}

func notifyMangaUpdateDetailed(mangaID int64, title string, changes []string, detailedChanges []interface{}) {
	url := os.Getenv("UDP_TRIGGER_URL")
	if url == "" {
		// call the dedicated manga-update trigger in the UDP server
		url = "http://udp-server:8085/notify/manga-update"
	}
	payload := map[string]interface{}{
		"manga_id":         mangaID,
		"title":            title,
		"changes":          changes,
		"detailed_changes": detailedChanges,
	}
	b, _ := json.Marshal(payload)
	_, _ = http.Post(url, "application/json", bytes.NewReader(b))
}

func (s *mangaService) Delete(ctx context.Context, id int64) error {
	// potential pre-delete checks (dependencies) could be here
	return s.repo.Delete(ctx, id)
}

// SearchByTitle returns mangas that match title (case-insensitive, partial)
func (s *mangaService) SearchByTitle(ctx context.Context, title string) ([]models.Manga, error) {
	return s.repo.SearchByTitle(ctx, title)
}

func (s *mangaService) ReplaceGenresForManga(ctx context.Context, mangaID int64, genreIDs []int64) error {
	// Validate genre IDs
	for _, id := range genreIDs {
		if id <= 0 {
			return fmt.Errorf("invalid genre id: %d", id)
		}
	}

	// Get current genres
	currentGenres, err := s.repo.GetGenresByManga(ctx, mangaID)
	if err != nil {
		return fmt.Errorf("failed to get current genres: %w", err)
	}

	// Remove all current genres
	if len(currentGenres) > 0 {
		currentGenreIDs := make([]int64, len(currentGenres))
		for i, g := range currentGenres {
			currentGenreIDs[i] = g.ID
		}
		if err := s.repo.RemoveGenresFromManga(ctx, mangaID, currentGenreIDs); err != nil {
			return fmt.Errorf("failed to remove current genres: %w", err)
		}
	}

	// Add new genres
	if len(genreIDs) > 0 {
		if err := s.repo.AddGenresToManga(ctx, mangaID, genreIDs); err != nil {
			return fmt.Errorf("failed to add new genres: %w", err)
		}
	}

	return nil
}

/* helper: generate slug-like string from title */
var nonAlnum = regexp.MustCompile(`[^a-z0-9\-]+`)

func generateSlug(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	// replace spaces with dash
	s = strings.ReplaceAll(s, " ", "-")
	// remove non-alnum/dash
	s = nonAlnum.ReplaceAllString(s, "")
	// collapse dashes
	s = strings.ReplaceAll(s, "--", "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "manga"
	}
	// limit length
	if len(s) > 50 {
		s = s[:50]
	}
	return s
}
