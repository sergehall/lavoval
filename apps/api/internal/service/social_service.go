package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

type SocialService struct {
	social *repository.SocialRepository
}

func NewSocialService(social *repository.SocialRepository) *SocialService {
	return &SocialService{social: social}
}

// ── Reviews ───────────────────────────────────────────────────────────────────

func (s *SocialService) ListReviews(ctx context.Context, skillID string) ([]domain.SkillReview, error) {
	items, err := s.social.ListReviews(ctx, skillID)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	return items, nil
}

type CreateReviewInput struct {
	Rating     int    `json:"rating" validate:"required,min=1,max=5"`
	ReviewText string `json:"reviewText"`
	RunID      string `json:"runId"`
}

func (s *SocialService) CreateReview(ctx context.Context, skillID, userID string, input CreateReviewInput) (domain.SkillReview, error) {
	rv := domain.SkillReview{
		ID:      uuid.NewString(),
		SkillID: skillID,
		UserID:  userID,
		Rating:  input.Rating,
	}
	if input.ReviewText != "" {
		rv.ReviewText = &input.ReviewText
	}
	if input.RunID != "" {
		rv.RunID = &input.RunID
	}

	created, err := s.social.CreateReview(ctx, rv)
	if err != nil {
		return domain.SkillReview{}, fmt.Errorf("create review: %w", err)
	}
	return created, nil
}

// ── Saves ─────────────────────────────────────────────────────────────────────

func (s *SocialService) SaveSkill(ctx context.Context, userID, skillID string) error {
	if err := s.social.SaveSkill(ctx, userID, skillID); err != nil {
		return fmt.Errorf("save skill: %w", err)
	}
	return nil
}

func (s *SocialService) UnsaveSkill(ctx context.Context, userID, skillID string) error {
	if err := s.social.UnsaveSkill(ctx, userID, skillID); err != nil {
		return fmt.Errorf("unsave skill: %w", err)
	}
	return nil
}

// ── Collections ───────────────────────────────────────────────────────────────

type CollectionInput struct {
	Title       string `json:"title" validate:"required,min=1"`
	Description string `json:"description"`
	Visibility  string `json:"visibility" validate:"required,oneof=public private"`
}

func (s *SocialService) ListCollections(ctx context.Context, ownerID string) ([]domain.Collection, error) {
	return s.social.ListCollections(ctx, ownerID)
}

func (s *SocialService) CreateCollection(ctx context.Context, ownerID string, input CollectionInput) (domain.Collection, error) {
	c := domain.Collection{
		ID:         uuid.NewString(),
		OwnerID:    ownerID,
		Title:      input.Title,
		Visibility: input.Visibility,
	}
	if input.Description != "" {
		c.Description = &input.Description
	}
	return s.social.CreateCollection(ctx, c)
}

func (s *SocialService) UpdateCollection(ctx context.Context, id, ownerID string, input CollectionInput) (domain.Collection, error) {
	c := domain.Collection{
		ID:         id,
		OwnerID:    ownerID,
		Title:      input.Title,
		Visibility: input.Visibility,
	}
	if input.Description != "" {
		c.Description = &input.Description
	}
	return s.social.UpdateCollection(ctx, c)
}

func (s *SocialService) DeleteCollection(ctx context.Context, id, ownerID string) error {
	return s.social.DeleteCollection(ctx, id, ownerID)
}

func (s *SocialService) AddCollectionItem(ctx context.Context, collectionID, skillID, ownerID string) error {
	return s.social.AddCollectionItem(ctx, collectionID, skillID, ownerID)
}

func (s *SocialService) RemoveCollectionItem(ctx context.Context, collectionID, skillID, ownerID string) error {
	return s.social.RemoveCollectionItem(ctx, collectionID, skillID, ownerID)
}

// ── Run feedback ─────────────────────────────────────────────────────────────

type RunFeedbackInput struct {
	Rating          int    `json:"rating" validate:"required,min=1,max=5"`
	UsefulnessScore *int   `json:"usefulnessScore"`
	WouldUseAgain   *bool  `json:"wouldUseAgain"`
	Comment         string `json:"comment"`
}

func (s *SocialService) CreateRunFeedback(ctx context.Context, runID, userID string, input RunFeedbackInput) error {
	fb := domain.RunFeedback{
		RunID:           runID,
		UserID:          userID,
		Rating:          input.Rating,
		UsefulnessScore: input.UsefulnessScore,
		WouldUseAgain:   input.WouldUseAgain,
	}
	if input.Comment != "" {
		fb.Comment = &input.Comment
	}
	return s.social.CreateRunFeedback(ctx, fb)
}
