package service

import (
    "context"
    "errors"
    "mangahub/internal/microservices/http-api/models"
    "mangahub/internal/microservices/http-api/repository"
)

var (
    ErrAlreadyInLibrary = errors.New("manga already in library")
    ErrNotInLibrary     = errors.New("manga not in library")
)

type LibraryService interface {
    Add(ctx context.Context, userID string, mangaID int64) error
    Remove(ctx context.Context, userID string, mangaID int64) error
    List(ctx context.Context, userID string) ([]models.UserLibrary, error)
}

type libraryService struct {
    repo      repository.LibraryRepository
    mangaRepo *repository.MangaRepo
}

func NewLibraryService(repo repository.LibraryRepository, mangaRepo *repository.MangaRepo) LibraryService {
    return &libraryService{
        repo:      repo,
        mangaRepo: mangaRepo,
    }
}

func (s *libraryService) Add(ctx context.Context, userID string, mangaID int64) error {
    // Check if manga exists
    if _, err := s.mangaRepo.GetByID(ctx, mangaID); err != nil {
        return errors.New("manga not found")
    }
    
    // Check if already in library
    exists, err := s.repo.Exists(ctx, userID, mangaID)
    if err != nil {
        return err
    }
    if exists {
        return ErrAlreadyInLibrary
    }
    
    return s.repo.Add(ctx, userID, mangaID)
}

func (s *libraryService) Remove(ctx context.Context, userID string, mangaID int64) error {
    return s.repo.Remove(ctx, userID, mangaID)
}

func (s *libraryService) List(ctx context.Context, userID string) ([]models.UserLibrary, error) {
    return s.repo.List(ctx, userID)
}