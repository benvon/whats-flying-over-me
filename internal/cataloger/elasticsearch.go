package cataloger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/example/whats-flying-over-me/internal/piaware"
)

// ElasticSearchConfig holds ElasticSearch connection settings
type ElasticSearchConfig struct {
	Enabled    bool
	URL        string
	Index      string
	Username   string
	Password   string
	Timeout    time.Duration
	MaxRetries int
}

// ElasticSearchCataloger implements the Cataloger interface for ElasticSearch
type ElasticSearchCataloger struct {
	config     ElasticSearchConfig
	httpClient *http.Client
	indexURL   string
}

// NewElasticSearchCataloger creates a new ElasticSearch cataloger
func NewElasticSearchCataloger(config ElasticSearchConfig) (*ElasticSearchCataloger, error) {
	if !config.Enabled {
		return &ElasticSearchCataloger{}, nil
	}

	if config.URL == "" {
		return nil, fmt.Errorf("ElasticSearch URL is required when enabled")
	}

	if config.Index == "" {
		config.Index = "aircraft"
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	// Build index URL
	indexURL := fmt.Sprintf("%s/%s/_doc", config.URL, config.Index)

	return &ElasticSearchCataloger{
		config:     config,
		httpClient: httpClient,
		indexURL:   indexURL,
	}, nil
}

// CatalogAircraft catalogs aircraft data to ElasticSearch
func (e *ElasticSearchCataloger) CatalogAircraft(ctx context.Context, aircraft []piaware.Aircraft, baseLat, baseLon float64) error {
	if !e.config.Enabled {
		return nil
	}

	if len(aircraft) == 0 {
		return nil
	}

	// Create bulk request body
	var bulkBody bytes.Buffer
	timestamp := time.Now()

	for _, a := range aircraft {
		// Calculate distance if coordinates are available
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

		// Add index action
		indexAction := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": e.config.Index,
			},
		}
		indexActionBytes, _ := json.Marshal(indexAction)
		bulkBody.Write(indexActionBytes)
		bulkBody.WriteString("\n")

		// Add document
		recordBytes, _ := json.Marshal(record)
		bulkBody.Write(recordBytes)
		bulkBody.WriteString("\n")
	}

	// Send bulk request
	return e.sendBulkRequest(ctx, bulkBody.Bytes())
}

// sendBulkRequest sends a bulk request to ElasticSearch with retry logic
func (e *ElasticSearchCataloger) sendBulkRequest(ctx context.Context, body []byte) error {
	var lastErr error

	for attempt := 0; attempt <= e.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry with exponential backoff
			waitTime := time.Duration(attempt) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
			}
		}

		err := e.doBulkRequest(ctx, body)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	return fmt.Errorf("failed to send bulk request after %d attempts: %w", e.config.MaxRetries+1, lastErr)
}

// doBulkRequest performs a single bulk request to ElasticSearch
func (e *ElasticSearchCataloger) doBulkRequest(ctx context.Context, body []byte) error {
	// Use bulk API endpoint
	bulkURL := fmt.Sprintf("%s/_bulk", e.config.URL)

	req, err := http.NewRequestWithContext(ctx, "POST", bulkURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-ndjson")

	// Add authentication if provided
	if e.config.Username != "" && e.config.Password != "" {
		req.SetBasicAuth(e.config.Username, e.config.Password)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for error details
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ElasticSearch request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// HealthCheck performs a health check on ElasticSearch
func (e *ElasticSearchCataloger) HealthCheck(ctx context.Context) error {
	if !e.config.Enabled {
		return nil
	}

	healthURL := fmt.Sprintf("%s/_cluster/health", e.config.URL)

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	// Add authentication if provided
	if e.config.Username != "" && e.config.Password != "" {
		req.SetBasicAuth(e.config.Username, e.config.Password)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ElasticSearch health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// Close closes the ElasticSearch cataloger
func (e *ElasticSearchCataloger) Close() error {
	// Nothing to close for HTTP client
	return nil
}

// calculateDistance calculates the haversine distance in kilometers between two points
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth radius in kilometers

	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	// Haversine formula
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
