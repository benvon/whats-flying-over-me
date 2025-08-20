package notifier

import (
	"testing"

	"github.com/example/whats-flying-over-me/internal/config"
)

func TestNewEmail(t *testing.T) {
	n, err := New(config.NotifierConfig{Method: "email"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := n.(*Email); !ok {
		t.Fatalf("expected *Email, got %T", n)
	}
}

func TestNewUnknown(t *testing.T) {
	if _, err := New(config.NotifierConfig{Method: "sms"}); err == nil {
		t.Fatal("expected error for unknown notifier")
	}
}
