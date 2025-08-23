package notifier

import (
	"sync"
	"time"
)

// Stats tracks program statistics and metrics.
type Stats struct {
	startTime      time.Time
	scrapeCount    int64
	scrapeFailures int64
	uniqueAircraft map[string]time.Time // hex -> first seen time
	mutex          sync.RWMutex
}

// NewStats creates a new statistics tracker.
func NewStats() *Stats {
	return &Stats{
		startTime:      time.Now(),
		uniqueAircraft: make(map[string]time.Time),
	}
}

// RecordScrape records a successful scrape.
func (s *Stats) RecordScrape() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.scrapeCount++
}

// RecordScrapeFailure records a failed scrape.
func (s *Stats) RecordScrapeFailure() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.scrapeFailures++
}

// RecordAircraft records a new aircraft sighting.
func (s *Stats) RecordAircraft(hex string) {
	if hex == "" {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.uniqueAircraft[hex]; !exists {
		s.uniqueAircraft[hex] = time.Now()
	}
}

// GetStats returns a copy of current statistics.
func (s *Stats) GetStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	uptime := time.Since(s.startTime)

	return map[string]interface{}{
		"uptime":          uptime.String(),
		"uptime_seconds":  int64(uptime.Seconds()),
		"scrape_count":    s.scrapeCount,
		"scrape_failures": s.scrapeFailures,
		"success_rate":    s.calculateSuccessRate(),
		"unique_aircraft": len(s.uniqueAircraft),
		"start_time":      s.startTime.Format(time.RFC3339),
	}
}

// calculateSuccessRate calculates the success rate as a percentage.
func (s *Stats) calculateSuccessRate() float64 {
	if s.scrapeCount == 0 {
		return 100.0
	}

	total := s.scrapeCount + s.scrapeFailures
	if total == 0 {
		return 100.0
	}

	successRate := float64(s.scrapeCount) / float64(total) * 100.0
	return successRate
}

// GetUniqueAircraftList returns a list of unique aircraft seen.
func (s *Stats) GetUniqueAircraftList() []map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var result []map[string]interface{}
	for hex, firstSeen := range s.uniqueAircraft {
		result = append(result, map[string]interface{}{
			"hex":        hex,
			"first_seen": firstSeen.Format(time.RFC3339),
		})
	}

	return result
}
