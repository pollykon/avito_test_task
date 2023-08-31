package add_segment

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pollykon/avito_test_task/internal/handlers"
	"github.com/pollykon/avito_test_task/internal/handlers/add_segment/mocks"
	"github.com/pollykon/avito_test_task/internal/logger"
	segmentService "github.com/pollykon/avito_test_task/internal/service/segment"
)

func TestSegmentHandler_AddSegment_Success(t *testing.T) {
	sentSlug := "AVITO"
	sentPercent := int64(10)

	jsonBodyRequest, _ := json.Marshal(map[string]interface{}{"slug": sentSlug, "percent": sentPercent})
	request, err := http.NewRequest(http.MethodPost, "", strings.NewReader(string(jsonBodyRequest)))
	if err != nil {
		t.Fatalf("error while sending request: %s", err)
	}

	w := httptest.NewRecorder()
	segmentServiceMock := mocks.NewSegmentService(t)

	segmentServiceMock.EXPECT().AddSegment(context.Background(), sentSlug, &sentPercent).Return(nil)

	handler := New(segmentServiceMock, slog.New(logger.NewNoopHandler()))
	handler.ServeHTTP(w, request)

	responseResult := w.Result()

	assert.Equal(t, handlers.ContentTypeJSON, responseResult.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, responseResult.StatusCode)

	var response HandlerResponse
	err = json.NewDecoder(responseResult.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, response.Status)
	assert.Nil(t, response.Error)
}

func TestSegmentHandler_AddSegment_Error(t *testing.T) {
	sentPercent := int64(10)

	tt := []struct {
		name string

		requestMethod string
		sentSlug      interface{}
		sentPercent   *int64

		buildSegmentServiceMock func(service *mocks.SegmentService)

		expectedStatusCode int
		expectedResponse   *HandlerResponse
	}{
		{
			name: "wrong_method",

			requestMethod: http.MethodGet,
			sentSlug:      nil,
			sentPercent:   nil,

			buildSegmentServiceMock: nil,

			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedResponse:   nil,
		},
		{
			name: "decode_error",

			requestMethod: http.MethodPost,
			sentSlug:      0,
			sentPercent:   nil,

			buildSegmentServiceMock: nil,

			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   nil,
		},
		{
			name: "empty_slug",

			requestMethod: http.MethodPost,
			sentSlug:      "",
			sentPercent:   nil,

			buildSegmentServiceMock: nil,

			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: &HandlerResponse{
				Status: http.StatusBadRequest,
				Error: &HandlerResponseError{
					Message: "slug shouldn't be empty",
				},
			},
		},
		{
			name: "service_error_segment_already_exists",

			requestMethod: http.MethodPost,
			sentSlug:      "AVITO",
			sentPercent:   &sentPercent,

			buildSegmentServiceMock: func(service *mocks.SegmentService) {
				service.EXPECT().AddSegment(context.Background(), "AVITO", &sentPercent).
					Return(segmentService.ErrSegmentAlreadyExists)
			},

			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: &HandlerResponse{
				Status: http.StatusBadRequest,
				Error: &HandlerResponseError{
					Message: handlers.ErrMsgBadRequest,
				},
			},
		},
		{
			name: "service_error_unexpected_error",

			requestMethod: http.MethodPost,
			sentSlug:      "AVITO",
			sentPercent:   &sentPercent,

			buildSegmentServiceMock: func(service *mocks.SegmentService) {
				service.EXPECT().AddSegment(context.Background(), "AVITO", &sentPercent).
					Return(fmt.Errorf("error from service"))
			},

			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse: &HandlerResponse{
				Status: http.StatusInternalServerError,
				Error: &HandlerResponseError{
					Message: handlers.ErrMsgInternal,
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			jsonBodyRequest, _ := json.Marshal(map[string]interface{}{"slug": tc.sentSlug, "percent": tc.sentPercent})
			request, err := http.NewRequest(tc.requestMethod, "", strings.NewReader(string(jsonBodyRequest)))
			if err != nil {
				t.Fatalf("error while sending request: %s", err)
			}

			w := httptest.NewRecorder()
			segmentServiceMock := mocks.NewSegmentService(t)

			if tc.buildSegmentServiceMock != nil {
				tc.buildSegmentServiceMock(segmentServiceMock)
			}

			handler := New(segmentServiceMock, slog.New(logger.NewNoopHandler()))
			handler.ServeHTTP(w, request)

			responseResult := w.Result()

			assert.Equal(t, handlers.ContentTypeJSON, responseResult.Header.Get("Content-Type"))
			assert.Equal(t, tc.expectedStatusCode, responseResult.StatusCode)

			if tc.expectedResponse != nil {
				var response HandlerResponse
				err = json.NewDecoder(responseResult.Body).Decode(&response)
				assert.NoError(t, err)

				assert.Equal(t, *tc.expectedResponse, response)
			}
		})
	}
}
