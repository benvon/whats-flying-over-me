package notifier

import (
	"sync"
	"time"

	"github.com/example/whats-flying-over-me/internal/config"
	"github.com/example/whats-flying-over-me/internal/piaware"
)

// AlertRecord tracks when an aircraft was last alerted.
type AlertRecord struct {
	TailNumber  string
	Transponder string
	LastAlerted time.Time
}

// Deduplicator prevents duplicate alerts for the same aircraft within a time window.
type Deduplicator struct {
	cfg     config.AlertDedupeConfig
	records map[string]*AlertRecord // key: tailNumber + ":" + transponder
	mutex   sync.RWMutex
}

// NewDeduplicator creates a new deduplicator.
func NewDeduplicator(cfg config.AlertDedupeConfig) *Deduplicator {
	return &Deduplicator{
		cfg:     cfg,
		records: make(map[string]*AlertRecord),
	}
}

// ShouldAlert determines if an alert should be sent for the given aircraft.
func (d *Deduplicator) ShouldAlert(aircraft piaware.NearbyAircraft) bool {
	if !d.cfg.Enabled {
		return true
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	now := time.Now()

	// Create keys for both tail number and transponder
	tailKey := d.makeKey(aircraft.Flight, aircraft.Hex)
	transponderKey := d.makeKey("", aircraft.Hex)

	// Check if we should alert based on tail number
	shouldAlertByTail := true
	if record, exists := d.records[tailKey]; exists {
		if now.Sub(record.LastAlerted) < d.cfg.BlockoutMin {
			shouldAlertByTail = false
		}
	}

	// Check if we should alert based on transponder
	shouldAlertByTransponder := true
	if record, exists := d.records[transponderKey]; exists {
		if now.Sub(record.LastAlerted) < d.cfg.BlockoutMin {
			shouldAlertByTransponder = false
		}
	}

	// Alert if either condition is met (new tail number OR new transponder)
	shouldAlert := shouldAlertByTail || shouldAlertByTransponder

	if shouldAlert {
		// Update records
		if aircraft.Flight != "" {
			d.records[tailKey] = &AlertRecord{
				TailNumber:  aircraft.Flight,
				Transponder: aircraft.Hex,
				LastAlerted: now,
			}
		}

		d.records[transponderKey] = &AlertRecord{
			TailNumber:  aircraft.Flight,
			Transponder: aircraft.Hex,
			LastAlerted: now,
		}

		// Clean up old records
		d.cleanup(now)
	}

	return shouldAlert
}

// makeKey creates a key for the records map.
func (d *Deduplicator) makeKey(tailNumber, transponder string) string {
	if tailNumber == "" {
		return ":" + transponder
	}
	return tailNumber + ":" + transponder
}

// cleanup removes old records to prevent memory leaks.
func (d *Deduplicator) cleanup(now time.Time) {
	cutoff := now.Add(-d.cfg.BlockoutMin * 2) // Keep records for 2x blockout time

	for key, record := range d.records {
		if record.LastAlerted.Before(cutoff) {
			delete(d.records, key)
		}
	}
}

// GetStats returns statistics about the deduplicator.
func (d *Deduplicator) GetStats() map[string]interface{} {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return map[string]interface{}{
		"enabled":        d.cfg.Enabled,
		"blockout_min":   d.cfg.BlockoutMin.String(),
		"active_records": len(d.records),
	}
}
