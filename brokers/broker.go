package brokers

import (
	"context"
	"errors"

	"github.com/ARGOeu/argo-messaging/messages"
	"time"
)

// Broker  Encapsulates the generic broker interface
type Broker interface {
	InitConfig()
	Initialize(peers []string)
	CloseConnections()
	Publish(ctx context.Context, topic string, payload messages.Message) (string, string, int, int64, error)
	GetMinOffset(ctx context.Context, topic string) int64
	GetMaxOffset(ctx context.Context, topic string) int64
	Consume(ctx context.Context, topic string, offset int64, imm bool, max int64) ([]string, error)
	DeleteTopic(ctx context.Context, topic string) error
	TimeToOffset(ctx context.Context, topic string, time time.Time) (int64, error)
}

var ErrOffsetOff = errors.New("Offset is off")
