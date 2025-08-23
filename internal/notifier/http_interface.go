package notifier

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient defines the interface for HTTP operations.
type HTTPClient interface {
	Post(url, contentType string, body []byte) (*HTTPResponse, error)
}

// HTTPResponse represents an HTTP response.
type HTTPResponse struct {
	StatusCode int
	Body       []byte
}

// RealHTTPClient implements the actual HTTP client.
type RealHTTPClient struct {
	timeout time.Duration
}

// NewRealHTTPClient creates a new real HTTP client.
func NewRealHTTPClient(timeout time.Duration) *RealHTTPClient {
	return &RealHTTPClient{timeout: timeout}
}

// Post sends a POST request.
func (c *RealHTTPClient) Post(url, contentType string, body []byte) (*HTTPResponse, error) {
	client := &http.Client{Timeout: c.timeout}

	resp, err := client.Post(url, contentType, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the error but don't fail the request
			fmt.Printf("warning: failed to close response body: %v\n", closeErr)
		}
	}()

	// Read response body
	respBody := make([]byte, 0)
	if resp.Body != nil {
		respBody, _ = io.ReadAll(resp.Body)
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
	}, nil
}
