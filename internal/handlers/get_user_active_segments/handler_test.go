package get_user_active_segments

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
	"github.com/pollykon/avito_test_task/internal/handlers/get_user_active_segments/mocks"
	"github.com/pollykon/avito_test_task/internal/logger"
)

func TestSegmentHandler_GetUserActiveSegments_Success(t *testing.T) {
	sentUserID := int64(10)
	expectedSegments := []string{"AVITO_TEST_1", "AVITO_TEST_2"}

	jsonBodyRequest, _ := json.Marshal(map[string]int64{"userId": sentUserID})
	request, err := http.NewRequest(http.MethodPost, "", strings.NewReader(string(jsonBodyRequest)))
	if err != nil {
		t.Fatalf("error while sending request: %s", err)
	}

	w := httptest.NewRecorder()
	segmentServiceMock := mocks.NewSegmentService(t)

	segmentServiceMock.EXPECT().GetUserActiveSegments(context.Background(), sentUserID).Return(expectedSegments, nil)

	handler := New(segmentServiceMock, slog.New(logger.NewNoopHandler()))
	handler.ServeHTTP(w, request)

	responseResult := w.Result()

	assert.Equal(t, handlers.ContentTypeJSON, responseResult.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, responseResult.StatusCode)

	var response HandlerResponse
	err = json.NewDecoder(responseResult.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, response.Status)
	assert.Equal(t, expectedSegments, response.Segments)
	assert.Nil(t, response.Error)
}

func TestSegmentHandler_GetUserActiveSegments_Error(t *testing.T) {
	tt := []struct {
		name string

		requestMethod string
		sentUserID    interface{}

		buildSegmentServiceMock func(service *mocks.SegmentService)

		gotSegments []string

		expectedStatusCode int
		expectedResponse   *HandlerResponse
	}{
		{
			name: "wrong_method",

			requestMethod: http.MethodGet,
			sentUserID:    0,

			buildSegmentServiceMock: nil,

			gotSegments: nil,

			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedResponse:   nil,
		},
		{
			name: "decode_error",

			requestMethod: http.MethodPost,
			sentUserID:    "0",

			buildSegmentServiceMock: nil,

			gotSegments: nil,

			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   nil,
		},
		{
			name: "wrong_userId",

			requestMethod: http.MethodPost,
			sentUserID:    -1,

			buildSegmentServiceMock: nil,

			gotSegments: nil,

			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: &HandlerResponse{
				Status: http.StatusBadRequest,
				Error: &HandlerResponseError{
					Message: "userId should be more than 0",
				},
			},
		},
		{
			name: "service_error_unexpected_error",

			requestMethod: http.MethodPost,
			sentUserID:    2,

			gotSegments: []string{"AVITO_TEST_1", "AVITO_TEST_2"},

			buildSegmentServiceMock: func(service *mocks.SegmentService) {
				service.EXPECT().GetUserActiveSegments(
					context.Background(), int64(2)).
					Return([]string{"AVITO_TEST_1", "AVITO_TEST_2"}, fmt.Errorf("error from service"))
			},

			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse: &HandlerResponse{
				Status: http.StatusInternalServerError,
				Error: &HandlerResponseError{
					Message: "error while getting active segment",
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			jsonBodyRequest, _ := json.Marshal(
				map[string]interface{}{"userId": tc.sentUserID},
			)
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
