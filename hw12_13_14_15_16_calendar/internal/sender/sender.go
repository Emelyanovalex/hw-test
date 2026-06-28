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

// StatusWriter records when a notification has been processed.
type StatusWriter interface {
	RecordSent(ctx context.Context, eventID string) error
}

// Sender reads notifications from the queue and logs them.
type Sender struct {
	consumer     queue.Consumer
	logger       Logger
	statusWriter StatusWriter
}

// New creates a Sender. sw may be nil — in that case no status is recorded.
func New(consumer queue.Consumer, logger Logger, sw StatusWriter) *Sender {
	return &Sender{consumer: consumer, logger: logger, statusWriter: sw}
}

// Run starts consuming notifications until ctx is cancelled.
func (s *Sender) Run(ctx context.Context) error {
	s.logger.Info("sender started, waiting for notifications...")
	return s.consumer.Consume(ctx, func(n queue.Notification) {
		s.logger.Info(fmt.Sprintf(
			"notification received: event_id=%s title=%q start_time=%s user_id=%s",
			n.EventID, n.Title, n.StartTime.Format(time.RFC3339), n.UserID,
		))
		if s.statusWriter != nil {
			if err := s.statusWriter.RecordSent(ctx, n.EventID); err != nil {
				s.logger.Error("failed to record sent notification: " + err.Error())
			}
		}
	})
}
