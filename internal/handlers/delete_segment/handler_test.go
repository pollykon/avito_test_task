package delete_segment

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
	"github.com/pollykon/avito_test_task/internal/handlers/delete_segment/mocks"
	"github.com/pollykon/avito_test_task/internal/logger"
	segmentService "github.com/pollykon/avito_test_task/internal/service/segment"
)

func TestSegmentHandler_DeleteSegment_Success(t *testing.T) {
	sentSlug := "AVITO"

	jsonBodyRequest, _ := json.Marshal(map[string]string{"slug": sentSlug})
	request, err := http.NewRequest(http.MethodPost, "", strings.NewReader(string(jsonBodyRequest)))
	if err != nil {
		t.Fatalf("error while sending request: %s", err)
	}

	w := httptest.NewRecorder()
	segmentServiceMock := mocks.NewSegmentService(t)

	segmentServiceMock.EXPECT().DeleteSegment(context.Background(), sentSlug).Return(nil)

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

func TestSegmentHandler_DeleteSegment_Error(t *testing.T) {
	tt := []struct {
		name string

		requestMethod string
		sentSlug      interface{}

		buildSegmentServiceMock func(service *mocks.SegmentService)

		expectedStatusCode int
		expectedResponse   *HandlerResponse
	}{
		{
			name: "wrong_method",

			requestMethod: http.MethodGet,
			sentSlug:      nil,

			buildSegmentServiceMock: nil,

			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedResponse:   nil,
		},
		{
			name: "decode_error",

			requestMethod: http.MethodPost,
			sentSlug:      0,

			buildSegmentServiceMock: nil,

			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   nil,
		},
		{
			name: "empty_slug",

			requestMethod: http.MethodPost,
			sentSlug:      "",

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
			name: "service_error_segment_not_exist",

			requestMethod: http.MethodPost,
			sentSlug:      "AVITO",

			buildSegmentServiceMock: func(service *mocks.SegmentService) {
				service.EXPECT().DeleteSegment(context.Background(), "AVITO").
					Return(segmentService.ErrSegmentNotExist)
			},

			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: &HandlerResponse{
				Status: http.StatusBadRequest,
				Error: &HandlerResponseError{
					Message: "segment doesn't exist",
				},
			},
		},
		{
			name: "service_error_unexpected_error",

			requestMethod: http.MethodPost,
			sentSlug:      "AVITO",

			buildSegmentServiceMock: func(service *mocks.SegmentService) {
				service.EXPECT().DeleteSegment(context.Background(), "AVITO").
					Return(fmt.Errorf("error from service"))
			},

			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse: &HandlerResponse{
				Status: http.StatusInternalServerError,
				Error: &HandlerResponseError{
					Message: "error while deleting segment",
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			jsonBodyRequest, _ := json.Marshal(map[string]interface{}{"slug": tc.sentSlug})
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
