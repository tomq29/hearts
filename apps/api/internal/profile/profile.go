package profile

import "time"

// Profile represents a user's profile information.
type Profile struct {
	ID                     string     `json:"id" db:"id"`
	UserID                 string     `json:"userId" db:"user_id"`
	FirstName              string     `json:"firstName" db:"first_name"`
	Bio                    string     `json:"bio" db:"bio"`
	Photos                 []string   `json:"photos" db:"photos"`
	SelfDescribedFlaws     []string   `json:"selfDescribedFlaws" db:"self_described_flaws"`
	SelfDescribedStrengths []string   `json:"selfDescribedStrengths" db:"self_described_strengths"`
	BirthDate              *time.Time `json:"birthDate" db:"birth_date"`
	Gender                 *string    `json:"gender" db:"gender"`
	Height                 *int       `json:"height" db:"height"`
	Latitude               *float64   `json:"latitude" db:"latitude"`
	Longitude              *float64   `json:"longitude" db:"longitude"`
	CreatedAt              time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt              time.Time  `json:"updatedAt" db:"updated_at"`
}
