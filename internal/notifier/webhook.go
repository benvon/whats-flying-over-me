package notifier

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/benvon/whats-flying-over-me/internal/config"
)

// Webhook implements Notifier using HTTP webhooks.
type Webhook struct {
	cfg    config.WebhookConfig
	client HTTPClient
}

// NewWebhook creates a new Webhook notifier.
func NewWebhook(cfg config.WebhookConfig) (*Webhook, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	client := NewRealHTTPClient(cfg.Timeout)

	return &Webhook{cfg: cfg, client: client}, nil
}

// NewWebhookWithClient creates a new Webhook notifier with a custom HTTP client (for testing).
func NewWebhookWithClient(cfg config.WebhookConfig, client HTTPClient) *Webhook {
	return &Webhook{cfg: cfg, client: client}
}

// Notify sends the alert to the webhook URL.
func (w *Webhook) Notify(alert AlertData) error {
	alertJSON, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert to JSON: %w", err)
	}

	resp, err := w.client.Post(w.cfg.URL, "application/json", alertJSON)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}
