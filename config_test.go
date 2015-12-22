package main

import (
	"testing"

	"github.com/ARGOeu/argo-messaging/config"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func (suite *ConfigTestSuite) TestLoadConfiguration() {
	config.Load()
	suite.Equal("localhost:9092", config.Kafka.Server)
	suite.Equal("topic1", config.Kafka.Topics[0])
	suite.Equal("topic2", config.Kafka.Topics[1])
}

func TestFactorsTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
