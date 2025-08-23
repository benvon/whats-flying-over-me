package notifier

import (
	"sync"
)

// MockNotifier is a mock implementation of Notifier for testing.
type MockNotifier struct {
	notifications []AlertData
	shouldFail    bool
	failMessage   string
	mutex         sync.RWMutex
}

// NewMockNotifier creates a new mock notifier.
func NewMockNotifier() *MockNotifier {
	return &MockNotifier{
		notifications: make([]AlertData, 0),
	}
}

// SetShouldFail configures the mock to simulate failures.
func (m *MockNotifier) SetShouldFail(shouldFail bool, message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.shouldFail = shouldFail
	m.failMessage = message
}

// Notify records the notification and returns an error if configured to fail.
func (m *MockNotifier) Notify(alert AlertData) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.notifications = append(m.notifications, alert)

	if m.shouldFail {
		return &mockError{message: m.failMessage}
	}
	return nil
}

// GetNotifications returns all notifications sent to this mock.
func (m *MockNotifier) GetNotifications() []AlertData {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]AlertData, len(m.notifications))
	copy(result, m.notifications)
	return result
}

// GetNotificationCount returns the number of notifications sent.
func (m *MockNotifier) GetNotificationCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.notifications)
}

// ClearNotifications clears the notification history.
func (m *MockNotifier) ClearNotifications() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.notifications = make([]AlertData, 0)
}

// mockError implements error interface for testing.
type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}
