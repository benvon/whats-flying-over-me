package notifier

import (
	"fmt"
	"testing"
	"time"

	"github.com/example/whats-flying-over-me/internal/config"
	"github.com/example/whats-flying-over-me/internal/piaware"
)

func TestNewDeduplicator(t *testing.T) {
	cfg := config.AlertDedupeConfig{
		Enabled:     true,
		BlockoutMin: 15 * time.Minute,
	}

	dedup := NewDeduplicator(cfg)

	if dedup.cfg.Enabled != true {
		t.Error("expected enabled to be true")
	}

	if dedup.cfg.BlockoutMin != 15*time.Minute {
		t.Error("expected blockout_min to be 15 minutes")
	}

	if len(dedup.records) != 0 {
		t.Error("expected empty records map")
	}
}

func TestShouldAlertNewAircraft(t *testing.T) {
	dedup := NewDeduplicator(config.AlertDedupeConfig{
		Enabled:     true,
		BlockoutMin: 15 * time.Minute,
	})

	aircraft := piaware.NearbyAircraft{
		Aircraft: piaware.Aircraft{
			Hex:    "ABC123",
			Flight: "TEST1",
		},
	}

	// First time seeing this aircraft - should alert
	if !dedup.ShouldAlert(aircraft) {
		t.Error("expected to alert for new aircraft")
	}

	// Check that records were created
	stats := dedup.GetStats()
	if stats["active_records"] != 2 { // One for hex, one for flight+hex
		t.Errorf("expected 2 active records, got %v", stats["active_records"])
	}
}

func TestShouldAlertDuplicateAircraft(t *testing.T) {
	dedup := NewDeduplicator(config.AlertDedupeConfig{
		Enabled:     true,
		BlockoutMin: 15 * time.Minute,
	})

	aircraft := piaware.NearbyAircraft{
		Aircraft: piaware.Aircraft{
			Hex:    "ABC123",
			Flight: "TEST1",
		},
	}

	// First time - should alert
	if !dedup.ShouldAlert(aircraft) {
		t.Error("expected to alert for new aircraft")
	}

	// Second time with same aircraft - should not alert
	if dedup.ShouldAlert(aircraft) {
		t.Error("expected not to alert for duplicate aircraft")
	}
}

func TestShouldAlertNewTransponderSameFlight(t *testing.T) {
	dedup := NewDeduplicator(config.AlertDedupeConfig{
		Enabled:     true,
		BlockoutMin: 15 * time.Minute,
	})

	aircraft1 := piaware.NearbyAircraft{
		Aircraft: piaware.Aircraft{
			Hex:    "ABC123",
			Flight: "TEST1",
		},
	}

	aircraft2 := piaware.NearbyAircraft{
		Aircraft: piaware.Aircraft{
			Hex:    "DEF456", // Different transponder
			Flight: "TEST1",  // Same flight
		},
	}

	// First aircraft - should alert
	if !dedup.ShouldAlert(aircraft1) {
		t.Error("expected to alert for first aircraft")
	}

	// Second aircraft with same flight but different transponder - should alert
	if !dedup.ShouldAlert(aircraft2) {
		t.Error("expected to alert for aircraft with new transponder")
	}
}

func TestShouldAlertAfterBlockoutExpires(t *testing.T) {
	dedup := NewDeduplicator(config.AlertDedupeConfig{
		Enabled:     true,
		BlockoutMin: 1 * time.Millisecond, // Very short blockout for testing
	})

	aircraft := piaware.NearbyAircraft{
		Aircraft: piaware.Aircraft{
			Hex:    "ABC123",
			Flight: "TEST1",
		},
	}

	// First time - should alert
	if !dedup.ShouldAlert(aircraft) {
		t.Error("expected to alert for new aircraft")
	}

	// Wait for blockout to expire
	time.Sleep(2 * time.Millisecond)

	// Should alert again after blockout expires
	if !dedup.ShouldAlert(aircraft) {
		t.Error("expected to alert after blockout expires")
	}
}

func TestShouldAlertWithDeduplicationDisabled(t *testing.T) {
	dedup := NewDeduplicator(config.AlertDedupeConfig{
		Enabled:     false,
		BlockoutMin: 15 * time.Minute,
	})

	aircraft := piaware.NearbyAircraft{
		Aircraft: piaware.Aircraft{
			Hex:    "ABC123",
			Flight: "TEST1",
		},
	}

	// Should always alert when deduplication is disabled
	if !dedup.ShouldAlert(aircraft) {
		t.Error("expected to alert when deduplication is disabled")
	}

	// Second time - should still alert
	if !dedup.ShouldAlert(aircraft) {
		t.Error("expected to alert again when deduplication is disabled")
	}
}

func TestShouldAlertAircraftWithoutFlight(t *testing.T) {
	dedup := NewDeduplicator(config.AlertDedupeConfig{
		Enabled:     true,
		BlockoutMin: 15 * time.Minute,
	})

	aircraft := piaware.NearbyAircraft{
		Aircraft: piaware.Aircraft{
			Hex:    "ABC123",
			Flight: "", // No flight number
		},
	}

	// Should alert for aircraft without flight number
	if !dedup.ShouldAlert(aircraft) {
		t.Error("expected to alert for aircraft without flight number")
	}

	// Should not alert again for same transponder
	if dedup.ShouldAlert(aircraft) {
		t.Error("expected not to alert for duplicate transponder")
	}
}

func TestCleanupOldRecords(t *testing.T) {
	dedup := NewDeduplicator(config.AlertDedupeConfig{
		Enabled:     true,
		BlockoutMin: 1 * time.Millisecond, // Very short for testing
	})

	aircraft := piaware.NearbyAircraft{
		Aircraft: piaware.Aircraft{
			Hex:    "ABC123",
			Flight: "TEST1",
		},
	}

	// Add some records
	dedup.ShouldAlert(aircraft)

	// Check initial record count
	stats := dedup.GetStats()
	if stats["active_records"] != 2 {
		t.Errorf("expected 2 active records, got %v", stats["active_records"])
	}

	// Wait for cleanup to trigger (2x blockout time)
	time.Sleep(3 * time.Millisecond)

	// Add another record to trigger cleanup
	dedup.ShouldAlert(aircraft)

	// Records should be cleaned up
	stats = dedup.GetStats()
	if stats["active_records"] != 2 { // Should still be 2 for current aircraft
		t.Errorf("expected 2 active records after cleanup, got %v", stats["active_records"])
	}
}

func TestMakeKey(t *testing.T) {
	dedup := NewDeduplicator(config.AlertDedupeConfig{
		Enabled:     true,
		BlockoutMin: 15 * time.Minute,
	})

	tests := []struct {
		name        string
		tailNumber  string
		transponder string
		expected    string
	}{
		{
			name:        "both tail and transponder",
			tailNumber:  "TEST1",
			transponder: "ABC123",
			expected:    "TEST1:ABC123",
		},
		{
			name:        "empty tail number",
			tailNumber:  "",
			transponder: "ABC123",
			expected:    ":ABC123",
		},
		{
			name:        "empty transponder",
			tailNumber:  "TEST1",
			transponder: "",
			expected:    "TEST1:",
		},
		{
			name:        "both empty",
			tailNumber:  "",
			transponder: "",
			expected:    ":",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dedup.makeKey(tt.tailNumber, tt.transponder)
			if result != tt.expected {
				t.Errorf("makeKey(%q, %q) = %q, expected %q", tt.tailNumber, tt.transponder, result, tt.expected)
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	dedup := NewDeduplicator(config.AlertDedupeConfig{
		Enabled:     true,
		BlockoutMin: 15 * time.Minute,
	})

	stats := dedup.GetStats()

	// Check that all expected fields are present
	expectedFields := []string{"enabled", "blockout_min", "active_records"}
	for _, field := range expectedFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("expected stats to contain field %q", field)
		}
	}

	// Check specific values
	if enabled, ok := stats["enabled"].(bool); !ok || !enabled {
		t.Error("expected enabled to be true")
	}

	if blockoutMin, ok := stats["blockout_min"].(string); !ok || blockoutMin != "15m0s" {
		t.Errorf("expected blockout_min to be '15m0s', got %q", blockoutMin)
	}

	if activeRecords, ok := stats["active_records"].(int); !ok || activeRecords != 0 {
		t.Errorf("expected active_records to be 0, got %v", activeRecords)
	}
}

func TestConcurrentAccess(t *testing.T) {
	dedup := NewDeduplicator(config.AlertDedupeConfig{
		Enabled:     true,
		BlockoutMin: 15 * time.Minute,
	})

	// Test concurrent access to ShouldAlert
	done := make(chan bool)
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			aircraft := piaware.NearbyAircraft{
				Aircraft: piaware.Aircraft{
					Hex:    fmt.Sprintf("ABC%d", id),
					Flight: fmt.Sprintf("TEST%d", id),
				},
			}
			dedup.ShouldAlert(aircraft)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Check that all records were created
	stats := dedup.GetStats()
	expectedRecords := numGoroutines * 2 // Each aircraft creates 2 records
	if activeRecords, ok := stats["active_records"].(int); !ok || activeRecords != expectedRecords {
		t.Errorf("expected %d active records, got %v", expectedRecords, activeRecords)
	}
}
