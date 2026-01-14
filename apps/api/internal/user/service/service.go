package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"github.com/kisssonik/hearts/internal/user"
	"github.com/kisssonik/hearts/internal/user/repository"
	"github.com/kisssonik/hearts/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials is returned when a login attempt fails.
var ErrInvalidCredentials = errors.New("invalid email or password")

// LoginResponse is the data returned upon a successful login.
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// UserService defines the interface for user-related business logic.
type UserService interface {
	RegisterUser(ctx context.Context, email, username, password string) (*user.User, error)
	Login(ctx context.Context, email, password string) (*LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
	GetUserByID(ctx context.Context, id string) (*user.User, error)
}

// userService is the implementation of UserService.
type userService struct {
	repo        repository.UserRepository
	authService auth.AuthService
}

// NewUserService creates a new instance of userService.
func NewUserService(repo repository.UserRepository, authService auth.AuthService) UserService {
	return &userService{
		repo:        repo,
		authService: authService,
	}
}

// RegisterUser handles the business logic for creating a new user.
func (s *userService) RegisterUser(ctx context.Context, email, username, password string) (*user.User, error) {
	// 1. Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 2. Create a new User object
	newUser := &user.User{
		Email:        email,
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	// 3. Call the repository to save the user
	err = s.repo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, err
	}

	// 4. Return the newly created user (which now has ID and CreatedAt from the DB)
	return newUser, nil
}

// Login handles the business logic for user authentication.
func (s *userService) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	// 1. Retrieve the user by email from the repository.
	u, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// 2. Compare the provided password with the stored hash.
	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// 3. Generate new access and refresh tokens, and the refresh token hash.
	accessToken, refreshToken, refreshTokenHash, refreshTokenExpiresAt, err := s.authService.GenerateTokens(u.ID)
	if err != nil {
		return nil, err
	}

	// 4. Store the refresh token hash and its expiry in the database.
	err = s.repo.SetRefreshToken(ctx, u.ID, refreshTokenHash, refreshTokenExpiresAt)
	if err != nil {
		return nil, err
	}

	// 5. Return the tokens to the handler.
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken validates a refresh token and issues new tokens if it's valid.
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// 1. Hash the incoming refresh token so we can look it up in the database.
	// We need a way to hash this, so we'll add a temporary helper in the auth package.
	// NOTE: This is a temporary solution. A better approach might be to have a separate
	// crypto package or keep the hashing consistent.
	hash := sha256.Sum256([]byte(refreshToken))
	refreshTokenHash := base64.URLEncoding.EncodeToString(hash[:])

	// 2. Get the user associated with the refresh token.
	// The repository method should also check if the token is expired.
	u, err := s.repo.GetUserByRefreshToken(ctx, refreshTokenHash)
	if err != nil {
		// If the token is not found or expired, return an error.
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("invalid or expired refresh token")
		}
		return nil, err
	}

	// 3. Generate a new pair of access and refresh tokens.
	newAccessToken, newRefreshToken, newRefreshTokenHash, newRefreshTokenExpiresAt, err := s.authService.GenerateTokens(u.ID)
	if err != nil {
		return nil, err
	}

	// 4. Update the refresh token in the database with the new hash and expiry.
	err = s.repo.SetRefreshToken(ctx, u.ID, newRefreshTokenHash, newRefreshTokenExpiresAt)
	if err != nil {
		return nil, err
	}

	// 5. Return the new tokens.
	return &LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// GetUserByID retrieves a user by their ID.
func (s *userService) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	return s.repo.GetUserByID(ctx, id)
}
