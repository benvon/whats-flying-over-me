package cataloger

import (
	"context"
	"testing"

	"github.com/example/whats-flying-over-me/internal/piaware"
)

func TestNewElasticSearchCataloger(t *testing.T) {
	tests := []struct {
		name    string
		config  ElasticSearchConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: ElasticSearchConfig{
				Enabled: true,
				URL:     "http://localhost:9200",
				Index:   "aircraft",
			},
			wantErr: false,
		},
		{
			name: "disabled config",
			config: ElasticSearchConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "enabled without URL",
			config: ElasticSearchConfig{
				Enabled: true,
				Index:   "aircraft",
			},
			wantErr: true,
		},
		{
			name: "default values",
			config: ElasticSearchConfig{
				Enabled: true,
				URL:     "http://localhost:9200",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewElasticSearchCataloger(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewElasticSearchCataloger() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestElasticSearchCatalogerCatalogAircraft(t *testing.T) {
	// Test with disabled cataloger
	config := ElasticSearchConfig{
		Enabled: false,
	}
	cataloger, err := NewElasticSearchCataloger(config)
	if err != nil {
		t.Fatalf("Failed to create cataloger: %v", err)
	}

	aircraft := []piaware.Aircraft{
		{
			Hex:     "ABC123",
			Flight:  "TEST123",
			Lat:     37.6213,
			Lon:     -122.3790,
			AltBaro: 5000,
		},
	}

	ctx := context.Background()
	err = cataloger.CatalogAircraft(ctx, aircraft, 37.6213, -122.3790)
	if err != nil {
		t.Errorf("CatalogAircraft() should not fail when disabled: %v", err)
	}
}

func TestElasticSearchCatalogerHealthCheck(t *testing.T) {
	// Test with disabled cataloger
	config := ElasticSearchConfig{
		Enabled: false,
	}
	cataloger, err := NewElasticSearchCataloger(config)
	if err != nil {
		t.Fatalf("Failed to create cataloger: %v", err)
	}

	ctx := context.Background()
	err = cataloger.HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck() should not fail when disabled: %v", err)
	}
}

func TestElasticSearchCatalogerClose(t *testing.T) {
	config := ElasticSearchConfig{
		Enabled: false,
	}
	cataloger, err := NewElasticSearchCataloger(config)
	if err != nil {
		t.Fatalf("Failed to create cataloger: %v", err)
	}

	err = cataloger.Close()
	if err != nil {
		t.Errorf("Close() should not fail: %v", err)
	}
}

func TestCalculateDistance(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		expected float64
	}{
		{
			name:     "same point",
			lat1:     37.6213,
			lon1:     -122.3790,
			lat2:     37.6213,
			lon2:     -122.3790,
			expected: 0.0,
		},
		{
			name:     "different points",
			lat1:     37.6213,
			lon1:     -122.3790,
			lat2:     37.7213,
			lon2:     -122.3790,
			expected: 11.1, // Approximately 11.1 km
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateDistance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			if tt.name == "same point" {
				if result > 0.1 { // Allow small floating point errors
					t.Errorf("calculateDistance() = %v, expected close to 0", result)
				}
			} else {
				// Allow some tolerance for distance calculations
				if result < tt.expected-1 || result > tt.expected+1 {
					t.Errorf("calculateDistance() = %v, expected close to %v", result, tt.expected)
				}
			}
		})
	}
}
