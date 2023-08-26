package segment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"strings"
	"time"
)

const errCodeUniqueViolation = "23505"

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) AddSegment(ctx context.Context, slug string) error {
	_, err := r.db.ExecContext(ctx, `insert into segment (id) values ($1)`, slug)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == errCodeUniqueViolation {
			return ErrSegmentAlreadyExists
		}

		return fmt.Errorf("error while inserting into segment: %w", err)
	}
	return nil
}

//удаляем несуществующий сегмент (наверное это ок)

func (r *Repository) DeleteSegment(ctx context.Context, slug string) error {
	_, err := r.db.ExecContext(ctx, `update segment set deleted = true where id = $1`, slug)
	if err != nil {
		return fmt.Errorf("error while deleting segment: %w", err)
	}
	return nil
}

// добавление юзера в сегмент когда сегмента нет
//нужна тут валидация? или её лучше в сервис

func (r *Repository) AddUserToSegment(ctx context.Context, userID int64, slugs []string, ttl *time.Duration) error {
	if len(slugs) == 0 {
		return nil
	}

	if ttl != nil && *ttl < time.Hour {
		return fmt.Errorf("error ttl less than hour")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	_, err = tx.ExecContext(ctx, `insert into "user" (id) values ($1) on conflict do nothing`, userID)
	if err != nil {
		return fmt.Errorf("error while inserting into user: %w", err)
	}
	values := make([]string, 0, len(slugs))
	slugsAny := make([]interface{}, 0, len(slugs))
	for i, slug := range slugs {
		if ttl != nil {
			values = append(values, fmt.Sprintf("(%d, $%d, (%d || ' hour')::interval)", userID, i+1, ttl.Hours()))
		} else {
			values = append(values, fmt.Sprintf("(%d, $%d)", userID, i+1))
		}
		slugsAny = append(slugsAny, slug)
	}

	var columns string
	if ttl != nil {
		columns = "(user_id, segment_id, ttl)"
	} else {
		columns = "(user_id, segment_id)"
	}

	query := fmt.Sprintf(`insert into user_segment %s values %s`, columns, strings.Join(values, ","))

	_, err = tx.ExecContext(ctx, query, slugsAny...)
	if err != nil {
		return fmt.Errorf("error while inserting into user_segment: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error while commiting transaction: %w", err)
	}

	return nil
}

// нужно проверять, что юзер/сегмент сущетсвует? можно использовать rowsAffected

func (r *Repository) DeleteUserFromSegment(ctx context.Context, userID int64, slugs []string) error {
	if len(slugs) == 0 {
		return nil
	}

	values := make([]string, 0, len(slugs))
	slugsAny := make([]interface{}, 0, len(slugs))
	for i, slug := range slugs {
		values = append(values, fmt.Sprintf("$%d", i+1))
		slugsAny = append(slugsAny, slug)
	}

	query := fmt.Sprintf(
		`delete from user_segment where user_id = %d and segment_id in (%s)`,
		userID,
		strings.Join(values, ","),
	)
	_, err := r.db.ExecContext(ctx, query, slugsAny...)
	if err != nil {
		return fmt.Errorf("error while deleting user from user_segment: %w", err)
	}

	return nil
}

func (r *Repository) GetUserActiveSegments(ctx context.Context, userID int64) ([]string, error) {
	query := `select id as segment_id from segment
 				inner join user_segment userseg on userseg.segment_id = segment.id
				where deleted = false 
				  and userseg.user_id = $1 
				  and (
				      userseg.ttl is null or now() < (userseg.insertDate + userseg.ttl)
				  )`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error while getting active user's segments: %w", err)
	}

	var segments []string
	for rows.Next() {
		var segment string

		err = rows.Scan(&segment)
		if err != nil {
			return nil, fmt.Errorf("error while scanning segments: %w", err)
		}

		segments = append(segments, segment)
	}

	return segments, nil
}
