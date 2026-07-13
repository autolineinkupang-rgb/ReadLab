package service

import (
	"fmt"

	"gorm.io/gorm"
	"readlab/backend/internal/model"
)

type ReviewService struct {
	DB *gorm.DB
}

func NewReviewService(db *gorm.DB) *ReviewService {
	return &ReviewService{DB: db}
}

func (s *ReviewService) CheckReviewGate(novelID uint) error {
	var count int64
	s.DB.Model(&model.Chapter{}).Where("novel_id = ?", novelID).Count(&count)
	if count < 5 {
		return fmt.Errorf("need at least 5 chapters to review")
	}
	return nil
}

func (s *ReviewService) CheckEditLimit(reviewID uint) (int, error) {
	var review model.Review
	if err := s.DB.First(&review, reviewID).Error; err != nil {
		return 0, err
	}
	if review.EditCount >= 5 {
		return int(review.EditCount), fmt.Errorf("edit limit reached")
	}
	return int(review.EditCount), nil
}

func (s *ReviewService) GetUserReview(userID, novelID uint) (*model.Review, error) {
	var review model.Review
	err := s.DB.Where("user_id = ? AND novel_id = ? AND parent_id IS NULL", userID, novelID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (s *ReviewService) CreateReview(userID, novelID uint, rating int, content string) (*model.Review, error) {
	// Unique index on (user_id, novel_id, parent_id) can't catch duplicates here:
	// NULL != NULL in SQL, so it never blocks two top-level (parent_id IS NULL)
	// reviews from the same user. Check explicitly instead.
	if _, err := s.GetUserReview(userID, novelID); err == nil {
		return nil, fmt.Errorf("you have already reviewed this novel")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	review := model.Review{
		UserID:  userID,
		NovelID: novelID,
		Rating:  uint(rating),
		Content: content,
	}
	if err := s.DB.Create(&review).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

func (s *ReviewService) GetNovelReviews(novelID uint, page, limit int) ([]model.Review, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var total int64
	s.DB.Model(&model.Review{}).Where("novel_id = ? AND parent_id IS NULL", novelID).Count(&total)

	var reviews []model.Review
	offset := (page - 1) * limit
	err := s.DB.Preload("User").
		Preload("Replies").
		Preload("Replies.User").
		Where("novel_id = ? AND parent_id IS NULL", novelID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&reviews).Error

	return reviews, total, err
}
