package cataloger

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/example/whats-flying-over-me/internal/piaware"
)

// MockCataloger is a mock implementation of the Cataloger interface for testing
type MockCataloger struct {
	mu                sync.RWMutex
	catalogedAircraft []AircraftRecord
	catalogCalls      int
	healthCheckCalls  int
	shouldFail        bool
	failMessage       string
	closeCalls        int
}

// NewMockCataloger creates a new mock cataloger
func NewMockCataloger() *MockCataloger {
	return &MockCataloger{
		catalogedAircraft: make([]AircraftRecord, 0),
	}
}

// SetShouldFail configures the mock to fail operations
func (m *MockCataloger) SetShouldFail(shouldFail bool, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
	m.failMessage = message
}

// CatalogAircraft catalogs aircraft data (mock implementation)
func (m *MockCataloger) CatalogAircraft(ctx context.Context, aircraft []piaware.Aircraft, baseLat, baseLon float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		return fmt.Errorf("mock cataloger failure: %s", m.failMessage)
	}

	m.catalogCalls++
	timestamp := time.Now()

	for _, a := range aircraft {
		var distanceKm float64
		if a.Lat != 0 && a.Lon != 0 {
			distanceKm = calculateDistance(baseLat, baseLon, a.Lat, a.Lon)
		}

		record := AircraftRecord{
			Hex:        a.Hex,
			Flight:     a.Flight,
			Lat:        a.Lat,
			Lon:        a.Lon,
			AltBaro:    a.AltBaro,
			Timestamp:  timestamp,
			DistanceKm: distanceKm,
			BaseLat:    baseLat,
			BaseLon:    baseLon,
		}

		m.catalogedAircraft = append(m.catalogedAircraft, record)
	}

	return nil
}

// HealthCheck performs a health check (mock implementation)
func (m *MockCataloger) HealthCheck(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.healthCheckCalls++

	if m.shouldFail {
		return fmt.Errorf("mock cataloger failure: %s", m.failMessage)
	}

	return nil
}

// Close closes the mock cataloger (mock implementation)
func (m *MockCataloger) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closeCalls++
	return nil
}

// GetCatalogedAircraft returns all cataloged aircraft records
func (m *MockCataloger) GetCatalogedAircraft() []AircraftRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]AircraftRecord, len(m.catalogedAircraft))
	copy(result, m.catalogedAircraft)
	return result
}

// GetCatalogCalls returns the number of catalog calls
func (m *MockCataloger) GetCatalogCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.catalogCalls
}

// GetHealthCheckCalls returns the number of health check calls
func (m *MockCataloger) GetHealthCheckCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.healthCheckCalls
}

// GetCloseCalls returns the number of close calls
func (m *MockCataloger) GetCloseCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.closeCalls
}

// Reset resets all counters and data
func (m *MockCataloger) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.catalogedAircraft = make([]AircraftRecord, 0)
	m.catalogCalls = 0
	m.healthCheckCalls = 0
	m.closeCalls = 0
	m.shouldFail = false
	m.failMessage = ""
}
