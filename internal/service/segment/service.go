package segment

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"
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

func (s Service) AddSegment(ctx context.Context, slug string, percent *int64) error {
	err := s.segmentRepo.AddSegment(ctx, slug, percent)
	if err != nil {
		if errors.Is(err, segmentRepository.ErrSegmentAlreadyExists) {
			return ErrSegmentAlreadyExists
		}
		return fmt.Errorf("error from segment service while inserting into segment: %w", err)
	}

	return nil
}

func (s Service) DeleteSegment(ctx context.Context, slug string) error {
	err := s.segmentRepo.DeleteSegment(ctx, slug)
	if err != nil {
		if errors.Is(err, segmentRepository.ErrSegmentNotExist) {
			return ErrSegmentNotExist
		}
		return fmt.Errorf("error from segment service while deleting from segment: %w", err)
	}

	return nil
}

func (s Service) AddUserToSegment(ctx context.Context, userID int64, slugs []string, ttl *time.Duration) error {
	err := s.segmentRepo.InTransaction(ctx, func(ctx context.Context) error {
		err := s.segmentRepo.AddUserToSegment(ctx, userID, slugs, ttl)
		if err != nil {
			if errors.Is(err, segmentRepository.ErrUserAlreadyInSegment) {
				return ErrUserAlreadyInSegment
			}
			return fmt.Errorf("error from segment service while adding user to segment: %w", err)
		}

		err = s.logRepo.Add(ctx, userID, slugs, logRepository.OperationTypeAdd)
		if err != nil {
			return fmt.Errorf("error from segment service while adding log: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error from segment service in trasaction: %w", err)
	}

	return nil
}

func (s Service) DeleteUserFromSegment(ctx context.Context, userID int64, slugs []string) error {
	err := s.segmentRepo.InTransaction(ctx, func(ctx context.Context) error {
		err := s.segmentRepo.DeleteUserFromSegment(ctx, userID, slugs)
		if err != nil {
			return fmt.Errorf("error from segment service while deleting user from segment: %w", err)
		}

		err = s.logRepo.Add(ctx, userID, slugs, logRepository.OperationTypeDelete)
		if err != nil {
			return fmt.Errorf("error from segment service while adding log: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error from segment service in transaction: %w", err)
	}

	return nil
}

func (s Service) GetUserActiveSegments(ctx context.Context, userID int64) ([]string, error) {
	var activeSegments []string
	err := s.segmentRepo.InTransaction(ctx, func(ctx context.Context) error {
		hashProcessor := fnv.New32a()
		_, _ = hashProcessor.Write([]byte(strconv.FormatInt(userID, 10)))
		userHash := int64(hashProcessor.Sum32())

		segments, err := s.segmentRepo.GetUserActiveSegments(ctx, userID, userHash)
		if err != nil {
			return fmt.Errorf("error from segment service while getting user's segments: %w", err)
		}

		if len(segments.NewSegments) != 0 {
			err = s.AddUserToSegment(ctx, userID, segments.NewSegments, nil)
			if err != nil {
				return fmt.Errorf("error from segment service while adding percent segments: %w", err)
			}
		}

		activeSegments = append(segments.ActiveSegments, segments.NewSegments...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error from segment service in transaction: %w", err)
	}

	return activeSegments, nil
}
