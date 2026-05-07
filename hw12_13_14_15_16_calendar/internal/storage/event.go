package storage

import (
	"errors"
	"time"
)

var (
	ErrDateBusy      = errors.New("date is already busy by another event")
	ErrEventNotFound = errors.New("event not found")
)

// Event represents a calendar event.
type Event struct {
	ID           string        `db:"id"            json:"id"`
	Title        string        `db:"title"         json:"title"`
	StartTime    time.Time     `db:"start_time"    json:"start_time"`
	Duration     time.Duration `db:"duration"      json:"duration"`
	Description  string        `db:"description"   json:"description,omitempty"`
	UserID       string        `db:"user_id"       json:"user_id"`
	NotifyBefore time.Duration `db:"notify_before" json:"notify_before,omitempty"`
}
