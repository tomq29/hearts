package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kisssonik/hearts/internal/like/handler"
	"github.com/kisssonik/hearts/internal/like/service"
	"github.com/kisssonik/hearts/internal/profile"
	"github.com/kisssonik/hearts/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockLikeService
type MockLikeService struct {
	mock.Mock
}

func (m *MockLikeService) LikeUser(ctx context.Context, fromUserID string, input service.LikeInput) (bool, error) {
	args := m.Called(ctx, fromUserID, input)
	return args.Bool(0), args.Error(1)
}

func (m *MockLikeService) GetMatches(ctx context.Context, userID string) ([]*profile.Profile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*profile.Profile), args.Error(1)
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

func TestLikeHandler_Like(t *testing.T) {
	mockService := new(MockLikeService)
	mockStorage := new(MockStorageProvider)
	logger := zap.NewNop()
	h := handler.NewLikeHandler(mockService, mockStorage, logger)

	input := service.LikeInput{TargetID: "u2", IsLike: true}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest("POST", "/likes", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, "u1"))
	w := httptest.NewRecorder()

	mockService.On("LikeUser", mock.Anything, "u1", input).Return(true, nil)

	h.Like(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp handler.LikeResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.True(t, resp.IsMatch)
}

func TestLikeHandler_GetMatches(t *testing.T) {
	mockService := new(MockLikeService)
	mockStorage := new(MockStorageProvider)
	logger := zap.NewNop()
	h := handler.NewLikeHandler(mockService, mockStorage, logger)

	req := httptest.NewRequest("GET", "/matches", nil)
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, "u1"))
	w := httptest.NewRecorder()

	matches := []*profile.Profile{
		{UserID: "u2", FirstName: "User 2", Photos: []string{"p1.jpg"}},
	}
	mockService.On("GetMatches", mock.Anything, "u1").Return(matches, nil)
	mockStorage.On("GetPresignedURL", mock.Anything, "p1.jpg").Return("http://p1.jpg", nil)

	h.GetMatches(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []*profile.Profile
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp, 1)
	assert.Equal(t, "http://p1.jpg", resp[0].Photos[0])
}
