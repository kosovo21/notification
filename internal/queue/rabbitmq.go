package queue

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

// RabbitMQ holds the connection and channel.
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

// NewRabbitMQ creates a new RabbitMQ connection.
func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
		url:     url,
	}, nil
}

// Close closes the connection and channel.
func (r *RabbitMQ) Close() {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close rabbitmq channel")
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close rabbitmq connection")
		}
	}
}

// DeclareExchange declares a topic exchange.
func (r *RabbitMQ) DeclareExchange(name string) error {
	return r.channel.ExchangeDeclare(
		name,    // name
		"topic", // type
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
}

// Channel returns the underlying channel (use with caution).
func (r *RabbitMQ) Channel() *amqp.Channel {
	return r.channel
}
