package service_test

import (
	"context"
	"testing"

	"github.com/kisssonik/hearts/internal/profile"
	"github.com/kisssonik/hearts/internal/profile/repository"
	"github.com/kisssonik/hearts/internal/profile/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func testBander(t *testing.T) *profile.Bander {
	t.Helper()
	b, err := profile.ParseBands([]string{"5", "10", "25", "50"})
	require.NoError(t, err)
	return b
}

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

func (m *MockProfileRepository) Search(ctx context.Context, currentUser *profile.Profile, params repository.SearchParams) ([]*profile.Profile, error) {
	args := m.Called(ctx, currentUser, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*profile.Profile), args.Error(1)
}

func TestCreateProfile_Success(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	s := service.NewProfileService(mockRepo, testBander(t))
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
	s := service.NewProfileService(mockRepo, testBander(t))
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
	s := service.NewProfileService(mockRepo, testBander(t))
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
	s := service.NewProfileService(mockRepo, testBander(t))
	ctx := context.Background()
	userID := "user-123"

	input := service.UpdateProfileInput{}

	mockRepo.On("GetByUserID", ctx, userID).Return(nil, repository.ErrNotFound)

	p, err := s.UpdateProfile(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, p)
	assert.Equal(t, repository.ErrNotFound, err)
}

func coords(lat, lon float64) (*float64, *float64) {
	return &lat, &lon
}

func TestGetProfileForViewer_AttachesBandNotCoordinates(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	s := service.NewProfileService(mockRepo, testBander(t))
	ctx := context.Background()

	// London centre and Heathrow, about 22 km apart.
	viewerLat, viewerLon := coords(51.5074, -0.1278)
	targetLat, targetLon := coords(51.4700, -0.4543)

	viewer := &profile.Profile{UserID: "viewer", Latitude: viewerLat, Longitude: viewerLon}
	target := &profile.Profile{UserID: "target", Latitude: targetLat, Longitude: targetLon}

	mockRepo.On("GetByUserID", ctx, "target").Return(target, nil)
	mockRepo.On("GetByUserID", ctx, "viewer").Return(viewer, nil)

	p, err := s.GetProfileForViewer(ctx, "viewer", "target")

	assert.NoError(t, err)
	require.NotNil(t, p.DistanceBand)
	assert.Equal(t, "10–25 km", *p.DistanceBand)
	mockRepo.AssertExpectations(t)
}

func TestGetProfileForViewer_NoBandWhenCoordinatesMissing(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	s := service.NewProfileService(mockRepo, testBander(t))
	ctx := context.Background()

	targetLat, targetLon := coords(51.4700, -0.4543)
	viewer := &profile.Profile{UserID: "viewer"} // never shared a location
	target := &profile.Profile{UserID: "target", Latitude: targetLat, Longitude: targetLon}

	mockRepo.On("GetByUserID", ctx, "target").Return(target, nil)
	mockRepo.On("GetByUserID", ctx, "viewer").Return(viewer, nil)

	p, err := s.GetProfileForViewer(ctx, "viewer", "target")

	assert.NoError(t, err)
	assert.Nil(t, p.DistanceBand)
}

// A viewer who has not created a profile can still read others.
func TestGetProfileForViewer_ViewerWithoutProfile(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	s := service.NewProfileService(mockRepo, testBander(t))
	ctx := context.Background()

	target := &profile.Profile{UserID: "target"}
	mockRepo.On("GetByUserID", ctx, "target").Return(target, nil)
	mockRepo.On("GetByUserID", ctx, "viewer").Return(nil, repository.ErrNotFound)

	p, err := s.GetProfileForViewer(ctx, "viewer", "target")

	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Nil(t, p.DistanceBand)
}

// A radius filter the viewer cannot satisfy must be dropped, not turned into
// an error that empties the feed.
func TestSearchProfiles_RadiusIgnoredWhenViewerHasNoLocation(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	s := service.NewProfileService(mockRepo, testBander(t))
	ctx := context.Background()

	viewer := &profile.Profile{UserID: "viewer"} // no coordinates
	mockRepo.On("GetByUserID", ctx, "viewer").Return(viewer, nil)
	mockRepo.On("Search", ctx, viewer, mock.MatchedBy(func(p repository.SearchParams) bool {
		return p.RadiusKM == nil // the radius was dropped
	})).Return([]*profile.Profile{{UserID: "other"}}, nil)

	radius := 10.0
	results, err := s.SearchProfiles(ctx, "viewer", service.SearchParams{RadiusKM: &radius})

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	mockRepo.AssertExpectations(t)
}
