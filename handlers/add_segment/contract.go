package add_segment

import "context"

type SegmentService interface {
	AddSegment(ctx context.Context, slug string) error
}

// можно вынести в папку handler?

type Logger interface {
	ErrorContext(ctx context.Context, msg string, args ...any)
}
