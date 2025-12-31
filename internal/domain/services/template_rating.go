package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type TemplateRatingService struct {
	db *gorm.DB
}

func NewTemplateRatingService(db *gorm.DB) *TemplateRatingService {
	return &TemplateRatingService{db: db}
}

type RateTemplateInput struct {
	TemplateID uuid.UUID
	UserID     uuid.UUID
	Rating     int
	Review     *string
}

func (s *TemplateRatingService) Rate(ctx context.Context, input RateTemplateInput) (*models.TemplateRating, error) {
	if input.Rating < 1 || input.Rating > 5 {
		return nil, fmt.Errorf("rating must be between 1 and 5")
	}

	var template models.TemplateMarketplace
	if err := s.db.WithContext(ctx).Where("id = ?", input.TemplateID).First(&template).Error; err != nil {
		return nil, fmt.Errorf("template not found")
	}

	var existing models.TemplateRating
	err := s.db.WithContext(ctx).
		Where("template_id = ? AND user_id = ?", input.TemplateID, input.UserID).
		First(&existing).Error

	if err == nil {
		existing.Rating = input.Rating
		existing.Review = input.Review
		if err := s.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return nil, err
		}
		s.updateTemplateRating(ctx, input.TemplateID)
		return &existing, nil
	}

	rating := &models.TemplateRating{
		TemplateID: input.TemplateID,
		UserID:     input.UserID,
		Rating:     input.Rating,
		Review:     input.Review,
	}

	if err := s.db.WithContext(ctx).Create(rating).Error; err != nil {
		return nil, err
	}

	s.updateTemplateRating(ctx, input.TemplateID)
	return rating, nil
}

func (s *TemplateRatingService) updateTemplateRating(ctx context.Context, templateID uuid.UUID) {
	var result struct {
		AvgRating float64
		Count     int
	}

	s.db.WithContext(ctx).
		Model(&models.TemplateRating{}).
		Select("AVG(rating) as avg_rating, COUNT(*) as count").
		Where("template_id = ?", templateID).
		Scan(&result)

	s.db.WithContext(ctx).
		Model(&models.TemplateMarketplace{}).
		Where("id = ?", templateID).
		Updates(map[string]interface{}{
			"rating":       result.AvgRating,
			"rating_count": result.Count,
		})
}

func (s *TemplateRatingService) Delete(ctx context.Context, templateID, userID uuid.UUID) error {
	result := s.db.WithContext(ctx).
		Where("template_id = ? AND user_id = ?", templateID, userID).
		Delete(&models.TemplateRating{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	s.updateTemplateRating(ctx, templateID)
	return nil
}

func (s *TemplateRatingService) GetUserRating(ctx context.Context, templateID, userID uuid.UUID) (*models.TemplateRating, error) {
	var rating models.TemplateRating
	err := s.db.WithContext(ctx).
		Where("template_id = ? AND user_id = ?", templateID, userID).
		First(&rating).Error
	if err != nil {
		return nil, ErrNotFound
	}
	return &rating, nil
}

func (s *TemplateRatingService) ListByTemplate(ctx context.Context, templateID uuid.UUID, limit, offset int) ([]models.TemplateRating, int64, error) {
	var ratings []models.TemplateRating
	var total int64

	query := s.db.WithContext(ctx).Model(&models.TemplateRating{}).Where("template_id = ?", templateID)
	query.Count(&total)

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&ratings).Error
	return ratings, total, err
}

func (s *TemplateRatingService) GetRatingStats(ctx context.Context, templateID uuid.UUID) (map[string]interface{}, error) {
	var stats struct {
		AvgRating float64 `gorm:"column:avg_rating"`
		Total     int64   `gorm:"column:total"`
	}

	err := s.db.WithContext(ctx).
		Model(&models.TemplateRating{}).
		Select("AVG(rating) as avg_rating, COUNT(*) as total").
		Where("template_id = ?", templateID).
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	var distribution []struct {
		Rating int   `gorm:"column:rating"`
		Count  int64 `gorm:"column:count"`
	}

	s.db.WithContext(ctx).
		Model(&models.TemplateRating{}).
		Select("rating, COUNT(*) as count").
		Where("template_id = ?", templateID).
		Group("rating").
		Order("rating DESC").
		Scan(&distribution)

	distMap := make(map[int]int64)
	for i := 1; i <= 5; i++ {
		distMap[i] = 0
	}
	for _, d := range distribution {
		distMap[d.Rating] = d.Count
	}

	return map[string]interface{}{
		"average":      stats.AvgRating,
		"total":        stats.Total,
		"distribution": distMap,
	}, nil
}
