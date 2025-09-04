package notifier

import "sync"

// MockHTTPClient is a mock implementation of HTTPClient for testing.
type MockHTTPClient struct {
	shouldFail   bool
	failMessage  string
	requests     []HTTPRequest
	responseCode int
	responseBody []byte
	mutex        sync.RWMutex
}

// HTTPRequest represents an HTTP request that was made.
type HTTPRequest struct {
	URL         string
	ContentType string
	Body        []byte
}

// NewMockHTTPClient creates a new mock HTTP client.
func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		requests:     make([]HTTPRequest, 0),
		responseCode: 200,
		responseBody: []byte("OK"),
	}
}

// SetShouldFail configures the mock to simulate failures.
func (m *MockHTTPClient) SetShouldFail(shouldFail bool, message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.shouldFail = shouldFail
	m.failMessage = message
}

// SetResponse sets the response that the mock should return.
func (m *MockHTTPClient) SetResponse(statusCode int, body []byte) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.responseCode = statusCode
	m.responseBody = body
}

// Post simulates sending an HTTP POST request.
func (m *MockHTTPClient) Post(url, contentType string, body []byte) (*HTTPResponse, error) {
	m.mutex.RLock()
	shouldFail := m.shouldFail
	failMessage := m.failMessage
	responseCode := m.responseCode
	responseBody := m.responseBody
	m.mutex.RUnlock()

	if shouldFail {
		return nil, &mockError{message: failMessage}
	}

	// Record the request
	req := HTTPRequest{
		URL:         url,
		ContentType: contentType,
		Body:        body,
	}

	m.mutex.Lock()
	m.requests = append(m.requests, req)
	m.mutex.Unlock()

	return &HTTPResponse{
		StatusCode: responseCode,
		Body:       responseBody,
	}, nil
}

// GetRequests returns all requests that were made.
func (m *MockHTTPClient) GetRequests() []HTTPRequest {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	result := make([]HTTPRequest, len(m.requests))
	copy(result, m.requests)
	return result
}

// GetRequestCount returns the number of requests made.
func (m *MockHTTPClient) GetRequestCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.requests)
}

// ClearRequests clears the request history.
func (m *MockHTTPClient) ClearRequests() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.requests = make([]HTTPRequest, 0)
}

// GetLastRequest returns the most recent request, or nil if none.
func (m *MockHTTPClient) GetLastRequest() *HTTPRequest {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if len(m.requests) == 0 {
		return nil
	}
	last := m.requests[len(m.requests)-1]
	return &last
}
