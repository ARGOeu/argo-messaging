package brokers

// MockBroker struct
type MockBroker struct {
	MsgList []string
}

// PopulateOne Adds three messages to the mock broker
func (b *MockBroker) PopulateOne() {
	msg1 := `{
  "messageId": "0",
  "attributes": [
    {
      "key": "foo",
      "value": "bar"
    }
  ],
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
  "attributes": [
    {
      "key": "foo",
      "value": "bar"
    }
  ],
  "data": "YmFzZTY0ZW5jb2RlZA==",
  "publishTime": "2016-02-24T11:55:09.786127994Z"
}`

	msg2 := `{
  "messageId": "1",
  "attributes": [
    {
      "key": "foo2",
      "value": "bar2"
    }
  ],
  "data": "YmFzZTY0ZW5jb2RlZA==",
  "publishTime": "2016-02-24T11:55:09.827678754Z"
}`

	msg3 := `{
  "messageId": "2",
  "attributes": [
    {
      "key": "foo2",
      "value": "bar2"
    }
  ],
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
func (b *MockBroker) Initialize(peer string) {
	b.MsgList = make([]string, 0)
}

// Publish function publish a message to the broker
func (b *MockBroker) Publish(topic string, payload string) (string, int, int64) {
	b.MsgList = append(b.MsgList, payload)
	return "ARGO.topic1", 0, int64(len(b.MsgList))
}

// GetOffset returns a current topic's offset
func (b *MockBroker) GetOffset(topic string) int64 {
	return int64(len(b.MsgList) + 1)
}

// Consume function to consume a message from the broker
func (b *MockBroker) Consume(topic string, offset int64) []string {
	return b.MsgList
}
