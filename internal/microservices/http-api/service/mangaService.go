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
	GetAll(ctx context.Context) ([]models.Manga, error)
	GetByID(ctx context.Context, id int64) (*models.Manga, error)
	Create(ctx context.Context, m *models.Manga) error
	Update(ctx context.Context, id int64, m *models.Manga) error
	Delete(ctx context.Context, id int64) error

	// new search method
	SearchByTitle(ctx context.Context, title string) ([]models.Manga, error)

	// manga <-> genres (for handler endpoints)
	GetGenresByManga(ctx context.Context, mangaID int64) ([]models.Genre, error)
	AddGenresToManga(ctx context.Context, mangaID int64, genreIDs []int64) error
	RemoveGenresFromManga(ctx context.Context, mangaID int64, genreIDs []int64) error
}

type mangaService struct {
	repo *repository.MangaRepo
}

func NewMangaService(r *repository.MangaRepo) MangaService {
	return &mangaService{repo: r}
}

func (s *mangaService) GetAll(ctx context.Context) ([]models.Manga, error) {
	return s.repo.GetAll(ctx)
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

	// Apply fields that are non-nil / non-zero in m to existing
	if m.Slug != nil {
		existing.Slug = m.Slug
	}
	if strings.TrimSpace(m.Title) != "" {
		existing.Title = m.Title
	}
	if m.Author != nil {
		existing.Author = m.Author
	}
	if m.Status != nil {
		existing.Status = m.Status
	}
	if m.TotalChapters != nil {
		existing.TotalChapters = m.TotalChapters
	}
	if m.Description != nil {
		existing.Description = m.Description
	}
	if m.CoverURL != nil {
		existing.CoverURL = m.CoverURL
	}

	// update updated_at business rule could be here

	if err := s.repo.Update(ctx, id, existing); err != nil {
		return err
	}

	// fire a best-effort notification about the update
	go notifyMangaUpdate(id, existing.Title)
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

func notifyMangaUpdate(mangaID int64, title string) {
	url := os.Getenv("UDP_TRIGGER_URL")
	if url == "" {
		// call the dedicated manga-update trigger in the UDP server
		url = "http://udp-server:8085/notify/manga-update"
	}
	payload := map[string]interface{}{"manga_id": mangaID, "title": title, "message": "Manga updated"}
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

func (s *mangaService) GetGenresByManga(ctx context.Context, mangaID int64) ([]models.Genre, error) {
	return s.repo.GetGenresByManga(ctx, mangaID)
}

func (s *mangaService) AddGenresToManga(ctx context.Context, mangaID int64, genreIDs []int64) error {
	if len(genreIDs) == 0 {
		return nil
	}
	for _, id := range genreIDs {
		if id <= 0 {
			return fmt.Errorf("invalid genre id: %d", id)
		}
	}
	return s.repo.AddGenresToManga(ctx, mangaID, genreIDs)
}

func (s *mangaService) RemoveGenresFromManga(ctx context.Context, mangaID int64, genreIDs []int64) error {
	if len(genreIDs) == 0 {
		return nil
	}
	for _, id := range genreIDs {
		if id <= 0 {
			return fmt.Errorf("invalid genre id: %d", id)
		}
	}
	return s.repo.RemoveGenresFromManga(ctx, mangaID, genreIDs)
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
