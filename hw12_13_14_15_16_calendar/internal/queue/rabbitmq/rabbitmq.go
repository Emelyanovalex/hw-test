package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const QueueName = "notifications"

func declareQueue(ch *amqp.Channel) error {
	_, err := ch.QueueDeclare(
		QueueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare queue %q: %w", QueueName, err)
	}
	return nil
}
