package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/pollykon/avito_test_task/cmd"
	logRepository "github.com/pollykon/avito_test_task/internal/repository/log"
	segmentRepository "github.com/pollykon/avito_test_task/internal/repository/segment"
	deletersService "github.com/pollykon/avito_test_task/internal/service/deleters"
	"github.com/pollykon/avito_test_task/internal/storage"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	config, err := cmd.Load()
	if err != nil {
		panic(err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.ErrorContext(context.Background(), "fail to connect to database", "error", err)
		return
	}

	defer func() { _ = db.Close() }()

	database := storage.New(db)

	logRepo := logRepository.New(database)

	segmentRepo := segmentRepository.New(database)

	cron := deletersService.New(segmentRepo, logRepo)
	ctx := context.Background()

	s := gocron.NewScheduler(time.UTC)

	//cron which deletes segments with flag 'deleted' = true
	_, err = s.Every(config.CronTimeInterval.DeleteSegments).Do(func() {
		logger.InfoContext(ctx, "starting to delete segments")
		err = cron.DeleteSegments(ctx, config.BatchSize.Segments)
		if err != nil {
			logger.ErrorContext(ctx, "error while deleting segments", "error", err)
			return
		}
	})
	if err != nil {
		logger.ErrorContext(ctx, "error while running cron which deletes segments", "error", err)
		return
	}

	// cron which deletes segments with expired ttl
	_, err = s.Every(config.CronTimeInterval.DeleteTTLSegments).Do(func() {
		logger.InfoContext(ctx, "starting to delete segments with ttl")
		err = cron.DeleteTTLSegments(ctx, config.BatchSize.TTLSegments)
		if err != nil {
			logger.ErrorContext(ctx, "error while deleting segments with ttl", "error", err)
			return
		}
	})
	if err != nil {
		logger.ErrorContext(ctx, "error while running cron which deletes segments from user segments", "error", err)
		return
	}
	// cron which deletes old logs (3 month)
	_, err = s.Every(config.CronTimeInterval.DeleteLogs).Do(func() {
		logger.InfoContext(ctx, "starting to delete logs")
		err = cron.DeleteLogs(ctx, config.BatchSize.Logs)
		if err != nil {
			logger.ErrorContext(ctx, "error while deleting logs", "error", err)
			return
		}
	})
	if err != nil {
		logger.ErrorContext(ctx, "error while running cron which deletes logs", "error", err)
		return
	}

	s.StartBlocking()
}
