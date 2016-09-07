package brokers

import "github.com/ARGOeu/argo-messaging/messages"

// Broker  Encapsulates the generic broker interface
type Broker interface {
	InitConfig()
	Initialize(peers []string)
	CloseConnections()
	Publish(topic string, payload messages.Message) (string, string, int, int64, error)
	GetOffset(topic string) int64
	Consume(topic string, offset int64, imm bool) []string
}
