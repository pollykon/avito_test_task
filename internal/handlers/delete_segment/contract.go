//go:generate mockery --all --output ./mocks --case underscore --with-expecter
package delete_segment

import "context"

type SegmentService interface {
	DeleteSegment(ctx context.Context, slug string) error
}
