package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kisssonik/hearts/internal/user"
	"github.com/kisssonik/hearts/internal/user/service"
	"github.com/kisssonik/hearts/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) SetRefreshToken(ctx context.Context, userID string, tokenHash string, expiresAt time.Time) error {
	args := m.Called(ctx, userID, tokenHash, expiresAt)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByRefreshToken(ctx context.Context, tokenHash string) (*user.User, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

// MockAuthService is a mock implementation of auth.AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) GenerateTokens(userID string) (string, string, string, time.Time, error) {
	args := m.Called(userID)
	return args.String(0), args.String(1), args.String(2), args.Get(3).(time.Time), args.Error(4)
}

func (m *MockAuthService) ValidateAccessToken(tokenString string) (*auth.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

func TestRegisterUser_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	// We don't need auth service for registration in the current implementation
	// but NewUserService requires it.
	mockAuth := new(MockAuthService)

	svc := service.NewUserService(mockRepo, mockAuth)

	ctx := context.Background()
	email := "test@example.com"
	username := "testuser"
	password := "password123"

	// Expect CreateUser to be called once.
	// We use mock.AnythingOfType to match the *user.User pointer.
	mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*user.User")).Return(nil)

	// Act
	createdUser, err := svc.RegisterUser(ctx, email, username, password)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.Equal(t, email, createdUser.Email)
	assert.Equal(t, username, createdUser.Username)
	assert.NotEmpty(t, createdUser.PasswordHash)           // Password should be hashed
	assert.NotEqual(t, password, createdUser.PasswordHash) // Hash should not be plain text

	mockRepo.AssertExpectations(t)
}

func TestRegisterUser_RepoError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockAuth := new(MockAuthService)
	svc := service.NewUserService(mockRepo, mockAuth)

	ctx := context.Background()
	expectedErr := errors.New("database error")

	mockRepo.On("CreateUser", ctx, mock.Anything).Return(expectedErr)

	// Act
	createdUser, err := svc.RegisterUser(ctx, "fail@example.com", "failuser", "pass")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, createdUser)

	mockRepo.AssertExpectations(t)
}
