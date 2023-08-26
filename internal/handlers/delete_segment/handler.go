package delete_segment

import (
	"context"
	"encoding/json"
	"errors"
	segmentService "github.com/pollykon/avito_test_task/internal/service/segment"
	"net/http"
)

type Handler struct {
	segmentService SegmentService
	logger         Logger
}

func New(s SegmentService, l Logger) Handler {
	return Handler{segmentService: s, logger: l}
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

	err := h.segmentService.DeleteSegment(ctx, request.SegmentSlug)
	if err != nil {
		if errors.Is(err, segmentService.ErrSegmentNotExist) {
			return HandlerResponse{
				Status: http.StatusBadRequest,
				Error: &HandlerResponseError{
					Message: "segment doesn't exist",
				},
			}
		}

		h.logger.ErrorContext(ctx, "error while deleting segment", "error", err, "request", request)
		return HandlerResponse{
			Status: http.StatusInternalServerError,
			Error: &HandlerResponseError{
				Message: "error while deleting segment",
			},
		}
	}
	return HandlerResponse{Status: http.StatusOK}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var request HandlerRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "error while parsing request", "error", err, "request", request)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := h.handle(r.Context(), request)
	w.WriteHeader(response.Status)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.ErrorContext(r.Context(), "error while encoding json", "error", err, "request", request)
		return
	}
}