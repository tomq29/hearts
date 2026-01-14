package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kisssonik/hearts/internal/review"
)

type ReviewRepository interface {
	Create(ctx context.Context, r *review.Review) error
	GetByTargetID(ctx context.Context, targetID string) ([]*review.Review, error)
}

type pgxReviewRepository struct {
	db *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) ReviewRepository {
	return &pgxReviewRepository{db: db}
}

func (r *pgxReviewRepository) Create(ctx context.Context, rev *review.Review) error {
	query := `
		INSERT INTO reviews (author_id, target_id, rating, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query,
		rev.AuthorID, rev.TargetID, rev.Rating, rev.Comment,
	).Scan(&rev.ID, &rev.CreatedAt)
}

func (r *pgxReviewRepository) GetByTargetID(ctx context.Context, targetID string) ([]*review.Review, error) {
	query := `
		SELECT id, author_id, target_id, rating, comment, created_at
		FROM reviews
		WHERE target_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*review.Review
	for rows.Next() {
		rev := &review.Review{}
		if err := rows.Scan(&rev.ID, &rev.AuthorID, &rev.TargetID, &rev.Rating, &rev.Comment, &rev.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, rev)
	}
	return reviews, nil
}
