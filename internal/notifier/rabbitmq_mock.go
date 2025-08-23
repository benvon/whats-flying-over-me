package notifier

// MockRabbitMQConnection is a mock implementation of RabbitMQConnection for testing.
type MockRabbitMQConnection struct {
	shouldFail    bool
	failMessage   string
	publishedMsgs []PublishedMessage
	isConnected   bool
}

// PublishedMessage represents a message that was published.
type PublishedMessage struct {
	Exchange   string
	RoutingKey string
	Body       []byte
}

// NewMockRabbitMQConnection creates a new mock RabbitMQ connection.
func NewMockRabbitMQConnection() *MockRabbitMQConnection {
	return &MockRabbitMQConnection{
		publishedMsgs: make([]PublishedMessage, 0),
		isConnected:   true,
	}
}

// SetShouldFail configures the mock to simulate failures.
func (m *MockRabbitMQConnection) SetShouldFail(shouldFail bool, message string) {
	m.shouldFail = shouldFail
	m.failMessage = message
}

// SetConnected sets the connection status.
func (m *MockRabbitMQConnection) SetConnected(connected bool) {
	m.isConnected = connected
}

// Publish simulates publishing a message.
func (m *MockRabbitMQConnection) Publish(exchange, routingKey string, body []byte) error {
	if m.shouldFail {
		return &mockError{message: m.failMessage}
	}

	msg := PublishedMessage{
		Exchange:   exchange,
		RoutingKey: routingKey,
		Body:       body,
	}
	m.publishedMsgs = append(m.publishedMsgs, msg)

	return nil
}

// IsConnected returns the connection status.
func (m *MockRabbitMQConnection) IsConnected() bool {
	return m.isConnected
}

// Close simulates closing the connection.
func (m *MockRabbitMQConnection) Close() error {
	m.isConnected = false
	return nil
}

// GetPublishedMessages returns all messages that were published.
func (m *MockRabbitMQConnection) GetPublishedMessages() []PublishedMessage {
	result := make([]PublishedMessage, len(m.publishedMsgs))
	copy(result, m.publishedMsgs)
	return result
}

// GetPublishedMessageCount returns the number of messages published.
func (m *MockRabbitMQConnection) GetPublishedMessageCount() int {
	return len(m.publishedMsgs)
}

// ClearPublishedMessages clears the published message history.
func (m *MockRabbitMQConnection) ClearPublishedMessages() {
	m.publishedMsgs = make([]PublishedMessage, 0)
}
