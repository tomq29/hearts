package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kisssonik/hearts/internal/user/handler"
	"github.com/kisssonik/hearts/internal/user/repository"
	"github.com/kisssonik/hearts/internal/user/service"

	_ "github.com/kisssonik/hearts/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	chatHandler "github.com/kisssonik/hearts/internal/chat/handler"
	chatRepo "github.com/kisssonik/hearts/internal/chat/repository"
	chatService "github.com/kisssonik/hearts/internal/chat/service"

	profileHandler "github.com/kisssonik/hearts/internal/profile/handler"
	profileRepo "github.com/kisssonik/hearts/internal/profile/repository"
	profileService "github.com/kisssonik/hearts/internal/profile/service"

	reviewHandler "github.com/kisssonik/hearts/internal/review/handler"
	reviewRepo "github.com/kisssonik/hearts/internal/review/repository"
	reviewService "github.com/kisssonik/hearts/internal/review/service"

	likeHandler "github.com/kisssonik/hearts/internal/like/handler"
	likeRepo "github.com/kisssonik/hearts/internal/like/repository"
	likeService "github.com/kisssonik/hearts/internal/like/service"

	notificationHandler "github.com/kisssonik/hearts/internal/notification/handler"
	notificationRepo "github.com/kisssonik/hearts/internal/notification/repository"
	notificationService "github.com/kisssonik/hearts/internal/notification/service"

	"github.com/kisssonik/hearts/pkg/auth"
	"github.com/kisssonik/hearts/pkg/config"
	"github.com/kisssonik/hearts/pkg/database"
	"github.com/kisssonik/hearts/pkg/logger"
	"github.com/kisssonik/hearts/pkg/middleware"
	"github.com/kisssonik/hearts/pkg/queue"
	"github.com/kisssonik/hearts/pkg/storage"
	"github.com/kisssonik/hearts/pkg/websocket"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// @title           Hearts API
// @version         1.0
// @description     This is the API for the Hearts dating application.
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	appLogger, err := logger.New(logger.Config{
		Level: cfg.Logger.Level,
		Mode:  cfg.Logger.Mode,
	})
	if err != nil {
		log.Fatalf("could not initialise logger: %v", err)
	}
	defer appLogger.Sync()

	// Run migrations
	db, err := sql.Open("pgx", cfg.Database.URL)
	if err != nil {
		appLogger.Fatal("Could not connect to database for migrations", zap.Error(err))
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		appLogger.Fatal("Failed to set goose dialect", zap.Error(err))
	}

	if err := goose.Up(db, "migrations"); err != nil {
		appLogger.Fatal("Failed to run migrations", zap.Error(err))
	}
	appLogger.Info("Migrations applied successfully")
	db.Close()

	dbPool, err := database.NewPostgresPool(context.Background(), cfg.Database)
	if err != nil {
		appLogger.Fatal("Could not connect to the database", zap.Error(err))
	}
	defer dbPool.Close()
	appLogger.Info("Successfully connected to the database")

	storageProvider, err := storage.NewMinioProvider(storage.Config{
		Endpoint:        cfg.Storage.Endpoint,
		AccessKeyID:     cfg.Storage.AccessKeyID,
		SecretAccessKey: cfg.Storage.SecretAccessKey,
		UseSSL:          cfg.Storage.UseSSL,
		BucketName:      cfg.Storage.BucketName,
		Location:        cfg.Storage.Location,
	})
	if err != nil {
		appLogger.Fatal("Could not initialise storage provider", zap.Error(err))
	}

	authService, err := auth.NewAuthService(cfg.Auth.JWTSecret)
	if err != nil {
		appLogger.Fatal("Could not create auth service", zap.Error(err))
	}

	// Kafka Setup
	kafkaProducer := queue.NewKafkaProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic, appLogger)
	defer kafkaProducer.Close()

	// Notification Producer (Topic: notifications)
	notificationProducer := queue.NewKafkaProducer(cfg.Kafka.Brokers, "notifications", appLogger)
	defer notificationProducer.Close()

	kafkaConsumer := queue.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID, appLogger)
	defer kafkaConsumer.Close()

	// WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	ticketStore := websocket.NewTicketStore()

	userRepo := repository.NewUserRepository(dbPool)
	userService := service.NewUserService(userRepo, authService)
	userHandler := handler.NewUserHandler(userService, authService, appLogger)

	pRepo := profileRepo.NewProfileRepository(dbPool)
	pService := profileService.NewProfileService(pRepo)
	pHandler := profileHandler.NewProfileHandler(pService, storageProvider, appLogger)

	nRepo := notificationRepo.NewNotificationRepository(dbPool)
	nService := notificationService.NewNotificationService(nRepo, wsHub, notificationProducer)
	nHandler := notificationHandler.NewNotificationHandler(nService, appLogger)

	lRepo := likeRepo.NewLikeRepository(dbPool)
	lService := likeService.NewLikeService(lRepo, pRepo, nService, kafkaProducer)
	lHandler := likeHandler.NewLikeHandler(lService, storageProvider, appLogger)

	rRepo := reviewRepo.NewReviewRepository(dbPool)
	rService := reviewService.NewReviewService(rRepo, lRepo)
	rHandler := reviewHandler.NewReviewHandler(rService, appLogger)

	cRepo := chatRepo.NewChatRepository(dbPool)
	cService := chatService.NewChatService(cRepo, lRepo, wsHub)
	cHandler := chatHandler.NewChatHandler(cService, appLogger)

	// Register WS Message Handler
	wsHub.SetMessageHandler(func(senderID string, message []byte) {
		var payload struct {
			Type     string `json:"type"`
			ToUserID string `json:"toUserId"`
			Content  string `json:"content"`
		}
		if err := json.Unmarshal(message, &payload); err != nil {
			appLogger.Error("Failed to unmarshal WS message", zap.Error(err))
			return
		}

		if payload.Type == "chat" {
			_, err := cService.SendMessage(context.Background(), senderID, payload.ToUserID, payload.Content)
			if err != nil {
				appLogger.Error("Failed to send chat message", zap.Error(err))
				// Optionally send error back to user
			}
		} else if payload.Type == "typing" {
			// Forward typing event to the target user
			wsHub.SendToUser(payload.ToUserID, message)
		}
	})

	// Start Match Worker
	go func() {
		appLogger.Info("Starting match worker")
		ctx := context.Background()
		err := kafkaConsumer.Subscribe(ctx, func(ctx context.Context, msg []byte) error {
			var matchCheck likeService.MatchCheckMessage
			if err := json.Unmarshal(msg, &matchCheck); err != nil {
				return err
			}

			appLogger.Info("Processing match check",
				zap.String("from", matchCheck.FromUserID),
				zap.String("target", matchCheck.TargetID))

			return lService.ProcessMatchCheck(ctx, matchCheck.FromUserID, matchCheck.TargetID)
		})
		if err != nil {
			appLogger.Error("Consumer stopped", zap.Error(err))
		}
	}()

	authMiddleware := auth.Middleware(authService, appLogger)

	mux := http.NewServeMux()

	// Health Check
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := dbPool.Ping(r.Context()); err != nil {
			appLogger.Error("Health check failed: database unreachable", zap.Error(err))
			http.Error(w, "Database unreachable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Swagger
	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	mux.HandleFunc("POST /api/v1/users/register", userHandler.Register)
	mux.HandleFunc("POST /api/v1/users/login", userHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", userHandler.RefreshToken)

	// Protected routes
	mux.Handle("GET /api/v1/users/me", authMiddleware(http.HandlerFunc(userHandler.Me)))

	mux.Handle("POST /api/v1/profiles", authMiddleware(http.HandlerFunc(pHandler.Create)))
	mux.Handle("PUT /api/v1/profiles", authMiddleware(http.HandlerFunc(pHandler.Update)))
	mux.Handle("POST /api/v1/profiles/upload", authMiddleware(http.HandlerFunc(pHandler.UploadPhoto)))
	mux.Handle("GET /api/v1/profiles/search", authMiddleware(http.HandlerFunc(pHandler.Search)))
	mux.HandleFunc("GET /api/v1/profiles/{userID}", pHandler.Get)

	mux.Handle("POST /api/v1/likes", authMiddleware(http.HandlerFunc(lHandler.Like)))
	mux.Handle("GET /api/v1/matches", authMiddleware(http.HandlerFunc(lHandler.GetMatches)))

	mux.Handle("GET /api/v1/notifications", authMiddleware(http.HandlerFunc(nHandler.List)))

	mux.Handle("POST /api/v1/reviews", authMiddleware(http.HandlerFunc(rHandler.Create)))
	mux.HandleFunc("GET /api/v1/users/{userID}/reviews", rHandler.List)

	mux.Handle("GET /api/v1/chats/{otherUserID}/messages", authMiddleware(http.HandlerFunc(cHandler.GetHistory)))

	// Ticket generation
	mux.Handle("POST /api/v1/chat/ticket", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(auth.UserIDKey).(string)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ticket := ticketStore.Create(userID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"ticket": ticket})
	})))

	// WebSocket
	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		// Try ticket first
		ticket := r.URL.Query().Get("ticket")
		if ticket != "" {
			userID, ok := ticketStore.Validate(ticket)
			if !ok {
				http.Error(w, "Invalid or expired ticket", http.StatusUnauthorized)
				return
			}
			websocket.ServeWs(wsHub, w, r, userID)
			return
		}

		tokenString := r.URL.Query().Get("token")
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := authService.ValidateAccessToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		websocket.ServeWs(wsHub, w, r, claims.UserID)
	})

	srv := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: middleware.CORS(mux),
	}

	go func() {
		appLogger.Info("Server starting", zap.String("address", cfg.Server.Address))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	appLogger.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	appLogger.Info("Server exiting")
}
