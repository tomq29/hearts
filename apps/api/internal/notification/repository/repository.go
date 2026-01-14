package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kisssonik/hearts/internal/notification"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *notification.Notification) error
	GetByUserID(ctx context.Context, userID string) ([]*notification.Notification, error)
	MarkAsRead(ctx context.Context, notificationID string) error
}

type pgxNotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) NotificationRepository {
	return &pgxNotificationRepository{db: db}
}

func (r *pgxNotificationRepository) Create(ctx context.Context, n *notification.Notification) error {
	query := `
		INSERT INTO notifications (user_id, type, message, is_read)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query,
		n.UserID, n.Type, n.Message, n.IsRead,
	).Scan(&n.ID, &n.CreatedAt)
}

func (r *pgxNotificationRepository) GetByUserID(ctx context.Context, userID string) ([]*notification.Notification, error) {
	query := `
		SELECT id, user_id, type, message, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*notification.Notification
	for rows.Next() {
		n := &notification.Notification{}
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Message, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (r *pgxNotificationRepository) MarkAsRead(ctx context.Context, notificationID string) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE id = $1`
	_, err := r.db.Exec(ctx, query, notificationID)
	return err
}
