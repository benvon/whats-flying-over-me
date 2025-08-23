package notifier

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/example/whats-flying-over-me/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ implements Notifier using RabbitMQ.
type RabbitMQ struct {
	cfg  config.RabbitMQConfig
	conn RabbitMQConnection
}

// NewRabbitMQ creates a new RabbitMQ notifier.
func NewRabbitMQ(cfg config.RabbitMQConfig) (*RabbitMQ, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("RabbitMQ URL is required")
	}
	if cfg.Exchange == "" {
		return nil, fmt.Errorf("RabbitMQ exchange is required")
	}
	if cfg.RoutingKey == "" {
		return nil, fmt.Errorf("RabbitMQ routing key is required")
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	// Create the actual connection
	conn, err := newRealRabbitMQConnection(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ connection: %w", err)
	}

	return &RabbitMQ{cfg: cfg, conn: conn}, nil
}

// NewRabbitMQWithConnection creates a new RabbitMQ notifier with a custom connection (for testing).
func NewRabbitMQWithConnection(cfg config.RabbitMQConfig, conn RabbitMQConnection) *RabbitMQ {
	return &RabbitMQ{cfg: cfg, conn: conn}
}

// Notify publishes the alert to RabbitMQ.
func (r *RabbitMQ) Notify(alert AlertData) error {
	alertJSON, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert to JSON: %w", err)
	}

	// Check if connection is still alive, reconnect if needed
	if !r.conn.IsConnected() {
		return fmt.Errorf("RabbitMQ connection is not available")
	}

	err = r.conn.Publish(r.cfg.Exchange, r.cfg.RoutingKey, alertJSON)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Close closes the RabbitMQ connection.
func (r *RabbitMQ) Close() error {
	return r.conn.Close()
}

// realRabbitMQConnection implements the actual RabbitMQ connection.
type realRabbitMQConnection struct {
	cfg  config.RabbitMQConfig
	conn *amqp.Connection
	ch   *amqp.Channel
}

// newRealRabbitMQConnection creates a new real RabbitMQ connection.
func newRealRabbitMQConnection(cfg config.RabbitMQConfig) (RabbitMQConnection, error) {
	conn := &realRabbitMQConnection{cfg: cfg}

	if err := conn.connect(); err != nil {
		return nil, err
	}

	return conn, nil
}

// connect establishes connection to RabbitMQ.
func (r *realRabbitMQConnection) connect() error {
	var err error

	r.conn, err = amqp.Dial(r.cfg.URL)
	if err != nil {
		return fmt.Errorf("failed to dial RabbitMQ: %w", err)
	}

	r.ch, err = r.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = r.ch.ExchangeDeclare(
		r.cfg.Exchange, // name
		"topic",        // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	return nil
}

// Publish publishes a message to RabbitMQ.
func (r *realRabbitMQConnection) Publish(exchange, routingKey string, body []byte) error {
	return r.ch.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)
}

// IsConnected checks if the connection is alive.
func (r *realRabbitMQConnection) IsConnected() bool {
	return r.conn != nil && !r.conn.IsClosed()
}

// Close closes the RabbitMQ connection.
func (r *realRabbitMQConnection) Close() error {
	if r.ch != nil {
		if err := r.ch.Close(); err != nil {
			return err
		}
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
