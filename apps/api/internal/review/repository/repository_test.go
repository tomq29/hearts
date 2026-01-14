package repository_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kisssonik/hearts/internal/review"
	"github.com/kisssonik/hearts/internal/review/repository"
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

	_, err = pool.Exec(context.Background(), "TRUNCATE TABLE reviews, users CASCADE")
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

func TestReviewRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewReviewRepository(db)
	ctx := context.Background()

	u1 := createTestUser(t, db, "u1@example.com", "u1")
	u2 := createTestUser(t, db, "u2@example.com", "u2")

	rev := &review.Review{
		AuthorID: u1.ID,
		TargetID: u2.ID,
		Rating:   5,
		Comment:  "Great person!",
	}

	err := repo.Create(ctx, rev)
	assert.NoError(t, err)
	assert.NotEmpty(t, rev.ID)

	fetched, err := repo.GetByTargetID(ctx, u2.ID)
	assert.NoError(t, err)
	assert.Len(t, fetched, 1)
	assert.Equal(t, "Great person!", fetched[0].Comment)
}
