package brokers

import (
	"log"
	"strconv"
	"sync"
	"time"
)

import (
	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/Shopify/sarama"
	"github.com/ARGOeu/argo-messaging/messages"
)

type kafkaLock struct {
	sync.Mutex
}

// KafkaBroker struct
type KafkaBroker struct {
	produceLock kafkaLock
	consumeLock kafkaLock
	Config      *sarama.Config
	Producer    sarama.SyncProducer
	Client      sarama.Client
	Consumer    sarama.Consumer
	Servers     []string
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
func NewKafkaBroker(peers []string) *KafkaBroker {
	brk := KafkaBroker{}
	brk.Initialize(peers)
	return &brk
}

// InitConfig creates a new configuration for kafka broker
func (b *KafkaBroker) InitConfig() {
	b.Config = sarama.NewConfig()
}

// Initialize the broker struct
func (b *KafkaBroker) Initialize(peers []string) {
	b.consumeLock = kafkaLock{}
	b.produceLock = kafkaLock{}

	b.Config = sarama.NewConfig()
	b.Config.Producer.RequiredAcks = sarama.WaitForAll
	b.Config.Producer.Retry.Max = 5
	b.Servers = peers

	var err error

	b.Client, err = sarama.NewClient(b.Servers, nil)
	if err != nil {
		// Should not reach here
		log.Fatalf("%s\t%s\t%s", "FATAL", "BROKER", err.Error())
	}

	b.Producer, err = sarama.NewSyncProducer(b.Servers, b.Config)
	if err != nil {
		// Should not reach here
		log.Fatalf("%s\t%s\t%s", "FATAL", "BROKER", err.Error())

	}

	b.Consumer, err = sarama.NewConsumer(b.Servers, nil)
	if err != nil {
		log.Fatalf("%s\t%s\t%s", "FATAL", "BROKER", err.Error())
	}

	log.Printf("%s\t%s\t%s:%s", "INFO", "BROKER", "Kafka Backend Initialized! Kafka node list", peers)

}

// Publish function publish a message to the broker
func (b *KafkaBroker) Publish(topic string, msg messages.Message) (string, string, int, int64) {
	b.produceLock.Lock()
	// Unlock after consumption
	defer b.produceLock.Unlock()
	off := b.GetOffset(topic)
	msg.ID = strconv.FormatInt(off, 10)
	// Stamp time to UTC Z to nanoseconds
	zNano := "2006-01-02T15:04:05.999999999Z"
	t := time.Now()
	msg.PubTime = t.Format(zNano)

	// Publish the message
	payload, _ := msg.ExportJSON()

	msgFinal := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(payload),
	}

	partition, offset, err := b.Producer.SendMessage(msgFinal)
	if err != nil {
		panic(err)
	}

	return msg.ID, topic, int(partition), offset

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
func (b *KafkaBroker) Consume(topic string, offset int64, imm bool) []string {
	b.consumeLock.Lock()
	// Unlock after consumption
	defer b.consumeLock.Unlock()
	// Fetch offset
	loff, err := b.Client.GetOffset(topic, 0, sarama.OffsetNewest)
	log.Println("consuming topic:", topic, "with offset:", loff)
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
	timeout := time.After(300 * time.Second)

	if imm {
		timeout = time.After(100 * time.Millisecond)
	}

ConsumerLoop:
	for {
		select {
		case <-timeout:
			{
				break ConsumerLoop
			}
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
