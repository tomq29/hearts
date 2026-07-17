package user

import "time"

// User represents a user in the system.
// Refresh tokens live in the refresh_tokens table (one row per device).
type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"` // Never expose this in JSON responses
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}
