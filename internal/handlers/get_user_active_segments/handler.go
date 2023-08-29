package get_user_active_segments

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/pollykon/avito_test_task/internal/handlers"
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
		_ = json.NewEncoder(w).Encode(HandlerResponse{
			Status: http.StatusMethodNotAllowed,
			Error: &HandlerResponseError{
				Message: handlers.ErrMsgMethodNotAllowed,
			},
		})
		return
	}

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

	response := h.handle(r.Context(), request)
	w.WriteHeader(response.Status)
	_ = json.NewEncoder(w).Encode(response)

	return
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

	activeSegments, err := h.segmentService.GetUserActiveSegments(ctx, request.UserID)
	if err != nil {
		h.logger.ErrorContext(ctx, "error while getting active segment", "error", err, "request", request)
		return HandlerResponse{
			Status: http.StatusInternalServerError,
			Error: &HandlerResponseError{
				Message: handlers.ErrMsgInternal,
			},
		}
	}

	return HandlerResponse{Status: http.StatusOK, Segments: activeSegments}
}
