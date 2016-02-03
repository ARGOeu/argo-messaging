package brokers

// MockBroker struct
type MockBroker struct {
	fakeOffset int64
}

// CloseConnections closes open producer, consumer and client
func (b *MockBroker) CloseConnections() {

}

// InitConfig creates a new configuration for kafka broker
func (b *MockBroker) InitConfig() {

}

// Initialize the broker struct
func (b *MockBroker) Initialize(peer string) {
	b.fakeOffset = 0
}

// Publish function publish a message to the broker
func (b *MockBroker) Publish(topic string, payload string) (string, int, int64) {
	b.fakeOffset = b.fakeOffset + 1
	return "mocktopic", 0, b.fakeOffset
}

// GetOffset returns a current topic's offset
func (b *MockBroker) GetOffset(topic string) int64 {
	return b.fakeOffset + 1
}

// Consume function to consume a message from the broker
func (b *MockBroker) Consume(topic string, offset int64) []string {

	return []string{"This is a test mock message"}
}
