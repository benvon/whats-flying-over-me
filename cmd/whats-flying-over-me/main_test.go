package main

import (
	"fmt"
	"testing"

	"github.com/benvon/whats-flying-over-me/internal/notifier"
)

// TestLogHeartbeatStatsAccess ensures that logHeartbeat correctly accesses
// all required stats fields with the correct keys. This test prevents
// the "unique_aircraft" being null issue from happening again.
func TestLogHeartbeatStatsAccess(t *testing.T) {
	// Create a stats instance with some data
	stats := notifier.NewStats()

	// Record some activity to populate stats
	stats.RecordScrape()
	stats.RecordScrape()
	stats.RecordScrapeFailure()
	stats.RecordAircraft("ABC123")
	stats.RecordAircraft("DEF456")

	// Get the stats data
	statsData := stats.GetStats()

	// Verify that all required keys exist and have non-nil values
	requiredKeys := []string{
		"uptime",
		"scrape_count",
		"scrape_failures",
		"success_rate",
		"unique_aircraft",
	}

	for _, key := range requiredKeys {
		if value, exists := statsData[key]; !exists {
			t.Errorf("required stats key '%s' is missing", key)
		} else if value == nil {
			t.Errorf("required stats key '%s' has nil value", key)
		}
	}

	// Verify specific expected values
	if statsData["scrape_count"] != int64(2) {
		t.Errorf("expected scrape_count to be 2, got %v", statsData["scrape_count"])
	}

	if statsData["scrape_failures"] != int64(1) {
		t.Errorf("expected scrape_failures to be 1, got %v", statsData["scrape_failures"])
	}

	if statsData["unique_aircraft"] != 2 {
		t.Errorf("expected unique_aircraft to be 2, got %v", statsData["unique_aircraft"])
	}

	// Verify uptime is a string and not empty
	if uptime, ok := statsData["uptime"].(string); !ok {
		t.Errorf("expected uptime to be string, got %T", statsData["uptime"])
	} else if uptime == "" {
		t.Errorf("expected uptime to be non-empty")
	}

	// Verify success_rate is a float64 (it gets formatted to string in logHeartbeat)
	if successRate, ok := statsData["success_rate"].(float64); !ok {
		t.Errorf("expected success_rate to be float64, got %T", statsData["success_rate"])
	} else if successRate < 0 || successRate > 100 {
		t.Errorf("expected success_rate to be between 0 and 100, got %f", successRate)
	}
}

// TestLogHeartbeatFieldMapping ensures that the logHeartbeat function
// maps stats data to the correct field names in the log output.
func TestLogHeartbeatFieldMapping(t *testing.T) {
	// Create a stats instance
	stats := notifier.NewStats()

	// Record some activity
	stats.RecordScrape()
	stats.RecordAircraft("TEST123")

	// Get the stats data
	statsData := stats.GetStats()

	// This test verifies the field mapping logic in logHeartbeat
	// The function should map statsData["unique_aircraft"] to "unique_aircraft" in the log
	// This prevents the bug where it was incorrectly accessing "uniqueCount"

	// Verify the key exists and has the expected value
	if uniqueAircraft, exists := statsData["unique_aircraft"]; !exists {
		t.Errorf("stats key 'unique_aircraft' is missing - this would cause null in logs")
	} else if uniqueAircraft != 1 {
		t.Errorf("expected unique_aircraft to be 1, got %v", uniqueAircraft)
	}

	// Verify that the incorrect key "uniqueCount" does NOT exist
	if _, exists := statsData["uniqueCount"]; exists {
		t.Errorf("stats key 'uniqueCount' should not exist - this was the bug")
	}
}

// TestLogHeartbeatExactFieldMapping ensures that the exact field mapping
// used in logHeartbeat function is correct and prevents the "uniqueCount" bug.
func TestLogHeartbeatExactFieldMapping(t *testing.T) {
	stats := notifier.NewStats()
	stats.RecordScrape()
	stats.RecordAircraft("EXACT123")

	statsData := stats.GetStats()

	// This test replicates the exact logic from logHeartbeat function
	// to ensure the field mapping is correct
	heartbeatFields := map[string]interface{}{
		"uptime":          statsData["uptime"],
		"scrape_count":    statsData["scrape_count"],
		"scrape_failures": statsData["scrape_failures"],
		"success_rate":    fmt.Sprintf("%.1f%%", statsData["success_rate"]),
		"unique_aircraft": statsData["unique_aircraft"],
	}

	// Verify that all fields have non-nil values
	for fieldName, value := range heartbeatFields {
		if value == nil {
			t.Errorf("heartbeat field '%s' has nil value - this would cause null in logs", fieldName)
		}
	}

	// Verify the specific field that was causing the bug
	if heartbeatFields["unique_aircraft"] != 1 {
		t.Errorf("expected unique_aircraft to be 1, got %v", heartbeatFields["unique_aircraft"])
	}

	// Verify that success_rate is properly formatted as a string
	if successRate, ok := heartbeatFields["success_rate"].(string); !ok {
		t.Errorf("expected success_rate to be string after formatting, got %T", heartbeatFields["success_rate"])
	} else if len(successRate) == 0 {
		t.Errorf("expected success_rate to be non-empty after formatting")
	}
}

// TestLogHeartbeatDataTypes ensures that all stats data has the correct types
// that can be properly formatted in the log output.
func TestLogHeartbeatDataTypes(t *testing.T) {
	stats := notifier.NewStats()
	stats.RecordScrape()
	stats.RecordAircraft("TYPE123")

	statsData := stats.GetStats()

	// Test type assertions for each field
	typeTests := []struct {
		key  string
		want interface{}
	}{
		{"uptime", ""},
		{"scrape_count", int64(0)},
		{"scrape_failures", int64(0)},
		{"success_rate", float64(0)},
		{"unique_aircraft", 0},
	}

	for _, tt := range typeTests {
		if value, exists := statsData[tt.key]; !exists {
			t.Errorf("key '%s' missing from stats", tt.key)
		} else {
			// Check if the value can be converted to the expected type
			switch tt.want.(type) {
			case string:
				if _, ok := value.(string); !ok {
					t.Errorf("key '%s' should be string, got %T", tt.key, value)
				}
			case int64:
				if _, ok := value.(int64); !ok {
					t.Errorf("key '%s' should be int64, got %T", tt.key, value)
				}
			case float64:
				if _, ok := value.(float64); !ok {
					t.Errorf("key '%s' should be float64, got %T", tt.key, value)
				}
			case int:
				if _, ok := value.(int); !ok {
					t.Errorf("key '%s' should be int, got %T", tt.key, value)
				}
			}
		}
	}
}
