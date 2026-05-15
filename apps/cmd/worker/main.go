package main

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emremy/socialqueue/internal/config"
	"github.com/emremy/socialqueue/internal/platform/database"
	"github.com/emremy/socialqueue/internal/platform/redis"
	"github.com/emremy/socialqueue/internal/queue/jobs"
	"github.com/emremy/socialqueue/internal/queue/worker"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

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

	publisher := jobs.NewPublisher(redisClient)

	jobStore := jobs.NewStore(db)
	jobService := jobs.NewService(jobStore, publisher)

	executor := worker.NewExecutor()

	executor.Register("test", func(ctx context.Context, job jobs.Job) error {
		n := rand.IntN(100) + 1
		println(n)
		return nil
	})

	workerApp := worker.New(worker.WorkerConfig{
		Concurrency:  5,
		PollInterval: 2 * time.Second,
		Logger:       logger,
		Redis:        redisClient,
	}, jobService, executor)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	workerApp.Run(ctx)
}
