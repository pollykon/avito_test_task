//go:generate mockery --all --output ./mocks --case underscore --with-expecter
package get_logs

import (
	"context"

	serviceLog "github.com/pollykon/avito_test_task/internal/service/log"
)

type LogService interface {
	GenerateCSV(ctx context.Context, request serviceLog.GetCSVRequest) (string, error)
}
