package service_test

import (
	"context"
	"testing"

	"github.com/kisssonik/hearts/internal/like"
	"github.com/kisssonik/hearts/internal/review"
	"github.com/kisssonik/hearts/internal/review/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockReviewRepository
type MockReviewRepository struct {
	mock.Mock
}

func (m *MockReviewRepository) Create(ctx context.Context, r *review.Review) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockReviewRepository) GetByTargetID(ctx context.Context, targetID string) ([]*review.Review, error) {
	args := m.Called(ctx, targetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*review.Review), args.Error(1)
}

// MockLikeRepository
type MockLikeRepository struct {
	mock.Mock
}

func (m *MockLikeRepository) Upsert(ctx context.Context, l *like.Like) error {
	args := m.Called(ctx, l)
	return args.Error(0)
}

func (m *MockLikeRepository) HasMutualLike(ctx context.Context, user1ID, user2ID string) (bool, error) {
	args := m.Called(ctx, user1ID, user2ID)
	return args.Bool(0), args.Error(1)
}

func (m *MockLikeRepository) GetMatches(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}

func TestReviewService_CreateReview(t *testing.T) {
	mockRepo := new(MockReviewRepository)
	mockLikeRepo := new(MockLikeRepository)
	s := service.NewReviewService(mockRepo, mockLikeRepo)

	ctx := context.Background()
	authorID := "u1"
	targetID := "u2"

	// Case 1: Success
	mockLikeRepo.On("HasMutualLike", ctx, authorID, targetID).Return(true, nil)
	mockRepo.On("Create", ctx, mock.Anything).Return(nil)

	rev, err := s.CreateReview(ctx, authorID, service.CreateReviewInput{
		TargetID: targetID,
		Rating:   5,
		Comment:  "Good",
	})
	assert.NoError(t, err)
	assert.NotNil(t, rev)

	// Case 2: No match
	mockLikeRepo.On("HasMutualLike", ctx, authorID, "u3").Return(false, nil)

	_, err = s.CreateReview(ctx, authorID, service.CreateReviewInput{
		TargetID: "u3",
		Rating:   5,
		Comment:  "Good",
	})
	assert.ErrorIs(t, err, service.ErrNoMatch)

	// Case 3: Self review
	_, err = s.CreateReview(ctx, authorID, service.CreateReviewInput{
		TargetID: authorID,
		Rating:   5,
		Comment:  "Good",
	})
	assert.ErrorIs(t, err, service.ErrSelfReview)
}
