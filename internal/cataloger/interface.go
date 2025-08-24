package cataloger

import (
	"context"
	"time"

	"github.com/example/whats-flying-over-me/internal/piaware"
)

// AircraftRecord represents a cataloged aircraft record
type AircraftRecord struct {
	Hex        string    `json:"hex"`
	Flight     string    `json:"flight"`
	Lat        float64   `json:"lat"`
	Lon        float64   `json:"lon"`
	AltBaro    int       `json:"alt_baro"`
	Timestamp  time.Time `json:"timestamp"`
	DistanceKm float64   `json:"distance_km"`
	BaseLat    float64   `json:"base_lat"`
	BaseLon    float64   `json:"base_lon"`
}

// Cataloger defines the interface for cataloging aircraft data
type Cataloger interface {
	// CatalogAircraft catalogs a batch of aircraft data
	CatalogAircraft(ctx context.Context, aircraft []piaware.Aircraft, baseLat, baseLon float64) error

	// HealthCheck performs a health check on the cataloging system
	HealthCheck(ctx context.Context) error

	// Close closes the cataloger and releases resources
	Close() error
}

// NoOpCataloger is a no-operation cataloger that does nothing
type NoOpCataloger struct{}

func (n *NoOpCataloger) CatalogAircraft(ctx context.Context, aircraft []piaware.Aircraft, baseLat, baseLon float64) error {
	return nil
}

func (n *NoOpCataloger) HealthCheck(ctx context.Context) error {
	return nil
}

func (n *NoOpCataloger) Close() error {
	return nil
}
