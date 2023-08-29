package add_segment

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/pollykon/avito_test_task/internal/handlers"
	segmentService "github.com/pollykon/avito_test_task/internal/service/segment"
	"log/slog"
	"net/http"
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
	defer func() { _ = r.Body.Close() }()

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
	if request.SegmentSlug == "" {
		return HandlerResponse{
			Status: http.StatusBadRequest,
			Error: &HandlerResponseError{
				Message: "slug shouldn't be empty",
			},
		}
	}

	err := h.segmentService.AddSegment(ctx, request.SegmentSlug)
	if err != nil {
		if errors.Is(err, segmentService.ErrSegmentAlreadyExists) {
			return HandlerResponse{
				Status: http.StatusBadRequest,
				Error: &HandlerResponseError{
					Message: "segment already exists",
				},
			}
		}

		h.logger.ErrorContext(ctx, "error while adding segment", "error", err, "request", request)
		return HandlerResponse{
			Status: http.StatusInternalServerError,
			Error: &HandlerResponseError{
				Message: handlers.ErrMsgInternal,
			},
		}
	}

	return HandlerResponse{Status: http.StatusOK}
}
