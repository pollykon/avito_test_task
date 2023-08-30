package delete_user_from_segment

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
	"github.com/pollykon/avito_test_task/internal/handlers/delete_user_from_segment/mocks"
	"github.com/pollykon/avito_test_task/internal/logger"
)

func TestSegmentHandler_DeleteUserFromSegment_Success(t *testing.T) {
	sentSlugs := []string{"AVITO"}
	sentUserID := 2

	jsonBodyRequest, _ := json.Marshal(map[string]interface{}{"slugs": sentSlugs, "userId": sentUserID})
	request, err := http.NewRequest(http.MethodPost, "", strings.NewReader(string(jsonBodyRequest)))
	if err != nil {
		t.Fatalf("error while sending request: %s", err)
	}

	w := httptest.NewRecorder()
	segmentServiceMock := mocks.NewSegmentService(t)

	segmentServiceMock.EXPECT().DeleteUserFromSegment(context.Background(), int64(sentUserID), sentSlugs).Return(nil)

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

func TestSegmentHandler_DeleteUserFromSegment_Error(t *testing.T) {
	tt := []struct {
		name string

		requestMethod string
		sentSlugs     []string
		sentUserID    interface{}

		buildSegmentServiceMock func(service *mocks.SegmentService)

		expectedStatusCode int
		expectedResponse   *HandlerResponse
	}{
		{
			name: "wrong_method",

			requestMethod: http.MethodGet,
			sentSlugs:     nil,
			sentUserID:    nil,

			buildSegmentServiceMock: nil,

			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedResponse:   nil,
		},
		{
			name: "decode_error",

			requestMethod: http.MethodPost,
			sentSlugs:     []string{"AVITO"},
			sentUserID:    "2",

			buildSegmentServiceMock: nil,

			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   nil,
		},
		{
			name: "wrong_userId",

			requestMethod: http.MethodPost,
			sentSlugs:     []string{"AVITO"},
			sentUserID:    -1,

			buildSegmentServiceMock: nil,

			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: &HandlerResponse{
				Status: http.StatusBadRequest,
				Error: &HandlerResponseError{
					Message: "userId should be more than 0",
				},
			},
		},
		{
			name: "empty_slugs",

			requestMethod: http.MethodPost,
			sentSlugs:     nil,
			sentUserID:    10,

			buildSegmentServiceMock: nil,

			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: &HandlerResponse{
				Status: http.StatusBadRequest,
				Error: &HandlerResponseError{
					Message: "slugs shouldn't be empty",
				},
			},
		},
		{
			name: "service_error_unexpected_error",

			requestMethod: http.MethodPost,
			sentSlugs:     []string{"AVITO"},
			sentUserID:    3,

			buildSegmentServiceMock: func(service *mocks.SegmentService) {
				service.EXPECT().DeleteUserFromSegment(context.Background(), int64(3), []string{"AVITO"}).
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
			jsonBodyRequest, _ := json.Marshal(map[string]interface{}{"slugs": tc.sentSlugs, "userId": tc.sentUserID})
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
