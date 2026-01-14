package service

import (
	"context"
	"errors"

	"github.com/kisssonik/hearts/internal/like"
	"github.com/kisssonik/hearts/internal/like/repository"
	"github.com/kisssonik/hearts/internal/notification/service"
	"github.com/kisssonik/hearts/internal/profile"
	profileRepo "github.com/kisssonik/hearts/internal/profile/repository"
	"github.com/kisssonik/hearts/pkg/queue"
)

var ErrSelfLike = errors.New("cannot like yourself")

type LikeService interface {
	LikeUser(ctx context.Context, fromUserID string, input LikeInput) (bool, error)
	GetMatches(ctx context.Context, userID string) ([]*profile.Profile, error)
	ProcessMatchCheck(ctx context.Context, fromUserID, targetID string) error
}

type LikeInput struct {
	TargetID string `json:"targetId"`
	IsLike   bool   `json:"isLike"`
}

type MatchCheckMessage struct {
	FromUserID string `json:"fromUserId"`
	TargetID   string `json:"targetId"`
}

type likeService struct {
	repo                repository.LikeRepository
	profileRepo         profileRepo.ProfileRepository
	notificationService service.NotificationService
	producer            queue.Producer
}

func NewLikeService(repo repository.LikeRepository, profileRepo profileRepo.ProfileRepository, notificationService service.NotificationService, producer queue.Producer) LikeService {
	return &likeService{
		repo:                repo,
		profileRepo:         profileRepo,
		notificationService: notificationService,
		producer:            producer,
	}
}

func (s *likeService) LikeUser(ctx context.Context, fromUserID string, input LikeInput) (bool, error) {
	if fromUserID == input.TargetID {
		return false, ErrSelfLike
	}

	l := &like.Like{
		FromUserID: fromUserID,
		ToUserID:   input.TargetID,
		IsLike:     input.IsLike,
	}

	if err := s.repo.Upsert(ctx, l); err != nil {
		return false, err
	}

	// If it's a pass (IsLike = false), it can't be a match.
	if !input.IsLike {
		return false, nil
	}

	// Publish match check to Kafka
	msg := MatchCheckMessage{
		FromUserID: fromUserID,
		TargetID:   input.TargetID,
	}
	if err := s.producer.Publish(ctx, msg); err != nil {
		// Log error but don't fail the request
		// In production, you might want to fallback to sync check or retry
		return false, nil
	}

	return false, nil
}

func (s *likeService) ProcessMatchCheck(ctx context.Context, fromUserID, targetID string) error {
	isMatch, err := s.repo.HasMutualLike(ctx, fromUserID, targetID)
	if err != nil {
		return err
	}
	if isMatch {
		return s.notificationService.NotifyMatch(ctx, fromUserID, targetID)
	}
	return nil
}

func (s *likeService) GetMatches(ctx context.Context, userID string) ([]*profile.Profile, error) {
	matchIDs, err := s.repo.GetMatches(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(matchIDs) == 0 {
		return []*profile.Profile{}, nil
	}

	return s.profileRepo.GetByUserIDs(ctx, matchIDs)
}
