package add_user_to_segments

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pollykon/avito_test_task/internal/handlers"
	"github.com/pollykon/avito_test_task/internal/handlers/add_user_to_segments/mocks"
	"github.com/pollykon/avito_test_task/internal/logger"
)

func TestSegmentHandler_AddUserToSegmentSegment_Success(t *testing.T) {
	sentSlugs := []string{"AVITO_TEST1", "AVITO_TEST2"}
	sentUserID := int64(10)
	sentTTL := int64(2)

	jsonBodyRequest, _ := json.Marshal(map[string]interface{}{"slugs": sentSlugs, "userId": sentUserID, "ttl": sentTTL})
	request, err := http.NewRequest(http.MethodPost, "", strings.NewReader(string(jsonBodyRequest)))
	if err != nil {
		t.Fatalf("error while sending request: %s", err)
	}

	w := httptest.NewRecorder()
	segmentServiceMock := mocks.NewSegmentService(t)

	sentTTLToDuration := time.Duration(sentTTL) * time.Hour

	segmentServiceMock.EXPECT().AddUserToSegment(context.Background(), sentUserID, sentSlugs, &sentTTLToDuration).Return(nil)

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

func TestSegmentHandler_AddUserToSegment_Error(t *testing.T) {
	negativeTTL := int64(-2)
	positiveTTL := int64(2)
	positiveTTLDuration := time.Duration(positiveTTL) * time.Hour

	tt := []struct {
		name string

		requestMethod string
		sentSlugs     []string
		sentUserID    interface{}
		sentTTL       *int64

		buildSegmentServiceMock func(service *mocks.SegmentService)

		expectedStatusCode int
		expectedResponse   *HandlerResponse
	}{
		{
			name: "wrong_method",

			requestMethod: http.MethodGet,
			sentSlugs:     nil,
			sentUserID:    0,
			sentTTL:       nil,

			buildSegmentServiceMock: nil,

			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedResponse:   nil,
		},
		{
			name: "decode_error",

			requestMethod: http.MethodPost,
			sentSlugs:     []string{"AVITO"},
			sentUserID:    "0",
			sentTTL:       nil,

			buildSegmentServiceMock: nil,

			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   nil,
		},
		{
			name: "wrong_userId",

			requestMethod: http.MethodPost,
			sentSlugs:     []string{"AVITO_TEST1", "AVITO_TEST2"},
			sentUserID:    -1,
			sentTTL:       nil,

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
			sentUserID:    2,
			sentTTL:       nil,

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
			name: "negative_ttl",

			requestMethod: http.MethodPost,
			sentSlugs:     []string{"AVITO_TEST1", "AVITO_TEST2"},
			sentUserID:    2,
			sentTTL:       &negativeTTL,

			buildSegmentServiceMock: nil,

			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: &HandlerResponse{
				Status: http.StatusBadRequest,
				Error: &HandlerResponseError{
					Message: "ttl should be positive",
				},
			},
		},
		{
			name: "service_error_segment_already_exists",

			requestMethod: http.MethodPost,
			sentSlugs:     []string{"AVITO_TEST1", "AVITO_TEST2"},
			sentUserID:    2,
			sentTTL:       &positiveTTL,

			buildSegmentServiceMock: func(service *mocks.SegmentService) {
				service.EXPECT().AddUserToSegment(
					context.Background(), int64(2), []string{"AVITO_TEST1", "AVITO_TEST2"}, &positiveTTLDuration,
				).
					Return(fmt.Errorf("error from service"))
			},

			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse: &HandlerResponse{
				Status: http.StatusInternalServerError,
				Error: &HandlerResponseError{
					Message: "error while adding user to segment",
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			jsonBodyRequest, _ := json.Marshal(
				map[string]interface{}{"slugs": tc.sentSlugs, "userId": tc.sentUserID, "ttl": tc.sentTTL},
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
