//go:generate mockery --all --output ./mocks --case underscore --with-expecter
package add_segment

import "context"

type SegmentService interface {
	AddSegment(ctx context.Context, slug string) error
}
