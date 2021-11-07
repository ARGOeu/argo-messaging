package handlers

import (
	"fmt"
	"github.com/ARGOeu/argo-messaging/version"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

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
	"push_worker_token": "push_token",
	"auth_option": "both"
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

func (suite *HandlerTestSuite) TestGetRequestTokenExtractStrategy() {

	// test the key extract strategy
	keyStrategy := GetRequestTokenExtractStrategy(config.UrlKey)
	u1, _ := url.Parse("https://host.com/v1/projects?key=tok3n")
	r1 := &http.Request{
		URL: u1,
	}
	suite.Equal("tok3n", keyStrategy(r1))

	// test the header extract strategy
	h1 := http.Header{}
	h1.Add("x-api-key", "tok3n")
	u2, _ := url.Parse("https://host.com/v1/projects")
	r2 := &http.Request{
		URL:    u2,
		Header: h1,
	}
	headerStrategy := GetRequestTokenExtractStrategy(config.HeaderKey)
	suite.Equal("tok3n", headerStrategy(r2))

	// test the key and header strategy when there is a x-api-key header present
	h2 := http.Header{}
	h2.Add("x-api-key", "tok3n-h")
	u3, _ := url.Parse("https://host.com/v1/projects?key=tok3n-url")
	r3 := &http.Request{
		URL:    u3,
		Header: h2,
	}
	bothStrategy := GetRequestTokenExtractStrategy(config.URLKeyAndHeaderKey)
	suite.Equal("tok3n-h", bothStrategy(r3))

	// test the key and header strategy when there is no a x-api-key header present but there is a key url value
	r3.Header = http.Header{}
	bothStrategy2 := GetRequestTokenExtractStrategy(0)
	suite.Equal("tok3n-url", bothStrategy2(r3))
}

func (suite *HandlerTestSuite) TestListVersion() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/version", nil)
	req.Header.Add("x-api-key", "st")
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
 "build_time": "%v",
 "golang": "%v",
 "compiler": "%v",
 "os": "%v",
 "architecture": "%v",
 "release": "%v"
}`
	expResp = fmt.Sprintf(expResp, version.BuildTime, version.GO, version.Compiler, version.OS, version.Arch, version.Release)

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushEnabled = true
	// add a wrong push worker token
	cfgKafka.PushWorkerToken = "missing"
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	str.UserList = append(str.UserList, stores.QUser{8, "uuid8", nil, "UserZ", "", "", "", "", "st", "foo-email", []string{"service_admin"}, time.Now(), time.Now(), ""})

	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/version", WrapMockAuthConfig(ListVersion, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func TestHandlersTestSuite(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	suite.Run(t, new(HandlerTestSuite))
}
