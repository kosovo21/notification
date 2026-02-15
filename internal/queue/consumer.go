package queue

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
)

// Consumer handles consuming messages from RabbitMQ.
type Consumer struct {
	rmq *RabbitMQ
}

// NewConsumer creates a new Consumer.
func NewConsumer(rmq *RabbitMQ) *Consumer {
	return &Consumer{rmq: rmq}
}

// HandlerFunc is the function signature for processing messages.
type HandlerFunc func(ctx context.Context, body []byte) error

// Consume starts consuming messages from the specified queue.
// This is a blocking call.
func (c *Consumer) Consume(ctx context.Context, queueName, routingKey string, handler HandlerFunc) error {
	ch := c.rmq.Channel()

	// Declare queue
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	if err := ch.QueueBind(
		q.Name,       // queue name
		routingKey,   // routing key
		ExchangeName, // exchange
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// Start consuming
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer tag (auto-generated)
		false,  // auto-ack (we will manual ack)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Info().Str("queue", queueName).Msg("started consuming messages")

	// Process messages
	for {
		select {
		case <-ctx.Done():
			log.Info().Str("queue", queueName).Msg("stopping consumer")
			return nil
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("channel closed")
			}

			// Process message
			if err := handler(ctx, d.Body); err != nil {
				log.Error().Err(err).Str("message_id", d.MessageId).Msg("failed to process message")
				// Nack (requeue or dead-letter)
				// For now, let's Nack with requeue=false (assuming DLQ configured or just drop)
				// Or requeue=true if transient?
				// Simple strategy: Nack(false)
				d.Nack(false, false)
			} else {
				d.Ack(false)
			}
		}
	}
}
