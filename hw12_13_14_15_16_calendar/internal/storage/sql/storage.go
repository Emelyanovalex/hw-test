package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Emelyanovalex/hw-test/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	dsn string
	db  *sqlx.DB
}

func New(dsn string) *Storage {
	return &Storage{dsn: dsn}
}

func (s *Storage) Connect(ctx context.Context) error {
	db, err := sqlx.ConnectContext(ctx, "pgx", s.dsn)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	s.db = db
	return nil
}

func (s *Storage) Close(_ context.Context) error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Storage) CreateEvent(ctx context.Context, e storage.Event) error {
	if err := s.assertNoOverlap(ctx, e, ""); err != nil {
		return err
	}
	const q = `
		INSERT INTO events (id, title, start_time, duration, description, user_id, notify_before)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.db.ExecContext(ctx, q,
		e.ID, e.Title, e.StartTime, int64(e.Duration), e.Description, e.UserID, int64(e.NotifyBefore))
	if err != nil {
		return fmt.Errorf("create event: %w", err)
	}
	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, id string, e storage.Event) error {
	e.ID = id
	if err := s.assertNoOverlap(ctx, e, id); err != nil {
		return err
	}
	const q = `
		UPDATE events
		SET title=$2, start_time=$3, duration=$4, description=$5, user_id=$6, notify_before=$7
		WHERE id=$1
	`
	res, err := s.db.ExecContext(ctx, q,
		id, e.Title, e.StartTime, int64(e.Duration), e.Description, e.UserID, int64(e.NotifyBefore))
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return storage.ErrEventNotFound
	}
	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM events WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return storage.ErrEventNotFound
	}
	return nil
}

func (s *Storage) ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	from := truncateToDay(date)
	return s.listInRange(ctx, from, from.AddDate(0, 0, 1))
}

func (s *Storage) ListEventsForWeek(ctx context.Context, weekStart time.Time) ([]storage.Event, error) {
	from := truncateToDay(weekStart)
	return s.listInRange(ctx, from, from.AddDate(0, 0, 7))
}

func (s *Storage) ListEventsForMonth(ctx context.Context, monthStart time.Time) ([]storage.Event, error) {
	from := truncateToDay(monthStart)
	return s.listInRange(ctx, from, from.AddDate(0, 1, 0))
}

func (s *Storage) listInRange(ctx context.Context, from, to time.Time) ([]storage.Event, error) {
	const q = `
		SELECT id, title, start_time, duration, description, user_id, notify_before
		FROM events
		WHERE start_time >= $1 AND start_time < $2
		ORDER BY start_time
	`
	rows, err := s.db.QueryxContext(ctx, q, from, to)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	result := make([]storage.Event, 0)
	for rows.Next() {
		var (
			e            storage.Event
			duration     int64
			notifyBefore int64
		)
		if err := rows.Scan(&e.ID, &e.Title, &e.StartTime, &duration,
			&e.Description, &e.UserID, &notifyBefore); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		e.Duration = time.Duration(duration)
		e.NotifyBefore = time.Duration(notifyBefore)
		result = append(result, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}
	return result, nil
}

func (s *Storage) assertNoOverlap(ctx context.Context, e storage.Event, ignoreID string) error {
	const q = `
		SELECT 1 FROM events
		WHERE user_id = $1
		  AND id <> $2
		  AND start_time < $3
		  AND (start_time + (duration || ' nanoseconds')::interval) > $4
		LIMIT 1
	`
	end := e.StartTime.Add(e.Duration)
	var exists int
	err := s.db.GetContext(ctx, &exists, q, e.UserID, ignoreID, end, e.StartTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("check overlap: %w", err)
	}
	return storage.ErrDateBusy
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
