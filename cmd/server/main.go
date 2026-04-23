package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

	_ "tracker/docs"
	"tracker/internal/config"
	"tracker/internal/db"
	"tracker/internal/handler"
	"tracker/internal/metrics"
	appmiddleware "tracker/internal/middleware"
	"tracker/internal/repository"
	"tracker/internal/service"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title Cohabit API
// @version 1.0
// @description API для веб-сервиса коллективного трекинга привычек
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Configure slog
	var level slog.Level
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()
	metrics.RegisterDBStats(database)

	// Run migrations
	migrationsDir := "/migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = "./internal/db/migrations"
	}
	if err := db.RunMigrations(database, migrationsDir); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Ensure uploads directory exists
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		slog.Error("failed to create uploads dir", "error", err)
		os.Exit(1)
	}

	// Repositories
	userRepo := repository.NewUserRepository(database)
	tokenRepo := repository.NewTokenRepository(database)
	categoryRepo := repository.NewCategoryRepository(database)
	challengeRepo := repository.NewChallengeRepository(database)
	participantRepo := repository.NewParticipantRepository(database)
	checkInRepo := repository.NewCheckInRepository(database)
	feedRepo := repository.NewFeedRepository(database)
	feedCommentRepo := repository.NewFeedCommentRepository(database)
	commentRepo := repository.NewCommentRepository(database)
	likeRepo := repository.NewLikeRepository(database)
	statsRepo := repository.NewStatsRepository(database)
	badgeRepo := repository.NewBadgeRepository(database)
	notifRepo := repository.NewNotificationRepository(database)

	// Services
	authSvc := service.NewAuthService(userRepo, tokenRepo, cfg.JWTSecret)
	challengeSvc := service.NewChallengeService(challengeRepo, participantRepo, feedRepo)
	checkInSvc := service.NewCheckInService(checkInRepo, challengeRepo, participantRepo, feedRepo)
	feedSvc := service.NewFeedService(feedRepo, participantRepo)
	interactionSvc := service.NewInteractionService(commentRepo, likeRepo)
	statsSvc := service.NewStatsService(statsRepo, challengeRepo)
	badgeSvc := service.NewBadgeService(badgeRepo, checkInRepo, challengeRepo, participantRepo, feedRepo)
	notifSvc := service.NewNotificationService(notifRepo)

	// Wire cross-service dependencies
	badgeSvc.SetNotificationService(notifSvc)
	checkInSvc.SetBadgeService(badgeSvc)

	// Background jobs
	statusUpdater := service.NewStatusUpdater(challengeRepo)
	statusUpdater.Start(context.Background())

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userRepo, statsSvc)
	challengeHandler := handler.NewChallengeHandler(challengeSvc, categoryRepo)
	checkInHandler := handler.NewCheckInHandler(checkInSvc)
	feedHandler := handler.NewFeedHandler(feedSvc)
	interactionHandler := handler.NewInteractionHandler(interactionSvc)
	statsHandler := handler.NewStatsHandler(statsSvc)
	badgeHandler := handler.NewBadgeHandler(badgeSvc)
	notifHandler := handler.NewNotificationHandler(notifSvc)
	feedCommentHandler := handler.NewFeedCommentHandler(feedCommentRepo)

	r := chi.NewRouter()
	r.Use(appmiddleware.CORS)
	r.Use(appmiddleware.Metrics)
	r.Use(appmiddleware.Logging(logger))
	r.Use(chimiddleware.Recoverer)

	// Prometheus metrics endpoint (no auth)
	r.Handle("/metrics", promhttp.Handler())

	// Serve uploaded files
	fileServer := http.StripPrefix("/static/", http.FileServer(http.Dir(uploadDir)))
	r.Handle("/static/*", fileServer)

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
		})

		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.Auth(authSvc))

			r.Get("/users/me", userHandler.GetMe)
			r.Patch("/users/me", userHandler.UpdateMe)
			r.Get("/users/me/stats", statsHandler.PersonalStats)
			r.Get("/users/me/badges", badgeHandler.MyBadges)
			r.Get("/users/{id}/profile", userHandler.GetProfile)
			r.Get("/users/{id}/badges", badgeHandler.UserBadges)

			r.Get("/badges", badgeHandler.ListDefinitions)

			r.Get("/notifications", notifHandler.List)
			r.Get("/notifications/unread-count", notifHandler.UnreadCount)
			r.Patch("/notifications/{id}/read", notifHandler.MarkRead)
			r.Patch("/notifications/read-all", notifHandler.MarkAllRead)

			r.Get("/categories", challengeHandler.ListCategories)

			r.Get("/challenges", challengeHandler.ListChallenges)
			r.Get("/challenges/my", challengeHandler.ListMy)
			r.Post("/challenges", challengeHandler.Create)
			r.Get("/challenges/{id}", challengeHandler.GetByID)
			r.Patch("/challenges/{id}", challengeHandler.Update)
			r.Post("/challenges/{id}/finish", challengeHandler.Finish)
			r.Get("/challenges/{id}/invite-link", challengeHandler.GetInviteLink)
			r.Post("/challenges/{id}/join", challengeHandler.JoinPublic)
			r.Post("/challenges/join/{inviteToken}", challengeHandler.JoinByInvite)
			r.Delete("/challenges/{id}/participants/{userID}", challengeHandler.RemoveParticipant)
			r.Get("/challenges/{id}/leaderboard", statsHandler.Leaderboard)
			r.Get("/challenges/{id}/stats", statsHandler.ChallengeStats)
			r.Get("/challenges/{id}/summary", statsHandler.ChallengeSummary)

			// New check-in endpoints
			r.Post("/challenges/{id}/checkin", checkInHandler.CheckIn)
			r.Delete("/challenges/{id}/checkin", checkInHandler.Undo)
			r.Get("/challenges/{id}/progress", checkInHandler.GetProgress)
			r.Get("/challenges/{id}/checkins", checkInHandler.ListAll)

			r.Get("/challenges/{id}/feed", feedHandler.GetFeed)

			r.Post("/feed/{eventId}/comments", feedCommentHandler.AddComment)
			r.Get("/feed/{eventId}/comments", feedCommentHandler.ListComments)
			r.Delete("/feed/comments/{id}", feedCommentHandler.DeleteComment)

			r.Post("/checkins/{id}/comments", interactionHandler.AddComment)
			r.Get("/checkins/{id}/comments", interactionHandler.GetComments)
			r.Post("/checkins/{id}/like", interactionHandler.ToggleLike)

			r.Post("/uploads", handler.UploadHandler(uploadDir))
		})
	})

	slog.Info("server starting", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
