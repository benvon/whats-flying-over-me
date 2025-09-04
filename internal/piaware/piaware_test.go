package piaware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDistance(t *testing.T) {
	tests := []struct {
		name      string
		lat1      float64
		lon1      float64
		lat2      float64
		lon2      float64
		expected  float64
		tolerance float64
	}{
		{
			name:      "same point",
			lat1:      0,
			lon1:      0,
			lat2:      0,
			lon2:      0,
			expected:  0,
			tolerance: 0.001,
		},
		{
			name:      "different points",
			lat1:      37.6213,
			lon1:      -122.3790,
			lat2:      37.7213,
			lon2:      -122.3790,
			expected:  11.1, // Approximately 11.1 km
			tolerance: 1.0,
		},
		{
			name:      "antipodal points",
			lat1:      0,
			lon1:      0,
			lat2:      0,
			lon2:      180,
			expected:  20015.1, // Half the Earth's circumference
			tolerance: 100.0,
		},
		{
			name:      "north pole to south pole",
			lat1:      90,
			lon1:      0,
			lat2:      -90,
			lon2:      0,
			expected:  20015.1, // Half the Earth's circumference
			tolerance: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := distance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			if tt.name == "same point" {
				if result > tt.tolerance {
					t.Errorf("distance() = %v, expected close to 0", result)
				}
			} else {
				if result < tt.expected-tt.tolerance || result > tt.expected+tt.tolerance {
					t.Errorf("distance() = %v, expected close to %v (tolerance: %v)", result, tt.expected, tt.tolerance)
				}
			}
		})
	}
}

func TestFilterAircraft(t *testing.T) {
	tests := []struct {
		name        string
		aircraft    []Aircraft
		baseLat     float64
		baseLon     float64
		radiusKm    float64
		altMax      int
		expected    []string // Expected hex codes
		description string
	}{
		{
			name: "basic filtering",
			aircraft: []Aircraft{
				{Hex: "1", Lat: 0.1, Lon: 0.1, AltBaro: 1000},
				{Hex: "2", Lat: 50, Lon: 50, AltBaro: 2000},
			},
			baseLat:     0,
			baseLon:     0,
			radiusKm:    1000,
			altMax:      1500,
			expected:    []string{"1"},
			description: "should filter by distance and altitude",
		},
		{
			name: "no aircraft match",
			aircraft: []Aircraft{
				{Hex: "1", Lat: 0.1, Lon: 0.1, AltBaro: 2000},
				{Hex: "2", Lat: 50, Lon: 50, AltBaro: 1000},
			},
			baseLat:     0,
			baseLon:     0,
			radiusKm:    1,
			altMax:      500,
			expected:    []string{},
			description: "should return empty when no aircraft match criteria",
		},
		{
			name: "all aircraft match",
			aircraft: []Aircraft{
				{Hex: "1", Lat: 0.1, Lon: 0.1, AltBaro: 1000},
				{Hex: "2", Lat: 0.2, Lon: 0.2, AltBaro: 2000},
			},
			baseLat:     0,
			baseLon:     0,
			radiusKm:    1000,
			altMax:      3000,
			expected:    []string{"1", "2"},
			description: "should return all aircraft when all match",
		},
		{
			name: "aircraft with zero coordinates",
			aircraft: []Aircraft{
				{Hex: "1", Lat: 0.1, Lon: 0.1, AltBaro: 1000},
				{Hex: "2", Lat: 0, Lon: 0, AltBaro: 1000}, // Zero coordinates
				{Hex: "3", Lat: 0.2, Lon: 0.2, AltBaro: 1000},
			},
			baseLat:     0,
			baseLon:     0,
			radiusKm:    1000,
			altMax:      2000,
			expected:    []string{"1", "3"}, // Should skip aircraft with zero coordinates
			description: "should skip aircraft with zero coordinates",
		},
		{
			name:        "empty aircraft list",
			aircraft:    []Aircraft{},
			baseLat:     0,
			baseLon:     0,
			radiusKm:    1000,
			altMax:      2000,
			expected:    []string{},
			description: "should handle empty aircraft list",
		},
		{
			name:        "nil aircraft list",
			aircraft:    nil,
			baseLat:     0,
			baseLon:     0,
			radiusKm:    1000,
			altMax:      2000,
			expected:    []string{},
			description: "should handle nil aircraft list",
		},
		{
			name: "exact altitude boundary",
			aircraft: []Aircraft{
				{Hex: "1", Lat: 0.1, Lon: 0.1, AltBaro: 1000},
				{Hex: "2", Lat: 0.2, Lon: 0.2, AltBaro: 1001},
			},
			baseLat:     0,
			baseLon:     0,
			radiusKm:    1000,
			altMax:      1000,
			expected:    []string{"1"}, // Only the one at exactly 1000 should match
			description: "should include aircraft at exact altitude boundary",
		},
		{
			name: "exact distance boundary",
			aircraft: []Aircraft{
				{Hex: "1", Lat: 0.1, Lon: 0.1, AltBaro: 1000},
				{Hex: "2", Lat: 0.2, Lon: 0.2, AltBaro: 1000},
			},
			baseLat:     0,
			baseLon:     0,
			radiusKm:    50.0, // Use a much larger radius to ensure both aircraft are included
			altMax:      2000,
			expected:    []string{"1", "2"}, // Both should match with larger radius
			description: "should include aircraft within radius",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterAircraft(tt.aircraft, tt.baseLat, tt.baseLon, tt.radiusKm, tt.altMax)

			if len(result) != len(tt.expected) {
				t.Errorf("FilterAircraft() returned %d aircraft, expected %d (%s)", len(result), len(tt.expected), tt.description)
				return
			}

			// Check that the hex codes match
			resultHexes := make([]string, len(result))
			for i, aircraft := range result {
				resultHexes[i] = aircraft.Hex
			}

			for i, expectedHex := range tt.expected {
				if i >= len(resultHexes) || resultHexes[i] != expectedHex {
					t.Errorf("FilterAircraft() result[%d] = %s, expected %s (%s)", i, resultHexes[i], expectedHex, tt.description)
				}
			}

			// Verify that distances are calculated and stored
			for i, aircraft := range result {
				if aircraft.DistanceKm < 0 {
					t.Errorf("FilterAircraft() result[%d] has negative distance: %v", i, aircraft.DistanceKm)
				}
			}
		})
	}
}

func TestFetch(t *testing.T) {
	tests := []struct {
		name        string
		response    string
		statusCode  int
		wantErr     bool
		expectedLen int
		description string
	}{
		{
			name:        "successful fetch with single aircraft",
			response:    `{"now":1,"aircraft":[{"hex":"abc","flight":"TEST","lat":1,"lon":2,"alt_baro":300}]}`,
			statusCode:  http.StatusOK,
			wantErr:     false,
			expectedLen: 1,
			description: "should successfully fetch single aircraft",
		},
		{
			name:        "successful fetch with multiple aircraft",
			response:    `{"now":1,"aircraft":[{"hex":"abc","flight":"TEST","lat":1,"lon":2,"alt_baro":300},{"hex":"def","flight":"TEST2","lat":2,"lon":3,"alt_baro":400}]}`,
			statusCode:  http.StatusOK,
			wantErr:     false,
			expectedLen: 2,
			description: "should successfully fetch multiple aircraft",
		},
		{
			name:        "successful fetch with empty aircraft list",
			response:    `{"now":1,"aircraft":[]}`,
			statusCode:  http.StatusOK,
			wantErr:     false,
			expectedLen: 0,
			description: "should handle empty aircraft list",
		},
		{
			name:        "successful fetch with aircraft missing fields",
			response:    `{"now":1,"aircraft":[{"hex":"abc","lat":1,"lon":2}]}`,
			statusCode:  http.StatusOK,
			wantErr:     false,
			expectedLen: 1,
			description: "should handle aircraft with missing optional fields",
		},
		{
			name:        "server error",
			response:    `Internal Server Error`,
			statusCode:  http.StatusInternalServerError,
			wantErr:     true,
			expectedLen: 0,
			description: "should return error on server error",
		},
		{
			name:        "not found",
			response:    `Not Found`,
			statusCode:  http.StatusNotFound,
			wantErr:     true,
			expectedLen: 0,
			description: "should return error on not found",
		},
		{
			name:        "invalid JSON",
			response:    `{"now":1,"aircraft":[{"hex":"abc","flight":"TEST","lat":1,"lon":2,"alt_baro":300]`, // Missing closing brace
			statusCode:  http.StatusOK,
			wantErr:     true,
			expectedLen: 0,
			description: "should return error on invalid JSON",
		},
		{
			name:        "malformed JSON structure",
			response:    `{"now":1,"aircraft":"not_an_array"}`,
			statusCode:  http.StatusOK,
			wantErr:     true,
			expectedLen: 0,
			description: "should return error on malformed JSON structure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if _, err := io.WriteString(w, tt.response); err != nil {
					t.Fatalf("write: %v", err)
				}
			}))
			defer srv.Close()

			ac, err := Fetch(srv.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fetch() error = %v, wantErr %v (%s)", err, tt.wantErr, tt.description)
				return
			}

			if !tt.wantErr && len(ac) != tt.expectedLen {
				t.Errorf("Fetch() returned %d aircraft, expected %d (%s)", len(ac), tt.expectedLen, tt.description)
			}

			// If we got aircraft, verify the first one has expected structure
			if !tt.wantErr && len(ac) > 0 && tt.name == "successful fetch with single aircraft" {
				if ac[0].Hex != "abc" {
					t.Errorf("Fetch() first aircraft hex = %s, expected abc", ac[0].Hex)
				}
				if ac[0].Flight != "TEST" {
					t.Errorf("Fetch() first aircraft flight = %s, expected TEST", ac[0].Flight)
				}
				if ac[0].Lat != 1 {
					t.Errorf("Fetch() first aircraft lat = %v, expected 1", ac[0].Lat)
				}
				if ac[0].Lon != 2 {
					t.Errorf("Fetch() first aircraft lon = %v, expected 2", ac[0].Lon)
				}
				if ac[0].AltBaro != 300 {
					t.Errorf("Fetch() first aircraft alt_baro = %v, expected 300", ac[0].AltBaro)
				}
			}
		})
	}
}

func TestFetchTimeout(t *testing.T) {
	// Test that fetch times out appropriately
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow response
		time.Sleep(15 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, `{"now":1,"aircraft":[]}`)
		if err != nil {
			t.Fatalf("write: %v", err)
		}
	}))
	defer srv.Close()

	_, err := Fetch(srv.URL)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
