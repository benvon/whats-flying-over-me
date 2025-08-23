package main

import (
	"fmt"
	"time"

	"github.com/example/whats-flying-over-me/internal/config"
	"github.com/example/whats-flying-over-me/internal/notifier"
	"github.com/example/whats-flying-over-me/internal/piaware"
)

// MonitorService handles the aircraft monitoring logic.
type MonitorService struct {
	cfg          config.Config
	notifier     notifier.Notifier
	deduplicator *notifier.Deduplicator
	stats        *notifier.Stats
	fetcher      AircraftFetcher
}

// NewMonitorService creates a new monitoring service.
func NewMonitorService(cfg config.Config, notifier notifier.Notifier, deduplicator *notifier.Deduplicator, stats *notifier.Stats, fetcher AircraftFetcher) *MonitorService {
	return &MonitorService{
		cfg:          cfg,
		notifier:     notifier,
		deduplicator: deduplicator,
		stats:        stats,
		fetcher:      fetcher,
	}
}

// RunMonitoringCycle executes one monitoring cycle.
func (m *MonitorService) RunMonitoringCycle() error {
	return m.check()
}

// GetStats returns the current statistics.
func (m *MonitorService) GetStats() map[string]interface{} {
	return m.stats.GetStats()
}

// check performs the aircraft check logic.
func (m *MonitorService) check() error {
	aircraft, err := m.fetcher(m.cfg.DataURL)
	if err != nil {
		return err
	}

	// Record all aircraft seen for statistics
	for _, a := range aircraft {
		m.stats.RecordAircraft(a.Hex)
	}

	nearby := piaware.FilterAircraft(aircraft, m.cfg.BaseLat, m.cfg.BaseLon, m.cfg.RadiusKm, m.cfg.AltitudeMax)

	if len(nearby) == 0 {
		// Log that no aircraft are in range
		return nil
	}

	alertCount := 0
	for _, a := range nearby {
		// Check if we should send an alert for this aircraft
		if !m.deduplicator.ShouldAlert(a) {
			// Skip duplicate alerts silently
			continue
		}

		// Create alert data
		alert := notifier.AlertData{
			Timestamp:   time.Now(),
			Aircraft:    a,
			AlertType:   "aircraft_nearby",
			Description: fmt.Sprintf("Aircraft %s detected within %.1f km at %d ft altitude", a.Hex, a.DistanceKm, a.AltBaro),
		}

		// Send notification
		if err := m.notifier.Notify(alert); err != nil {
			// Log notification failure but continue with other aircraft
			continue
		}

		alertCount++
	}

	return nil
}

// GetAircraftCounts returns the counts of aircraft seen and in range.
func (m *MonitorService) GetAircraftCounts() (totalSeen, inRange int) {
	// This would be implemented to return current counts
	// For now, we'll return the stats data
	stats := m.stats.GetStats()
	if uniqueCount, ok := stats["unique_aircraft"].(int); ok {
		totalSeen = uniqueCount
	}
	// Note: inRange would need to be tracked separately or calculated
	return totalSeen, 0
}
