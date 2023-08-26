//go:generate mockery --all --output ./mocks --case underscore --with-expecter
package add_user_to_segments

import (
	"context"
	"time"
)

type SegmentService interface {
	AddUserToSegment(ctx context.Context, userID int64, slugs []string, ttl *time.Duration) error
}
