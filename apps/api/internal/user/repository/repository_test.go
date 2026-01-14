package repository_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kisssonik/hearts/internal/user"
	"github.com/kisssonik/hearts/internal/user/repository"
	"github.com/kisssonik/hearts/pkg/config"
	"github.com/kisssonik/hearts/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a connection to the test database.
// WARNING: This connects to the running Docker database and truncates the users table!
func setupTestDB(t *testing.T) *pgxpool.Pool {
	cfg := config.DatabaseConfig{
		URL: "postgres://postgres:password@localhost:5433/hearts?sslmode=disable",
	}
	pool, err := database.NewPostgresPool(context.Background(), cfg)
	require.NoError(t, err)

	// Clean up
	_, err = pool.Exec(context.Background(), "TRUNCATE TABLE users CASCADE")
	require.NoError(t, err)

	return pool
}

func TestCreateUser_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	u := &user.User{
		Email:        "integration@example.com",
		Username:     "integration_user",
		PasswordHash: "hashed_secret",
	}

	err := repo.CreateUser(ctx, u)
	assert.NoError(t, err)
	assert.NotEmpty(t, u.ID)
	assert.NotZero(t, u.CreatedAt)

	// Verify it's in the DB
	savedUser, err := repo.GetUserByID(ctx, u.ID)
	assert.NoError(t, err)
	assert.Equal(t, u.Email, savedUser.Email)
}

func TestCreateUser_DuplicateEmail_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	u1 := &user.User{
		Email:        "duplicate@example.com",
		Username:     "user1",
		PasswordHash: "hash",
	}
	err := repo.CreateUser(ctx, u1)
	require.NoError(t, err)

	u2 := &user.User{
		Email:        "duplicate@example.com", // Same email
		Username:     "user2",
		PasswordHash: "hash",
	}
	err = repo.CreateUser(ctx, u2)

	assert.Error(t, err)
	assert.Equal(t, repository.ErrDuplicateEmailOrUsername, err)
}
