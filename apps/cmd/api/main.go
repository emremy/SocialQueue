package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	appauth "github.com/emremy/socialqueue/internal/app/auth"
	"github.com/emremy/socialqueue/internal/config"
	"github.com/emremy/socialqueue/internal/platform/database"
	"github.com/emremy/socialqueue/internal/platform/redis"
	"github.com/emremy/socialqueue/internal/transport/httpserver"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()
	accessTTL, err := time.ParseDuration(cfg.AccessTokenTTL)
	if err != nil {
		log.Fatal(err)
	}

	refreshTTL, err := time.ParseDuration(cfg.RefreshTokenTTL)
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewPostgres(cfg)
	if err != nil {
		slog.Error("failed to connect postgres", "error", err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("failed to get sql db", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	redisClient := redis.NewClient(cfg.RedisAddr)

	authStore := appauth.NewStore(db)

	tokenManager := appauth.CreateTokenManager(
		cfg.JwtSecret,
		accessTTL,
		refreshTTL,
	)

	router := httpserver.CreateRouter(db, redisClient, authStore, tokenManager)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.AppPort),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("social queue api started", "env", cfg.AppEnv, "addr", server.Addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
