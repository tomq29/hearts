package chat

import "time"

type Message struct {
	ID         string    `json:"id" db:"id"`
	SenderID   string    `json:"senderId" db:"sender_id"`
	ReceiverID string    `json:"receiverId" db:"receiver_id"`
	Content    string    `json:"content" db:"content"`
	IsRead     bool      `json:"isRead" db:"is_read"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}
