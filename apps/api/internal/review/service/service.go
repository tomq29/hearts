package service

import (
	"context"
	"errors"

	"github.com/kisssonik/hearts/internal/like/repository"
	"github.com/kisssonik/hearts/internal/review"
	reviewRepo "github.com/kisssonik/hearts/internal/review/repository"
)

var (
	ErrSelfReview    = errors.New("cannot review yourself")
	ErrInvalidRating = errors.New("rating must be between 1 and 5")
	ErrNoMatch       = errors.New("you can only review users you have matched with")
)

type ReviewService interface {
	CreateReview(ctx context.Context, authorID string, input CreateReviewInput) (*review.Review, error)
	GetReviewsForUser(ctx context.Context, targetID string) ([]*review.Review, error)
}

type CreateReviewInput struct {
	TargetID string `json:"targetId"`
	Rating   int    `json:"rating"`
	Comment  string `json:"comment"`
}

type reviewService struct {
	repo     reviewRepo.ReviewRepository
	likeRepo repository.LikeRepository
}

func NewReviewService(repo reviewRepo.ReviewRepository, likeRepo repository.LikeRepository) ReviewService {
	return &reviewService{
		repo:     repo,
		likeRepo: likeRepo,
	}
}

func (s *reviewService) CreateReview(ctx context.Context, authorID string, input CreateReviewInput) (*review.Review, error) {
	if authorID == input.TargetID {
		return nil, ErrSelfReview
	}
	if input.Rating < 1 || input.Rating > 5 {
		return nil, ErrInvalidRating
	}

	// Check if users are matched
	isMatch, err := s.likeRepo.HasMutualLike(ctx, authorID, input.TargetID)
	if err != nil {
		return nil, err
	}
	if !isMatch {
		return nil, ErrNoMatch
	}

	rev := &review.Review{
		AuthorID: authorID,
		TargetID: input.TargetID,
		Rating:   input.Rating,
		Comment:  input.Comment,
	}

	if err := s.repo.Create(ctx, rev); err != nil {
		return nil, err
	}
	return rev, nil
}

func (s *reviewService) GetReviewsForUser(ctx context.Context, targetID string) ([]*review.Review, error) {
	return s.repo.GetByTargetID(ctx, targetID)
}
