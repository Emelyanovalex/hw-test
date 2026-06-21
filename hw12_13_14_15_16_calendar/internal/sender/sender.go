package sender

import (
	"context"
	"fmt"
	"time"

	"github.com/Emelyanovalex/hw12_calendar/internal/queue"
)

// Logger is the minimal logging surface used by Sender.
type Logger interface {
	Info(msg string)
	Error(msg string)
}

// Sender reads notifications from the queue and logs them.
type Sender struct {
	consumer queue.Consumer
	logger   Logger
}

// New creates a Sender.
func New(consumer queue.Consumer, logger Logger) *Sender {
	return &Sender{consumer: consumer, logger: logger}
}

// Run starts consuming notifications until ctx is cancelled.
func (s *Sender) Run(ctx context.Context) error {
	s.logger.Info("sender started, waiting for notifications...")
	return s.consumer.Consume(ctx, func(n queue.Notification) {
		s.logger.Info(fmt.Sprintf(
			"notification received: event_id=%s title=%q start_time=%s user_id=%s",
			n.EventID, n.Title, n.StartTime.Format(time.RFC3339), n.UserID,
		))
	})
}
