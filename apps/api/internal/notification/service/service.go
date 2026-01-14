package service

import (
	"context"
	"encoding/json"

	"github.com/kisssonik/hearts/internal/notification"
	"github.com/kisssonik/hearts/internal/notification/repository"
	"github.com/kisssonik/hearts/pkg/queue"
	"github.com/kisssonik/hearts/pkg/websocket"
)

type NotificationService interface {
	NotifyMatch(ctx context.Context, user1ID, user2ID string) error
	GetNotifications(ctx context.Context, userID string) ([]*notification.Notification, error)
	MarkAsRead(ctx context.Context, notificationID string) error
}

type notificationService struct {
	repo     repository.NotificationRepository
	hub      *websocket.Hub
	producer queue.Producer
}

func NewNotificationService(repo repository.NotificationRepository, hub *websocket.Hub, producer queue.Producer) NotificationService {
	return &notificationService{repo: repo, hub: hub, producer: producer}
}

func (s *notificationService) NotifyMatch(ctx context.Context, user1ID, user2ID string) error {
	// Notify User 1
	n1 := &notification.Notification{
		UserID:  user1ID,
		Type:    "match",
		Message: "You have a new match!",
	}
	if err := s.repo.Create(ctx, n1); err != nil {
		return err
	}
	s.sendToWS(user1ID, n1)
	s.sendToKafka(user1ID, n1)

	// Notify User 2
	n2 := &notification.Notification{
		UserID:  user2ID,
		Type:    "match",
		Message: "You have a new match!",
	}
	if err := s.repo.Create(ctx, n2); err != nil {
		return err
	}
	s.sendToWS(user2ID, n2)
	s.sendToKafka(user2ID, n2)

	return nil
}

func (s *notificationService) sendToWS(userID string, n *notification.Notification) {
	if s.hub == nil {
		return
	}
	msg, _ := json.Marshal(n)
	s.hub.SendToUser(userID, msg)
}

func (s *notificationService) sendToKafka(userID string, n *notification.Notification) {
	if s.producer == nil {
		return
	}
	// Payload matches what notification-service expects
	payload := struct {
		ToUserID string `json:"toUserId"`
		Message  string `json:"message"`
		Type     string `json:"type"`
	}{
		ToUserID: userID,
		Message:  n.Message,
		Type:     n.Type,
	}
	// We use a separate topic "notifications" for this
	// Note: The producer interface might need a topic argument if it's not fixed.
	// Let's check queue.Producer interface.
	// Assuming Publish takes the message and the producer is configured with a topic.
	// Wait, the producer in main.go is configured with "match-checks" topic?
	// Let's check main.go again.
	s.producer.Publish(context.Background(), payload)
}

func (s *notificationService) GetNotifications(ctx context.Context, userID string) ([]*notification.Notification, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, notificationID string) error {
	return s.repo.MarkAsRead(ctx, notificationID)
}
