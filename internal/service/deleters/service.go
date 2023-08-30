package deleters

import (
	"context"
)

type Cron struct {
	segmentRepo SegmentRepository
	logRepo     LogRepository
}

func New(segmentRepo SegmentRepository, logRepo LogRepository) *Cron {
	return &Cron{segmentRepo: segmentRepo, logRepo: logRepo}
}

func (c *Cron) DeleteSegments(ctx context.Context, batchSize int64) error {
	err := c.segmentRepo.DeleteSegments(ctx, batchSize)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cron) DeleteTTLSegments(ctx context.Context, batchSize int64) error {
	err := c.segmentRepo.DeleteUserSegmentsWithBadTTL(ctx, batchSize)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cron) DeleteLogs(ctx context.Context, batchSize int64) error {
	err := c.logRepo.Delete(ctx, batchSize)
	if err != nil {
		return err
	}

	return nil
}
