package notifier

import (
	"testing"
	"time"
)

func TestNewStats(t *testing.T) {
	stats := NewStats()

	if stats.startTime.IsZero() {
		t.Error("startTime should be set")
	}

	if stats.scrapeCount != 0 {
		t.Error("scrapeCount should start at 0")
	}

	if stats.scrapeFailures != 0 {
		t.Error("scrapeFailures should start at 0")
	}

	if len(stats.uniqueAircraft) != 0 {
		t.Error("uniqueAircraft should start empty")
	}
}

func TestRecordScrape(t *testing.T) {
	stats := NewStats()

	stats.RecordScrape()
	if stats.scrapeCount != 1 {
		t.Errorf("expected scrapeCount 1, got %d", stats.scrapeCount)
	}

	stats.RecordScrape()
	if stats.scrapeCount != 2 {
		t.Errorf("expected scrapeCount 2, got %d", stats.scrapeCount)
	}
}

func TestRecordScrapeFailure(t *testing.T) {
	stats := NewStats()

	stats.RecordScrapeFailure()
	if stats.scrapeFailures != 1 {
		t.Errorf("expected scrapeFailures 1, got %d", stats.scrapeFailures)
	}

	stats.RecordScrapeFailure()
	if stats.scrapeFailures != 2 {
		t.Errorf("expected scrapeFailures 2, got %d", stats.scrapeFailures)
	}
}

func TestRecordAircraft(t *testing.T) {
	stats := NewStats()

	// Record first aircraft
	stats.RecordAircraft("ABC123")
	if len(stats.uniqueAircraft) != 1 {
		t.Errorf("expected 1 unique aircraft, got %d", len(stats.uniqueAircraft))
	}

	// Record same aircraft again (should not increase count)
	stats.RecordAircraft("ABC123")
	if len(stats.uniqueAircraft) != 1 {
		t.Errorf("expected 1 unique aircraft after duplicate, got %d", len(stats.uniqueAircraft))
	}

	// Record different aircraft
	stats.RecordAircraft("DEF456")
	if len(stats.uniqueAircraft) != 2 {
		t.Errorf("expected 2 unique aircraft, got %d", len(stats.uniqueAircraft))
	}

	// Record empty hex (should be ignored)
	stats.RecordAircraft("")
	if len(stats.uniqueAircraft) != 2 {
		t.Errorf("expected 2 unique aircraft after empty hex, got %d", len(stats.uniqueAircraft))
	}
}

func TestGetStats(t *testing.T) {
	stats := NewStats()

	// Wait a moment to ensure uptime is measurable
	time.Sleep(10 * time.Millisecond)

	stats.RecordScrape()
	stats.RecordScrape()
	stats.RecordScrapeFailure()
	stats.RecordAircraft("ABC123")
	stats.RecordAircraft("DEF456")

	statsData := stats.GetStats()

	if statsData["scrape_count"] != int64(2) {
		t.Errorf("expected scrape_count 2, got %v", statsData["scrape_count"])
	}

	if statsData["scrape_failures"] != int64(1) {
		t.Errorf("expected scrape_failures 1, got %v", statsData["scrape_failures"])
	}

	if statsData["unique_aircraft"] != 2 {
		t.Errorf("expected unique_aircraft 2, got %v", statsData["unique_aircraft"])
	}

	// Check that uptime is reasonable (should be at least 10ms due to our sleep)
	if uptime, ok := statsData["uptime_seconds"].(int64); !ok || uptime < 0 {
		t.Errorf("uptime_seconds should be a positive integer, got %v", statsData["uptime_seconds"])
	}

	if statsData["success_rate"] != 66.66666666666666 {
		t.Errorf("expected success_rate ~66.7%%, got %v", statsData["success_rate"])
	}
}

func TestCalculateSuccessRate(t *testing.T) {
	stats := NewStats()

	// Test with no scrapes
	statsData := stats.GetStats()
	if statsData["success_rate"] != 100.0 {
		t.Errorf("expected success_rate 100.0 with no scrapes, got %v", statsData["success_rate"])
	}

	// Test with only successful scrapes
	stats.RecordScrape()
	stats.RecordScrape()
	statsData = stats.GetStats()
	if statsData["success_rate"] != 100.0 {
		t.Errorf("expected success_rate 100.0 with only successful scrapes, got %v", statsData["success_rate"])
	}

	// Test with mixed results
	stats.RecordScrapeFailure()
	statsData = stats.GetStats()
	if statsData["success_rate"] != 66.66666666666666 {
		t.Errorf("expected success_rate ~66.7%%, got %v", statsData["success_rate"])
	}
}

func TestGetUniqueAircraftList(t *testing.T) {
	stats := NewStats()

	stats.RecordAircraft("ABC123")
	stats.RecordAircraft("DEF456")

	list := stats.GetUniqueAircraftList()

	if len(list) != 2 {
		t.Errorf("expected 2 aircraft in list, got %d", len(list))
	}

	// Check that both aircraft are present
	foundABC := false
	foundDEF := false

	for _, aircraft := range list {
		if aircraft["hex"] == "ABC123" {
			foundABC = true
		}
		if aircraft["hex"] == "DEF456" {
			foundDEF = true
		}
	}

	if !foundABC {
		t.Error("ABC123 not found in aircraft list")
	}

	if !foundDEF {
		t.Error("DEF456 not found in aircraft list")
	}
}
