package get_logs

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/pollykon/avito_test_task/internal/handlers"
	logService "github.com/pollykon/avito_test_task/internal/service/log"
)

type Handler struct {
	logService      LogService
	staticURIPrefix string
	logger          *slog.Logger
}

func New(logService LogService, staticURIPrefix string, logger *slog.Logger) Handler {
	return Handler{
		logService:      logService,
		staticURIPrefix: staticURIPrefix,
		logger:          logger,
	}
}

const (
	defaultSeparator = ","
	schema           = "http://"
)

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", handlers.ContentTypeJSON)

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(HandlerResponse{
			Status: http.StatusMethodNotAllowed,
			Error: &HandlerResponseError{
				Message: handlers.ErrMsgMethodNotAllowed,
			},
		})
		return
	}

	defer func() {
		_ = r.Body.Close()
	}()

	var request HandlerRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(HandlerResponse{
			Status: http.StatusBadRequest,
			Error: &HandlerResponseError{
				Message: handlers.ErrMsgBadRequest,
			},
		})
		return
	}

	response := h.handle(r.Context(), request, r.Host)
	w.WriteHeader(response.Status)
	_ = json.NewEncoder(w).Encode(response)
	return
}

func (h Handler) handle(ctx context.Context, request HandlerRequest, host string) HandlerResponse {
	if request.UserID <= 0 {
		return HandlerResponse{
			Status: http.StatusBadRequest,
			Error: &HandlerResponseError{
				Message: "userId should be more than 0",
			},
			URL: "",
		}
	}

	fromTimeRFC, errFrom := time.Parse(time.RFC3339, request.From)
	toTimeRFC, errTo := time.Parse(time.RFC3339, request.To)
	if errFrom != nil || errTo != nil {
		return HandlerResponse{
			Status: http.StatusBadRequest,
			Error: &HandlerResponseError{
				Message: "time must be in RFC3339",
			},
			URL: "",
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
				Message: handlers.ErrMsgInternal,
			},
			URL: "",
		}
	}
	return HandlerResponse{
		Status: http.StatusOK,
		Error:  nil,
		URL:    schema + host + h.staticURIPrefix + "/" + URI,
	}
}
