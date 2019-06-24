package brokers

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/ARGOeu/argo-messaging/messages"
	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
)

type topicLock struct {
	sync.Mutex
}

type TopicOffset struct {
	Offset int64 `json:"offset"`
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
	ClusterAdmin    sarama.ClusterAdmin
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
	// Close Cluster Admin
	if err := b.ClusterAdmin.Close(); err != nil {
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

// Initialize method is a retry wrapper for init (which attempts to connect to broker backend)
func (b *KafkaBroker) Initialize(peers []string) {
	for {
		// Try to initialize broker backend
		log.Info("BROKER", "\t", "Attempting to connect to kafka backend: ", peers)
		err := b.init(peers)
		// if err happened log it and retry in 3sec - else all is ok so return
		if err != nil {
			log.Error("BROKER", "\t", err.Error())
			time.Sleep(3 * time.Second)
		} else {
			log.Info("BROKER", "\t", "Kafka Backend Initialized! Kafka node list", peers)
			return
		}
	}
}

// init attempts to connect to broker backend and initialize local broker-related structures
func (b *KafkaBroker) init(peers []string) error {
	b.createTopicLock = topicLock{}
	b.consumeLock = make(map[string]*topicLock)
	b.Config = sarama.NewConfig()
	b.Config.Consumer.Fetch.Default = 1000000
	b.Config.Producer.RequiredAcks = sarama.WaitForAll
	b.Config.Producer.Retry.Max = 5
	b.Config.Producer.Return.Successes = true
	b.Config.Version = sarama.V2_1_0_0
	b.Servers = peers

	var err error

	b.Client, err = sarama.NewClient(b.Servers, b.Config)
	if err != nil {
		return err
	}

	b.Producer, err = sarama.NewSyncProducer(b.Servers, b.Config)
	if err != nil {
		return err
	}

	b.Consumer, err = sarama.NewConsumer(b.Servers, b.Config)
	if err != nil {
		return err
	}

	b.ClusterAdmin, err = sarama.NewClusterAdmin(b.Servers, b.Config)
	if err != nil {
		return err
	}

	return nil
}

// Publish function publish a message to the broker
func (b *KafkaBroker) Publish(topic string, msg messages.Message) (string, string, int, int64, error) {

	off := b.GetMaxOffset(topic)
	msg.ID = strconv.FormatInt(off, 10)
	// Stamp time to UTC Z to nanoseconds
	zNano := "2006-01-02T15:04:05.999999999Z"
	// Timestamp on publish time -- should be in UTC
	t := time.Now().UTC()
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

// TimeToOffset returns the offset of the first message with a timestamp equal or
// greater than the time given.
func (b *KafkaBroker) TimeToOffset(topic string, t time.Time) (int64, error) {
	return b.Client.GetOffset(topic, 0, t.UnixNano()/int64(time.Millisecond))
}

// DeleteTopic deletes the topic from the Kafka cluster
func (b *KafkaBroker) DeleteTopic(topic string) error {

	b.lockForTopic(topic)

	defer b.unlockForTopic(topic)

	return b.ClusterAdmin.DeleteTopic(topic)
}

// Consume function to consume a message from the broker
func (b *KafkaBroker) Consume(ctx context.Context, topic string, offset int64, imm bool, max int64) ([]string, error) {

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
		// If the http client cancels the http request break consume loop
		case <-ctx.Done():
			{
				break ConsumerLoop
			}
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
