//go:generate mockery --all --output ./mocks --case underscore --with-expecter
package delete_user_from_segment

import "context"

type SegmentService interface {
	DeleteUserFromSegment(ctx context.Context, userID int64, slugs []string) error
}
