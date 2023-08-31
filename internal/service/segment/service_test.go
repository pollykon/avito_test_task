package segment

import (
	"context"
	"fmt"
	"hash/fnv"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	logRepository "github.com/pollykon/avito_test_task/internal/repository/log"
	segmentRepository "github.com/pollykon/avito_test_task/internal/repository/segment"
	"github.com/pollykon/avito_test_task/internal/service/segment/mocks"
)

func TestService_GetUserActiveSegments_Success(t *testing.T) {
	sentUserID := int64(10)
	hashProcessor := fnv.New32a()
	_, _ = hashProcessor.Write([]byte(strconv.FormatInt(sentUserID, 10)))
	userHash := int64(hashProcessor.Sum32())
	expectedSegments := segmentRepository.UserSegments{
		NewSegments:    []string{"AVITO_VOICE_MESSAGES"},
		ActiveSegments: []string{"AVITO_DISCOUNT_50"},
	}

	segmentRepoMock := mocks.NewSegmentRepository(t)
	logRepoMock := mocks.NewLogRepository(t)

	segmentRepoMock.EXPECT().InTransaction(context.Background(), mock.Anything).
		Run(func(ctx context.Context, f func(context.Context) error) {
			assert.NoError(t, f(ctx))
		}).Return(nil)

	segmentRepoMock.EXPECT().
		GetUserActiveSegments(context.Background(), sentUserID, userHash).
		Return(expectedSegments, nil)

	segmentRepoMock.EXPECT().
		AddUserToSegment(context.Background(), sentUserID, expectedSegments.NewSegments, (*time.Duration)(nil)).
		Return(nil)

	logRepoMock.EXPECT().
		Add(context.Background(), sentUserID, expectedSegments.NewSegments, logRepository.OperationTypeAdd).
		Return(nil)

	service := New(logRepoMock, segmentRepoMock)

	currentSegments, err := service.GetUserActiveSegments(context.Background(), sentUserID)

	expectedActiveSegments := append(expectedSegments.ActiveSegments, expectedSegments.NewSegments...)

	assert.NoError(t, err)
	assert.Equal(t, expectedActiveSegments, currentSegments)
}

func TestService_GetUserActiveSegments_Error(t *testing.T) {
	sentUserID := int64(10)
	hashProcessor := fnv.New32a()
	_, _ = hashProcessor.Write([]byte(strconv.FormatInt(sentUserID, 10)))
	sentUserHash := int64(hashProcessor.Sum32())
	expectedSegments := segmentRepository.UserSegments{
		NewSegments:    []string{"AVITO_VOICE_MESSAGES"},
		ActiveSegments: []string{"AVITO_DISCOUNT_50"},
	}
	expectedErrorFromRepo := fmt.Errorf("error from repository")

	tt := []struct {
		name string

		sentUserID   int64
		sentUserHash int64

		buildSegmentRepoMock func(mock *mocks.SegmentRepository)
		buildLogRepoMock     func(mock *mocks.LogRepository)

		expectedSegments segmentRepository.UserSegments
		expectedError    error
	}{
		{
			name: "unexpected_error_from_transaction",

			sentUserID:   sentUserID,
			sentUserHash: sentUserHash,

			buildSegmentRepoMock: func(repo *mocks.SegmentRepository) {
				repo.EXPECT().InTransaction(context.Background(), mock.Anything).
					Run(func(ctx context.Context, f func(context.Context) error) {
						assert.NoError(t, f(ctx))
					}).Return(expectedErrorFromRepo)

				repo.EXPECT().GetUserActiveSegments(context.Background(), sentUserID, sentUserHash).
					Return(segmentRepository.UserSegments{}, nil)
			},

			expectedSegments: segmentRepository.UserSegments{},
			expectedError:    expectedErrorFromRepo,
		},
		{
			name: "unexpected_error_from_get",

			sentUserID:   sentUserID,
			sentUserHash: sentUserHash,

			buildSegmentRepoMock: func(repo *mocks.SegmentRepository) {
				repo.EXPECT().InTransaction(context.Background(), mock.Anything).
					Run(func(ctx context.Context, f func(context.Context) error) {
						assert.ErrorIs(t, f(ctx), expectedErrorFromRepo)
					}).Return(expectedErrorFromRepo)

				repo.EXPECT().GetUserActiveSegments(context.Background(), sentUserID, sentUserHash).
					Return(segmentRepository.UserSegments{}, expectedErrorFromRepo)
			},

			expectedSegments: segmentRepository.UserSegments{},
			expectedError:    expectedErrorFromRepo,
		},
		{
			name: "unexpected_error_from_add",

			sentUserID:   sentUserID,
			sentUserHash: sentUserHash,

			buildSegmentRepoMock: func(repo *mocks.SegmentRepository) {
				repo.EXPECT().InTransaction(context.Background(), mock.Anything).
					Run(func(ctx context.Context, f func(context.Context) error) {
						assert.ErrorIs(t, f(ctx), expectedErrorFromRepo)
					}).Return(expectedErrorFromRepo)

				repo.EXPECT().GetUserActiveSegments(context.Background(), sentUserID, sentUserHash).
					Return(expectedSegments, nil)

				repo.EXPECT().AddUserToSegment(
					context.Background(),
					sentUserID,
					expectedSegments.NewSegments,
					(*time.Duration)(nil)).
					Return(expectedErrorFromRepo)
			},

			buildLogRepoMock: func(repo *mocks.LogRepository) {
				repo.EXPECT().
					Add(context.Background(), sentUserID, expectedSegments.NewSegments, logRepository.OperationTypeAdd).
					Return(expectedErrorFromRepo)
			},

			expectedSegments: segmentRepository.UserSegments{},
			expectedError:    expectedErrorFromRepo,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			segmentRepoMock := mocks.NewSegmentRepository(t)

			if tc.buildSegmentRepoMock != nil {
				tc.buildSegmentRepoMock(segmentRepoMock)
			}

			service := New(mocks.NewLogRepository(t), segmentRepoMock)

			currentSegments, err := service.GetUserActiveSegments(context.Background(), tc.sentUserID)

			expectedActiveSegments := append(tc.expectedSegments.ActiveSegments, tc.expectedSegments.NewSegments...)

			assert.ErrorIs(t, err, tc.expectedError)
			assert.Equal(t, currentSegments, expectedActiveSegments)
		})
	}
}

func TestService_AddSegment_Success(t *testing.T) {
	sentSlug := "AVITO"
	sentPercent := int64(2)

	segmentRepoMock := mocks.NewSegmentRepository(t)
	segmentRepoMock.EXPECT().AddSegment(context.Background(), sentSlug, &sentPercent).Return(nil)

	service := New(mocks.NewLogRepository(t), segmentRepoMock)

	err := service.AddSegment(context.Background(), sentSlug, &sentPercent)

	assert.NoError(t, err)
}

func TestService_AddSegment_Error(t *testing.T) {
	expectedErrorFromRepo := fmt.Errorf("error from repository")
	sentPercent := int64(10)

	tt := []struct {
		name string

		sentSlug    string
		sentPercent *int64

		buildMockSegmentRepo func(mock *mocks.SegmentRepository)

		expectedError error
	}{
		{
			name: "unexpected_error_from_repo",

			sentSlug:    "AVITO",
			sentPercent: &sentPercent,

			buildMockSegmentRepo: func(repo *mocks.SegmentRepository) {
				repo.EXPECT().AddSegment(context.Background(), "AVITO", &sentPercent).Return(expectedErrorFromRepo)
			},

			expectedError: expectedErrorFromRepo,
		}, {
			name: "error_from_repo_segment_exists",

			sentSlug:    "AVITO",
			sentPercent: &sentPercent,

			buildMockSegmentRepo: func(repo *mocks.SegmentRepository) {
				repo.EXPECT().AddSegment(context.Background(), "AVITO", &sentPercent).
					Return(segmentRepository.ErrSegmentAlreadyExists)
			},

			expectedError: ErrSegmentAlreadyExists,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			segmentRepoMock := mocks.NewSegmentRepository(t)

			if tc.buildMockSegmentRepo != nil {
				tc.buildMockSegmentRepo(segmentRepoMock)
			}

			service := New(mocks.NewLogRepository(t), segmentRepoMock)

			err := service.AddSegment(context.Background(), tc.sentSlug, tc.sentPercent)

			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestService_DeleteSegment_Success(t *testing.T) {
	sentSlug := "AVITO"

	segmentRepoMock := mocks.NewSegmentRepository(t)
	segmentRepoMock.EXPECT().DeleteSegment(context.Background(), sentSlug).Return(nil)
	service := New(mocks.NewLogRepository(t), segmentRepoMock)

	err := service.DeleteSegment(context.Background(), sentSlug)

	assert.NoError(t, err)
}

func TestService_DeleteSegment_Error(t *testing.T) {
	expectedErrorFromRepo := fmt.Errorf("error from repository")

	tt := []struct {
		name string

		sentSlug string

		buildRepositoryMock func(mock *mocks.SegmentRepository)

		expectedError error
	}{
		{
			name: "unexpected_error_from_repo",

			sentSlug: "AVITO",

			buildRepositoryMock: func(repo *mocks.SegmentRepository) {
				repo.EXPECT().DeleteSegment(context.Background(), "AVITO").
					Return(expectedErrorFromRepo)
			},

			expectedError: expectedErrorFromRepo,
		},
		{
			name: "error_from_repo_segment_not_exist",

			sentSlug: "AVITO",

			buildRepositoryMock: func(repo *mocks.SegmentRepository) {
				repo.EXPECT().DeleteSegment(context.Background(), "AVITO").
					Return(segmentRepository.ErrSegmentNotExist)
			},

			expectedError: ErrSegmentNotExist,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			segmentRepoMock := mocks.NewSegmentRepository(t)

			if segmentRepoMock != nil {
				tc.buildRepositoryMock(segmentRepoMock)
			}

			service := New(mocks.NewLogRepository(t), segmentRepoMock)

			err := service.DeleteSegment(context.Background(), tc.sentSlug)

			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestService_AddUserToSegment_Success(t *testing.T) {
	sentUserID := int64(10)
	sentSlugs := []string{"AVITO"}
	sentTTLToDuration := time.Duration(2) * time.Hour

	segmentRepoMock := mocks.NewSegmentRepository(t)
	segmentRepoMock.EXPECT().InTransaction(context.Background(), mock.Anything).
		Run(func(ctx context.Context, f func(context.Context) error) {
			assert.NoError(t, f(ctx))
		}).Return(nil)

	segmentRepoMock.EXPECT().
		AddUserToSegment(context.Background(), sentUserID, sentSlugs, &sentTTLToDuration).
		Return(nil)

	logRepoMock := mocks.NewLogRepository(t)
	logRepoMock.EXPECT().
		Add(context.Background(), sentUserID, sentSlugs, logRepository.OperationTypeAdd).Return(nil)

	service := New(logRepoMock, segmentRepoMock)

	err := service.AddUserToSegment(context.Background(), int64(sentUserID), sentSlugs, &sentTTLToDuration)

	assert.NoError(t, err)
}

func TestService_AddUserToSegment_Error(t *testing.T) {
	positiveTTL := int64(2)
	positiveTTLDuration := time.Duration(positiveTTL) * time.Hour
	expectedErrorFromRepo := fmt.Errorf("error from segment repository")

	tt := []struct {
		name string

		sentUserID int64
		sentSlugs  []string
		sentTTL    *int64

		buildSegmentRepoMock func(mock *mocks.SegmentRepository)
		buildLogRepoMock     func(mock *mocks.LogRepository)

		expectedErrorFromRepo error
	}{
		{
			name: "unexpected_error_from_transaction",

			sentUserID: int64(2),
			sentSlugs:  []string{"AVITO"},
			sentTTL:    &positiveTTL,

			buildSegmentRepoMock: func(repo *mocks.SegmentRepository) {
				repo.EXPECT().InTransaction(context.Background(), mock.Anything).
					Run(func(ctx context.Context, f func(context.Context) error) {
						assert.NoError(t, f(ctx))
					}).Return(expectedErrorFromRepo)

				repo.EXPECT().AddUserToSegment(context.Background(), int64(2), []string{"AVITO"}, &positiveTTLDuration).
					Return(nil)
			},
			buildLogRepoMock: func(repo *mocks.LogRepository) {
				repo.EXPECT().Add(context.Background(), int64(2), []string{"AVITO"}, logRepository.OperationTypeAdd).
					Return(nil)
			},

			expectedErrorFromRepo: expectedErrorFromRepo,
		},
		{
			name: "unexpected_error_from_segment_repo",

			sentUserID: int64(2),
			sentSlugs:  []string{"AVITO"},
			sentTTL:    &positiveTTL,

			buildSegmentRepoMock: func(repo *mocks.SegmentRepository) {
				repo.EXPECT().InTransaction(context.Background(), mock.Anything).
					Run(func(ctx context.Context, f func(context.Context) error) {
						assert.ErrorIs(t, f(ctx), expectedErrorFromRepo)
					}).Return(expectedErrorFromRepo)

				repo.EXPECT().AddUserToSegment(context.Background(), int64(2), []string{"AVITO"}, &positiveTTLDuration).
					Return(expectedErrorFromRepo)
			},
			buildLogRepoMock: nil,

			expectedErrorFromRepo: expectedErrorFromRepo,
		},
		{
			name: "unexpected_error_from_log_repo",

			sentUserID: int64(2),
			sentSlugs:  []string{"AVITO"},
			sentTTL:    &positiveTTL,

			buildSegmentRepoMock: func(repo *mocks.SegmentRepository) {
				repo.EXPECT().InTransaction(context.Background(), mock.Anything).
					Run(func(ctx context.Context, f func(context.Context) error) {
						assert.ErrorIs(t, f(ctx), expectedErrorFromRepo)
					}).Return(expectedErrorFromRepo)

				repo.EXPECT().AddUserToSegment(context.Background(), int64(2), []string{"AVITO"}, &positiveTTLDuration).
					Return(nil)
			},
			buildLogRepoMock: func(repo *mocks.LogRepository) {
				repo.EXPECT().Add(context.Background(), int64(2), []string{"AVITO"}, logRepository.OperationTypeAdd).
					Return(expectedErrorFromRepo)
			},

			expectedErrorFromRepo: expectedErrorFromRepo,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			positiveTTLDuration := time.Duration(*tc.sentTTL) * time.Hour

			segmentRepoMock := mocks.NewSegmentRepository(t)

			if tc.buildSegmentRepoMock != nil {
				tc.buildSegmentRepoMock(segmentRepoMock)
			}

			logRepoMock := mocks.NewLogRepository(t)

			if tc.buildLogRepoMock != nil {
				tc.buildLogRepoMock(logRepoMock)
			}

			service := New(logRepoMock, segmentRepoMock)

			err := service.AddUserToSegment(context.Background(), tc.sentUserID, tc.sentSlugs, &positiveTTLDuration)

			assert.ErrorIs(t, err, tc.expectedErrorFromRepo)
		})
	}
}

func TestService_DeleteUserFromSegments_Success(t *testing.T) {
	sentUserID := int64(10)
	sentSlugs := []string{"AVITO"}

	segmentRepoMock := mocks.NewSegmentRepository(t)
	segmentRepoMock.EXPECT().InTransaction(context.Background(), mock.Anything).
		Run(func(ctx context.Context, f func(context.Context) error) {
			assert.NoError(t, f(ctx))
		}).Return(nil)

	segmentRepoMock.EXPECT().
		DeleteUserFromSegment(context.Background(), sentUserID, sentSlugs).
		Return(nil)

	logRepoMock := mocks.NewLogRepository(t)
	logRepoMock.EXPECT().Add(context.Background(), sentUserID, sentSlugs, logRepository.OperationTypeDelete).
		Return(nil)

	service := New(logRepoMock, segmentRepoMock)

	err := service.DeleteUserFromSegment(context.Background(), sentUserID, sentSlugs)

	assert.NoError(t, err)
}

func TestService_DeleteUserFromSegments_Error(t *testing.T) {
	expectedErrorFromRepo := fmt.Errorf("error from log repository")

	tt := []struct {
		name string

		sentUserID int64
		sentSlugs  []string

		buildSegmentRepoMock func(mock *mocks.SegmentRepository)
		buildLogRepoMock     func(mock *mocks.LogRepository)

		expectedErrorFromRepo error
	}{
		{
			name: "unexpected_error_from_transaction",

			sentUserID: int64(2),
			sentSlugs:  []string{"AVITO"},

			buildSegmentRepoMock: func(repo *mocks.SegmentRepository) {
				repo.EXPECT().InTransaction(context.Background(), mock.Anything).
					Run(func(ctx context.Context, f func(context.Context) error) {
						assert.NoError(t, f(ctx))
					}).Return(expectedErrorFromRepo)

				repo.EXPECT().DeleteUserFromSegment(context.Background(), int64(2), []string{"AVITO"}).
					Return(nil)
			},
			buildLogRepoMock: func(repo *mocks.LogRepository) {
				repo.EXPECT().Add(context.Background(), int64(2), []string{"AVITO"}, logRepository.OperationTypeDelete).
					Return(nil)
			},

			expectedErrorFromRepo: expectedErrorFromRepo,
		},
		{
			name: "unexpected_error_from_segment_repo",

			sentUserID: int64(2),
			sentSlugs:  []string{"AVITO"},

			buildSegmentRepoMock: func(repo *mocks.SegmentRepository) {
				repo.EXPECT().InTransaction(context.Background(), mock.Anything).
					Run(func(ctx context.Context, f func(context.Context) error) {
						assert.ErrorIs(t, f(ctx), expectedErrorFromRepo)
					}).Return(expectedErrorFromRepo)

				repo.EXPECT().DeleteUserFromSegment(context.Background(), int64(2), []string{"AVITO"}).
					Return(expectedErrorFromRepo)
			},
			buildLogRepoMock: nil,

			expectedErrorFromRepo: expectedErrorFromRepo,
		},
		{
			name: "unexpected_error_from_log_repo",

			sentUserID: int64(2),
			sentSlugs:  []string{"AVITO"},

			buildSegmentRepoMock: func(repo *mocks.SegmentRepository) {
				repo.EXPECT().InTransaction(context.Background(), mock.Anything).
					Run(func(ctx context.Context, f func(context.Context) error) {
						assert.ErrorIs(t, f(ctx), expectedErrorFromRepo)
					}).Return(expectedErrorFromRepo)

				repo.EXPECT().DeleteUserFromSegment(context.Background(), int64(2), []string{"AVITO"}).
					Return(nil)
			},
			buildLogRepoMock: func(repo *mocks.LogRepository) {
				repo.EXPECT().Add(context.Background(), int64(2), []string{"AVITO"}, logRepository.OperationTypeDelete).
					Return(expectedErrorFromRepo)
			},

			expectedErrorFromRepo: expectedErrorFromRepo,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			segmentRepoMock := mocks.NewSegmentRepository(t)

			if tc.buildSegmentRepoMock != nil {
				tc.buildSegmentRepoMock(segmentRepoMock)
			}

			logRepoMock := mocks.NewLogRepository(t)

			if tc.buildLogRepoMock != nil {
				tc.buildLogRepoMock(logRepoMock)
			}

			service := New(logRepoMock, segmentRepoMock)

			err := service.DeleteUserFromSegment(context.Background(), tc.sentUserID, tc.sentSlugs)

			assert.ErrorIs(t, err, tc.expectedErrorFromRepo)
		})
	}
}
