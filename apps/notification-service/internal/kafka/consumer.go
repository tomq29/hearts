package kafka

import (
	"context"
	"encoding/json"

	"github.com/kisssonik/hearts/apps/notification-service/internal/hub"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Consumer struct {
	reader *kafka.Reader
	logger *zap.Logger
	hub    *hub.Hub
}

func NewConsumer(brokers []string, topic string, groupID string, logger *zap.Logger, hub *hub.Hub) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})

	return &Consumer{
		reader: r,
		logger: logger,
		hub:    hub,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	defer c.reader.Close()

	c.logger.Info("Kafka consumer started", zap.String("topic", c.reader.Config().Topic))

	for {
		select {
		case <-ctx.Done():
			return
		default:
			m, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.logger.Error("Failed to read message", zap.Error(err))
				continue
			}

			var payload struct {
				ToUserID string `json:"toUserId"`
				Message  string `json:"message"`
				Type     string `json:"type"`
			}

			if err := json.Unmarshal(m.Value, &payload); err != nil {
				c.logger.Error("Failed to unmarshal", zap.Error(err))
				continue
			}

			c.logger.Info("Received notification",
				zap.String("to", payload.ToUserID),
				zap.String("msg", payload.Message))

			// Send to WebSocket
			c.hub.Send(payload.ToUserID, m.Value)
		}
	}
}
