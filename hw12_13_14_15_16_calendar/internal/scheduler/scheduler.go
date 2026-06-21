package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/Emelyanovalex/hw12_calendar/internal/queue"
	"github.com/Emelyanovalex/hw12_calendar/internal/storage"
)

// Logger is the minimal logging surface used by Scheduler.
type Logger interface {
	Info(msg string)
	Error(msg string)
}

// Storage is the storage interface required by Scheduler.
type Storage interface {
	ListEventsToNotify(ctx context.Context, now time.Time) ([]storage.Event, error)
	MarkEventNotified(ctx context.Context, id string) error
	DeleteOldEvents(ctx context.Context, before time.Time) (int64, error)
}

// Scheduler periodically scans storage for events that need notifications
// and removes events older than one year.
type Scheduler struct {
	storage  Storage
	producer queue.Producer
	logger   Logger
	interval time.Duration
}

// New creates a Scheduler.
func New(stor Storage, producer queue.Producer, logger Logger, interval time.Duration) *Scheduler {
	return &Scheduler{
		storage:  stor,
		producer: producer,
		logger:   logger,
		interval: interval,
	}
}

// Run starts the scheduler loop; it returns when ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) error {
	s.logger.Info(fmt.Sprintf("scheduler started, interval=%s", s.interval))
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.sendNotifications(ctx)
			s.cleanupOldEvents(ctx)
		}
	}
}

func (s *Scheduler) sendNotifications(ctx context.Context) {
	events, err := s.storage.ListEventsToNotify(ctx, time.Now())
	if err != nil {
		s.logger.Error("list events to notify: " + err.Error())
		return
	}
	for _, e := range events {
		n := queue.Notification{
			EventID:   e.ID,
			Title:     e.Title,
			StartTime: e.StartTime,
			UserID:    e.UserID,
		}
		if err := s.producer.Publish(ctx, n); err != nil {
			s.logger.Error(fmt.Sprintf("publish notification for event %s: %s", e.ID, err))
			continue
		}
		if err := s.storage.MarkEventNotified(ctx, e.ID); err != nil {
			s.logger.Error(fmt.Sprintf("mark event %s notified: %s", e.ID, err))
		} else {
			s.logger.Info(fmt.Sprintf("notification queued: event_id=%s title=%q", e.ID, e.Title))
		}
	}
}

func (s *Scheduler) cleanupOldEvents(ctx context.Context) {
	before := time.Now().AddDate(-1, 0, 0)
	n, err := s.storage.DeleteOldEvents(ctx, before)
	if err != nil {
		s.logger.Error("delete old events: " + err.Error())
		return
	}
	if n > 0 {
		s.logger.Info(fmt.Sprintf("deleted %d old event(s)", n))
	}
}
