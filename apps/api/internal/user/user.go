package user

import "time"

// User represents a user in the system.
type User struct {
	ID                    string     `json:"id" db:"id"`
	Email                 string     `json:"email" db:"email"`
	Username              string     `json:"username" db:"username"`
	PasswordHash          string     `json:"-" db:"password_hash"` // Never expose this in JSON responses
	RefreshTokenHash      *string    `json:"-" db:"refresh_token_hash"`
	RefreshTokenExpiresAt *time.Time `json:"-" db:"refresh_token_expires_at"`
	CreatedAt             time.Time  `json:"createdAt" db:"created_at"`
}
