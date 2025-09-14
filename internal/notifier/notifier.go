package notifier

import (
	"fmt"
	"time"

	"github.com/benvon/whats-flying-over-me/internal/config"
	"github.com/benvon/whats-flying-over-me/internal/piaware"
)

// AlertData represents the data structure for notifications.
type AlertData struct {
	Timestamp   time.Time              `json:"timestamp"`
	Aircraft    piaware.NearbyAircraft `json:"aircraft"`
	AlertType   string                 `json:"alert_type"`
	Description string                 `json:"description"`
}

// Notifier defines a mechanism for sending notifications.
type Notifier interface {
	Notify(alert AlertData) error
}

// MultiNotifier sends notifications to multiple backends.
type MultiNotifier struct {
	notifiers []Notifier
}

// New creates a notifier based on the configuration.
func New(cfg config.NotifierConfig) (Notifier, error) {
	var notifiers []Notifier

	// Always add console notifier
	if cfg.Console {
		notifiers = append(notifiers, NewConsole())
	}

	// Add webhook notifier if enabled
	if cfg.Webhook.Enabled {
		webhook, err := NewWebhook(cfg.Webhook)
		if err != nil {
			return nil, fmt.Errorf("failed to create webhook notifier: %w", err)
		}
		notifiers = append(notifiers, webhook)
	}

	// Add RabbitMQ notifier if enabled
	if cfg.RabbitMQ.Enabled {
		rabbitmq, err := NewRabbitMQ(cfg.RabbitMQ)
		if err != nil {
			return nil, fmt.Errorf("failed to create RabbitMQ notifier: %w", err)
		}
		notifiers = append(notifiers, rabbitmq)
	}

	if len(notifiers) == 0 {
		return nil, fmt.Errorf("no notifiers configured")
	}

	if len(notifiers) == 1 {
		return notifiers[0], nil
	}

	return &MultiNotifier{notifiers: notifiers}, nil
}

// Notify sends notifications to all configured backends.
func (m *MultiNotifier) Notify(alert AlertData) error {
	var lastErr error
	for _, n := range m.notifiers {
		if err := n.Notify(alert); err != nil {
			lastErr = err
			// Continue with other notifiers even if one fails
		}
	}
	return lastErr
}
