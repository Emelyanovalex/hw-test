package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Emelyanovalex/hw12_calendar/internal/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer reads notifications from a RabbitMQ queue.
type Consumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewConsumer connects to RabbitMQ and prepares the notifications queue.
func NewConsumer(dsn string) (*Consumer, error) {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, fmt.Errorf("amqp dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}
	if err := declareQueue(ch); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	return &Consumer{conn: conn, ch: ch}, nil
}

// Consume delivers messages to handler until ctx is cancelled or the channel closes.
func (c *Consumer) Consume(ctx context.Context, handler func(queue.Notification)) error {
	msgs, err := c.ch.Consume(
		QueueName,
		"",    // consumer tag (auto-generated)
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("start consume: %w", err)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("amqp channel closed")
			}
			var n queue.Notification
			if err := json.Unmarshal(msg.Body, &n); err != nil {
				_ = msg.Nack(false, false)
				continue
			}
			handler(n)
			_ = msg.Ack(false)
		}
	}
}

// Close releases the channel and connection.
func (c *Consumer) Close() error {
	_ = c.ch.Close()
	return c.conn.Close()
}
