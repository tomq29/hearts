package main

import (
	"context"
	"net/http"
	"os"

	"github.com/kisssonik/hearts/apps/notification-service/internal/handler"
	"github.com/kisssonik/hearts/apps/notification-service/internal/hub"
	"github.com/kisssonik/hearts/apps/notification-service/internal/kafka"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Initialize Hub
	wsHub := hub.NewHub()

	// Initialize Kafka Consumer
	brokers := []string{"kafka:9092"}
	if envBrokers := os.Getenv("KAFKA_BROKERS"); envBrokers != "" {
		brokers = []string{envBrokers}
	}

	consumer := kafka.NewConsumer(
		brokers,
		"notifications",
		"notification-service-group",
		logger,
		wsHub,
	)

	// Start Consumer in background
	go consumer.Start(context.Background())

	// Initialize WebSocket Handler
	wsHandler := handler.NewWebSocketHandler(wsHub, logger)

	// Start HTTP Server
	http.HandleFunc("/ws/notifications", wsHandler.Handle)

	logger.Info("Notification Service running on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
}
