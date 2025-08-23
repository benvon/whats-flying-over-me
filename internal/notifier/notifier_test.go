package notifier

import (
	"testing"
	"time"

	"github.com/example/whats-flying-over-me/internal/config"
	"github.com/example/whats-flying-over-me/internal/piaware"
)

func TestNewConsole(t *testing.T) {
	cfg := config.NotifierConfig{Console: true}
	n, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := n.(*Console); !ok {
		t.Fatalf("expected *Console, got %T", n)
	}
}

func TestNewWebhook(t *testing.T) {
	cfg := config.NotifierConfig{
		Webhook: config.WebhookConfig{
			Enabled: true,
			URL:     "http://localhost:8080/webhook",
		},
	}
	n, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := n.(*Webhook); !ok {
		t.Fatalf("expected *Webhook, got %T", n)
	}
}

func TestNewWebhookMissingURL(t *testing.T) {
	cfg := config.NotifierConfig{
		Webhook: config.WebhookConfig{
			Enabled: true,
			URL:     "",
		},
	}
	if _, err := New(cfg); err == nil {
		t.Fatal("expected error for missing webhook URL")
	}
}

func TestNewRabbitMQConfigValidation(t *testing.T) {
	// Test missing URL
	cfg := config.NotifierConfig{
		RabbitMQ: config.RabbitMQConfig{
			Enabled:    true,
			URL:        "",
			Exchange:   "test_exchange",
			RoutingKey: "test_key",
		},
	}
	if _, err := New(cfg); err == nil {
		t.Fatal("expected error for missing RabbitMQ URL")
	}

	// Test missing exchange
	cfg = config.NotifierConfig{
		RabbitMQ: config.RabbitMQConfig{
			Enabled:    true,
			URL:        "amqp://localhost:5672",
			Exchange:   "",
			RoutingKey: "test_key",
		},
	}
	if _, err := New(cfg); err == nil {
		t.Fatal("expected error for missing RabbitMQ exchange")
	}

	// Test missing routing key
	cfg = config.NotifierConfig{
		RabbitMQ: config.RabbitMQConfig{
			Enabled:    true,
			URL:        "amqp://localhost:5672",
			Exchange:   "test_exchange",
			RoutingKey: "",
		},
	}
	if _, err := New(cfg); err == nil {
		t.Fatal("expected error for missing RabbitMQ routing key")
	}
}

func TestMultiNotifier(t *testing.T) {
	cfg := config.NotifierConfig{
		Console: true,
		Webhook: config.WebhookConfig{
			Enabled: true,
			URL:     "http://localhost:8080/webhook",
		},
	}
	n, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := n.(*MultiNotifier); !ok {
		t.Fatalf("expected *MultiNotifier, got %T", n)
	}
}

func TestNoNotifiers(t *testing.T) {
	cfg := config.NotifierConfig{
		Console:  false,
		Webhook:  config.WebhookConfig{Enabled: false},
		RabbitMQ: config.RabbitMQConfig{Enabled: false},
	}
	if _, err := New(cfg); err == nil {
		t.Fatal("expected error for no notifiers configured")
	}
}

func TestConsoleNotify(t *testing.T) {
	console := NewConsole()
	alert := AlertData{
		Timestamp:   time.Now(),
		Aircraft:    piaware.NearbyAircraft{},
		AlertType:   "test",
		Description: "test alert",
	}
	if err := console.Notify(alert); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
