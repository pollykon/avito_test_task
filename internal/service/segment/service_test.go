package segment

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	segmentRepository "github.com/pollykon/avito_test_task/internal/repository/segment"
	"github.com/pollykon/avito_test_task/internal/service/segment/mocks"
)

func TestService_GetUserActiveSegments_Success(t *testing.T) {
	sentUserID := int64(10)
	expectedSegments := []string{"AVITO_VOICE_MESSAGES", "AVITO_DISCOUNT_50"}

	segmentRepoMock := mocks.NewSegmentRepository(t)
	segmentRepoMock.EXPECT().
		GetUserActiveSegments(context.Background(), sentUserID).
		Return(expectedSegments, nil)

	service := New(mocks.NewLogRepository(t), segmentRepoMock)

	currentSegments, err := service.GetUserActiveSegments(context.Background(), sentUserID)

	assert.NoError(t, err)
	assert.Equal(t, expectedSegments, currentSegments)
}

func TestService_GetUserActiveSegments_Error(t *testing.T) {
	sentUserID := int64(10)
	expectedErrorFromRepository := errors.New("error from repository")

	segmentRepoMock := mocks.NewSegmentRepository(t)
	segmentRepoMock.EXPECT().
		GetUserActiveSegments(context.Background(), sentUserID).
		Return(nil, expectedErrorFromRepository)

	service := New(mocks.NewLogRepository(t), segmentRepoMock)

	_, err := service.GetUserActiveSegments(context.Background(), sentUserID)

	assert.ErrorIs(t, err, expectedErrorFromRepository)
}

func TestService_AddSegment_Success(t *testing.T) {
	expectedSlug := "AVITO"

	segmentRepoMock := mocks.NewSegmentRepository(t)
	segmentRepoMock.EXPECT().AddSegment(context.Background(), expectedSlug).Return(nil)
	service := New(mocks.NewLogRepository(t), segmentRepoMock)

	err := service.AddSegment(context.Background(), expectedSlug)

	assert.NoError(t, err)
}

func TestService_AddSegment_Error(t *testing.T) {
	expectedSlug := "AVITO"

	segmentRepoMock := mocks.NewSegmentRepository(t)
	segmentRepoMock.EXPECT().
		AddSegment(context.Background(), expectedSlug).
		Return(segmentRepository.ErrSegmentAlreadyExists)

	service := New(mocks.NewLogRepository(t), segmentRepoMock)

	err := service.AddSegment(context.Background(), expectedSlug)

	assert.ErrorIs(t, err, ErrSegmentAlreadyExists)
}
