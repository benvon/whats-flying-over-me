package notifier

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// TestServer is a simple HTTP server for testing webhook delivery.
type TestServer struct {
	server     *http.Server
	received   []ReceivedWebhook
	mutex      sync.RWMutex
	port       int
	shouldFail bool
	failCode   int
}

// ReceivedWebhook represents a webhook that was received by the test server.
type ReceivedWebhook struct {
	Timestamp time.Time
	URL       string
	Headers   map[string]string
	Body      []byte
}

// NewTestServer creates a new test server.
func NewTestServer(port int) *TestServer {
	return &TestServer{
		port:     port,
		received: make([]ReceivedWebhook, 0),
	}
}

// Start starts the test server.
func (ts *TestServer) Start() error {
	mux := http.NewServeMux()

	// Handle webhook endpoint
	mux.HandleFunc("/webhook", ts.handleWebhook)

	// Handle health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			fmt.Printf("warning: failed to write health response: %v\n", err)
		}
	})

	ts.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", ts.port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second, // Prevent Slowloris attacks
	}

	go func() {
		if err := ts.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Test server error: %v\n", err)
		}
	}()

	// Wait a moment for server to start
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Stop stops the test server.
func (ts *TestServer) Stop() error {
	if ts.server != nil {
		return ts.server.Close()
	}
	return nil
}

// GetURL returns the base URL of the test server.
func (ts *TestServer) GetURL() string {
	return fmt.Sprintf("http://localhost:%d", ts.port)
}

// GetWebhookURL returns the webhook endpoint URL.
func (ts *TestServer) GetWebhookURL() string {
	return fmt.Sprintf("http://localhost:%d/webhook", ts.port)
}

// SetShouldFail configures the server to return error responses.
func (ts *TestServer) SetShouldFail(shouldFail bool, failCode int) {
	ts.shouldFail = shouldFail
	ts.failCode = failCode
}

// handleWebhook handles incoming webhook requests.
func (ts *TestServer) handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Check if we should simulate failure
	if ts.shouldFail {
		w.WriteHeader(ts.failCode)
		return
	}

	// Read request body
	body := make([]byte, 0)
	if r.Body != nil {
		body, _ = io.ReadAll(r.Body)
		defer func() {
			if closeErr := r.Body.Close(); closeErr != nil {
				fmt.Printf("warning: failed to close request body: %v\n", closeErr)
			}
		}()
	}

	// Record headers
	headers := make(map[string]string)
	for name, values := range r.Header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}

	// Record the webhook
	webhook := ReceivedWebhook{
		Timestamp: time.Now(),
		URL:       r.URL.String(),
		Headers:   headers,
		Body:      body,
	}

	ts.mutex.Lock()
	ts.received = append(ts.received, webhook)
	ts.mutex.Unlock()

	// Return success response
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		fmt.Printf("warning: failed to write webhook response: %v\n", err)
	}
}

// GetReceivedWebhooks returns all webhooks that were received.
func (ts *TestServer) GetReceivedWebhooks() []ReceivedWebhook {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	result := make([]ReceivedWebhook, len(ts.received))
	copy(result, ts.received)
	return result
}

// GetReceivedWebhookCount returns the number of webhooks received.
func (ts *TestServer) GetReceivedWebhookCount() int {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	return len(ts.received)
}

// ClearReceivedWebhooks clears the webhook history.
func (ts *TestServer) ClearReceivedWebhooks() {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	ts.received = make([]ReceivedWebhook, 0)
}

// GetLastWebhook returns the most recent webhook, or nil if none.
func (ts *TestServer) GetLastWebhook() *ReceivedWebhook {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	if len(ts.received) == 0 {
		return nil
	}
	last := ts.received[len(ts.received)-1]
	return &last
}

// WaitForWebhooks waits for a specific number of webhooks to be received.
func (ts *TestServer) WaitForWebhooks(expectedCount int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if ts.GetReceivedWebhookCount() >= expectedCount {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}

	return false
}
