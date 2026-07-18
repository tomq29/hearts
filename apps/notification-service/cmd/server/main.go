package main

import (
	"context"
	"net/http"
	"os"
	"strings"

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
		brokers = strings.Split(envBrokers, ",")
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

	// The compose file passes PORT; default matches its published port.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	logger.Info("Notification Service running", zap.String("port", port))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
}
