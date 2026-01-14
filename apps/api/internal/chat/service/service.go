package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/kisssonik/hearts/internal/chat"
	"github.com/kisssonik/hearts/internal/chat/repository"
	likeRepo "github.com/kisssonik/hearts/internal/like/repository"
	"github.com/kisssonik/hearts/pkg/websocket"
)

var ErrNotMatched = errors.New("users are not matched")

type ChatService interface {
	SendMessage(ctx context.Context, senderID, receiverID, content string) (*chat.Message, error)
	GetHistory(ctx context.Context, userID, otherUserID string) ([]*chat.Message, error)
}

type chatService struct {
	repo     repository.ChatRepository
	likeRepo likeRepo.LikeRepository
	hub      *websocket.Hub
}

func NewChatService(repo repository.ChatRepository, likeRepo likeRepo.LikeRepository, hub *websocket.Hub) ChatService {
	return &chatService{repo: repo, likeRepo: likeRepo, hub: hub}
}

func (s *chatService) SendMessage(ctx context.Context, senderID, receiverID, content string) (*chat.Message, error) {
	// 1. Check if matched
	isMatch, err := s.likeRepo.HasMutualLike(ctx, senderID, receiverID)
	if err != nil {
		return nil, err
	}
	if !isMatch {
		return nil, ErrNotMatched
	}

	// 2. Save message
	msg := &chat.Message{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
	}
	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}

	// 3. Send via WebSocket
	// We send a JSON payload with type "chat"
	payload := map[string]interface{}{
		"type":    "chat",
		"message": msg,
	}
	bytes, _ := json.Marshal(payload)

	// Send to receiver
	s.hub.SendToUser(receiverID, bytes)
	// Send echo to sender (optional, but good for consistency)
	s.hub.SendToUser(senderID, bytes)

	return msg, nil
}

func (s *chatService) GetHistory(ctx context.Context, userID, otherUserID string) ([]*chat.Message, error) {
	// Check match first? Maybe not strictly necessary for history, but good for privacy.
	isMatch, err := s.likeRepo.HasMutualLike(ctx, userID, otherUserID)
	if err != nil {
		return nil, err
	}
	if !isMatch {
		return nil, ErrNotMatched
	}

	return s.repo.GetMessages(ctx, userID, otherUserID)
}
