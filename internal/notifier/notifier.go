package notifier

import (
	"fmt"

	"github.com/example/whats-flying-over-me/internal/config"
)

// Notifier defines a mechanism for sending notifications.
type Notifier interface {
	Notify(subject, body string) error
}

// New creates a notifier based on the configuration.
func New(cfg config.NotifierConfig) (Notifier, error) {
	switch cfg.Method {
	case "email":
		return NewEmail(cfg.Email), nil
	default:
		return nil, fmt.Errorf("unknown notifier method: %s", cfg.Method)
	}
}
