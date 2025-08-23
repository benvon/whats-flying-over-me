package notifier

// RabbitMQConnection defines the interface for RabbitMQ operations.
type RabbitMQConnection interface {
	Publish(exchange, routingKey string, body []byte) error
	IsConnected() bool
	Close() error
}

// RabbitMQPublisher defines the interface for publishing messages.
type RabbitMQPublisher interface {
	PublishMessage(exchange, routingKey string, body []byte) error
}
