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
	tests := []struct {
		name     string
		config   ElasticSearchConfig
		aircraft []piaware.Aircraft
		baseLat  float64
		baseLon  float64
		wantErr  bool
	}{
		{
			name: "disabled cataloger",
			config: ElasticSearchConfig{
				Enabled: false,
			},
			aircraft: []piaware.Aircraft{
				{
					Hex:     "ABC123",
					Flight:  "TEST123",
					Lat:     37.6213,
					Lon:     -122.3790,
					AltBaro: 5000,
				},
			},
			baseLat: 37.6213,
			baseLon: -122.3790,
			wantErr: false,
		},
		{
			name: "empty aircraft list",
			config: ElasticSearchConfig{
				Enabled: true,
				URL:     "http://localhost:9200",
			},
			aircraft: []piaware.Aircraft{},
			baseLat:  37.6213,
			baseLon:  -122.3790,
			wantErr:  false, // Empty list should not fail
		},
		{
			name: "nil aircraft list",
			config: ElasticSearchConfig{
				Enabled: true,
				URL:     "http://localhost:9200",
			},
			aircraft: nil,
			baseLat:  37.6213,
			baseLon:  -122.3790,
			wantErr:  false, // Nil list should not fail
		},
		{
			name: "aircraft with coordinates",
			config: ElasticSearchConfig{
				Enabled: true,
				URL:     "http://localhost:9200",
			},
			aircraft: []piaware.Aircraft{
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
					Lat:     0, // No coordinates
					Lon:     0,
					AltBaro: 3000,
				},
			},
			baseLat: 37.6213,
			baseLon: -122.3790,
			wantErr: true, // Will fail due to no real server
		},
		{
			name: "aircraft without coordinates",
			config: ElasticSearchConfig{
				Enabled: true,
				URL:     "http://localhost:9200",
			},
			aircraft: []piaware.Aircraft{
				{
					Hex:     "GHI789",
					Flight:  "TEST789",
					Lat:     0,
					Lon:     0,
					AltBaro: 2000,
				},
			},
			baseLat: 37.6213,
			baseLon: -122.3790,
			wantErr: true, // Will fail due to no real server
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cataloger, err := NewElasticSearchCataloger(tt.config)
			if err != nil {
				t.Fatalf("Failed to create cataloger: %v", err)
			}

			ctx := context.Background()
			err = cataloger.CatalogAircraft(ctx, tt.aircraft, tt.baseLat, tt.baseLon)
			if (err != nil) != tt.wantErr {
				t.Errorf("CatalogAircraft() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestElasticSearchCatalogerHealthCheck(t *testing.T) {
	tests := []struct {
		name    string
		config  ElasticSearchConfig
		wantErr bool
	}{
		{
			name: "disabled cataloger",
			config: ElasticSearchConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "enabled cataloger with auth",
			config: ElasticSearchConfig{
				Enabled:  true,
				URL:      "http://localhost:9200",
				Username: "user",
				Password: "pass",
			},
			wantErr: true, // Will fail due to no real server
		},
		{
			name: "enabled cataloger without auth",
			config: ElasticSearchConfig{
				Enabled: true,
				URL:     "http://localhost:9200",
			},
			wantErr: true, // Will fail due to no real server
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cataloger, err := NewElasticSearchCataloger(tt.config)
			if err != nil {
				t.Fatalf("Failed to create cataloger: %v", err)
			}

			ctx := context.Background()
			err = cataloger.HealthCheck(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("HealthCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
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

func TestElasticSearchCatalogerSendBulkRequest(t *testing.T) {
	tests := []struct {
		name        string
		config      ElasticSearchConfig
		body        []byte
		wantErr     bool
		description string
	}{
		{
			name: "enabled cataloger with retries",
			config: ElasticSearchConfig{
				Enabled:    true,
				URL:        "http://localhost:9200",
				MaxRetries: 2,
			},
			body:        []byte("test body"),
			wantErr:     true, // Will fail due to no real server
			description: "should attempt retries and fail",
		},
		{
			name: "enabled cataloger with no retries",
			config: ElasticSearchConfig{
				Enabled:    true,
				URL:        "http://localhost:9200",
				MaxRetries: 0,
			},
			body:        []byte("test body"),
			wantErr:     true, // Will fail due to no real server
			description: "should fail immediately with no retries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cataloger, err := NewElasticSearchCataloger(tt.config)
			if err != nil {
				t.Fatalf("Failed to create cataloger: %v", err)
			}

			ctx := context.Background()
			err = cataloger.sendBulkRequest(ctx, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("sendBulkRequest() error = %v, wantErr %v (%s)", err, tt.wantErr, tt.description)
			}
		})
	}
}

func TestElasticSearchCatalogerDoBulkRequest(t *testing.T) {
	tests := []struct {
		name        string
		config      ElasticSearchConfig
		body        []byte
		wantErr     bool
		description string
	}{
		{
			name: "enabled cataloger",
			config: ElasticSearchConfig{
				Enabled: true,
				URL:     "http://localhost:9200",
			},
			body:        []byte("test body"),
			wantErr:     true, // Will fail due to no real server
			description: "should fail without real server",
		},
		{
			name: "enabled cataloger with auth",
			config: ElasticSearchConfig{
				Enabled:  true,
				URL:      "http://localhost:9200",
				Username: "user",
				Password: "pass",
			},
			body:        []byte("test body"),
			wantErr:     true, // Will fail due to no real server
			description: "should fail without real server even with auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cataloger, err := NewElasticSearchCataloger(tt.config)
			if err != nil {
				t.Fatalf("Failed to create cataloger: %v", err)
			}

			ctx := context.Background()
			err = cataloger.doBulkRequest(ctx, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("doBulkRequest() error = %v, wantErr %v (%s)", err, tt.wantErr, tt.description)
			}
		})
	}
}
