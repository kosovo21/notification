package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

// Publisher handles publishing messages to RabbitMQ.
type Publisher struct {
	rmq *RabbitMQ
}

// NewPublisher creates a new Publisher.
func NewPublisher(rmq *RabbitMQ) *Publisher {
	return &Publisher{rmq: rmq}
}

// Publish publishes a message to the specified exchange and routing key.
func (p *Publisher) Publish(ctx context.Context, exchange, routingKey string, body interface{}) error {
	bytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = p.rmq.Channel().PublishWithContext(ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         bytes,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		log.Error().Err(err).
			Str("exchange", exchange).
			Str("routing_key", routingKey).
			Msg("failed to publish message")
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Debug().
		Str("exchange", exchange).
		Str("routing_key", routingKey).
		Msg("published message to queue")

	return nil
}
