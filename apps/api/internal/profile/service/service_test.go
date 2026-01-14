package service_test

import (
	"context"
	"testing"

	"github.com/kisssonik/hearts/internal/profile"
	"github.com/kisssonik/hearts/internal/profile/repository"
	"github.com/kisssonik/hearts/internal/profile/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProfileRepository is a mock implementation of repository.ProfileRepository
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

func TestCreateProfile_Success(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	s := service.NewProfileService(mockRepo)
	ctx := context.Background()
	userID := "user-123"

	input := service.CreateProfileInput{
		FirstName: "John",
		Bio:       "Hello world",
	}

	// Expect GetByUserID to return ErrNotFound (profile shouldn't exist)
	mockRepo.On("GetByUserID", ctx, userID).Return(nil, repository.ErrNotFound)

	// Expect Create to be called
	mockRepo.On("Create", ctx, mock.MatchedBy(func(p *profile.Profile) bool {
		return p.UserID == userID && p.FirstName == "John" && p.Bio == "Hello world"
	})).Return(nil)

	p, err := s.CreateProfile(ctx, userID, input)

	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "John", p.FirstName)
	mockRepo.AssertExpectations(t)
}

func TestCreateProfile_AlreadyExists(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	s := service.NewProfileService(mockRepo)
	ctx := context.Background()
	userID := "user-123"

	input := service.CreateProfileInput{FirstName: "John"}

	// Expect GetByUserID to return an existing profile
	existingProfile := &profile.Profile{ID: "profile-1", UserID: userID}
	mockRepo.On("GetByUserID", ctx, userID).Return(existingProfile, nil)

	p, err := s.CreateProfile(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, p)
	assert.Equal(t, service.ErrProfileAlreadyExists, err)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestUpdateProfile_Success(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	s := service.NewProfileService(mockRepo)
	ctx := context.Background()
	userID := "user-123"

	newName := "Johnny"
	input := service.UpdateProfileInput{
		FirstName: &newName,
	}

	existingProfile := &profile.Profile{
		ID:        "profile-1",
		UserID:    userID,
		FirstName: "John",
		Bio:       "Old bio",
	}

	mockRepo.On("GetByUserID", ctx, userID).Return(existingProfile, nil)

	mockRepo.On("Update", ctx, mock.MatchedBy(func(p *profile.Profile) bool {
		return p.FirstName == "Johnny" && p.Bio == "Old bio"
	})).Return(nil)

	p, err := s.UpdateProfile(ctx, userID, input)

	assert.NoError(t, err)
	assert.Equal(t, "Johnny", p.FirstName)
	mockRepo.AssertExpectations(t)
}

func TestUpdateProfile_NotFound(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	s := service.NewProfileService(mockRepo)
	ctx := context.Background()
	userID := "user-123"

	input := service.UpdateProfileInput{}

	mockRepo.On("GetByUserID", ctx, userID).Return(nil, repository.ErrNotFound)

	p, err := s.UpdateProfile(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, p)
	assert.Equal(t, repository.ErrNotFound, err)
}
