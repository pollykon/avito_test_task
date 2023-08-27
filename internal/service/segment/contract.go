//go:generate mockery --all --output ./mocks --case underscore --with-expecter
package segment

import (
	"context"
	"time"
)

type SegmentRepository interface {
	AddSegment(ctx context.Context, slug string) error
	DeleteSegment(ctx context.Context, slug string) error
	AddUserToSegment(ctx context.Context, userID int64, slugs []string, ttl *time.Duration) error
	DeleteUserFromSegment(ctx context.Context, userID int64, slugs []string) error
	GetUserActiveSegments(ctx context.Context, userID int64) ([]string, error)
}

type LogRepository interface {
	Add(ctx context.Context, userID int64, segment []string, operation string) error
}
