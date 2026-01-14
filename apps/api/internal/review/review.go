package review

import "time"

// Review represents a review left by one user for another.
type Review struct {
	ID        string    `json:"id" db:"id"`
	AuthorID  string    `json:"authorId" db:"author_id"`
	TargetID  string    `json:"targetId" db:"target_id"`
	Rating    int       `json:"rating" db:"rating"`
	Comment   string    `json:"comment" db:"comment"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}
