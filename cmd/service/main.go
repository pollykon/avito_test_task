package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/pollykon/avito_test_task/cmd"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	handlerAddSegment "github.com/pollykon/avito_test_task/internal/handlers/add_segment"
	handlerAddUserToSegment "github.com/pollykon/avito_test_task/internal/handlers/add_user_to_segments"
	handlerDeleteSegment "github.com/pollykon/avito_test_task/internal/handlers/delete_segment"
	handlerDeleteUserFromSegment "github.com/pollykon/avito_test_task/internal/handlers/delete_user_from_segment"
	handlerGetLogs "github.com/pollykon/avito_test_task/internal/handlers/get_logs"
	handlerGetUserActiveSegment "github.com/pollykon/avito_test_task/internal/handlers/get_user_active_segments"
	csvRepository "github.com/pollykon/avito_test_task/internal/repository/csv"
	logRepository "github.com/pollykon/avito_test_task/internal/repository/log"
	segmentRepository "github.com/pollykon/avito_test_task/internal/repository/segment"
	serviceLog "github.com/pollykon/avito_test_task/internal/service/log"
	serviceSegment "github.com/pollykon/avito_test_task/internal/service/segment"
	"github.com/pollykon/avito_test_task/internal/storage"
)

const (
	servicePort     = "1101"
	staticURIPrefix = "/static"
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

	logCSVDirectory := config.CSV.LogCSVDirectory

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

	if err = db.Ping(); err != nil {
		logger.ErrorContext(context.Background(), "database doesn't response", "error", err)
		return
	}

	database := storage.New(db)

	segmentRepo := segmentRepository.New(database)
	logRepo := logRepository.New(database)
	csvRepo := csvRepository.New(logCSVDirectory)

	segmentService := serviceSegment.New(logRepo, segmentRepo)
	logService := serviceLog.New(logRepo, csvRepo)

	segmentAddHandler := handlerAddSegment.New(segmentService, logger)

	segmentDeleteHandler := handlerDeleteSegment.New(segmentService, logger)

	segmentAddUserToSegment := handlerAddUserToSegment.New(segmentService, logger)

	segmentDeleteUserFromSegment := handlerDeleteUserFromSegment.New(segmentService, logger)

	segmentGetUserActiveSegments := handlerGetUserActiveSegment.New(segmentService, logger)

	logGetLogsHandler := handlerGetLogs.New(logService, staticURIPrefix, logger)

	mux := http.NewServeMux()

	mux.Handle("/add_segment_v1", segmentAddHandler)
	mux.Handle("/delete_segment_v1", segmentDeleteHandler)
	mux.Handle("/add_user_to_segments_v1", segmentAddUserToSegment)
	mux.Handle("/delete_user_from_segment_v1", segmentDeleteUserFromSegment)
	mux.Handle("/get_user_active_segments_v1", segmentGetUserActiveSegments)
	mux.Handle("/get_user_logs_v1", logGetLogsHandler)

	staticHandler := http.StripPrefix(staticURIPrefix, http.FileServer(http.Dir(logCSVDirectory)))
	mux.Handle(staticURIPrefix+"/", staticHandler)

	server := http.Server{
		Addr:    ":" + servicePort,
		Handler: mux,
	}

	go func() {
		logger.InfoContext(context.Background(), "service started", "port", servicePort)
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.ErrorContext(context.Background(), "error while starting server", "error", err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logger.InfoContext(context.Background(), "shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		logger.ErrorContext(context.Background(), "error while shutting down", "error", err)
	}
}
