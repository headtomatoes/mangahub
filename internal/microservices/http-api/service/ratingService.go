package service

import (
	"context"
	"errors"

	"mangahub/internal/microservices/http-api/dto"
	"mangahub/internal/microservices/http-api/models"
	"mangahub/internal/microservices/http-api/repository"

	"gorm.io/gorm"
)

type RatingService interface {
	CreateOrUpdateRating(userID string, mangaID int64, ratingValue int) (*dto.RatingResponse, error)
	DeleteRating(userID string, mangaID int64) error
	GetUserRating(userID string, mangaID int64) (*dto.UserRatingResponse, error)
	GetMangaRatings(mangaID int64, page, pageSize int) (*dto.PaginatedRatingResponse, error)
	GetMangaAverageRating(mangaID int64) (float64, int64, error)
}

type ratingService struct {
	ratingRepo repository.RatingRepository
	mangaRepo  *repository.MangaRepo
}

func NewRatingService(ratingRepo repository.RatingRepository, mangaRepo *repository.MangaRepo) RatingService {
	return &ratingService{
		ratingRepo: ratingRepo,
		mangaRepo:  mangaRepo,
	}
}

// CreateOrUpdateRating creates or updates a user's rating for a manga and updates the average rating
func (s *ratingService) CreateOrUpdateRating(userID string, mangaID int64, ratingValue int) (*dto.RatingResponse, error) {
	ctx := context.Background()

	// Check if manga exists
	_, err := s.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("manga not found")
		}
		return nil, err
	}

	// Check if rating already exists
	existingRating, err := s.ratingRepo.GetByUserAndManga(userID, mangaID)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var rating *models.Rating

	if existingRating != nil {
		// Update existing rating
		existingRating.Rating = ratingValue
		if err := s.ratingRepo.Update(existingRating); err != nil {
			return nil, err
		}
		rating = existingRating
	} else {
		// Create new rating
		newRating := &models.Rating{
			UserID:  userID,
			MangaID: mangaID,
			Rating:  ratingValue,
		}
		if err := s.ratingRepo.Create(newRating); err != nil {
			return nil, err
		}
		// Reload with user data
		rating, err = s.ratingRepo.GetByUserAndManga(userID, mangaID)
		if err != nil {
			return nil, err
		}
	}

	// Update manga's average rating
	if err := s.updateMangaAverageRating(mangaID); err != nil {
		// Log error but don't fail the request
		// In production, you might want to handle this differently
	}

	return dto.FromModelToRatingResponse(rating), nil
}

// DeleteRating deletes a user's rating and updates the average rating
func (s *ratingService) DeleteRating(userID string, mangaID int64) error {
	ctx := context.Background()

	// Check if manga exists
	_, err := s.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("manga not found")
		}
		return err
	}

	// Delete the rating
	if err := s.ratingRepo.Delete(userID, mangaID); err != nil {
		return err
	}

	// Update manga's average rating
	if err := s.updateMangaAverageRating(mangaID); err != nil {
		// Log error but don't fail the request
	}

	return nil
}

// GetUserRating retrieves a user's rating for a specific manga
func (s *ratingService) GetUserRating(userID string, mangaID int64) (*dto.UserRatingResponse, error) {
	rating, err := s.ratingRepo.GetByUserAndManga(userID, mangaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("rating not found")
		}
		return nil, err
	}

	return &dto.UserRatingResponse{
		Rating:    rating.Rating,
		CreatedAt: rating.CreatedAt,
		UpdatedAt: rating.UpdatedAt,
	}, nil
}

// GetMangaRatings retrieves all ratings for a manga with pagination
func (s *ratingService) GetMangaRatings(mangaID int64, page, pageSize int) (*dto.PaginatedRatingResponse, error) {
	ctx := context.Background()

	// Check if manga exists
	_, err := s.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("manga not found")
		}
		return nil, err
	}

	ratings, total, err := s.ratingRepo.GetByManga(mangaID, page, pageSize)
	if err != nil {
		return nil, err
	}

	ratingResponses := make([]dto.RatingResponse, 0, len(ratings))
	for _, rating := range ratings {
		ratingResponses = append(ratingResponses, *dto.FromModelToRatingResponse(&rating))
	}

	return dto.NewPaginatedRatingResponse(ratingResponses, int(total), page, pageSize), nil
}

// GetMangaAverageRating retrieves the average rating and count for a manga
func (s *ratingService) GetMangaAverageRating(mangaID int64) (float64, int64, error) {
	ctx := context.Background()

	// Check if manga exists
	_, err := s.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, errors.New("manga not found")
		}
		return 0, 0, err
	}

	avg, err := s.ratingRepo.CalculateAverageRating(mangaID)
	if err != nil {
		return 0, 0, err
	}

	count, err := s.ratingRepo.CountRatings(mangaID)
	if err != nil {
		return 0, 0, err
	}

	return avg, count, nil
}

// updateMangaAverageRating updates the average_rating field in the manga table
func (s *ratingService) updateMangaAverageRating(mangaID int64) error {
	ctx := context.Background()

	avg, err := s.ratingRepo.CalculateAverageRating(mangaID)
	if err != nil {
		return err
	}

	manga, err := s.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		return err
	}

	manga.AverageRating = &avg
	return s.mangaRepo.Update(ctx, mangaID, manga)
}
