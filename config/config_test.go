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
		"certificate_authorities_dir": "/etc/grid-security/certificates",
		"per_resource_auth":"true",
		"service_token":"S3CR3T",
		"push_tls_enabled": "true",
		"push_server_host": "localhost",
		"push_server_port": 5555,
		"verify_push_server": "true",
        "push_worker_token": "pw-token",
		"log_facilities": ["SYSLOG", "CONSOLE"],
        "auth_option": "header"
	}`
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
	suite.Equal("/etc/grid-security/certificates", APIcfg2.CertificateAuthoritiesDir)
	suite.Equal(true, APIcfg.ResAuth)
	suite.Equal("S3CR3T", APIcfg.ServiceToken)
	suite.True(APIcfg2.PushTlsEnabled)
	suite.Equal("localhost", APIcfg2.PushServerHost)
	suite.Equal(5555, APIcfg2.PushServerPort)
	suite.Equal("pw-token", APIcfg2.PushWorkerToken)
	suite.True(APIcfg2.VerifyPushServer)
	suite.Equal(0, len(APIcfg2.LogFacilities))
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
	suite.Equal("/etc/grid-security/certificates", APIcfg.CertificateAuthoritiesDir)
	suite.Equal(true, APIcfg.ResAuth)
	suite.True(APIcfg.PushTlsEnabled)
	suite.Equal("localhost", APIcfg.PushServerHost)
	suite.Equal(5555, APIcfg.PushServerPort)
	suite.True(APIcfg.VerifyPushServer)
	suite.Equal("pw-token", APIcfg.PushWorkerToken)
	suite.Equal([]string{"SYSLOG", "CONSOLE"}, APIcfg.LogFacilities)
	suite.Equal(HeaderKey, int(APIcfg.AuthOption()))
}

func (suite *ConfigTestSuite) TestSetAuthOption() {
	cfg := APICfg{}

	cfg.setAuthOption("bOth")
	suite.Equal(URLKeyAndHeaderKey, int(cfg.authOption))

	cfg.setAuthOption("KEY")
	suite.Equal(UrlKey, int(cfg.authOption))

	cfg.setAuthOption("header")
	suite.Equal(HeaderKey, int(cfg.authOption))

	cfg.authOption = 0
	cfg.setAuthOption("")
	suite.Equal(UrlKey, int(cfg.authOption))
}

func (suite *ConfigTestSuite) TestAuthOption() {

	a1 := AuthOption(UrlKey)
	suite.Equal("key", a1.String())

	a2 := AuthOption(HeaderKey)
	suite.Equal("header", a2.String())

	a3 := AuthOption(URLKeyAndHeaderKey)
	suite.Equal("both", a3.String())
}

func TestConfigTestSuite(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	suite.Run(t, new(ConfigTestSuite))
}
