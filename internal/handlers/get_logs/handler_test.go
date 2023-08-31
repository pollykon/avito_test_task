package get_logs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pollykon/avito_test_task/internal/handlers"
	"github.com/pollykon/avito_test_task/internal/handlers/get_logs/mocks"
	"github.com/pollykon/avito_test_task/internal/logger"
	logService "github.com/pollykon/avito_test_task/internal/service/log"
)

const (
	staticURIPrefix = "/static"
)

func TestLogHandler_GetLogs_Success(t *testing.T) {
	sentUserID := 13
	sentFrom := "2023-08"
	sentTo := "2023-09"
	parsedFrom, _ := time.Parse("2006-01", sentFrom)
	parsedTo, _ := time.Parse("2006-01", sentTo)
	separator := ","

	sentRequest := logService.GetCSVRequest{
		UserID:    13,
		From:      parsedFrom,
		To:        parsedTo,
		Separator: ",",
	}

	jsonBodyRequest, _ := json.Marshal(map[string]interface{}{
		"userId":    sentUserID,
		"from":      sentFrom,
		"to":        sentTo,
		"separator": separator,
	})
	request, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:1011/get_user_logs",
		strings.NewReader(string(jsonBodyRequest)),
	)
	if err != nil {
		t.Fatalf("error while sending request: %s", err)
	}

	expectedURI := "http://localhost:1011/static/ef8cde3a-89a1-4fd5-81e2-34dac98a4740.csv"

	w := httptest.NewRecorder()
	logServiceMock := mocks.NewLogService(t)
	logServiceMock.EXPECT().GenerateCSV(context.Background(), sentRequest).
		Return("ef8cde3a-89a1-4fd5-81e2-34dac98a4740.csv", nil)

	handler := New(logServiceMock, staticURIPrefix, slog.New(logger.NewNoopHandler()))
	handler.ServeHTTP(w, request)

	responseResult := w.Result()

	assert.Equal(t, handlers.ContentTypeJSON, responseResult.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, responseResult.StatusCode)

	var response HandlerResponse
	err = json.NewDecoder(responseResult.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, response.Status)
	assert.Equal(t, expectedURI, response.URL)
	assert.Nil(t, response.Error)
}

func TestLogHandler_GetLogs_Error(t *testing.T) {
	sentUserID := 13
	sentFrom := "2023-08"
	sentTo := "2023-09"
	parsedFrom, _ := time.Parse("2006-01", sentFrom)
	parsedTo, _ := time.Parse("2006-01", sentTo)
	separator := ","

	sentRequest := logService.GetCSVRequest{
		UserID:    13,
		From:      parsedFrom,
		To:        parsedTo,
		Separator: separator,
	}

	wrongSentFrom := "123"

	tt := []struct {
		name string

		sentMethod    string
		sentUserID    interface{}
		sentFrom      interface{}
		sentTo        string
		sentSeparator string

		buildLogServiceMock func(mock *mocks.LogService)

		expectedStatusCode int
		expectedResponse   *HandlerResponse
	}{
		{
			name: "wrong_method",

			sentMethod:    http.MethodGet,
			sentUserID:    0,
			sentFrom:      nil,
			sentTo:        "",
			sentSeparator: "",

			buildLogServiceMock: nil,

			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedResponse:   nil,
		},
		{
			name: "decode_error",

			sentMethod:    http.MethodPost,
			sentUserID:    "0",
			sentFrom:      &sentFrom,
			sentTo:        sentTo,
			sentSeparator: separator,

			buildLogServiceMock: nil,

			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   nil,
		},
		{
			name: "wrong_userId",

			sentMethod:    http.MethodPost,
			sentUserID:    -1,
			sentFrom:      &sentFrom,
			sentTo:        sentTo,
			sentSeparator: separator,

			buildLogServiceMock: nil,

			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   nil,
		},
		{
			name: "wrong_from",

			sentMethod:    http.MethodPost,
			sentUserID:    sentUserID,
			sentFrom:      &wrongSentFrom,
			sentTo:        sentTo,
			sentSeparator: separator,

			buildLogServiceMock: nil,

			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   nil,
		},
		{
			name: "from_equal_or_after_to",

			sentMethod:    http.MethodPost,
			sentUserID:    sentUserID,
			sentFrom:      &sentTo,
			sentTo:        sentTo,
			sentSeparator: separator,

			buildLogServiceMock: nil,

			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   nil,
		},
		{
			name: "service_error_unexpected_error",

			sentMethod:    http.MethodPost,
			sentUserID:    sentUserID,
			sentFrom:      &sentFrom,
			sentTo:        sentTo,
			sentSeparator: separator,

			buildLogServiceMock: func(repo *mocks.LogService) {
				repo.EXPECT().GenerateCSV(context.Background(), sentRequest).
					Return("", fmt.Errorf("error from service"))
			},

			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			jsonBodyRequest, _ := json.Marshal(map[string]interface{}{
				"userId":    tc.sentUserID,
				"from":      tc.sentFrom,
				"to":        tc.sentTo,
				"separator": tc.sentSeparator,
			})
			request, err := http.NewRequest(
				tc.sentMethod,
				"http://localhost:1011/get_user_logs",
				strings.NewReader(string(jsonBodyRequest)))
			if err != nil {
				t.Fatalf("error while sending request: %s", err)
			}

			w := httptest.NewRecorder()
			logServiceMock := mocks.NewLogService(t)

			if tc.buildLogServiceMock != nil {
				tc.buildLogServiceMock(logServiceMock)
			}

			handler := New(logServiceMock, staticURIPrefix, slog.New(logger.NewNoopHandler()))
			handler.ServeHTTP(w, request)

			responseResult := w.Result()

			assert.Equal(t, handlers.ContentTypeJSON, responseResult.Header.Get("Content-Type"))
			assert.Equal(t, tc.expectedStatusCode, responseResult.StatusCode)

			if tc.expectedResponse != nil {
				var response HandlerResponse
				err = json.NewDecoder(responseResult.Body).Decode(&response)
				assert.NoError(t, err)

				assert.Equal(t, tc.expectedResponse, response)
			}
		})
	}
}
