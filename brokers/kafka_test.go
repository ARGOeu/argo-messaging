package brokers

import (
	"testing"

	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/Shopify/sarama/mocks"
	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/stretchr/testify/suite"
)

type BrokerTestSuite struct {
	suite.Suite
}

func (suite *BrokerTestSuite) TestPublish() {

	var broker KafkaBroker
	broker.InitConfig()
	pr := mocks.NewSyncProducer(suite.T(), broker.Config)
	pr.ExpectSendMessageAndSucceed()
	broker.Producer = pr
	broker.Publish("test-topic", "test-message")

}

func TestBrokersTestSuite(t *testing.T) {
	suite.Run(t, new(BrokerTestSuite))
}
