package get_user_active_segments

import "context"

type SegmentService interface {
	GetUserActiveSegments(ctx context.Context, userID int64) ([]string, error)
}

type Logger interface {
	ErrorContext(ctx context.Context, msg string, args ...any)
}
