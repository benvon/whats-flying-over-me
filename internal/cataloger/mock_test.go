package cataloger

import (
	"context"
	"testing"

	"github.com/example/whats-flying-over-me/internal/piaware"
)

func TestMockCatalogerCatalogAircraft(t *testing.T) {
	mock := NewMockCataloger()
	ctx := context.Background()

	aircraft := []piaware.Aircraft{
		{
			Hex:     "ABC123",
			Flight:  "TEST123",
			Lat:     37.6213,
			Lon:     -122.3790,
			AltBaro: 5000,
		},
		{
			Hex:     "DEF456",
			Flight:  "TEST456",
			Lat:     37.7213,
			Lon:     -122.3790,
			AltBaro: 8000,
		},
	}

	baseLat, baseLon := 37.6213, -122.3790

	// Test successful cataloging
	err := mock.CatalogAircraft(ctx, aircraft, baseLat, baseLon)
	if err != nil {
		t.Errorf("CatalogAircraft() failed: %v", err)
	}

	if mock.GetCatalogCalls() != 1 {
		t.Errorf("Expected 1 catalog call, got %d", mock.GetCatalogCalls())
	}

	cataloged := mock.GetCatalogedAircraft()
	if len(cataloged) != 2 {
		t.Errorf("Expected 2 aircraft records, got %d", len(cataloged))
	}

	// Verify first aircraft
	if cataloged[0].Hex != "ABC123" {
		t.Errorf("Expected hex ABC123, got %s", cataloged[0].Hex)
	}
	if cataloged[0].Flight != "TEST123" {
		t.Errorf("Expected flight TEST123, got %s", cataloged[0].Flight)
	}
	if cataloged[0].DistanceKm != 0.0 {
		t.Errorf("Expected distance 0.0, got %f", cataloged[0].DistanceKm)
	}

	// Verify second aircraft
	if cataloged[1].Hex != "DEF456" {
		t.Errorf("Expected hex DEF456, got %s", cataloged[1].Hex)
	}
	if cataloged[1].DistanceKm <= 0.0 {
		t.Errorf("Expected positive distance, got %f", cataloged[1].DistanceKm)
	}
}

func TestMockCatalogerHealthCheck(t *testing.T) {
	mock := NewMockCataloger()
	ctx := context.Background()

	// Test successful health check
	err := mock.HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck() failed: %v", err)
	}

	if mock.GetHealthCheckCalls() != 1 {
		t.Errorf("Expected 1 health check call, got %d", mock.GetHealthCheckCalls())
	}
}

func TestMockCatalogerClose(t *testing.T) {
	mock := NewMockCataloger()

	err := mock.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	if mock.GetCloseCalls() != 1 {
		t.Errorf("Expected 1 close call, got %d", mock.GetCloseCalls())
	}
}

func TestMockCatalogerSetShouldFail(t *testing.T) {
	mock := NewMockCataloger()
	ctx := context.Background()

	// Set to fail
	mock.SetShouldFail(true, "test failure")

	aircraft := []piaware.Aircraft{
		{
			Hex:     "ABC123",
			Flight:  "TEST123",
			Lat:     37.6213,
			Lon:     -122.3790,
			AltBaro: 5000,
		},
	}

	// Test that cataloging fails
	err := mock.CatalogAircraft(ctx, aircraft, 37.6213, -122.3790)
	if err == nil {
		t.Error("Expected CatalogAircraft to fail when shouldFail is true")
	}
	if err.Error() != "mock cataloger failure: test failure" {
		t.Errorf("Expected error 'mock cataloger failure: test failure', got '%v'", err)
	}

	// Test that health check fails
	err = mock.HealthCheck(ctx)
	if err == nil {
		t.Error("Expected HealthCheck to fail when shouldFail is true")
	}
	if err.Error() != "mock cataloger failure: test failure" {
		t.Errorf("Expected error 'mock cataloger failure: test failure', got '%v'", err)
	}
}

func TestMockCatalogerReset(t *testing.T) {
	mock := NewMockCataloger()
	ctx := context.Background()

	// Perform some operations
	aircraft := []piaware.Aircraft{
		{
			Hex:     "ABC123",
			Flight:  "TEST123",
			Lat:     37.6213,
			Lon:     -122.3790,
			AltBaro: 5000,
		},
	}

	mock.CatalogAircraft(ctx, aircraft, 37.6213, -122.3790)
	mock.HealthCheck(ctx)
	mock.Close()

	// Verify operations were recorded
	if mock.GetCatalogCalls() != 1 || mock.GetHealthCheckCalls() != 1 || mock.GetCloseCalls() != 1 {
		t.Error("Expected operations to be recorded before reset")
	}

	// Reset
	mock.Reset()

	// Verify everything is reset
	if mock.GetCatalogCalls() != 0 {
		t.Errorf("Expected 0 catalog calls after reset, got %d", mock.GetCatalogCalls())
	}
	if mock.GetHealthCheckCalls() != 0 {
		t.Errorf("Expected 0 health check calls after reset, got %d", mock.GetHealthCheckCalls())
	}
	if mock.GetCloseCalls() != 0 {
		t.Errorf("Expected 0 close calls after reset, got %d", mock.GetCloseCalls())
	}
	if len(mock.GetCatalogedAircraft()) != 0 {
		t.Errorf("Expected 0 cataloged aircraft after reset, got %d", len(mock.GetCatalogedAircraft()))
	}
}
