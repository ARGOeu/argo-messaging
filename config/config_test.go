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
	suite.cfgStr = `{
	  "bind_ip":"",
	  "port":8080,
		"broker_hosts":["localhost:9092"],
		"store_host":"localhost",
		"store_db":"argo_msg",
		"use_authorization":true,
		"use_authentication":true,
		"use_ack":true,
		"certificate":"/etc/pki/tls/certs/localhost.crt",
		"certificate_key":"/etc/pki/tls/private/localhost.key"
	}`

	log.SetOutput(ioutil.Discard)
}

func (suite *ConfigTestSuite) TestLoadConfiguration() {
	APIcfg := NewAPICfg()
	suite.Equal([]string(nil), APIcfg.BrokerHosts)
	APIcfg.Load()
	suite.Equal([]string{"localhost:9092"}, APIcfg.BrokerHosts)
	suite.Equal("localhost", APIcfg.StoreHost)
	suite.Equal("argo_msg", APIcfg.StoreDB)
	suite.Equal(true, APIcfg.Authen)
	suite.Equal(true, APIcfg.Author)
	suite.Equal(8080, APIcfg.Port)
	suite.Equal(true, APIcfg.Ack)

	// test "LOAD" param
	APIcfg2 := NewAPICfg("LOAD")
	suite.Equal([]string{"localhost:9092"}, APIcfg2.BrokerHosts)
	suite.Equal("localhost", APIcfg2.StoreHost)
	suite.Equal("argo_msg", APIcfg2.StoreDB)
	suite.Equal(true, APIcfg2.Authen)
	suite.Equal(true, APIcfg2.Author)
	suite.Equal(8080, APIcfg.Port)
	suite.Equal(true, APIcfg.Ack)
	suite.Equal("", APIcfg.BindIP)
	suite.Equal("/etc/pki/tls/certs/localhost.crt", APIcfg.Cert)
	suite.Equal("/etc/pki/tls/private/localhost.key", APIcfg.CertKey)
}

func (suite *ConfigTestSuite) TestLoadStringJSON() {
	APIcfg := NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	suite.Equal([]string{"localhost:9092"}, APIcfg.BrokerHosts)
	suite.Equal("localhost", APIcfg.StoreHost)
	suite.Equal("argo_msg", APIcfg.StoreDB)
	suite.Equal(true, APIcfg.Authen)
	suite.Equal(true, APIcfg.Author)
	suite.Equal(8080, APIcfg.Port)
	suite.Equal(true, APIcfg.Ack)
	suite.Equal("", APIcfg.BindIP)
	suite.Equal("/etc/pki/tls/certs/localhost.crt", APIcfg.Cert)
	suite.Equal("/etc/pki/tls/private/localhost.key", APIcfg.CertKey)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
