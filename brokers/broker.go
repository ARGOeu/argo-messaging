package brokers

import (
	"context"
	"errors"

	"github.com/ARGOeu/argo-messaging/messages"
)

// Broker  Encapsulates the generic broker interface
type Broker interface {
	InitConfig()
	Initialize(peers []string)
	CloseConnections()
	Publish(topic string, payload messages.Message) (string, string, int, int64, error)
	GetMinOffset(topic string) int64
	GetMaxOffset(topic string) int64
	Consume(ctx context.Context, topic string, offset int64, imm bool, max int64) ([]string, error)
	DeleteTopic(topic string) error
}

var ErrOffsetOff = errors.New("Offset is off")
