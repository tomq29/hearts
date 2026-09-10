package service

import (
	"context"
	"errors"
	"time"

	"github.com/kisssonik/hearts/internal/profile"
	"github.com/kisssonik/hearts/internal/profile/repository"
)

var ErrProfileAlreadyExists = errors.New("profile already exists")

type ProfileService interface {
	CreateProfile(ctx context.Context, userID string, input CreateProfileInput) (*profile.Profile, error)
	GetProfileByUserID(ctx context.Context, userID string) (*profile.Profile, error)
	GetProfileForViewer(ctx context.Context, viewerID, targetUserID string) (*profile.Profile, error)
	UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*profile.Profile, error)
	SearchProfiles(ctx context.Context, userID string, params SearchParams) ([]*profile.Profile, error)
}

type SearchParams struct {
	MinAge    *int
	MaxAge    *int
	RadiusKM  *float64
	Gender    *string
	MinHeight *int
	MaxHeight *int
}

type CreateProfileInput struct {
	FirstName              string     `json:"firstName"`
	Bio                    string     `json:"bio"`
	Photos                 []string   `json:"photos"`
	SelfDescribedFlaws     []string   `json:"selfDescribedFlaws"`
	SelfDescribedStrengths []string   `json:"selfDescribedStrengths"`
	BirthDate              *time.Time `json:"birthDate"`
	Gender                 *string    `json:"gender"`
	Height                 *int       `json:"height"`
	Latitude               *float64   `json:"latitude"`
	Longitude              *float64   `json:"longitude"`
}

type UpdateProfileInput struct {
	FirstName              *string    `json:"firstName"`
	Bio                    *string    `json:"bio"`
	Photos                 []string   `json:"photos"`
	SelfDescribedFlaws     []string   `json:"selfDescribedFlaws"`
	SelfDescribedStrengths []string   `json:"selfDescribedStrengths"`
	BirthDate              *time.Time `json:"birthDate"`
	Gender                 *string    `json:"gender"`
	Height                 *int       `json:"height"`
	Latitude               *float64   `json:"latitude"`
	Longitude              *float64   `json:"longitude"`
}

type profileService struct {
	repo   repository.ProfileRepository
	bander *profile.Bander
}

func NewProfileService(repo repository.ProfileRepository, bander *profile.Bander) ProfileService {
	return &profileService{repo: repo, bander: bander}
}

func (s *profileService) CreateProfile(ctx context.Context, userID string, input CreateProfileInput) (*profile.Profile, error) {
	// Check if profile already exists
	_, err := s.repo.GetByUserID(ctx, userID)
	if err == nil {
		return nil, ErrProfileAlreadyExists
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	p := &profile.Profile{
		UserID:                 userID,
		FirstName:              input.FirstName,
		Bio:                    input.Bio,
		Photos:                 input.Photos,
		SelfDescribedFlaws:     input.SelfDescribedFlaws,
		SelfDescribedStrengths: input.SelfDescribedStrengths,
		BirthDate:              input.BirthDate,
		Gender:                 input.Gender,
		Height:                 input.Height,
		Latitude:               input.Latitude,
		Longitude:              input.Longitude,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *profileService) GetProfileByUserID(ctx context.Context, userID string) (*profile.Profile, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// GetProfileForViewer returns someone else's profile as that viewer may see it:
// coordinates are never serialised, so a coarse distance band takes their place.
// Kept separate from GetProfileByUserID, which serves the caller their own
// profile and should return everything.
func (s *profileService) GetProfileForViewer(ctx context.Context, viewerID, targetUserID string) (*profile.Profile, error) {
	target, err := s.repo.GetByUserID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	viewer, err := s.repo.GetByUserID(ctx, viewerID)
	if err != nil {
		// A viewer without a profile still gets to read others, just without
		// a distance to compare against.
		if errors.Is(err, repository.ErrNotFound) {
			return target, nil
		}
		return nil, err
	}

	s.attachDistanceBand(viewer, target)
	return target, nil
}

// attachDistanceBand labels how far target is from viewer. Profiles missing
// either coordinate are left unlabelled, so omitempty drops the field.
func (s *profileService) attachDistanceBand(viewer, target *profile.Profile) {
	if viewer == nil || target == nil {
		return
	}
	if viewer.Latitude == nil || viewer.Longitude == nil ||
		target.Latitude == nil || target.Longitude == nil {
		return
	}
	if viewer.UserID == target.UserID {
		return
	}

	km := profile.DistanceKM(*viewer.Latitude, *viewer.Longitude, *target.Latitude, *target.Longitude)
	band := s.bander.Band(km)
	target.DistanceBand = &band
}

func (s *profileService) UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*profile.Profile, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.FirstName != nil {
		p.FirstName = *input.FirstName
	}
	if input.Bio != nil {
		p.Bio = *input.Bio
	}
	if input.Photos != nil {
		p.Photos = input.Photos
	}
	if input.SelfDescribedFlaws != nil {
		p.SelfDescribedFlaws = input.SelfDescribedFlaws
	}
	if input.SelfDescribedStrengths != nil {
		p.SelfDescribedStrengths = input.SelfDescribedStrengths
	}
	if input.BirthDate != nil {
		p.BirthDate = input.BirthDate
	}
	if input.Gender != nil {
		p.Gender = input.Gender
	}
	if input.Height != nil {
		p.Height = input.Height
	}
	if input.Latitude != nil {
		p.Latitude = input.Latitude
	}
	if input.Longitude != nil {
		p.Longitude = input.Longitude
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *profileService) SearchProfiles(ctx context.Context, userID string, params SearchParams) ([]*profile.Profile, error) {
	// Get current user's profile to know their location
	currentUserProfile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// A radius filter needs the viewer's own location. Without it, drop the
	// filter rather than failing: a filter that cannot apply should not empty
	// the feed.
	radius := params.RadiusKM
	if currentUserProfile.Latitude == nil || currentUserProfile.Longitude == nil {
		radius = nil
	}

	results, err := s.repo.Search(ctx, currentUserProfile, repository.SearchParams{
		MinAge:    params.MinAge,
		MaxAge:    params.MaxAge,
		RadiusKM:  radius,
		Gender:    params.Gender,
		MinHeight: params.MinHeight,
		MaxHeight: params.MaxHeight,
	})
	if err != nil {
		return nil, err
	}

	for _, p := range results {
		s.attachDistanceBand(currentUserProfile, p)
	}
	return results, nil
}
