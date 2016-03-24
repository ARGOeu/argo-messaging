package brokers

// Broker  Encapsulates the generic broker interface
type Broker interface {
	InitConfig()
	Initialize(peer string)
	CloseConnections()
	Publish(topic string, payload string) (string, int, int64)
	GetOffset(topic string) int64
	Consume(topic string, offset int64) []string
}
