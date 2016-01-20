package brokers

import (
	"testing"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func (suite *ConfigTestSuite) TestPublish() {

	var broker KafkaBroker
	broker.Config = sarama.NewConfig()
	pr := mocks.NewSyncProducer(suite.T(), broker.Config)
	pr.ExpectSendMessageAndSucceed()
	broker.Producer = pr
	broker.Publish("test-topic", "test-message")

}

func TestFactorsTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
