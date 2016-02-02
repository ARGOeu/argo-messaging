package config

import (
	"testing"

	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *ConfigTestSuite) SetupTest() {
	suite.cfgStr = `
	{
	  "server":"localhost:9092",
	  "topics":["topic1","topic2"]
	}
	`
}

func (suite *ConfigTestSuite) TestLoadConfiguration() {
	kafkaCfg := NewKafkaCfg()
	kafkaCfg.Load()
	suite.Equal("localhost:9092", kafkaCfg.Server)
	suite.Equal("topic1", kafkaCfg.Topics[0])
	suite.Equal("topic2", kafkaCfg.Topics[1])
}

func (suite *ConfigTestSuite) TestLoadStringJSON() {
	kafkaCfg := NewKafkaCfg()
	kafkaCfg.LoadStrJSON(suite.cfgStr)
	suite.Equal("localhost:9092", kafkaCfg.Server)
	suite.Equal("topic1", kafkaCfg.Topics[0])
	suite.Equal("topic2", kafkaCfg.Topics[1])
}

func TestFactorsTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
