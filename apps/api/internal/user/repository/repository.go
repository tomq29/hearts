package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kisssonik/hearts/internal/user"
)

var (
	// ErrDuplicateEmailOrUsername is a custom error for unique constraint violations.
	ErrDuplicateEmailOrUsername = errors.New("email or username already exists")
	// ErrNotFound is returned when a requested record is not found.
	ErrNotFound = errors.New("not found")
)

// UserRepository defines the interface for user-related database operations.
type UserRepository interface {
	CreateUser(ctx context.Context, u *user.User) error
	GetUserByEmail(ctx context.Context, email string) (*user.User, error)
	GetUserByID(ctx context.Context, id string) (*user.User, error)
	CreateRefreshToken(ctx context.Context, userID string, tokenHash string, expiresAt time.Time) error
	GetUserByRefreshToken(ctx context.Context, tokenHash string) (*user.User, error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
}

// pgxUserRepository is the implementation of UserRepository using pgx.
type pgxUserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new instance of pgxUserRepository.
func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &pgxUserRepository{db: db}
}

// CreateUser inserts a new user into the database.
func (r *pgxUserRepository) CreateUser(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query, u.Email, u.Username, u.PasswordHash).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		// If the error is a unique violation, return our custom error.
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateEmailOrUsername
		}
		return err
	}
	return nil
}

// GetUserByEmail retrieves a user from the database by their email address.
func (r *pgxUserRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, email, username, password_hash, created_at
		FROM users
		WHERE email = $1
	`
	u := &user.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		// If no user is found, pgx returns ErrNoRows. We wrap this in our custom error.
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

// GetUserByID retrieves a user from the database by their ID.
func (r *pgxUserRepository) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	query := `
		SELECT id, email, username, password_hash, created_at
		FROM users
		WHERE id = $1
	`
	u := &user.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

// CreateRefreshToken stores a new refresh token for the user. Each login or
// rotation adds its own row, so a user can hold one token per device.
func (r *pgxUserRepository) CreateRefreshToken(ctx context.Context, userID string, tokenHash string, expiresAt time.Time) error {
	// Opportunistic cleanup: expired tokens for this user will never be
	// looked up again, so drop them while we're here.
	if _, err := r.db.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE user_id = $1 AND expires_at <= NOW()`, userID); err != nil {
		return err
	}

	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, userID, tokenHash, expiresAt)
	return err
}

// GetUserByRefreshToken retrieves a user by a valid (unexpired) refresh token hash.
func (r *pgxUserRepository) GetUserByRefreshToken(ctx context.Context, tokenHash string) (*user.User, error) {
	query := `
		SELECT u.id, u.email, u.username, u.password_hash, u.created_at
		FROM users u
		JOIN refresh_tokens rt ON rt.user_id = u.id
		WHERE rt.token_hash = $1 AND rt.expires_at > NOW()
	`
	u := &user.User{}
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

// DeleteRefreshToken removes a refresh token (rotation or logout).
// Deleting a token that no longer exists is not an error.
func (r *pgxUserRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
	return err
}
