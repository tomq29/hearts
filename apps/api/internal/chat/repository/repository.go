package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kisssonik/hearts/internal/chat"
)

type ChatRepository interface {
	CreateMessage(ctx context.Context, msg *chat.Message) error
	GetMessages(ctx context.Context, user1ID, user2ID string) ([]*chat.Message, error)
}

type pgxChatRepository struct {
	db *pgxpool.Pool
}

func NewChatRepository(db *pgxpool.Pool) ChatRepository {
	return &pgxChatRepository{db: db}
}

func (r *pgxChatRepository) CreateMessage(ctx context.Context, msg *chat.Message) error {
	query := `
		INSERT INTO messages (sender_id, receiver_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, is_read, created_at
	`
	return r.db.QueryRow(ctx, query, msg.SenderID, msg.ReceiverID, msg.Content).Scan(
		&msg.ID, &msg.IsRead, &msg.CreatedAt,
	)
}

func (r *pgxChatRepository) GetMessages(ctx context.Context, user1ID, user2ID string) ([]*chat.Message, error) {
	query := `
		SELECT id, sender_id, receiver_id, content, is_read, created_at
		FROM messages
		WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, user1ID, user2ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*chat.Message
	for rows.Next() {
		msg := &chat.Message{}
		if err := rows.Scan(
			&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Content, &msg.IsRead, &msg.CreatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
