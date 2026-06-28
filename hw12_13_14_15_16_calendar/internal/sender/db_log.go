package sender

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver
)

// DBLog records sent notifications in the notifications_sent table.
type DBLog struct {
	db *sql.DB
}

// NewDBLog opens a PostgreSQL connection for recording sent notifications.
func NewDBLog(dsn string) (*DBLog, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	return &DBLog{db: db}, nil
}

// RecordSent inserts a record that the notification for eventID was processed.
func (d *DBLog) RecordSent(ctx context.Context, eventID string) error {
	_, err := d.db.ExecContext(ctx, `INSERT INTO notifications_sent (event_id) VALUES ($1)`, eventID)
	if err != nil {
		return fmt.Errorf("record sent: %w", err)
	}
	return nil
}

// Close releases the database connection.
func (d *DBLog) Close() error {
	return d.db.Close()
}
