package handlers

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	oldPush "github.com/ARGOeu/argo-messaging/push"
	push "github.com/ARGOeu/argo-messaging/push/grpc/client"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.cfgStr = `{
	"bind_ip":"",
	"port":8080,
	"zookeeper_hosts":["localhost"],
	"kafka_znode":"",
	"store_host":"localhost",
	"store_db":"argo_msg",
	"certificate":"/etc/pki/tls/certs/localhost.crt",
	"certificate_key":"/etc/pki/tls/private/localhost.key",
	"per_resource_auth":"true",
	"push_enabled": "true",
	"push_worker_token": "push_token"
	}`
}

func (suite *HandlerTestSuite) TestHealthCheck() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/status", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
 "status": "ok",
 "push_servers": [
  {
   "endpoint": "localhost:5555",
   "status": "SERVING"
  }
 ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushEnabled = true
	cfgKafka.PushWorkerToken = "push_token"
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/status", WrapMockAuthConfig(HealthCheck, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)

	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestHealthCheckDetails() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/status?details=true&key=admin-viewer-token", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
 "status": "ok",
 "push_servers": [
  {
   "endpoint": "localhost:5555",
   "status": "SERVING"
  }
 ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushEnabled = true
	cfgKafka.PushWorkerToken = "push_token"
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	str.UserList = append(str.UserList, stores.QUser{
		UUID:         "admin-viewer-id",
		Name:         "admin-viewer",
		Token:        "admin-viewer-token",
		ServiceRoles: []string{"admin_viewer"},
	})

	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/status", WrapMockAuthConfig(HealthCheck, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)

	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestHealthCheckPushDisabled() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/status", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
 "status": "ok",
 "push_functionality": "disabled"
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushEnabled = false
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/status", WrapMockAuthConfig(HealthCheck, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestHealthCheckPushWorkerMissing() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/status", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
 "status": "warning",
 "push_servers": [
  {
   "endpoint": "localhost:5555",
   "status": "SERVING"
  }
 ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushEnabled = true
	// add a wrong push worker token
	cfgKafka.PushWorkerToken = "missing"
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/status", WrapMockAuthConfig(HealthCheck, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func TestHandlersTestSuite(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	suite.Run(t, new(HandlerTestSuite))
}
