//go:generate mockery --all --output ./mocks --case underscore --with-expecter
package log

import (
	"context"
	"time"

	"github.com/pollykon/avito_test_task/internal/repository/log"
)

type LogRepository interface {
	Get(ctx context.Context, userID int64, from time.Time, to time.Time) ([]log.Log, error)
}

type CSVRepository interface {
	Save(csv string) (string, error)
}
