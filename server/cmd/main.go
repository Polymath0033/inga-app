package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"example.com/inga/features/auth"
	"example.com/inga/internal/config"
	"example.com/inga/internal/database"
	"example.com/inga/internal/database/models"
	"example.com/inga/internal/logger"
	"example.com/inga/internal/notification/emails"
	"example.com/inga/internal/notification/queue"
	"example.com/inga/server"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	_ = godotenv.Load()

	cfg := config.Config{}
	err := envconfig.Process(ctx, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	logger.Info(ctx, "Configuration loaded", zap.Any("config", cfg))

	db, err := database.Open(ctx, cfg.Database.DSN)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	querier := models.New(db)

	emailer, err := emails.New(cfg.Email)
	if err != nil {
		logger.Warn(ctx, "email sender is disabled", zap.Error(err))
	}

	// Create email queue with 2 worker goroutines
	emailQueue := queue.NewEmailQueue(emailer, 2)
	defer func() {
		// Graceful shutdown: wait up to 5 seconds for pending emails
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := emailQueue.Close(shutdownCtx); err != nil {
			logger.Warn(ctx, "error closing email queue", zap.Error(err))
		}
	}()

	authService := auth.NewService(querier, emailQueue, cfg.Email.VerificationURL)

	authHandler := auth.NewHandler(authService)

	srvHandler := server.New(cfg, querier, authHandler)
	routes, err := srvHandler.BuildRoute()
	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info(ctx, "starting server", zap.String("addr", addr))
	if err := routes.Run(addr); err != nil {
		log.Fatal(err)
	}
}
