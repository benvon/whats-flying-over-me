package main

import (
	"testing"
	"time"

	"github.com/example/whats-flying-over-me/internal/config"
	"github.com/example/whats-flying-over-me/internal/notifier"
	"github.com/example/whats-flying-over-me/internal/piaware"
)

func TestNewMonitorService(t *testing.T) {
	cfg := config.Config{
		BaseLat:     40.7128,
		BaseLon:     -74.0060,
		RadiusKm:    25.0,
		AltitudeMax: 10000,
		DataURL:     "http://test.com",
	}

	mockNotifier := notifier.NewMockNotifier()
	deduplicator := notifier.NewDeduplicator(config.AlertDedupeConfig{
		Enabled:     true,
		BlockoutMin: 15 * time.Minute,
	})
	stats := notifier.NewStats()

	mockFetcher := func(url string) ([]piaware.Aircraft, error) {
		return []piaware.Aircraft{}, nil
	}

	service := NewMonitorService(cfg, mockNotifier, deduplicator, stats, mockFetcher)

	if service == nil {
		t.Fatal("expected service to be created")
	}

	if service.cfg.DataURL != "http://test.com" {
		t.Errorf("expected DataURL 'http://test.com', got %s", service.cfg.DataURL)
	}
}

func TestMonitorServiceRunMonitoringCycleNoAircraft(t *testing.T) {
	cfg := config.Config{
		BaseLat:     40.7128,
		BaseLon:     -74.0060,
		RadiusKm:    25.0,
		AltitudeMax: 10000,
		DataURL:     "http://test.com",
	}

	mockNotifier := notifier.NewMockNotifier()
	deduplicator := notifier.NewDeduplicator(config.AlertDedupeConfig{
		Enabled:     true,
		BlockoutMin: 15 * time.Minute,
	})
	stats := notifier.NewStats()

	mockFetcher := func(url string) ([]piaware.Aircraft, error) {
		return []piaware.Aircraft{}, nil
	}

	service := NewMonitorService(cfg, mockNotifier, deduplicator, stats, mockFetcher)

	err := service.RunMonitoringCycle()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should have no notifications
	if mockNotifier.GetNotificationCount() != 0 {
		t.Errorf("expected 0 notifications, got %d", mockNotifier.GetNotificationCount())
	}
}

func TestMonitorServiceRunMonitoringCycleWithAircraft(t *testing.T) {
	cfg := config.Config{
		BaseLat:     40.7128,
		BaseLon:     -74.0060,
		RadiusKm:    25.0,
		AltitudeMax: 10000,
		DataURL:     "http://test.com",
	}

	mockNotifier := notifier.NewMockNotifier()
	deduplicator := notifier.NewDeduplicator(config.AlertDedupeConfig{
		Enabled:     true,
		BlockoutMin: 15 * time.Minute,
	})
	stats := notifier.NewStats()

	mockFetcher := func(url string) ([]piaware.Aircraft, error) {
		return piaware.CreateNearbyAircraft(), nil
	}

	service := NewMonitorService(cfg, mockNotifier, deduplicator, stats, mockFetcher)

	err := service.RunMonitoringCycle()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should have sent notifications for both aircraft
	notifications := mockNotifier.GetNotifications()
	if len(notifications) != 2 {
		t.Errorf("expected 2 notifications, got %d", len(notifications))
	}

	// Check notification content
	for i, notification := range notifications {
		if notification.AlertType != "aircraft_nearby" {
			t.Errorf("notification %d: expected alert_type 'aircraft_nearby', got %s", i, notification.AlertType)
		}
		if notification.Aircraft.Hex == "" {
			t.Errorf("notification %d: expected aircraft hex, got empty", i)
		}
	}
}

func TestMonitorServiceRunMonitoringCycleWithFetcherError(t *testing.T) {
	cfg := config.Config{
		BaseLat:     40.7128,
		BaseLon:     -74.0060,
		RadiusKm:    25.0,
		AltitudeMax: 10000,
		DataURL:     "http://test.com",
	}

	mockNotifier := notifier.NewMockNotifier()
	deduplicator := notifier.NewDeduplicator(config.AlertDedupeConfig{
		Enabled:     true,
		BlockoutMin: 15 * time.Minute,
	})
	stats := notifier.NewStats()

	mockFetcher := func(url string) ([]piaware.Aircraft, error) {
		return nil, &mockError{message: "network error"}
	}

	service := NewMonitorService(cfg, mockNotifier, deduplicator, stats, mockFetcher)

	err := service.RunMonitoringCycle()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "network error" {
		t.Errorf("expected error 'network error', got %v", err.Error())
	}

	// Should have no notifications due to error
	if mockNotifier.GetNotificationCount() != 0 {
		t.Errorf("expected 0 notifications, got %d", mockNotifier.GetNotificationCount())
	}
}

func TestMonitorServiceGetStats(t *testing.T) {
	cfg := config.Config{
		BaseLat:     40.7128,
		BaseLon:     -74.0060,
		RadiusKm:    25.0,
		AltitudeMax: 10000,
		DataURL:     "http://test.com",
	}

	mockNotifier := notifier.NewMockNotifier()
	deduplicator := notifier.NewDeduplicator(config.AlertDedupeConfig{
		Enabled:     true,
		BlockoutMin: 15 * time.Minute,
	})
	stats := notifier.NewStats()

	mockFetcher := func(url string) ([]piaware.Aircraft, error) {
		return []piaware.Aircraft{}, nil
	}

	service := NewMonitorService(cfg, mockNotifier, deduplicator, stats, mockFetcher)

	// Record some stats
	stats.RecordScrape()
	stats.RecordAircraft("ABC123")

	serviceStats := service.GetStats()

	if serviceStats["scrape_count"] != int64(1) {
		t.Errorf("expected scrape_count 1, got %v", serviceStats["scrape_count"])
	}

	if serviceStats["unique_aircraft"] != 1 {
		t.Errorf("expected unique_aircraft 1, got %v", serviceStats["unique_aircraft"])
	}
}

func TestMonitorServiceGetAircraftCounts(t *testing.T) {
	cfg := config.Config{
		BaseLat:     40.7128,
		BaseLon:     -74.0060,
		RadiusKm:    25.0,
		AltitudeMax: 10000,
		DataURL:     "http://test.com",
	}

	mockNotifier := notifier.NewMockNotifier()
	deduplicator := notifier.NewDeduplicator(config.AlertDedupeConfig{
		Enabled:     true,
		BlockoutMin: 15 * time.Minute,
	})
	stats := notifier.NewStats()

	mockFetcher := func(url string) ([]piaware.Aircraft, error) {
		return []piaware.Aircraft{}, nil
	}

	service := NewMonitorService(cfg, mockNotifier, deduplicator, stats, mockFetcher)

	// Record some aircraft
	stats.RecordAircraft("ABC123")
	stats.RecordAircraft("DEF456")

	totalSeen, inRange := service.GetAircraftCounts()

	if totalSeen != 2 {
		t.Errorf("expected totalSeen 2, got %d", totalSeen)
	}

	if inRange != 0 {
		t.Errorf("expected inRange 0, got %d", inRange)
	}
}

// mockError implements error interface for testing.
type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}
