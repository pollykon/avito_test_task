package delete_user_from_segment

import (
	"context"
	"encoding/json"
	"github.com/pollykon/avito_test_task/internal/handlers"
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

	if request.SegmentSlugs == nil {
		return HandlerResponse{
			Status: http.StatusBadRequest,
			Error: &HandlerResponseError{
				Message: "slugs shouldn't be empty",
			},
		}
	}

	err := h.segmentService.DeleteUserFromSegment(ctx, request.UserID, request.SegmentSlugs)
	if err != nil {
		h.logger.ErrorContext(ctx, "error while deleting user from segment", "error", err, "request", request)
		return HandlerResponse{
			Status: http.StatusInternalServerError,
			Error: &HandlerResponseError{
				Message: handlers.ErrMsgInternal,
			},
		}
	}

	return HandlerResponse{Status: http.StatusOK}
}
