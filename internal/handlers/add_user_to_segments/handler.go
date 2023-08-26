package add_user_to_segments

import (
	"context"
	"encoding/json"
	"github.com/pollykon/avito_test_task/internal/handlers"
	"log/slog"
	"net/http"
	"time"
)

type Handler struct {
	segmentService SegmentService
	logger         *slog.Logger
}

func New(s SegmentService, l *slog.Logger) Handler {
	return Handler{segmentService: s, logger: l}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", handlers.ContentTypeJSON)

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer func() { _ = r.Body.Close() }()

	var request HandlerRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "error while parse request", "error", err, "request", request)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := h.handle(r.Context(), request)
	w.WriteHeader(response.Status)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.ErrorContext(r.Context(), "error while encoding response", "error", err, "request", request)
		return
	}
}

func (h Handler) handle(ctx context.Context, request HandlerRequest) HandlerResponse {
	if request.UserID <= 0 {
		return HandlerResponse{
			Status: http.StatusBadRequest,
			Error: &HandlerResponseError{
				Message: "userId should be more than 0",
			},
		}
	}

	if request.SegmentSlugs == nil {
		return HandlerResponse{
			Status: http.StatusBadRequest,
			Error: &HandlerResponseError{
				Message: "slugs shouldn't be empty",
			},
		}
	}

	if request.TTLHours != nil && *request.TTLHours <= 0 {
		return HandlerResponse{
			Status: http.StatusBadRequest,
			Error: &HandlerResponseError{
				Message: "ttl should be positive",
			},
		}
	}

	var ttlDuration *time.Duration
	if request.TTLHours != nil {
		ttl := time.Duration(*request.TTLHours) * time.Hour
		ttlDuration = &ttl
	}
	err := h.segmentService.AddUserToSegment(context.Background(), request.UserID, request.SegmentSlugs, ttlDuration)
	if err != nil {
		h.logger.ErrorContext(ctx, "error while adding user to segment", "error", err, "request", request)
		return HandlerResponse{
			Status: http.StatusInternalServerError,
			Error: &HandlerResponseError{
				Message: "error while adding user to segment",
			},
		}
	}

	return HandlerResponse{Status: http.StatusOK}
}
