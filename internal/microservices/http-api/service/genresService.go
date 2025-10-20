package service

import (
    "context"
    "errors"
    "strings"

    "mangahub/internal/microservices/http-api/models"
    "mangahub/internal/microservices/http-api/repository"
)

type GenreService interface {
    GetAll(ctx context.Context) ([]models.Genre, error)
    Create(ctx context.Context, g *models.Genre) error
}

type genreService struct {
    repo *repository.GenreRepo
}

func NewGenreService(r *repository.GenreRepo) GenreService {
    return &genreService{repo: r}
}

func (s *genreService) GetAll(ctx context.Context) ([]models.Genre, error) {
    return s.repo.GetAll(ctx)
}

func (s *genreService) Create(ctx context.Context, g *models.Genre) error {
    if strings.TrimSpace(g.Name) == "" {
        return errors.New("genre name required")
    }
    g.Name = strings.TrimSpace(g.Name)
    return s.repo.Create(ctx, g)
}