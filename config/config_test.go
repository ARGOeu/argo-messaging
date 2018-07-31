package config

import (
	"io/ioutil"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *ConfigTestSuite) SetupTest() {
	suite.cfgStr = `{
	  "bind_ip":"",
	  "port":8080,
		"zookeeper_hosts":["localhost"],
		"kafka_znode":"/argo-messaging",
		"store_host":"localhost",
		"store_db":"argo_msg",
		"certificate":"/etc/pki/tls/certs/localhost.crt",
		"certificate_key":"/etc/pki/tls/private/localhost.key",
		"per_resource_auth":"true",
		"service_token":"S3CR3T"
	}`

	log.SetOutput(ioutil.Discard)
}

func (suite *ConfigTestSuite) TestLoadConfiguration() {
	APIcfg := NewAPICfg()
	suite.Equal([]string(nil), APIcfg.ZooHosts)
	APIcfg.LoadTest()
	suite.Equal([]string{"localhost"}, APIcfg.ZooHosts)
	suite.Equal("", APIcfg.KafkaZnode)
	suite.Equal("localhost", APIcfg.StoreHost)
	suite.Equal("argo_msg", APIcfg.StoreDB)
	suite.Equal(8080, APIcfg.Port)

	// test "LOADTEST" param
	APIcfg2 := NewAPICfg("LOADTEST")
	suite.Equal([]string{"localhost"}, APIcfg2.ZooHosts)
	suite.Equal("", APIcfg2.KafkaZnode)
	suite.Equal("localhost", APIcfg2.StoreHost)
	suite.Equal("argo_msg", APIcfg2.StoreDB)
	suite.Equal(8080, APIcfg.Port)
	suite.Equal("", APIcfg.BindIP)
	suite.Equal("/etc/pki/tls/certs/localhost.crt", APIcfg.Cert)
	suite.Equal("/etc/pki/tls/private/localhost.key", APIcfg.CertKey)
	suite.Equal(true, APIcfg.ResAuth)
	suite.Equal("S3CR3T", APIcfg.ServiceToken)
}

func (suite *ConfigTestSuite) TestLoadStringJSON() {
	APIcfg := NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	suite.Equal([]string{"localhost"}, APIcfg.ZooHosts)
	suite.Equal("/argo-messaging", APIcfg.KafkaZnode)
	suite.Equal("localhost", APIcfg.StoreHost)
	suite.Equal("argo_msg", APIcfg.StoreDB)
	suite.Equal(8080, APIcfg.Port)
	suite.Equal("", APIcfg.BindIP)
	suite.Equal("/etc/pki/tls/certs/localhost.crt", APIcfg.Cert)
	suite.Equal("/etc/pki/tls/private/localhost.key", APIcfg.CertKey)
	suite.Equal(true, APIcfg.ResAuth)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
