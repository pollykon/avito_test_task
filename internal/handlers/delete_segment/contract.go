package delete_segment

import "context"

type SegmentService interface {
	DeleteSegment(ctx context.Context, slug string) error
}

type Logger interface {
	ErrorContext(ctx context.Context, msg string, args ...any)
}
