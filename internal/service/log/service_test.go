package log

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	logRepo "github.com/pollykon/avito_test_task/internal/repository/log"
	"github.com/pollykon/avito_test_task/internal/service/log/mocks"
)

func TestLogService_GenerateCSV_Success(t *testing.T) {
	sentFrom := "2023-08"
	sentTo := "2023-09"
	parsedFrom, _ := time.Parse("2006-01", sentFrom)
	parsedTo, _ := time.Parse("2006-01", sentTo)

	sentRequest := GetCSVRequest{
		UserID:    13,
		From:      parsedFrom,
		To:        parsedTo,
		Separator: ",",
	}

	sentCSV := "logId,userId,segmentId,operation,insertTime\n1,12,AVITO,add,2023-08-01T00:00:00Z"

	expectedLogs := []logRepo.Log{
		{
			ID:         1,
			UserID:     int64(12),
			SegmentID:  "AVITO",
			Operation:  logRepo.OperationTypeAdd,
			InsertTime: parsedFrom,
		},
	}

	expectedFileName := "log.csv"

	logRepoMock := mocks.NewLogRepository(t)
	logRepoMock.EXPECT().Get(context.Background(), sentRequest.UserID, sentRequest.From, sentRequest.To).
		Return(expectedLogs, nil)

	csvRepoMock := mocks.NewCSVRepository(t)
	csvRepoMock.EXPECT().Save(sentCSV).Return(expectedFileName, nil)

	service := New(logRepoMock, csvRepoMock)

	fileName, err := service.GenerateCSV(context.Background(), sentRequest)

	assert.NoError(t, err)
	assert.Equal(t, expectedFileName, fileName)
}

func TestLogService_GenerateCSV_Error(t *testing.T) {
	parsedFrom, _ := time.Parse("2006-01", "2023-08")
	parsedTo, _ := time.Parse("2006-01", "2023-09")
	sentRequest := GetCSVRequest{
		UserID:    13,
		From:      parsedFrom,
		To:        parsedTo,
		Separator: ",",
	}
	sentCSV := "logId,userId,segmentId,operation,insertTime\n1,12,AVITO,add,2023-08-01T00:00:00Z"

	errFromLogRepo := fmt.Errorf("error from log repo")
	errFromCSVRepo := fmt.Errorf("error from csv repo")

	expectedLogs := []logRepo.Log{
		{
			ID:         1,
			UserID:     int64(12),
			SegmentID:  "AVITO",
			Operation:  logRepo.OperationTypeAdd,
			InsertTime: parsedFrom,
		},
	}

	tt := []struct {
		name string

		sentRequest GetCSVRequest
		sentCSV     string

		buildLogRepoMock func(mock *mocks.LogRepository)
		buildCSVRepoMock func(mock *mocks.CSVRepository)

		expectedLogs     []logRepo.Log
		expectedFileName string
		expectedError    error
	}{
		{
			name: "unexpected_error_from_log_repo",

			sentRequest: sentRequest,
			sentCSV:     sentCSV,

			buildLogRepoMock: func(repo *mocks.LogRepository) {
				repo.EXPECT().Get(context.Background(), sentRequest.UserID, sentRequest.From, sentRequest.To).
					Return(nil, errFromLogRepo)
			},
			buildCSVRepoMock: nil,

			expectedLogs:     nil,
			expectedFileName: "",
			expectedError:    errFromLogRepo,
		},
		{
			name: "unexpected_error_from_csv_repo",

			sentRequest: sentRequest,
			sentCSV:     sentCSV,

			buildLogRepoMock: func(repo *mocks.LogRepository) {
				repo.EXPECT().Get(context.Background(), sentRequest.UserID, sentRequest.From, sentRequest.To).
					Return(expectedLogs, nil)
			},
			buildCSVRepoMock: func(repo *mocks.CSVRepository) {
				repo.EXPECT().Save(sentCSV).
					Return("", errFromCSVRepo)
			},

			expectedLogs:     expectedLogs,
			expectedFileName: "",
			expectedError:    errFromCSVRepo,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			logRepoMock := mocks.NewLogRepository(t)
			if tc.buildLogRepoMock != nil {
				tc.buildLogRepoMock(logRepoMock)
			}

			csvRepoMock := mocks.NewCSVRepository(t)
			if tc.buildCSVRepoMock != nil {
				tc.buildCSVRepoMock(csvRepoMock)
			}

			service := New(logRepoMock, csvRepoMock)

			fileName, err := service.GenerateCSV(context.Background(), tc.sentRequest)

			assert.ErrorIs(t, err, tc.expectedError)
			assert.Equal(t, "", fileName)
		})
	}
}
