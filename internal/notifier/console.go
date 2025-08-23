package notifier

import (
	"encoding/json"
	"fmt"
	"log"
)

// Console implements Notifier using console logging.
type Console struct{}

// NewConsole creates a new Console notifier.
func NewConsole() *Console {
	return &Console{}
}

// Notify logs the alert to the console.
func (c *Console) Notify(alert AlertData) error {
	// Convert alert to JSON for structured logging
	alertJSON, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert to JSON: %w", err)
	}

	log.Printf("ALERT: %s - %s", alert.AlertType, string(alertJSON))
	return nil
}
