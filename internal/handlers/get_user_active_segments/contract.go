//go:generate mockery --all --output ./mocks --case underscore --with-expecter
package get_user_active_segments

import "context"

type SegmentService interface {
	GetUserActiveSegments(ctx context.Context, userID int64) ([]string, error)
}
