package notifier

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/example/whats-flying-over-me/internal/config"
	"github.com/example/whats-flying-over-me/internal/piaware"
)

func TestNewWebhookWithClient(t *testing.T) {
	cfg := config.WebhookConfig{
		Enabled: true,
		URL:     "http://localhost:8080/webhook",
		Timeout: 5 * time.Second,
	}

	mockClient := NewMockHTTPClient()
	webhook := NewWebhookWithClient(cfg, mockClient)

	if webhook == nil {
		t.Fatal("expected webhook to be created")
	}

	if webhook.client != mockClient {
		t.Error("expected webhook to use the provided client")
	}
}

func TestWebhookNotifySuccess(t *testing.T) {
	cfg := config.WebhookConfig{
		Enabled: true,
		URL:     "http://localhost:8080/webhook",
		Timeout: 5 * time.Second,
	}

	mockClient := NewMockHTTPClient()
	webhook := NewWebhookWithClient(cfg, mockClient)

	alert := AlertData{
		Timestamp:   time.Now(),
		Aircraft:    piaware.NearbyAircraft{},
		AlertType:   "test",
		Description: "test alert",
	}

	err := webhook.Notify(alert)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check that request was made
	if mockClient.GetRequestCount() != 1 {
		t.Errorf("expected 1 request, got %d", mockClient.GetRequestCount())
	}

	lastRequest := mockClient.GetLastRequest()
	if lastRequest == nil {
		t.Fatal("expected request to be recorded")
	}

	if lastRequest.URL != "http://localhost:8080/webhook" {
		t.Errorf("expected URL 'http://localhost:8080/webhook', got %s", lastRequest.URL)
	}

	if lastRequest.ContentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %s", lastRequest.ContentType)
	}

	// Check that alert data was sent
	var receivedAlert AlertData
	if err := json.Unmarshal(lastRequest.Body, &receivedAlert); err != nil {
		t.Fatalf("failed to unmarshal request body: %v", err)
	}

	if receivedAlert.AlertType != "test" {
		t.Errorf("expected alert_type 'test', got %s", receivedAlert.AlertType)
	}
}

func TestWebhookNotifyHTTPFailure(t *testing.T) {
	cfg := config.WebhookConfig{
		Enabled: true,
		URL:     "http://localhost:8080/webhook",
		Timeout: 5 * time.Second,
	}

	mockClient := NewMockHTTPClient()
	mockClient.SetShouldFail(true, "network error")

	webhook := NewWebhookWithClient(cfg, mockClient)

	alert := AlertData{
		Timestamp:   time.Now(),
		Aircraft:    piaware.NearbyAircraft{},
		AlertType:   "test",
		Description: "test alert",
	}

	err := webhook.Notify(alert)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "failed to send webhook: network error" {
		t.Errorf("expected error 'failed to send webhook: network error', got %v", err.Error())
	}
}

func TestWebhookNotifyHTTPErrorResponse(t *testing.T) {
	cfg := config.WebhookConfig{
		Enabled: true,
		URL:     "http://localhost:8080/webhook",
		Timeout: 5 * time.Second,
	}

	mockClient := NewMockHTTPClient()
	mockClient.SetResponse(500, []byte("Internal Server Error"))

	webhook := NewWebhookWithClient(cfg, mockClient)

	alert := AlertData{
		Timestamp:   time.Now(),
		Aircraft:    piaware.NearbyAircraft{},
		AlertType:   "test",
		Description: "test alert",
	}

	err := webhook.Notify(alert)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "webhook returned status 500" {
		t.Errorf("expected error 'webhook returned status 500', got %v", err.Error())
	}
}

func TestWebhookNotifyWithRealAircraftData(t *testing.T) {
	cfg := config.WebhookConfig{
		Enabled: true,
		URL:     "http://localhost:8080/webhook",
		Timeout: 5 * time.Second,
	}

	mockClient := NewMockHTTPClient()
	webhook := NewWebhookWithClient(cfg, mockClient)

	aircraft := piaware.NearbyAircraft{
		Aircraft: piaware.Aircraft{
			Hex:     "ABC123",
			Flight:  "TEST1",
			Lat:     40.7128,
			Lon:     -74.0060,
			AltBaro: 5000,
		},
		DistanceKm: 15.2,
	}

	alert := AlertData{
		Timestamp:   time.Now(),
		Aircraft:    aircraft,
		AlertType:   "aircraft_nearby",
		Description: "Aircraft ABC123 detected within 15.2 km at 5000 ft altitude",
	}

	err := webhook.Notify(alert)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check that request was made
	lastRequest := mockClient.GetLastRequest()
	if lastRequest == nil {
		t.Fatal("expected request to be recorded")
	}

	// Verify the JSON payload
	var receivedAlert AlertData
	if err := json.Unmarshal(lastRequest.Body, &receivedAlert); err != nil {
		t.Fatalf("failed to unmarshal request body: %v", err)
	}

	if receivedAlert.Aircraft.Hex != "ABC123" {
		t.Errorf("expected aircraft hex 'ABC123', got %s", receivedAlert.Aircraft.Hex)
	}

	if receivedAlert.Aircraft.Flight != "TEST1" {
		t.Errorf("expected aircraft flight 'TEST1', got %s", receivedAlert.Aircraft.Flight)
	}

	if receivedAlert.AlertType != "aircraft_nearby" {
		t.Errorf("expected alert_type 'aircraft_nearby', got %s", receivedAlert.AlertType)
	}
}

func TestWebhookIntegrationWithTestServer(t *testing.T) {
	// Start test server
	server := NewTestServer(8081)
	if err := server.Start(); err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer func() {
		if stopErr := server.Stop(); stopErr != nil {
			t.Logf("warning: failed to stop test server: %v", stopErr)
		}
	}()

	// Create webhook with real HTTP client
	cfg := config.WebhookConfig{
		Enabled: true,
		URL:     server.GetWebhookURL(),
		Timeout: 5 * time.Second,
	}

	webhook, err := NewWebhook(cfg)
	if err != nil {
		t.Fatalf("failed to create webhook: %v", err)
	}

	// Send test alert
	alert := AlertData{
		Timestamp:   time.Now(),
		Aircraft:    piaware.NearbyAircraft{},
		AlertType:   "integration_test",
		Description: "integration test alert",
	}

	err = webhook.Notify(alert)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Wait for webhook to be received
	if !server.WaitForWebhooks(1, 2*time.Second) {
		t.Fatal("webhook was not received within timeout")
	}

	// Verify webhook was received
	webhooks := server.GetReceivedWebhooks()
	if len(webhooks) != 1 {
		t.Errorf("expected 1 webhook, got %d", len(webhooks))
	}

	webhookData := webhooks[0]
	if webhookData.URL != "/webhook" {
		t.Errorf("expected URL '/webhook', got %s", webhookData.URL)
	}

	// Verify headers
	if webhookData.Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %s", webhookData.Headers["Content-Type"])
	}

	// Verify body
	var receivedAlert AlertData
	if err := json.Unmarshal(webhookData.Body, &receivedAlert); err != nil {
		t.Fatalf("failed to unmarshal webhook body: %v", err)
	}

	if receivedAlert.AlertType != "integration_test" {
		t.Errorf("expected alert_type 'integration_test', got %s", receivedAlert.AlertType)
	}
}

func TestWebhookIntegrationWithServerFailure(t *testing.T) {
	// Start test server
	server := NewTestServer(8082)
	if err := server.Start(); err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer func() {
		if stopErr := server.Stop(); stopErr != nil {
			t.Logf("warning: failed to stop test server: %v", stopErr)
		}
	}()

	// Configure server to return error
	server.SetShouldFail(true, 500)

	// Create webhook with real HTTP client
	cfg := config.WebhookConfig{
		Enabled: true,
		URL:     server.GetWebhookURL(),
		Timeout: 5 * time.Second,
	}

	webhook, err := NewWebhook(cfg)
	if err != nil {
		t.Fatalf("failed to create webhook: %v", err)
	}

	// Send test alert
	alert := AlertData{
		Timestamp:   time.Now(),
		Aircraft:    piaware.NearbyAircraft{},
		AlertType:   "failure_test",
		Description: "failure test alert",
	}

	err = webhook.Notify(alert)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "webhook returned status 500" {
		t.Errorf("expected error 'webhook returned status 500', got %v", err.Error())
	}
}
