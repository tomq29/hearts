package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kisssonik/hearts/internal/profile"
)

var ErrNotFound = errors.New("profile not found")

type SearchParams struct {
	MinAge    *int
	MaxAge    *int
	RadiusKM  *float64
	Gender    *string
	MinHeight *int
	MaxHeight *int
}

type ProfileRepository interface {
	Create(ctx context.Context, p *profile.Profile) error
	GetByUserID(ctx context.Context, userID string) (*profile.Profile, error)
	GetByUserIDs(ctx context.Context, userIDs []string) ([]*profile.Profile, error)
	Update(ctx context.Context, p *profile.Profile) error
	Search(ctx context.Context, currentUser *profile.Profile, params SearchParams) ([]*profile.Profile, error)
}

type pgxProfileRepository struct {
	db *pgxpool.Pool
}

func NewProfileRepository(db *pgxpool.Pool) ProfileRepository {
	return &pgxProfileRepository{db: db}
}

func (r *pgxProfileRepository) Create(ctx context.Context, p *profile.Profile) error {
	query := `
		INSERT INTO profiles (user_id, first_name, bio, photos, self_described_flaws, self_described_strengths, birth_date, gender, height, latitude, longitude)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		p.UserID, p.FirstName, p.Bio, p.Photos, p.SelfDescribedFlaws, p.SelfDescribedStrengths, p.BirthDate, p.Gender, p.Height, p.Latitude, p.Longitude,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *pgxProfileRepository) GetByUserID(ctx context.Context, userID string) (*profile.Profile, error) {
	query := `
		SELECT id, user_id, first_name, bio, photos, self_described_flaws, self_described_strengths, birth_date, gender, height, latitude, longitude, created_at, updated_at
		FROM profiles
		WHERE user_id = $1
	`
	p := &profile.Profile{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&p.ID, &p.UserID, &p.FirstName, &p.Bio, &p.Photos, &p.SelfDescribedFlaws, &p.SelfDescribedStrengths, &p.BirthDate, &p.Gender, &p.Height, &p.Latitude, &p.Longitude, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *pgxProfileRepository) GetByUserIDs(ctx context.Context, userIDs []string) ([]*profile.Profile, error) {
	query := `
		SELECT id, user_id, first_name, bio, photos, self_described_flaws, self_described_strengths, birth_date, gender, height, latitude, longitude, created_at, updated_at
		FROM profiles
		WHERE user_id = ANY($1)
	`
	rows, err := r.db.Query(ctx, query, userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*profile.Profile
	for rows.Next() {
		p := &profile.Profile{}
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.FirstName, &p.Bio, &p.Photos, &p.SelfDescribedFlaws, &p.SelfDescribedStrengths, &p.BirthDate, &p.Gender, &p.Height, &p.Latitude, &p.Longitude, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

func (r *pgxProfileRepository) Update(ctx context.Context, p *profile.Profile) error {
	query := `
		UPDATE profiles
		SET first_name = $1, bio = $2, photos = $3, self_described_flaws = $4, self_described_strengths = $5, birth_date = $6, gender = $7, height = $8, latitude = $9, longitude = $10, updated_at = NOW()
		WHERE user_id = $11
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		p.FirstName, p.Bio, p.Photos, p.SelfDescribedFlaws, p.SelfDescribedStrengths, p.BirthDate, p.Gender, p.Height, p.Latitude, p.Longitude, p.UserID,
	).Scan(&p.UpdatedAt)
}

func (r *pgxProfileRepository) Search(ctx context.Context, currentUser *profile.Profile, params SearchParams) ([]*profile.Profile, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	// Exclude self
	conditions = append(conditions, fmt.Sprintf("user_id != $%d", argIdx))
	args = append(args, currentUser.UserID)
	argIdx++

	if params.MinAge != nil {
		conditions = append(conditions, fmt.Sprintf("EXTRACT(YEAR FROM AGE(birth_date)) >= $%d", argIdx))
		args = append(args, *params.MinAge)
		argIdx++
	}

	if params.MaxAge != nil {
		conditions = append(conditions, fmt.Sprintf("EXTRACT(YEAR FROM AGE(birth_date)) <= $%d", argIdx))
		args = append(args, *params.MaxAge)
		argIdx++
	}

	if params.Gender != nil {
		conditions = append(conditions, fmt.Sprintf("gender = $%d", argIdx))
		args = append(args, *params.Gender)
		argIdx++
	}

	if params.MinHeight != nil {
		conditions = append(conditions, fmt.Sprintf("height >= $%d", argIdx))
		args = append(args, *params.MinHeight)
		argIdx++
	}

	if params.MaxHeight != nil {
		conditions = append(conditions, fmt.Sprintf("height <= $%d", argIdx))
		args = append(args, *params.MaxHeight)
		argIdx++
	}

	if params.RadiusKM != nil && currentUser.Latitude != nil && currentUser.Longitude != nil {
		// Haversine formula
		// 6371 is Earth radius in km
		haversine := fmt.Sprintf(`
			(6371 * acos(
				cos(radians($%d)) * cos(radians(latitude)) * cos(radians(longitude) - radians($%d)) +
				sin(radians($%d)) * sin(radians(latitude))
			)) <= $%d
		`, argIdx, argIdx+1, argIdx, argIdx+2)

		conditions = append(conditions, haversine)
		args = append(args, *currentUser.Latitude, *currentUser.Longitude, *params.RadiusKM)
		argIdx += 3
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, first_name, bio, photos, self_described_flaws, self_described_strengths, birth_date, gender, height, latitude, longitude, created_at, updated_at
		FROM profiles
		%s
		LIMIT 50
	`, whereClause)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*profile.Profile
	for rows.Next() {
		p := &profile.Profile{}
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.FirstName, &p.Bio, &p.Photos, &p.SelfDescribedFlaws, &p.SelfDescribedStrengths, &p.BirthDate, &p.Gender, &p.Height, &p.Latitude, &p.Longitude, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}
