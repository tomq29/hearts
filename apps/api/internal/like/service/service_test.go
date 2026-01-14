package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/kisssonik/hearts/internal/like"
	"github.com/kisssonik/hearts/internal/like/service"
	"github.com/kisssonik/hearts/internal/notification"
	"github.com/kisssonik/hearts/internal/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

// MockProfileRepository
type MockProfileRepository struct {
	mock.Mock
}

func (m *MockProfileRepository) Create(ctx context.Context, p *profile.Profile) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockProfileRepository) GetByUserID(ctx context.Context, userID string) (*profile.Profile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*profile.Profile), args.Error(1)
}

func (m *MockProfileRepository) GetByUserIDs(ctx context.Context, userIDs []string) ([]*profile.Profile, error) {
	args := m.Called(ctx, userIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*profile.Profile), args.Error(1)
}

func (m *MockProfileRepository) Update(ctx context.Context, p *profile.Profile) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

// MockNotificationService
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) NotifyMatch(ctx context.Context, user1ID, user2ID string) error {
	args := m.Called(ctx, user1ID, user2ID)
	return args.Error(0)
}

func (m *MockNotificationService) GetNotifications(ctx context.Context, userID string) ([]*notification.Notification, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*notification.Notification), args.Error(1)
}

func (m *MockNotificationService) MarkAsRead(ctx context.Context, notificationID string) error {
	args := m.Called(ctx, notificationID)
	return args.Error(0)
}

func TestLikeService_LikeUser(t *testing.T) {
	ctx := context.Background()
	fromID := "u1"
	toID := "u2"

	t.Run("Like, no mutual", func(t *testing.T) {
		mockRepo := new(MockLikeRepository)
		mockProfileRepo := new(MockProfileRepository)
		mockNotifService := new(MockNotificationService)
		s := service.NewLikeService(mockRepo, mockProfileRepo, mockNotifService)

		mockRepo.On("Upsert", ctx, mock.MatchedBy(func(l *like.Like) bool {
			return l.FromUserID == fromID && l.ToUserID == toID && l.IsLike == true
		})).Return(nil)
		// HasMutualLike is called in goroutine, so we use mock.Anything for context
		mockRepo.On("HasMutualLike", mock.Anything, fromID, toID).Return(false, nil)

		isMatch, err := s.LikeUser(ctx, fromID, service.LikeInput{TargetID: toID, IsLike: true})
		assert.NoError(t, err)
		assert.False(t, isMatch) // Always false now

		// Wait for goroutine
		time.Sleep(100 * time.Millisecond)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Like, mutual", func(t *testing.T) {
		mockRepo := new(MockLikeRepository)
		mockProfileRepo := new(MockProfileRepository)
		mockNotifService := new(MockNotificationService)
		s := service.NewLikeService(mockRepo, mockProfileRepo, mockNotifService)

		mockRepo.On("Upsert", ctx, mock.MatchedBy(func(l *like.Like) bool {
			return l.FromUserID == fromID && l.ToUserID == toID && l.IsLike == true
		})).Return(nil)
		mockRepo.On("HasMutualLike", mock.Anything, fromID, toID).Return(true, nil)
		mockNotifService.On("NotifyMatch", mock.Anything, fromID, toID).Return(nil)

		isMatch, err := s.LikeUser(ctx, fromID, service.LikeInput{TargetID: toID, IsLike: true})
		assert.NoError(t, err)
		assert.False(t, isMatch) // Always false now

		// Wait for goroutine
		time.Sleep(100 * time.Millisecond)
		mockRepo.AssertExpectations(t)
		mockNotifService.AssertExpectations(t)
	})

	t.Run("Pass", func(t *testing.T) {
		mockRepo := new(MockLikeRepository)
		mockProfileRepo := new(MockProfileRepository)
		mockNotifService := new(MockNotificationService)
		s := service.NewLikeService(mockRepo, mockProfileRepo, mockNotifService)

		mockRepo.On("Upsert", ctx, mock.MatchedBy(func(l *like.Like) bool {
			return l.FromUserID == fromID && l.ToUserID == toID && l.IsLike == false
		})).Return(nil)

		isMatch, err := s.LikeUser(ctx, fromID, service.LikeInput{TargetID: toID, IsLike: false})
		assert.NoError(t, err)
		assert.False(t, isMatch)
		mockRepo.AssertExpectations(t)
	})
}

func TestLikeService_GetMatches(t *testing.T) {
	mockRepo := new(MockLikeRepository)
	mockProfileRepo := new(MockProfileRepository)
	mockNotifService := new(MockNotificationService)
	s := service.NewLikeService(mockRepo, mockProfileRepo, mockNotifService)

	ctx := context.Background()
	userID := "u1"
	matchIDs := []string{"u2", "u3"}
	profiles := []*profile.Profile{
		{UserID: "u2", FirstName: "User 2"},
		{UserID: "u3", FirstName: "User 3"},
	}

	mockRepo.On("GetMatches", ctx, userID).Return(matchIDs, nil)
	mockProfileRepo.On("GetByUserIDs", ctx, matchIDs).Return(profiles, nil)

	result, err := s.GetMatches(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "User 2", result[0].FirstName)
}
