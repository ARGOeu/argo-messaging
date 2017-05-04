package brokers

import (
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
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
	b.Config.Consumer.Fetch.Default = 1000000
	b.Config.Producer.RequiredAcks = sarama.WaitForAll
	b.Config.Producer.Retry.Max = 5
	b.Servers = peers

	var err error

	b.Client, err = sarama.NewClient(b.Servers, nil)
	if err != nil {
		// Should not reach here
		log.Fatal("BROKER", "\t", err.Error())
	}

	b.Producer, err = sarama.NewSyncProducer(b.Servers, b.Config)
	if err != nil {
		// Should not reach here
		log.Fatal("BROKER", "\t", err.Error())

	}

	b.Consumer, err = sarama.NewConsumer(b.Servers, b.Config)

	if err != nil {
		log.Fatal("BROKER", "\t", err.Error())
	}

	log.Info("BROKER", "\t", "Kafka Backend Initialized! Kafka node list", peers)

}

// Publish function publish a message to the broker
func (b *KafkaBroker) Publish(topic string, msg messages.Message) (string, string, int, int64, error) {

	off := b.GetMaxOffset(topic)
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
func (b *KafkaBroker) GetMaxOffset(topic string) int64 {
	// Fetch offset
	loff, err := b.Client.GetOffset(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Error(err.Error())
	}
	return loff
}

// GetOffset returns a current topic's offset
func (b *KafkaBroker) GetMinOffset(topic string) int64 {
	// Fetch offset
	loff, err := b.Client.GetOffset(topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Error(err.Error())
	}
	return loff
}

// Consume function to consume a message from the broker
func (b *KafkaBroker) Consume(topic string, offset int64, imm bool, max int64) ([]string, error) {

	b.lockForTopic(topic)

	defer b.unlockForTopic(topic)
	// Fetch offsets
	loff, err := b.Client.GetOffset(topic, 0, sarama.OffsetNewest)

	if err != nil {
		log.Error(err.Error())
	}

	oldOff, err := b.Client.GetOffset(topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Error(err.Error())
	}

	log.Debug("consuming topic:", topic, " min_offset:", oldOff, " max_offset:", loff, " current offset:", offset)

	// If tracked offset is equal or bigger than topic offset means no new messages
	if offset >= loff {
		return []string{}, nil
	}

	// If tracked offset is left behind increment it to topic's min. offset
	if offset < oldOff {
		log.Debug("Tracked offset is off for topic:", topic, " broker offset:", offset, " tracked offset:", oldOff)
		return []string{}, ErrOffsetOff
	}

	partitionConsumer, err := b.Consumer.ConsumePartition(topic, 0, offset)

	if err != nil {
		log.Debug("Unable to consume")
		log.Debug(err.Error())
		return []string{}, err

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

			log.Debug("consumed:" + string(consumed))
			log.Debug("max:" + string(max))
			log.Debug(msg)
			// if we pass over the available messages and still want more

			if consumed >= max {
				break ConsumerLoop
			}

			if offset+consumed > loff-1 {
				// if returnImmediately is set dont wait for more
				if imm {
					break ConsumerLoop
				}

			}

		}
	}

	return messages, nil
}
