package piaware

import (
	"errors"
)

// MockClient is a mock implementation for testing piaware data fetching.
type MockClient struct {
	aircraft    []Aircraft
	shouldFail  bool
	failMessage string
	callCount   int
}

// NewMockClient creates a new mock piaware client.
func NewMockClient() *MockClient {
	return &MockClient{
		aircraft: make([]Aircraft, 0),
	}
}

// SetAircraft sets the aircraft data that the mock will return.
func (m *MockClient) SetAircraft(aircraft []Aircraft) {
	m.aircraft = aircraft
}

// SetShouldFail configures the mock to simulate failures.
func (m *MockClient) SetShouldFail(shouldFail bool, message string) {
	m.shouldFail = shouldFail
	m.failMessage = message
}

// Fetch simulates fetching aircraft data.
func (m *MockClient) Fetch(url string) ([]Aircraft, error) {
	m.callCount++

	if m.shouldFail {
		return nil, errors.New(m.failMessage)
	}

	return m.aircraft, nil
}

// GetCallCount returns the number of times Fetch was called.
func (m *MockClient) GetCallCount() int {
	return m.callCount
}

// CreateTestAircraft creates test aircraft data for testing.
func CreateTestAircraft() []Aircraft {
	return []Aircraft{
		{
			Hex:     "ABC123",
			Flight:  "TEST1",
			Lat:     40.7128,
			Lon:     -74.0060,
			AltBaro: 5000,
		},
		{
			Hex:     "DEF456",
			Flight:  "TEST2",
			Lat:     40.7500,
			Lon:     -74.0000,
			AltBaro: 8000,
		},
		{
			Hex:     "GHI789",
			Flight:  "TEST3",
			Lat:     40.8000,
			Lon:     -74.1000,
			AltBaro: 12000,
		},
	}
}

// CreateNearbyAircraft creates aircraft that would be in range of the test coordinates.
func CreateNearbyAircraft() []Aircraft {
	return []Aircraft{
		{
			Hex:     "NEAR1",
			Flight:  "NEAR1",
			Lat:     40.7128,
			Lon:     -74.0060,
			AltBaro: 5000,
		},
		{
			Hex:     "NEAR2",
			Flight:  "NEAR2",
			Lat:     40.7200,
			Lon:     -74.0100,
			AltBaro: 3000,
		},
	}
}

// CreateFarAircraft creates aircraft that would be out of range.
func CreateFarAircraft() []Aircraft {
	return []Aircraft{
		{
			Hex:     "FAR1",
			Flight:  "FAR1",
			Lat:     41.0000,
			Lon:     -74.5000,
			AltBaro: 5000,
		},
		{
			Hex:     "FAR2",
			Flight:  "FAR2",
			Lat:     40.5000,
			Lon:     -73.5000,
			AltBaro: 8000,
		},
	}
}
