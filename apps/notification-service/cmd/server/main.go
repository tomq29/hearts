package main

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/kisssonik/hearts/apps/notification-service/internal/auth"
	"github.com/kisssonik/hearts/apps/notification-service/internal/handler"
	"github.com/kisssonik/hearts/apps/notification-service/internal/hub"
	"github.com/kisssonik/hearts/apps/notification-service/internal/kafka"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Verifies access tokens minted by the API. Refusing to start without the
	// secret beats silently accepting every request.
	verifier, err := auth.NewVerifier(os.Getenv("JWT_SECRET_KEY"))
	if err != nil {
		logger.Fatal("JWT_SECRET_KEY must be set", zap.Error(err))
	}

	// Initialize Hub
	notificationHub := hub.NewHub()

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
		notificationHub,
	)

	// Start Consumer in background
	go consumer.Start(context.Background())

	// Browser origins allowed to open a notification stream (comma-separated).
	allowedOrigins := []string{"http://localhost:5173", "http://localhost:4173"}
	if envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); envOrigins != "" {
		allowedOrigins = strings.Split(envOrigins, ",")
	}

	// Start HTTP Server
	http.Handle("/notifications/stream", handler.NewStreamHandler(notificationHub, verifier, logger, allowedOrigins))

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
