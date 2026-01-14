package like

import "time"

// Like represents a user interaction (like or pass).
type Like struct {
	ID         string    `json:"id" db:"id"`
	FromUserID string    `json:"fromUserId" db:"from_user_id"`
	ToUserID   string    `json:"toUserId" db:"to_user_id"`
	IsLike     bool      `json:"isLike" db:"is_like"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}
