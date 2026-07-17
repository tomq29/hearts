package service_test

import (
	"context"
	"testing"

	"github.com/kisssonik/hearts/internal/like"
	"github.com/kisssonik/hearts/internal/like/service"
	"github.com/kisssonik/hearts/internal/notification"
	"github.com/kisssonik/hearts/internal/profile"
	profileRepo "github.com/kisssonik/hearts/internal/profile/repository"
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

func (m *MockProfileRepository) Search(ctx context.Context, currentUser *profile.Profile, params profileRepo.SearchParams) ([]*profile.Profile, error) {
	args := m.Called(ctx, currentUser, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*profile.Profile), args.Error(1)
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

// MockProducer
type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) Publish(ctx context.Context, message interface{}) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestLikeService_LikeUser(t *testing.T) {
	ctx := context.Background()
	fromID := "u1"
	toID := "u2"

	t.Run("Like publishes match check", func(t *testing.T) {
		mockRepo := new(MockLikeRepository)
		mockProfileRepo := new(MockProfileRepository)
		mockNotifService := new(MockNotificationService)
		mockProducer := new(MockProducer)
		s := service.NewLikeService(mockRepo, mockProfileRepo, mockNotifService, mockProducer)

		mockRepo.On("Upsert", ctx, mock.MatchedBy(func(l *like.Like) bool {
			return l.FromUserID == fromID && l.ToUserID == toID && l.IsLike == true
		})).Return(nil)
		mockProducer.On("Publish", ctx, service.MatchCheckMessage{
			FromUserID: fromID,
			TargetID:   toID,
		}).Return(nil)

		isMatch, err := s.LikeUser(ctx, fromID, service.LikeInput{TargetID: toID, IsLike: true})
		assert.NoError(t, err)
		assert.False(t, isMatch) // Match detection is async via Kafka

		mockRepo.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})

	t.Run("Pass does not publish", func(t *testing.T) {
		mockRepo := new(MockLikeRepository)
		mockProfileRepo := new(MockProfileRepository)
		mockNotifService := new(MockNotificationService)
		mockProducer := new(MockProducer)
		s := service.NewLikeService(mockRepo, mockProfileRepo, mockNotifService, mockProducer)

		mockRepo.On("Upsert", ctx, mock.MatchedBy(func(l *like.Like) bool {
			return l.FromUserID == fromID && l.ToUserID == toID && l.IsLike == false
		})).Return(nil)

		isMatch, err := s.LikeUser(ctx, fromID, service.LikeInput{TargetID: toID, IsLike: false})
		assert.NoError(t, err)
		assert.False(t, isMatch)
		mockRepo.AssertExpectations(t)
		mockProducer.AssertNotCalled(t, "Publish")
	})

	t.Run("Self-like is rejected", func(t *testing.T) {
		mockRepo := new(MockLikeRepository)
		mockProfileRepo := new(MockProfileRepository)
		mockNotifService := new(MockNotificationService)
		mockProducer := new(MockProducer)
		s := service.NewLikeService(mockRepo, mockProfileRepo, mockNotifService, mockProducer)

		_, err := s.LikeUser(ctx, fromID, service.LikeInput{TargetID: fromID, IsLike: true})
		assert.ErrorIs(t, err, service.ErrSelfLike)
		mockRepo.AssertNotCalled(t, "Upsert")
	})
}

func TestLikeService_ProcessMatchCheck(t *testing.T) {
	ctx := context.Background()
	fromID := "u1"
	toID := "u2"

	t.Run("Mutual like notifies", func(t *testing.T) {
		mockRepo := new(MockLikeRepository)
		mockProfileRepo := new(MockProfileRepository)
		mockNotifService := new(MockNotificationService)
		mockProducer := new(MockProducer)
		s := service.NewLikeService(mockRepo, mockProfileRepo, mockNotifService, mockProducer)

		mockRepo.On("HasMutualLike", ctx, fromID, toID).Return(true, nil)
		mockNotifService.On("NotifyMatch", ctx, fromID, toID).Return(nil)

		err := s.ProcessMatchCheck(ctx, fromID, toID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockNotifService.AssertExpectations(t)
	})

	t.Run("No mutual like, no notification", func(t *testing.T) {
		mockRepo := new(MockLikeRepository)
		mockProfileRepo := new(MockProfileRepository)
		mockNotifService := new(MockNotificationService)
		mockProducer := new(MockProducer)
		s := service.NewLikeService(mockRepo, mockProfileRepo, mockNotifService, mockProducer)

		mockRepo.On("HasMutualLike", ctx, fromID, toID).Return(false, nil)

		err := s.ProcessMatchCheck(ctx, fromID, toID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockNotifService.AssertNotCalled(t, "NotifyMatch")
	})
}

func TestLikeService_GetMatches(t *testing.T) {
	mockRepo := new(MockLikeRepository)
	mockProfileRepo := new(MockProfileRepository)
	mockNotifService := new(MockNotificationService)
	mockProducer := new(MockProducer)
	s := service.NewLikeService(mockRepo, mockProfileRepo, mockNotifService, mockProducer)

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
