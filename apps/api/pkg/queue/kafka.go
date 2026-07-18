package queue

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer interface {
	Publish(ctx context.Context, message interface{}) error
	Close() error
}

type Consumer interface {
	Subscribe(ctx context.Context, handler func(ctx context.Context, message []byte) error) error
	Close() error
}

type kafkaProducer struct {
	writer *kafka.Writer
	logger *zap.Logger
}

func NewKafkaProducer(brokers []string, topic string, logger *zap.Logger) Producer {
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &kafkaProducer{writer: w, logger: logger}
}

func (p *kafkaProducer) Publish(ctx context.Context, message interface{}) error {
	bytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx,
		kafka.Message{
			Value: bytes,
		},
	)
}

func (p *kafkaProducer) Close() error {
	return p.writer.Close()
}

type kafkaConsumer struct {
	reader *kafka.Reader
	logger *zap.Logger
}

// EnsureTopic creates the topic if it does not exist yet. A consumer group
// that joins before its topic exists is assigned zero partitions and never
// receives anything — even after the topic is auto-created by a producer —
// so consumers must call this before joining.
func EnsureTopic(brokers []string, topic string, logger *zap.Logger) {
	client := &kafka.Client{Addr: kafka.TCP(brokers...)}
	resp, err := client.CreateTopics(context.Background(), &kafka.CreateTopicsRequest{
		Topics: []kafka.TopicConfig{{Topic: topic, NumPartitions: 1, ReplicationFactor: 1}},
	})
	if err != nil {
		logger.Warn("Failed to ensure Kafka topic exists", zap.String("topic", topic), zap.Error(err))
		return
	}
	for name, topicErr := range resp.Errors {
		// Already existing is the expected steady state, not a problem.
		if topicErr != nil && !errors.Is(topicErr, kafka.TopicAlreadyExists) {
			logger.Warn("Failed to create Kafka topic", zap.String("topic", name), zap.Error(topicErr))
		}
	}
}

func NewKafkaConsumer(brokers []string, topic, groupID string, logger *zap.Logger) Consumer {
	EnsureTopic(brokers, topic, logger)
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	return &kafkaConsumer{reader: r, logger: logger}
}

func (c *kafkaConsumer) Subscribe(ctx context.Context, handler func(ctx context.Context, message []byte) error) error {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			c.logger.Error("Failed to read message", zap.Error(err))
			// If context is cancelled, stop
			if ctx.Err() != nil {
				return ctx.Err()
			}
			time.Sleep(time.Second) // Backoff
			continue
		}

		if err := handler(ctx, m.Value); err != nil {
			c.logger.Error("Failed to handle message", zap.Error(err))
			// In a real app, you might want to retry or DLQ
		}
	}
}

func (c *kafkaConsumer) Close() error {
	return c.reader.Close()
}
