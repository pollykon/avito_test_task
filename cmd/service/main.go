package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	handlerAddSegment "github.com/pollykon/avito_test_task/internal/handlers/add_segment"
	handlerAddUserToSegment "github.com/pollykon/avito_test_task/internal/handlers/add_user_to_segments"
	handlerDeleteSegment "github.com/pollykon/avito_test_task/internal/handlers/delete_segment"
	handlerDeleteUserFromSegment "github.com/pollykon/avito_test_task/internal/handlers/delete_user_from_segment"
	handlerGetUserActiveSegment "github.com/pollykon/avito_test_task/internal/handlers/get_user_active_segments"
	logRepository "github.com/pollykon/avito_test_task/internal/repository/log"
	segmentRepository "github.com/pollykon/avito_test_task/internal/repository/segment"
	serviceSegment "github.com/pollykon/avito_test_task/internal/service/segment"
	"log/slog"
	"net/http"
	"os"
)

const servicePort = "1101"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	err := godotenv.Load()
	if err != nil {
		logger.ErrorContext(context.Background(), "fail to load .env", "error", err)
	}

	databaseHost := os.Getenv("PG_HOST")
	databasePort := os.Getenv("PG_PORT")
	databaseUser := os.Getenv("PG_USER")
	databasePassword := os.Getenv("PG_PASSWORD")
	databaseName := os.Getenv("PG_DATABASE_NAME")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		databaseHost,
		databasePort,
		databaseUser,
		databasePassword,
		databaseName,
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

	segmentRepo := segmentRepository.New(db)
	logRepo := logRepository.New(db)

	segmentService := serviceSegment.New(logRepo, segmentRepo)

	segmentAddHandler := handlerAddSegment.New(segmentService, logger)

	segmentDeleteHandler := handlerDeleteSegment.New(segmentService, logger)

	segmentAddUserToSegment := handlerAddUserToSegment.New(segmentService, logger)

	segmentDeleteUserFromSegment := handlerDeleteUserFromSegment.New(segmentService, logger)

	segmentGetUserActiveSegments := handlerGetUserActiveSegment.New(segmentService, logger)

	mux := http.NewServeMux()

	mux.Handle("/add_segment", &segmentAddHandler)
	mux.Handle("/delete_segment", &segmentDeleteHandler)
	mux.Handle("/add_user_to_segments", &segmentAddUserToSegment)
	mux.Handle("/delete_user_from_segment", &segmentDeleteUserFromSegment)
	mux.Handle("/get_user_active_segments", &segmentGetUserActiveSegments)

	logger.InfoContext(context.Background(), "service started", "port", servicePort)
	server := http.Server{
		Addr:    ":" + servicePort,
		Handler: mux,
	}
	err = server.ListenAndServe()
	if err != nil {
		logger.ErrorContext(context.Background(), "error listening", "error", err)
		return
	}
}
