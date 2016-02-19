package config

import (
	"io/ioutil"
	"log"
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
	  "broker_host":"localhost:9092",
		"store_host":"localhost",
		"store_db":"argo_msg"
	}
	`

	log.SetOutput(ioutil.Discard)
}

func (suite *ConfigTestSuite) TestLoadConfiguration() {
	APIcfg := NewAPICfg()
	suite.Equal("", APIcfg.BrokerHost)
	APIcfg.Load()
	suite.Equal("localhost:9092", APIcfg.BrokerHost)
	suite.Equal("localhost", APIcfg.StoreHost)
	suite.Equal("argo_msg", APIcfg.StoreDB)

	// test "LOAD" param
	APIcfg2 := NewAPICfg("LOAD")
	suite.Equal("localhost:9092", APIcfg2.BrokerHost)
	suite.Equal("localhost", APIcfg2.StoreHost)
	suite.Equal("argo_msg", APIcfg2.StoreDB)

}

func (suite *ConfigTestSuite) TestLoadStringJSON() {
	APIcfg := NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	suite.Equal("localhost:9092", APIcfg.BrokerHost)
	suite.Equal("localhost", APIcfg.StoreHost)
	suite.Equal("argo_msg", APIcfg.StoreDB)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
