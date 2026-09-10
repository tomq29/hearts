package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
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

func (m *MockProfileService) GetProfileForViewer(ctx context.Context, viewerID, targetUserID string) (*profile.Profile, error) {
	args := m.Called(ctx, viewerID, targetUserID)
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

func (m *MockProfileService) SearchProfiles(ctx context.Context, userID string, params service.SearchParams) ([]*profile.Profile, error) {
	args := m.Called(ctx, userID, params)
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
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, "viewer1"))
	w := httptest.NewRecorder()

	expectedProfile := &profile.Profile{ID: "p1", UserID: "user1", FirstName: "John"}
	mockService.On("GetProfileForViewer", mock.Anything, "viewer1", "user1").Return(expectedProfile, nil)

	h.Get(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// The security property itself: a viewer must never receive another user's
// coordinates, only a coarse band. Asserted on the serialised bytes rather
// than the struct, because the JSON is what actually leaves the process.
func TestProfileHandler_Get_NeverSerialisesCoordinates(t *testing.T) {
	mockService := new(MockProfileService)
	mockStorage := new(MockStorageProvider)
	logger := zap.NewNop()
	h := handler.NewProfileHandler(mockService, mockStorage, logger)

	req := httptest.NewRequest("GET", "/profiles/user1", nil)
	req.SetPathValue("userID", "user1")
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, "viewer1"))
	w := httptest.NewRecorder()

	lat, lon := 51.4700, -0.4543
	band := "10–25 km"
	target := &profile.Profile{
		ID: "p1", UserID: "user1", FirstName: "John",
		Latitude: &lat, Longitude: &lon, DistanceBand: &band,
	}
	mockService.On("GetProfileForViewer", mock.Anything, "viewer1", "user1").Return(target, nil)

	h.Get(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.NotContains(t, body, "latitude")
	assert.NotContains(t, body, "longitude")
	assert.NotContains(t, body, "51.47")
	assert.NotContains(t, body, "-0.4543")
	assert.Contains(t, body, "10–25 km")
}

func TestProfileHandler_Get_Unauthenticated(t *testing.T) {
	mockService := new(MockProfileService)
	mockStorage := new(MockStorageProvider)
	logger := zap.NewNop()
	h := handler.NewProfileHandler(mockService, mockStorage, logger)

	req := httptest.NewRequest("GET", "/profiles/user1", nil)
	req.SetPathValue("userID", "user1")
	w := httptest.NewRecorder()

	h.Get(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertNotCalled(t, "GetProfileForViewer")
}

func multipartBody(t *testing.T, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("photo", filename)
	assert.NoError(t, err)
	_, err = part.Write(content)
	assert.NoError(t, err)
	assert.NoError(t, writer.Close())
	return body, writer.FormDataContentType()
}

func TestProfileHandler_UploadPhoto_RejectsNonImage(t *testing.T) {
	mockService := new(MockProfileService)
	mockStorage := new(MockStorageProvider)
	logger := zap.NewNop()
	h := handler.NewProfileHandler(mockService, mockStorage, logger)

	// A script disguised with an image filename must be rejected by sniffing.
	body, contentType := multipartBody(t, "photo.png", []byte("<script>alert(1)</script>"))
	req := httptest.NewRequest("POST", "/profiles/upload", body)
	req.Header.Set("Content-Type", contentType)
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, "user1"))
	w := httptest.NewRecorder()

	h.UploadPhoto(w, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
	mockStorage.AssertNotCalled(t, "Upload")
}

func TestProfileHandler_UploadPhoto_AcceptsPNG(t *testing.T) {
	mockService := new(MockProfileService)
	mockStorage := new(MockStorageProvider)
	logger := zap.NewNop()
	h := handler.NewProfileHandler(mockService, mockStorage, logger)

	// Minimal PNG signature is enough for http.DetectContentType.
	pngBytes := append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 16)...)
	body, contentType := multipartBody(t, "anything.txt", pngBytes)
	req := httptest.NewRequest("POST", "/profiles/upload", body)
	req.Header.Set("Content-Type", contentType)
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, "user1"))
	w := httptest.NewRecorder()

	// The stored key must use the sniffed extension, not the client filename.
	mockStorage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "image/png",
		mock.MatchedBy(func(key string) bool { return strings.HasSuffix(key, ".png") }),
	).Return("user1/key.png", nil)
	mockStorage.On("GetPresignedURL", mock.Anything, "user1/key.png").Return("http://url", nil)

	h.UploadPhoto(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStorage.AssertExpectations(t)
}

func TestProfileHandler_Get_NotFound(t *testing.T) {
	mockService := new(MockProfileService)
	mockStorage := new(MockStorageProvider)
	logger := zap.NewNop()
	h := handler.NewProfileHandler(mockService, mockStorage, logger)

	req := httptest.NewRequest("GET", "/profiles/user1", nil)
	req.SetPathValue("userID", "user1")
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, "viewer1"))
	w := httptest.NewRecorder()

	mockService.On("GetProfileForViewer", mock.Anything, "viewer1", "user1").Return(nil, repository.ErrNotFound)

	h.Get(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}
