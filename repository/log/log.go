package log

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Log struct {
	db *sql.DB
}

func New(db *sql.DB) *Log {
	return &Log{
		db: db,
	}
}

func (l *Log) AddLog(ctx context.Context, userID int64, segments []string, operation string) error {
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

	query := fmt.Sprintf(`insert into "log" (user_id, segment_id, operation) values %s`, strings.Join(values, ","))
	_, err := l.db.ExecContext(ctx, query, queryArgs...)
	if err != nil {
		return fmt.Errorf("error while inserting into log: %w", err)
	}

	return nil
}

// нужно?

func (l *Log) DeleteLog(ctx context.Context, logID int64) error {
	_, err := l.db.ExecContext(ctx, `delete from log where log_id = $1`, logID)
	if err != nil {
		return fmt.Errorf("error while deleting from log: %w", err)
	}

	return nil
}
