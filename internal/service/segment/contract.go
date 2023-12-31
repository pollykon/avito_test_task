//go:generate mockery --all --output ./mocks --case underscore --with-expecter
package segment

import (
	"context"
	"time"

	segmentRepo "github.com/pollykon/avito_test_task/internal/repository/segment"
)

type SegmentRepository interface {
	AddSegment(ctx context.Context, slug string, percent *int64) error
	DeleteSegment(ctx context.Context, slug string) error
	AddUserToSegment(ctx context.Context, userID int64, slugs []string, ttl *time.Duration) error
	DeleteUserFromSegment(ctx context.Context, userID int64, slugs []string) error
	GetUserActiveSegments(ctx context.Context, userID int64, userHash int64) (segmentRepo.UserSegments, error)
	InTransaction(ctx context.Context, f func(ctx context.Context) error) error
}

type LogRepository interface {
	Add(ctx context.Context, userID int64, segment []string, operation string) error
}

type Transaction interface {
	TransactionWrapper(ctx context.Context, f func(ctx context.Context) error) error
}
