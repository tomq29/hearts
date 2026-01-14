package repository_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kisssonik/hearts/internal/like"
	"github.com/kisssonik/hearts/internal/like/repository"
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

	_, err = pool.Exec(context.Background(), "TRUNCATE TABLE likes, users CASCADE")
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

func TestLike_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewLikeRepository(db)
	ctx := context.Background()

	u1 := createTestUser(t, db, "u1@example.com", "u1")
	u2 := createTestUser(t, db, "u2@example.com", "u2")

	// u1 likes u2
	l1 := &like.Like{FromUserID: u1.ID, ToUserID: u2.ID, IsLike: true}
	err := repo.Upsert(ctx, l1)
	assert.NoError(t, err)

	// Check mutual like (should be false)
	mutual, err := repo.HasMutualLike(ctx, u1.ID, u2.ID)
	assert.NoError(t, err)
	assert.False(t, mutual)

	// u2 likes u1
	l2 := &like.Like{FromUserID: u2.ID, ToUserID: u1.ID, IsLike: true}
	err = repo.Upsert(ctx, l2)
	assert.NoError(t, err)

	// Check mutual like (should be true)
	mutual, err = repo.HasMutualLike(ctx, u1.ID, u2.ID)
	assert.NoError(t, err)
	assert.True(t, mutual)

	// Check matches
	matches, err := repo.GetMatches(ctx, u1.ID)
	assert.NoError(t, err)
	assert.Contains(t, matches, u2.ID)
}
