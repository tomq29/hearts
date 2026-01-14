package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kisssonik/hearts/internal/profile"
	"github.com/kisssonik/hearts/internal/profile/repository"
	"github.com/kisssonik/hearts/internal/user"
	userRepo "github.com/kisssonik/hearts/internal/user/repository"
	"github.com/kisssonik/hearts/pkg/config"
	"github.com/kisssonik/hearts/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	cfg := config.DatabaseConfig{
		URL: "postgres://postgres:password@localhost:5433/hearts?sslmode=disable",
	}
	pool, err := database.NewPostgresPool(context.Background(), cfg)
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(), "TRUNCATE TABLE profiles, users CASCADE")
	require.NoError(t, err)

	return pool
}

func createTestUser(t *testing.T, db *pgxpool.Pool, email, username string) *user.User {
	repo := userRepo.NewUserRepository(db)
	u := &user.User{
		Email:        email,
		Username:     username,
		PasswordHash: "hash",
	}
	err := repo.CreateUser(context.Background(), u)
	require.NoError(t, err)
	return u
}

func TestProfileRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	u := createTestUser(t, db, "test@example.com", "testuser")
	repo := repository.NewProfileRepository(db)

	p := &profile.Profile{
		UserID:                 u.ID,
		FirstName:              "John",
		Bio:                    "Hello world",
		Photos:                 []string{"photo1.jpg"},
		SelfDescribedFlaws:     []string{"flaw1"},
		SelfDescribedStrengths: []string{"strength1"},
	}

	err := repo.Create(context.Background(), p)
	assert.NoError(t, err)
	assert.NotEmpty(t, p.ID)
	assert.WithinDuration(t, time.Now(), p.CreatedAt, 2*time.Second)
	assert.WithinDuration(t, time.Now(), p.UpdatedAt, 2*time.Second)
}

func TestProfileRepository_GetByUserID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	u := createTestUser(t, db, "test@example.com", "testuser")
	repo := repository.NewProfileRepository(db)

	p := &profile.Profile{
		UserID:    u.ID,
		FirstName: "John",
	}
	err := repo.Create(context.Background(), p)
	require.NoError(t, err)

	fetched, err := repo.GetByUserID(context.Background(), u.ID)
	assert.NoError(t, err)
	assert.Equal(t, p.ID, fetched.ID)
	assert.Equal(t, "John", fetched.FirstName)
}

func TestProfileRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	u := createTestUser(t, db, "test@example.com", "testuser")
	repo := repository.NewProfileRepository(db)

	p := &profile.Profile{
		UserID:    u.ID,
		FirstName: "John",
	}
	err := repo.Create(context.Background(), p)
	require.NoError(t, err)

	p.FirstName = "Jane"
	p.Bio = "Updated bio"
	err = repo.Update(context.Background(), p)
	assert.NoError(t, err)

	fetched, err := repo.GetByUserID(context.Background(), u.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Jane", fetched.FirstName)
	assert.Equal(t, "Updated bio", fetched.Bio)
}
