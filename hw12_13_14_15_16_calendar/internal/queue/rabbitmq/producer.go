package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Emelyanovalex/hw12_calendar/internal/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Producer publishes notifications to a RabbitMQ queue.
type Producer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewProducer connects to RabbitMQ and prepares the notifications queue.
func NewProducer(dsn string) (*Producer, error) {
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
	return &Producer{conn: conn, ch: ch}, nil
}

// Publish serialises n as JSON and puts it into the queue.
func (p *Producer) Publish(_ context.Context, n queue.Notification) error {
	body, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}
	return p.ch.Publish(
		"",        // default exchange
		QueueName, // routing key = queue name
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

// Close releases the channel and connection.
func (p *Producer) Close() error {
	_ = p.ch.Close()
	return p.conn.Close()
}
