package get_logs

import (
	"context"
	"encoding/json"
	"github.com/pollykon/avito_test_task/internal/handlers"
	"log/slog"
	"net/http"
	"time"

	logService "github.com/pollykon/avito_test_task/internal/service/log"
)

type Handler struct {
	logService      logService.Service
	staticURIPrefix string
	logger          *slog.Logger
}

func New(logService logService.Service, logger *slog.Logger, staticURIPrefix string) Handler {
	return Handler{
		logService:      logService,
		staticURIPrefix: staticURIPrefix,
		logger:          logger,
	}
}

const defaultSeparator = ","

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", handlers.ContentTypeJSON)

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer func() {
		_ = r.Body.Close()
	}()

	var request HandlerRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "error while parsing request", "error", err, "request", request)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := h.handle(r.Context(), request, r.Host)
	w.WriteHeader(response.Status)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.ErrorContext(r.Context(), "error while encoding response", "error", err, "request", request)
		return
	}
	return
}

func (h Handler) handle(ctx context.Context, request HandlerRequest, host string) HandlerResponse {
	fromTimeRFC, errFrom := time.Parse(time.RFC3339, request.From)
	toTimeRFC, errTo := time.Parse(time.RFC3339, request.To)
	if errFrom != nil || errTo != nil {
		return HandlerResponse{
			Status: http.StatusBadRequest,
			Error: &HandlerResponseError{
				Message: "time must be in RFC3339",
			},
			URI: "",
		}
	}

	requestSeparator := defaultSeparator

	if request.Separator != nil {
		requestSeparator = *request.Separator
	}

	logServiceRequest := logService.GetCSVRequest{
		UserID:    request.UserID,
		From:      fromTimeRFC,
		To:        toTimeRFC,
		Separator: requestSeparator,
	}

	URI, err := h.logService.GenerateCSV(ctx, logServiceRequest)
	if err != nil {
		h.logger.ErrorContext(ctx, "error while getting logs", "error", err, "request", request)
		return HandlerResponse{
			Status: http.StatusInternalServerError,
			Error: &HandlerResponseError{
				Message: "error while getting logs",
			},
			URI: "",
		}
	}

	h.logger.Info(host)

	return HandlerResponse{
		Status: http.StatusOK,
		Error:  nil,
		URI:    "http://" + host + h.staticURIPrefix + "/" + URI,
	}
}
