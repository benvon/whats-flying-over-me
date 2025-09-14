package notifier

import (
	"testing"
	"time"

	"github.com/benvon/whats-flying-over-me/internal/config"
	"github.com/benvon/whats-flying-over-me/internal/piaware"
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

func TestMultiNotifierNotify(t *testing.T) {
	tests := []struct {
		name           string
		notifiers      []Notifier
		alert          AlertData
		wantErr        bool
		expectedCounts []int
		description    string
	}{
		{
			name: "all notifiers succeed",
			notifiers: []Notifier{
				NewMockNotifier(),
				NewMockNotifier(),
				NewMockNotifier(),
			},
			alert: AlertData{
				Timestamp:   time.Now(),
				Aircraft:    piaware.NearbyAircraft{},
				AlertType:   "test",
				Description: "test alert",
			},
			wantErr:        false,
			expectedCounts: []int{1, 1, 1},
			description:    "should succeed when all notifiers work",
		},
		{
			name: "some notifiers fail",
			notifiers: []Notifier{
				NewMockNotifier(),
				func() *MockNotifier {
					m := NewMockNotifier()
					m.SetShouldFail(true, "notifier 2 failed")
					return m
				}(),
				NewMockNotifier(),
			},
			alert: AlertData{
				Timestamp:   time.Now(),
				Aircraft:    piaware.NearbyAircraft{},
				AlertType:   "test",
				Description: "test alert",
			},
			wantErr:        true,           // Should return error from last failing notifier
			expectedCounts: []int{1, 1, 1}, // All should still receive the notification
			description:    "should continue with other notifiers even if one fails",
		},
		{
			name: "all notifiers fail",
			notifiers: []Notifier{
				func() *MockNotifier {
					m := NewMockNotifier()
					m.SetShouldFail(true, "notifier 1 failed")
					return m
				}(),
				func() *MockNotifier {
					m := NewMockNotifier()
					m.SetShouldFail(true, "notifier 2 failed")
					return m
				}(),
				func() *MockNotifier {
					m := NewMockNotifier()
					m.SetShouldFail(true, "notifier 3 failed")
					return m
				}(),
			},
			alert: AlertData{
				Timestamp:   time.Now(),
				Aircraft:    piaware.NearbyAircraft{},
				AlertType:   "test",
				Description: "test alert",
			},
			wantErr:        true,           // Should return error from last failing notifier
			expectedCounts: []int{1, 1, 1}, // All should still receive the notification
			description:    "should return error from last failing notifier",
		},
		{
			name: "single notifier",
			notifiers: []Notifier{
				NewMockNotifier(),
			},
			alert: AlertData{
				Timestamp:   time.Now(),
				Aircraft:    piaware.NearbyAircraft{},
				AlertType:   "test",
				Description: "test alert",
			},
			wantErr:        false,
			expectedCounts: []int{1},
			description:    "should work with single notifier",
		},
		{
			name:      "empty notifiers list",
			notifiers: []Notifier{},
			alert: AlertData{
				Timestamp:   time.Now(),
				Aircraft:    piaware.NearbyAircraft{},
				AlertType:   "test",
				Description: "test alert",
			},
			wantErr:        false,
			expectedCounts: []int{},
			description:    "should not fail with empty notifiers list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			multiNotifier := &MultiNotifier{notifiers: tt.notifiers}

			err := multiNotifier.Notify(tt.alert)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultiNotifier.Notify() error = %v, wantErr %v (%s)", err, tt.wantErr, tt.description)
			}

			// Verify all mock notifiers received the notification
			for i, notifier := range tt.notifiers {
				if mockNotifier, ok := notifier.(*MockNotifier); ok {
					count := mockNotifier.GetNotificationCount()
					if i < len(tt.expectedCounts) && count != tt.expectedCounts[i] {
						t.Errorf("Notifier %d received %d notifications, expected %d", i, count, tt.expectedCounts[i])
					}
				}
			}
		})
	}
}

func TestMultiNotifierNotifyMultipleAlerts(t *testing.T) {
	// Test that multiple alerts are handled correctly
	notifier1 := NewMockNotifier()
	notifier2 := NewMockNotifier()

	multiNotifier := &MultiNotifier{notifiers: []Notifier{notifier1, notifier2}}

	alerts := []AlertData{
		{
			Timestamp:   time.Now(),
			Aircraft:    piaware.NearbyAircraft{},
			AlertType:   "alert1",
			Description: "first alert",
		},
		{
			Timestamp:   time.Now(),
			Aircraft:    piaware.NearbyAircraft{},
			AlertType:   "alert2",
			Description: "second alert",
		},
		{
			Timestamp:   time.Now(),
			Aircraft:    piaware.NearbyAircraft{},
			AlertType:   "alert3",
			Description: "third alert",
		},
	}

	// Send multiple alerts
	for _, alert := range alerts {
		if err := multiNotifier.Notify(alert); err != nil {
			t.Fatalf("unexpected error sending alert: %v", err)
		}
	}

	// Verify both notifiers received all alerts
	if count1 := notifier1.GetNotificationCount(); count1 != len(alerts) {
		t.Errorf("Notifier 1 received %d notifications, expected %d", count1, len(alerts))
	}
	if count2 := notifier2.GetNotificationCount(); count2 != len(alerts) {
		t.Errorf("Notifier 2 received %d notifications, expected %d", count2, len(alerts))
	}

	// Verify the actual notifications
	notifications1 := notifier1.GetNotifications()
	notifications2 := notifier2.GetNotifications()

	if len(notifications1) != len(alerts) {
		t.Errorf("Notifier 1 has %d notifications, expected %d", len(notifications1), len(alerts))
	}
	if len(notifications2) != len(alerts) {
		t.Errorf("Notifier 2 has %d notifications, expected %d", len(notifications2), len(alerts))
	}

	// Verify alert types match
	for i, alert := range alerts {
		if notifications1[i].AlertType != alert.AlertType {
			t.Errorf("Notifier 1 alert %d type = %s, expected %s", i, notifications1[i].AlertType, alert.AlertType)
		}
		if notifications2[i].AlertType != alert.AlertType {
			t.Errorf("Notifier 2 alert %d type = %s, expected %s", i, notifications2[i].AlertType, alert.AlertType)
		}
	}
}
