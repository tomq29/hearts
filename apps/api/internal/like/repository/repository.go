package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kisssonik/hearts/internal/like"
)

type LikeRepository interface {
	Upsert(ctx context.Context, l *like.Like) error
	HasMutualLike(ctx context.Context, user1ID, user2ID string) (bool, error)
	GetMatches(ctx context.Context, userID string) ([]string, error)
}

type pgxLikeRepository struct {
	db *pgxpool.Pool
}

func NewLikeRepository(db *pgxpool.Pool) LikeRepository {
	return &pgxLikeRepository{db: db}
}

func (r *pgxLikeRepository) Upsert(ctx context.Context, l *like.Like) error {
	query := `
		INSERT INTO likes (from_user_id, to_user_id, is_like)
		VALUES ($1, $2, $3)
		ON CONFLICT (from_user_id, to_user_id) 
		DO UPDATE SET is_like = EXCLUDED.is_like, created_at = NOW()
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query,
		l.FromUserID, l.ToUserID, l.IsLike,
	).Scan(&l.ID, &l.CreatedAt)
}

func (r *pgxLikeRepository) HasMutualLike(ctx context.Context, user1ID, user2ID string) (bool, error) {
	// Check if user1 liked user2 AND user2 liked user1
	query := `
		SELECT COUNT(*)
		FROM likes l1
		JOIN likes l2 ON l1.to_user_id = l2.from_user_id AND l1.from_user_id = l2.to_user_id
		WHERE l1.from_user_id = $1 AND l1.to_user_id = $2
		  AND l1.is_like = TRUE AND l2.is_like = TRUE
	`
	var count int
	err := r.db.QueryRow(ctx, query, user1ID, user2ID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *pgxLikeRepository) GetMatches(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT l1.to_user_id
		FROM likes l1
		JOIN likes l2 ON l1.to_user_id = l2.from_user_id
		WHERE l1.from_user_id = $1 
		  AND l1.is_like = TRUE 
		  AND l2.to_user_id = $1 
		  AND l2.is_like = TRUE
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []string
	for rows.Next() {
		var matchID string
		if err := rows.Scan(&matchID); err != nil {
			return nil, err
		}
		matches = append(matches, matchID)
	}
	return matches, nil
}
