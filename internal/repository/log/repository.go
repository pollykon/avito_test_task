package log

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pollykon/avito_test_task/internal/storage"
)

type Repository struct {
	db storage.Database
}

func New(db storage.Database) *Repository {
	return &Repository{
		db: db,
	}
}

func (l *Repository) Add(ctx context.Context, userID int64, segments []string, operation string) error {
	if len(segments) == 0 {
		return nil
	}

	values := make([]string, 0, len(segments))
	queryArgs := make([]interface{}, 0, len(segments)+1)
	queryArgs = append(queryArgs, operation)
	for i, segment := range segments {
		values = append(values, fmt.Sprintf("(%d, $%d, $1)", userID, i+2))
		queryArgs = append(queryArgs, segment)
	}

	query := fmt.Sprintf(`insert into log (user_id, segment_id, operation) values %s`, strings.Join(values, ","))
	_, err := l.db.ExecContext(ctx, query, queryArgs...)
	if err != nil {
		return fmt.Errorf("error while inserting into log: %w", err)
	}

	return nil
}

func (l *Repository) Delete(ctx context.Context, limit int64) error {
	query := `delete from log where id in (select id from log where insert_time + interval '3 months' < now() limit $1)`
	_, err := l.db.ExecContext(ctx, query, limit)
	if err != nil {
		return fmt.Errorf("error while deleting from log: %w", err)
	}

	return nil
}

func (l *Repository) Get(ctx context.Context, userID int64, from time.Time, to time.Time) ([]Log, error) {
	query := `select id, user_id, segment_id, operation, insert_time from log
                  where user_id = $1 
				  and insert_time >= $2
				  and insert_time < $3`

	rows, err := l.db.QueryContext(ctx, query, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("error while getting logs: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var logs []Log
	for rows.Next() {
		var log = Log{}

		err = rows.Scan(&log.ID, &log.UserID, &log.SegmentID, &log.Operation, &log.InsertTime)
		if err != nil {
			return nil, fmt.Errorf("error while scanning rows: %w", err)
		}

		logs = append(logs, log)
	}

	return logs, nil
}
