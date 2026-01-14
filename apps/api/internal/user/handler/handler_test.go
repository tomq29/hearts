package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kisssonik/hearts/internal/user"
	"github.com/kisssonik/hearts/internal/user/handler"
	"github.com/kisssonik/hearts/internal/user/repository"
	"github.com/kisssonik/hearts/internal/user/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockUserService is a mock implementation of service.UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) RegisterUser(ctx context.Context, email, username, password string) (*user.User, error) {
	args := m.Called(ctx, email, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, email, password string) (*service.LoginResponse, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.LoginResponse), args.Error(1)
}

func (m *MockUserService) RefreshToken(ctx context.Context, refreshToken string) (*service.LoginResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.LoginResponse), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func TestRegister_Success(t *testing.T) {
	// Setup
	mockService := new(MockUserService)
	logger := zap.NewNop()
	h := handler.NewUserHandler(mockService, nil, logger)

	input := map[string]string{
		"email":    "test@example.com",
		"username": "testuser",
		"password": "password123",
	}
	body, _ := json.Marshal(input)

	expectedUser := &user.User{
		ID:       "user-123",
		Email:    "test@example.com",
		Username: "testuser",
	}

	mockService.On("RegisterUser", mock.Anything, "test@example.com", "testuser", "password123").Return(expectedUser, nil)

	// Request
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	// Execute
	h.Register(rr, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rr.Code)

	var responseUser user.User
	json.Unmarshal(rr.Body.Bytes(), &responseUser)
	assert.Equal(t, expectedUser.ID, responseUser.ID)
	assert.Equal(t, expectedUser.Email, responseUser.Email)

	mockService.AssertExpectations(t)
}

func TestRegister_InvalidInput(t *testing.T) {
	mockService := new(MockUserService)
	logger := zap.NewNop()
	h := handler.NewUserHandler(mockService, nil, logger)

	input := map[string]string{
		"email": "", // Invalid
	}
	body, _ := json.Marshal(input)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRegister_Conflict(t *testing.T) {
	mockService := new(MockUserService)
	logger := zap.NewNop()
	h := handler.NewUserHandler(mockService, nil, logger)

	input := map[string]string{
		"email":    "existing@example.com",
		"username": "existinguser",
		"password": "password123",
	}
	body, _ := json.Marshal(input)

	mockService.On("RegisterUser", mock.Anything, "existing@example.com", "existinguser", "password123").
		Return(nil, repository.ErrDuplicateEmailOrUsername)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	mockService.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	mockService := new(MockUserService)
	logger := zap.NewNop()
	h := handler.NewUserHandler(mockService, nil, logger)

	input := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(input)

	loginResp := &service.LoginResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	mockService.On("Login", mock.Anything, "test@example.com", "password123").Return(loginResp, nil)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Check response body for access token
	var resp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.Equal(t, "access-token", resp["accessToken"])

	// Check cookie for refresh token
	cookies := rr.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "refreshToken" {
			assert.Equal(t, "refresh-token", c.Value)
			assert.True(t, c.HttpOnly)
			found = true
			break
		}
	}
	assert.True(t, found, "RefreshToken cookie should be set")

	mockService.AssertExpectations(t)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	mockService := new(MockUserService)
	logger := zap.NewNop()
	h := handler.NewUserHandler(mockService, nil, logger)

	input := map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}
	body, _ := json.Marshal(input)

	mockService.On("Login", mock.Anything, "test@example.com", "wrongpassword").
		Return(nil, service.ErrInvalidCredentials)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	mockService.AssertExpectations(t)
}
