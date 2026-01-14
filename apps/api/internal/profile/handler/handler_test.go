package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kisssonik/hearts/internal/profile"
	"github.com/kisssonik/hearts/internal/profile/handler"
	"github.com/kisssonik/hearts/internal/profile/repository"
	"github.com/kisssonik/hearts/internal/profile/service"
	"github.com/kisssonik/hearts/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockProfileService
type MockProfileService struct {
	mock.Mock
}

func (m *MockProfileService) CreateProfile(ctx context.Context, userID string, input service.CreateProfileInput) (*profile.Profile, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*profile.Profile), args.Error(1)
}

func (m *MockProfileService) GetProfileByUserID(ctx context.Context, userID string) (*profile.Profile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*profile.Profile), args.Error(1)
}

func (m *MockProfileService) UpdateProfile(ctx context.Context, userID string, input service.UpdateProfileInput) (*profile.Profile, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*profile.Profile), args.Error(1)
}

// MockStorageProvider
type MockStorageProvider struct {
	mock.Mock
}

func (m *MockStorageProvider) Upload(ctx context.Context, file io.Reader, size int64, contentType, key string) (string, error) {
	args := m.Called(ctx, file, size, contentType, key)
	return args.String(0), args.Error(1)
}

func (m *MockStorageProvider) GetPresignedURL(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func TestProfileHandler_Create(t *testing.T) {
	mockService := new(MockProfileService)
	mockStorage := new(MockStorageProvider)
	logger := zap.NewNop()
	h := handler.NewProfileHandler(mockService, mockStorage, logger)

	input := service.CreateProfileInput{
		FirstName: "John",
	}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest("POST", "/profiles", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, "user1"))
	w := httptest.NewRecorder()

	expectedProfile := &profile.Profile{ID: "p1", UserID: "user1", FirstName: "John"}
	mockService.On("CreateProfile", mock.Anything, "user1", input).Return(expectedProfile, nil)

	h.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestProfileHandler_Get(t *testing.T) {
	mockService := new(MockProfileService)
	mockStorage := new(MockStorageProvider)
	logger := zap.NewNop()
	h := handler.NewProfileHandler(mockService, mockStorage, logger)

	req := httptest.NewRequest("GET", "/profiles/user1", nil)
	req.SetPathValue("userID", "user1")
	w := httptest.NewRecorder()

	expectedProfile := &profile.Profile{ID: "p1", UserID: "user1", FirstName: "John"}
	mockService.On("GetProfileByUserID", mock.Anything, "user1").Return(expectedProfile, nil)

	h.Get(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestProfileHandler_Get_NotFound(t *testing.T) {
	mockService := new(MockProfileService)
	mockStorage := new(MockStorageProvider)
	logger := zap.NewNop()
	h := handler.NewProfileHandler(mockService, mockStorage, logger)

	req := httptest.NewRequest("GET", "/profiles/user1", nil)
	req.SetPathValue("userID", "user1")
	w := httptest.NewRecorder()

	mockService.On("GetProfileByUserID", mock.Anything, "user1").Return(nil, repository.ErrNotFound)

	h.Get(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}
