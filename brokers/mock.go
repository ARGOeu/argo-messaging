package brokers

import (
	"context"
	"strconv"

	"errors"
	"fmt"
	"github.com/ARGOeu/argo-messaging/messages"
	"strings"
	"time"
)

// MockBroker struct
type MockBroker struct {
	MsgList          []string
	Topics           map[string]string
	TopicTimeIndices map[string][]TimeToOffset
}

type TimeToOffset struct {
	Timestamp time.Time
	Offset    int64
}

// PopulateOne Adds three messages to the mock broker
func (b *MockBroker) PopulateOne() {
	msg1 := `{
  "messageId": "0",
  "attributes":
    {
      "foo":"bar"
    },
  "data": "YmFzZTY0ZW5jb2RlZA==",
  "publishTime": "2016-02-24T11:55:09.786127994Z"
}`

	b.MsgList = make([]string, 0)
	b.MsgList = append(b.MsgList, msg1)

}

// PopulateThree Adds three messages to the mock broker
func (b *MockBroker) PopulateThree() {
	msg1 := `{
  "messageId": "0",
  "attributes":
    {
      "foo":"bar"
    },
  "data": "YmFzZTY0ZW5jb2RlZA==",
  "publishTime": "2016-02-24T11:55:09.786127994Z"
}`

	msg2 := `{
  "messageId": "1",
  "attributes":
    {
    	"foo2":"bar2"
    },
  "data": "YmFzZTY0ZW5jb2RlZA==",
  "publishTime": "2016-02-24T11:55:09.827678754Z"
}`

	msg3 := `{
  "messageId": "2",
  "attributes":
    {
      "foo2":"bar2"
    },
  "data": "YmFzZTY0ZW5jb2RlZA==",
  "publishTime": "2016-02-24T11:55:09.830417467Z"
}`
	b.MsgList = make([]string, 0)
	b.MsgList = append(b.MsgList, msg1)
	b.MsgList = append(b.MsgList, msg2)
	b.MsgList = append(b.MsgList, msg3)
}

// CloseConnections closes open producer, consumer and client
func (b *MockBroker) CloseConnections() {

}

// InitConfig creates a new configuration for kafka broker
func (b *MockBroker) InitConfig() {

}

// Initialize the broker struct
func (b *MockBroker) Initialize(peers []string) {
	b.MsgList = make([]string, 0)
}

// Publish function publish a message to the broker
func (b *MockBroker) Publish(ctx context.Context, topic string, msg messages.Message) (string, string, int, int64, error) {
	payload, _ := msg.ExportJSON()
	b.MsgList = append(b.MsgList, payload)
	off := b.GetMaxOffset(ctx, topic) - 1
	msgID := strconv.FormatInt(off, 10)
	// split the name that SHOULD come in the form of project_uuid.topic_name
	s := strings.Split(topic, ".")
	return msgID, fmt.Sprintf("%s.%s", s[0], s[1]), 0, int64(len(b.MsgList)), nil
}

// GetOffset returns a current topic's offset
func (b *MockBroker) GetMaxOffset(ctx context.Context, topic string) int64 {
	return int64(len(b.MsgList) + 1)
}

// GetOffset returns a current topic's offset
func (b *MockBroker) GetMinOffset(ctx context.Context, topic string) int64 {
	return int64(len(b.MsgList))
}

// Consume function to consume a message from the broker
func (b *MockBroker) Consume(ctx context.Context, topic string, offset int64, imm bool, max int64) ([]string, error) {
	return b.MsgList, nil
}

// Delete topic from the broker
func (b *MockBroker) DeleteTopic(ctx context.Context, topic string) error {

	_, ok := b.Topics[topic]
	if !ok {
		return errors.New("topic not found on the broker")
	}

	delete(b.Topics, topic)
	return nil
}

func (b *MockBroker) TimeToOffset(ctx context.Context, topic string, time time.Time) (int64, error) {

	topicTimeIndices, ok := b.TopicTimeIndices[topic]

	if !ok {
		return -1, errors.New("topic not found on the broker")
	}

	for _, item := range topicTimeIndices {
		if item.Timestamp.Equal(time) || item.Timestamp.After(time) {
			return item.Offset, nil
		}
	}

	return -1, nil
}
