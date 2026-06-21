package queue

import (
	"context"
	"time"
)

// Notification is sent from Scheduler to Sender via the message queue.
type Notification struct {
	EventID   string    `json:"event_id"`
	Title     string    `json:"title"`
	StartTime time.Time `json:"start_time"`
	UserID    string    `json:"user_id"`
}

// Producer publishes notifications to the queue.
type Producer interface {
	Publish(ctx context.Context, n Notification) error
	Close() error
}

// Consumer reads notifications from the queue.
type Consumer interface {
	Consume(ctx context.Context, handler func(n Notification)) error
	Close() error
}
