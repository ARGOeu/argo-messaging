package brokers

import (
	"log"
	"strconv"
	"sync"
	"time"
)

import (
	"github.com/ARGOeu/argo-messaging/messages"
	"github.com/Shopify/sarama"
)

type topicLock struct {
	sync.Mutex
}

// KafkaBroker struct
type KafkaBroker struct {
	sync.Mutex
	createTopicLock topicLock
	consumeLock     map[string]*topicLock
	Config          *sarama.Config
	Producer        sarama.SyncProducer
	Client          sarama.Client
	Consumer        sarama.Consumer
	Servers         []string
}

func (b *KafkaBroker) lockForTopic(topic string) {
	// Check if lock for topic exists
	_, present := b.consumeLock[topic]
	if present == false {
		// TopicLock is not in list so add it
		b.createTopicLock.Lock()
		_, nowPresent := b.consumeLock[topic]
		if nowPresent == false {
			b.consumeLock[topic] = &topicLock{}
			b.consumeLock[topic].Lock()
		}
		b.createTopicLock.Unlock()
	} else {
		b.consumeLock[topic].Lock()
	}
}

func (b *KafkaBroker) unlockForTopic(topic string) {
	// Check if lock for topic exists
	_, present := b.consumeLock[topic]
	if present == false {
		return
	}

	b.consumeLock[topic].Unlock()

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
	b.createTopicLock = topicLock{}
	b.consumeLock = make(map[string]*topicLock)
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
func (b *KafkaBroker) Publish(topic string, msg messages.Message) (string, string, int, int64, error) {

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
		return msg.ID, topic, int(partition), offset, err
	}

	return msg.ID, topic, int(partition), offset, nil

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

	b.lockForTopic(topic)

	defer b.unlockForTopic(topic)
	// Fetch offset

	// consumer, _ := sarama.NewConsumer(b.Servers, b.Config)

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
		log.Println("Partition already consumed aborting try")
		return []string{}
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
			if imm {
				break ConsumerLoop
			}
			if offset+consumed > loff-1 {
				break ConsumerLoop
			}

		}
	}

	return messages
}
