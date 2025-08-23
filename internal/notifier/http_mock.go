package notifier

// MockHTTPClient is a mock implementation of HTTPClient for testing.
type MockHTTPClient struct {
	shouldFail   bool
	failMessage  string
	requests     []HTTPRequest
	responseCode int
	responseBody []byte
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
	m.shouldFail = shouldFail
	m.failMessage = message
}

// SetResponse sets the response that the mock should return.
func (m *MockHTTPClient) SetResponse(statusCode int, body []byte) {
	m.responseCode = statusCode
	m.responseBody = body
}

// Post simulates sending an HTTP POST request.
func (m *MockHTTPClient) Post(url, contentType string, body []byte) (*HTTPResponse, error) {
	if m.shouldFail {
		return nil, &mockError{message: m.failMessage}
	}

	// Record the request
	req := HTTPRequest{
		URL:         url,
		ContentType: contentType,
		Body:        body,
	}
	m.requests = append(m.requests, req)

	return &HTTPResponse{
		StatusCode: m.responseCode,
		Body:       m.responseBody,
	}, nil
}

// GetRequests returns all requests that were made.
func (m *MockHTTPClient) GetRequests() []HTTPRequest {
	result := make([]HTTPRequest, len(m.requests))
	copy(result, m.requests)
	return result
}

// GetRequestCount returns the number of requests made.
func (m *MockHTTPClient) GetRequestCount() int {
	return len(m.requests)
}

// ClearRequests clears the request history.
func (m *MockHTTPClient) ClearRequests() {
	m.requests = make([]HTTPRequest, 0)
}

// GetLastRequest returns the most recent request, or nil if none.
func (m *MockHTTPClient) GetLastRequest() *HTTPRequest {
	if len(m.requests) == 0 {
		return nil
	}
	last := m.requests[len(m.requests)-1]
	return &last
}
