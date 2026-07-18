package kafka

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/kisssonik/hearts/apps/notification-service/internal/hub"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Consumer struct {
	reader *kafka.Reader
	logger *zap.Logger
	hub    *hub.Hub
}

// ensureTopic creates the topic if it does not exist yet. A consumer group
// that joins before its topic exists is assigned zero partitions and never
// receives anything — even after the topic is auto-created by a producer.
func ensureTopic(brokers []string, topic string, logger *zap.Logger) {
	client := &kafka.Client{Addr: kafka.TCP(brokers...)}
	resp, err := client.CreateTopics(context.Background(), &kafka.CreateTopicsRequest{
		Topics: []kafka.TopicConfig{{Topic: topic, NumPartitions: 1, ReplicationFactor: 1}},
	})
	if err != nil {
		logger.Warn("Failed to ensure Kafka topic exists", zap.String("topic", topic), zap.Error(err))
		return
	}
	for name, topicErr := range resp.Errors {
		if topicErr != nil && !errors.Is(topicErr, kafka.TopicAlreadyExists) {
			logger.Warn("Failed to create Kafka topic", zap.String("topic", name), zap.Error(topicErr))
		}
	}
}

func NewConsumer(brokers []string, topic string, groupID string, logger *zap.Logger, hub *hub.Hub) *Consumer {
	ensureTopic(brokers, topic, logger)
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
