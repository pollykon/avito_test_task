package segment

import (
	"context"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"hash/fnv"
	"strconv"
	"strings"
	"time"

	"github.com/pollykon/avito_test_task/internal/storage"
)

const errCodeUniqueViolation = "23505"

type Repository struct {
	db storage.Database
}

func New(db storage.Database) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) AddSegment(ctx context.Context, slug string, percent *int64) error {
	query := `insert into segment (id) values ($1)`
	if percent != nil {
		query = fmt.Sprintf("insert into segment (id, percent) values ($1, %d)", *percent)
	}
	_, err := r.db.ExecContext(ctx, query, slug)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == errCodeUniqueViolation {
			return ErrSegmentAlreadyExists
		}

		return fmt.Errorf("error while inserting into segment: %w", err)
	}

	return nil
}

// DeleteSegment sets segments' flags 'deleted' = true
func (r *Repository) DeleteSegment(ctx context.Context, slug string) error {
	res, err := r.db.ExecContext(ctx, `update segment set deleted = true where id = $1`, slug)
	if err != nil {
		return fmt.Errorf("error while deleting segment: %w", err)
	}

	numberOfDeletedSegments, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error while getting affected rows: %w", err)
	}

	if numberOfDeletedSegments == 0 {
		return ErrSegmentNotExist
	}

	return nil
}

func (r *Repository) AddUserToSegment(ctx context.Context, userID int64, slugs []string, ttl *time.Duration) error {
	if len(slugs) == 0 {
		return nil
	}

	err := r.db.WithTransaction(ctx, func(ctx context.Context) error {
		_, err := r.db.ExecContext(ctx, `insert into "user" (id) values ($1) on conflict do nothing`, userID)
		if err != nil {
			return fmt.Errorf("error while inserting into user: %w", err)
		}

		values := make([]string, 0, len(slugs))
		slugsAny := make([]interface{}, 0, len(slugs))
		for i, slug := range slugs {
			if ttl != nil {
				values = append(values, fmt.Sprintf("(%d, $%d, (%d || ' hour')::interval)", userID, i+1, int64(ttl.Hours())))
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

		_, err = r.db.ExecContext(ctx, query, slugsAny...)
		if err != nil {
			return fmt.Errorf("error while inserting into user_segment: %w", err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error while beddining transaction: %w", err)
	}

	return nil
}

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

// GetUserActiveSegments returns segments which:
// 1. were added to user and weren't deleted
// 2. were added with ttl, and they are still actual
// 3. were added to user by counting segment's percent
func (r *Repository) GetUserActiveSegments(ctx context.Context, userID int64) ([]string, error) {
	hashProcessor := fnv.New32a()
	_, _ = hashProcessor.Write([]byte(strconv.FormatInt(userID, 10)))
	hashedUserID := hashProcessor.Sum32()

	participationUserSign := int64(hashedUserID % 100)

	query := `select segment.id as segment_id from segment
 				left join user_segment userseg on userseg.segment_id = segment.id
				where segment.deleted = false
				  and (userseg.user_id = $1 or segment.percent >= $2)
				  and (
				      userseg.ttl is null or now() < (userseg.insert_time + userseg.ttl)
				  )`
	rows, err := r.db.QueryContext(ctx, query, userID, participationUserSign)
	if err != nil {
		return nil, fmt.Errorf("error while getting active user's segments: %w", err)
	}

	defer func() { _ = rows.Close() }()

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

func (r *Repository) InTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	return r.db.WithTransaction(ctx, f)
}

func (r *Repository) DeleteUserSegmentsWithBadTTL(ctx context.Context, limit int64) error {
	query := `delete from user_segment where id in (
				select id from user_segment where ttl is not null and ttl + insert_time < now() limit $1
			  )`
	_, err := r.db.ExecContext(ctx, query, limit)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteSegments(ctx context.Context, limit int64) error {
	query := `with deleted_rows AS (
				delete from segment where id in (
				  select id from segment where deleted = true limit $1
				)
			    returning id
			  )
			  delete from user_segment
			  where segment_id in (select id from deleted_rows)`
	_, err := r.db.ExecContext(ctx, query, limit)
	if err != nil {
		return err
	}

	return nil
}
