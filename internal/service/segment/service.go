package segment

import (
	"context"
	"errors"
	"fmt"
	"time"

	logRepository "github.com/pollykon/avito_test_task/internal/repository/log"
	segmentRepository "github.com/pollykon/avito_test_task/internal/repository/segment"
)

type Service struct {
	logRepo     LogRepository
	segmentRepo SegmentRepository
}

func New(logRepo LogRepository, segmentRepo SegmentRepository) Service {
	return Service{logRepo: logRepo, segmentRepo: segmentRepo}
}

func (s Service) AddSegment(ctx context.Context, slug string) error {
	err := s.segmentRepo.AddSegment(ctx, slug)
	if err != nil {
		if errors.Is(err, segmentRepository.ErrSegmentAlreadyExists) {
			return ErrSegmentAlreadyExists
		}
		return fmt.Errorf("error in service while inserting into segment: %w", err)
	}

	return nil
}

func (s Service) DeleteSegment(ctx context.Context, slug string) error {
	err := s.segmentRepo.DeleteSegment(ctx, slug)
	if err != nil {
		if errors.Is(err, segmentRepository.ErrSegmentNotExist) {
			return ErrSegmentNotExist
		}
		return fmt.Errorf("error in service while deleting from segment: %w", err)
	}

	return nil
}

func (s Service) AddUserToSegment(ctx context.Context, userID int64, slugs []string, ttl *time.Duration) error {
	err := s.segmentRepo.AddUserToSegment(ctx, userID, slugs, ttl)
	if err != nil {
		return fmt.Errorf("error in service while adding user to segment: %w", err)
	}

	err = s.logRepo.Add(ctx, userID, slugs, logRepository.OperationTypeAdd)
	if err != nil {
		return fmt.Errorf("error in service while adding log: %w", err)
	}

	return nil
}

func (s Service) DeleteUserFromSegment(ctx context.Context, userID int64, slugs []string) error {
	err := s.segmentRepo.DeleteUserFromSegment(ctx, userID, slugs)
	if err != nil {
		return fmt.Errorf("error in service while deleting user from segment: %w", err)
	}

	err = s.logRepo.Add(ctx, userID, slugs, logRepository.OperationTypeDelete)
	if err != nil {
		return fmt.Errorf("error in service while adding log: %w", err)
	}

	return nil
}

func (s Service) GetUserActiveSegments(ctx context.Context, userID int64) ([]string, error) {
	segments, err := s.segmentRepo.GetUserActiveSegments(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error in service while getting user's segments: %w", err)
	}

	return segments, nil
}
