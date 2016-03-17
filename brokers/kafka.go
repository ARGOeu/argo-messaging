package brokers

import "log"

import "github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/Shopify/sarama"

// KafkaBroker struct
type KafkaBroker struct {
	Config   *sarama.Config
	Producer sarama.SyncProducer
	Client   sarama.Client
	Consumer sarama.Consumer
	Servers  []string
}

// CloseConnections closes open producer, consumer and client
func (b *KafkaBroker) CloseConnections() {
	// Close Producer
	if err := b.Producer.Close(); err != nil {
		log.Fatalln(err)
	}
	// Close Consumer
	if err := b.Consumer.Close(); err != nil {
		log.Fatalln(err)
	}
	// Close Client
	if err := b.Client.Close(); err != nil {
		log.Fatalln(err)
	}

}

// NewKafkaBroker creates a new kafka broker object
func NewKafkaBroker(peer string) *KafkaBroker {
	brk := KafkaBroker{}
	brk.Initialize(peer)
	return &brk
}

// InitConfig creates a new configuration for kafka broker
func (b *KafkaBroker) InitConfig() {
	b.Config = sarama.NewConfig()
}

// Initialize the broker struct
func (b *KafkaBroker) Initialize(peer string) {
	b.Config = sarama.NewConfig()
	b.Config.Producer.RequiredAcks = sarama.WaitForAll
	b.Config.Producer.Retry.Max = 5
	b.Servers = []string{peer}

	var err error

	b.Producer, err = sarama.NewSyncProducer(b.Servers, b.Config)
	if err != nil {
		// Should not reach here
		panic(err)
	}

	b.Client, err = sarama.NewClient(b.Servers, nil)
	if err != nil {
		// Should not reach here
		panic(err)
	}

	b.Consumer, err = sarama.NewConsumer(b.Servers, nil)
	if err != nil {
		panic(err)
	}

	log.Printf("%s\t%s\t%s:%s", "INFO", "BROKER", "Kafka Backend Initialized! Kafka server", peer)

}

// Publish function publish a message to the broker
func (b *KafkaBroker) Publish(topic string, payload string) (string, int, int64) {

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(payload),
	}

	partition, offset, err := b.Producer.SendMessage(msg)
	if err != nil {
		panic(err)
	}

	return topic, int(partition), offset
}

// GetOffset returns a current topic's offset
func (b *KafkaBroker) GetOffset(topic string) int64 {
	// Fetch offset
	loff, err := b.Client.GetOffset(topic, 0, sarama.OffsetNewest)
	if err != nil {
		panic(err)
	}
	return loff
}

// Consume function to consume a message from the broker
func (b *KafkaBroker) Consume(topic string, offset int64) []string {

	// Fetch offset
	loff, err := b.Client.GetOffset(topic, 0, sarama.OffsetNewest)
	if err != nil {
		panic(err)
	}

	// If tracked offset is equal or bigger than topic offset means no new messages
	if offset >= loff {
		return []string{}
	}

	partitionConsumer, err := b.Consumer.ConsumePartition(topic, 0, offset)

	if err != nil {
		panic(err)
	}

	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	messages := []string{}
	var consumed int64
ConsumerLoop:
	for {
		select {
		case msg := <-partitionConsumer.Messages():

			messages = append(messages, string(msg.Value[:]))
			consumed++
			if offset+consumed > loff-1 {
				break ConsumerLoop
			}

		}
	}

	return messages
}
