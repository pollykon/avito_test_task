package delete_user_from_segment

import "context"

type SegmentService interface {
	DeleteUserFromSegment(ctx context.Context, userID int64, slugs []string) error
}

type Logger interface {
	ErrorContext(ctx context.Context, msg string, args ...any)
}
