package deleters

import "context"

type SegmentRepository interface {
	DeleteUserSegmentsWithBadTTL(ctx context.Context, limit int64) error
	DeleteSegments(ctx context.Context, limit int64) error
}

type LogRepository interface {
	Delete(ctx context.Context, limit int64) error
}
