package handlers

import (
	"bytes"
	"fmt"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	oldPush "github.com/ARGOeu/argo-messaging/push"
	push "github.com/ARGOeu/argo-messaging/push/grpc/client"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/ARGOeu/argo-messaging/subscriptions"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type SubscriptionsHandlersTestSuite struct {
	cfgStr string
	suite.Suite
}

func (suite *SubscriptionsHandlersTestSuite) SetupTest() {
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

func (suite *SubscriptionsHandlersTestSuite) TestSubModPushConfigError() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1",
	"pushConfig": {
		 "pushEndpoint": "http://www.example.com",
		 "retryPolicy": {}
	}
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1:modifyPushConfig", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 400,
      "message": "Push endpoint should be addressed by a valid https url",
      "status": "INVALID_ARGUMENT"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modifyPushConfig", WrapMockAuthConfig(SubModPush, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubModPushInvalidRetPol() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1",
	"pushConfig": {
		 "pushEndpoint": "https://www.example.com",
		 "retryPolicy": {
			"type": "unknown"
		 }
	}
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1:modifyPushConfig", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 400,
      "message": "Retry policy can only be of 'linear' or 'slowstart' type",
      "status": "INVALID_ARGUMENT"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modifyPushConfig", WrapMockAuthConfig(SubModPush, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expResp, w.Body.String())

}

// TestSubModPushConfigToActive tests the case where the user modifies the push configuration,
// in order to activate the subscription on the push server
// the push configuration was empty before the api call
func (suite *SubscriptionsHandlersTestSuite) TestSubModPushConfigToActive() {

	postJSON := `{
	"pushConfig": {
		 "pushEndpoint": "https://www.example.com",
		 "retryPolicy": {}
	}
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1:modifyPushConfig", strings.NewReader(postJSON))
	if err != nil {
		log.Fatal(err)
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modifyPushConfig", WrapMockAuthConfig(SubModPush, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	sub, _ := str.QueryOneSub("argo_uuid", "sub1")
	suite.Equal(200, w.Code)
	suite.Equal("", w.Body.String())
	suite.Equal("https://www.example.com", sub.PushEndpoint)
	suite.Equal(int64(1), sub.MaxMessages)
	suite.Equal(3000, sub.RetPeriod)
	suite.Equal("linear", sub.RetPolicy)
	suite.False(sub.Verified)
	suite.NotEqual("", sub.VerificationHash)
	suite.Equal(subscriptions.AutoGenerationAuthorizationHeader, sub.AuthorizationType)
	suite.NotEqual("", sub.AuthorizationHeader)
}

// TestSubModPushConfigToInactive tests the use case where the user modifies the push configuration
// in order to deactivate the subscription on the push server
// the push configuration has values before the call and turns into an empty one by the end of the call
func (suite *SubscriptionsHandlersTestSuite) TestSubModPushConfigToInactive() {

	postJSON := `{
	"pushConfig": {}
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub4:modifyPushConfig", strings.NewReader(postJSON))
	if err != nil {
		log.Fatal(err)
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modifyPushConfig", WrapMockAuthConfig(SubModPush, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	sub, _ := str.QueryOneSub("argo_uuid", "sub4")
	suite.Equal(200, w.Code)
	suite.Equal("", w.Body.String())
	suite.Equal("", sub.PushEndpoint)
	suite.Equal(0, sub.RetPeriod)
	suite.Equal("", sub.RetPolicy)
	suite.Equal("", sub.VerificationHash)
	suite.False(sub.Verified)
	// check to see that the push worker user has been removed from the subscription's acl
	a1, _ := str.QueryACL("argo_uuid", "subscriptions", "sub4")
	suite.Equal([]string{"uuid2", "uuid4"}, a1.ACL)
}

// TestSubModPushConfigToInactivePushDisabled tests the use case where the user modifies the push configuration
// in order to deactivate the subscription on the push server
// the push configuration has values before the call and turns into an empty one by the end of the call
// push enabled is false, but turning a subscription from push to pull should always be available as an api action
func (suite *SubscriptionsHandlersTestSuite) TestSubModPushConfigToInactivePushDisabled() {

	postJSON := `{
	"pushConfig": {}
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub4:modifyPushConfig", strings.NewReader(postJSON))
	if err != nil {
		log.Fatal(err)
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushEnabled = false
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modifyPushConfig", WrapMockAuthConfig(SubModPush, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	sub, _ := str.QueryOneSub("argo_uuid", "sub4")
	suite.Equal(200, w.Code)
	suite.Equal("", w.Body.String())
	suite.Equal("", sub.PushEndpoint)
	suite.Equal(0, sub.RetPeriod)
	suite.Equal("", sub.RetPolicy)
}

// TestSubModPushConfigToInactiveMissingPushWorker tests the use case where the user modifies the push configuration
// in order to deactivate the subscription on the push server
// the push configuration has values before the call and turns into an empty one by the end of the call
// push enabled is true, we cannot retrieve the push worker user in order to remove him from the subscription's acl
// but turning a subscription from push to pull should always be available as an api action
func (suite *SubscriptionsHandlersTestSuite) TestSubModPushConfigToInactiveMissingPushWorker() {

	postJSON := `{
	"pushConfig": {}
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub4:modifyPushConfig", strings.NewReader(postJSON))
	if err != nil {
		log.Fatal(err)
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushWorkerToken = "missing"
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modifyPushConfig", WrapMockAuthConfig(SubModPush, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	sub, _ := str.QueryOneSub("argo_uuid", "sub4")
	suite.Equal(200, w.Code)
	suite.Equal("", w.Body.String())
	suite.Equal("", sub.PushEndpoint)
	suite.Equal(0, sub.RetPeriod)
	suite.Equal("", sub.RetPolicy)
}

// TestSubModPushConfigToActive tests the case where the user modifies the push configuration,
// in order to activate the subscription on the push server
// the push configuration was empty before the api call
// since the push endpoint that has been registered is different from the previous verified one
// the sub will be deactivated on the push server and turn into unverified
func (suite *SubscriptionsHandlersTestSuite) TestSubModPushConfigUpdate() {

	postJSON := `{
	"pushConfig": {
		 "pushEndpoint": "https://www.example2.com",
         "maxMessages": 5,
		 "authorizationHeader": {
			"type": "autogen"
         },
		 "retryPolicy": {
             "type":"linear",
             "period": 5000
         }
	}
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub4:modifyPushConfig", strings.NewReader(postJSON))
	if err != nil {
		log.Fatal(err)
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	subBeforeUpdate, _ := str.QueryOneSub("argo_uuid", "sub4")
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modifyPushConfig", WrapMockAuthConfig(SubModPush, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	sub, _ := str.QueryOneSub("argo_uuid", "sub4")
	suite.Equal(200, w.Code)
	suite.Equal("", w.Body.String())
	suite.Equal("https://www.example2.com", sub.PushEndpoint)
	suite.Equal(int64(5), sub.MaxMessages)
	suite.Equal(5000, sub.RetPeriod)
	suite.Equal("linear", sub.RetPolicy)
	suite.False(sub.Verified)
	suite.NotEqual("", sub.VerificationHash)
	suite.NotEqual("push-id-1", sub.VerificationHash)
	suite.NotEqual(subBeforeUpdate.AuthorizationHeader, sub.AuthorizationHeader)
}

// TestSubModPushConfigToActiveORUpdatePushDisabled tests the case where the user modifies the push configuration,
// in order to activate the subscription on the push server
// the push enabled config option is set to false
func (suite *SubscriptionsHandlersTestSuite) TestSubModPushConfigToActiveORUpdatePushDisabled() {

	postJSON := `{
	"pushConfig": {
		 "pushEndpoint": "https://www.example2.com",
		 "retryPolicy": {
             "type":"linear",
             "period": 5000
         }
	}
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub4:modifyPushConfig", strings.NewReader(postJSON))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 409,
      "message": "Push functionality is currently disabled",
      "status": "CONFLICT"
   }
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
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modifyPushConfig", WrapMockAuthConfig(SubModPush, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(409, w.Code)
	suite.Equal(expResp, w.Body.String())
}

// TestSubModPushConfigToActiveORUpdateMissingWorker tests the case where the user modifies the push configuration,
// in order to activate the subscription on the push server
// push enabled is true, but ams can't retrieve the push worker user
func (suite *SubscriptionsHandlersTestSuite) TestSubModPushConfigToActiveORUpdateMissingWorker() {

	postJSON := `{
	"pushConfig": {
		 "pushEndpoint": "https://www.example2.com",
		 "retryPolicy": {
             "type":"linear",
             "period": 5000
         }
	}
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub4:modifyPushConfig", strings.NewReader(postJSON))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 500,
      "message": "Push functionality is currently unavailable",
      "status": "INTERNAL_SERVER_ERROR"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushWorkerToken = "missing"
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modifyPushConfig", WrapMockAuthConfig(SubModPush, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(500, w.Code)
	suite.Equal(expResp, w.Body.String())
}

// TestSubModPushConfigUpdateAuthzDisabled tests the case where the user modifies the push configuration,
// in order to activate the subscription on the push server
// the push configuration was empty before the api call
// since the push endpoint that has been registered is different from the previous verified one
// the sub will be deactivated on the push server and turn into unverified
func (suite *SubscriptionsHandlersTestSuite) TestSubModPushConfigUpdateAuthzDisabled() {

	postJSON := `{
	"pushConfig": {
		 "pushEndpoint": "https://www.example2.com",
         "maxMessages": 5,
  		"authorizationHeader": {
  		"type": "disabled"
		},
		 "retryPolicy": {
             "type":"linear",
             "period": 5000
         }
	}
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub4:modifyPushConfig", strings.NewReader(postJSON))
	if err != nil {
		log.Fatal(err)
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modifyPushConfig", WrapMockAuthConfig(SubModPush, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	sub, _ := str.QueryOneSub("argo_uuid", "sub4")
	suite.Equal(200, w.Code)
	suite.Equal("", w.Body.String())
	suite.Equal("https://www.example2.com", sub.PushEndpoint)
	suite.Equal(int64(5), sub.MaxMessages)
	suite.Equal(5000, sub.RetPeriod)
	suite.Equal("linear", sub.RetPolicy)
	suite.False(sub.Verified)
	suite.NotEqual("", sub.VerificationHash)
	suite.NotEqual("push-id-1", sub.VerificationHash)
	suite.Equal(subscriptions.DisabledAuthorizationHeader, sub.AuthorizationType)
	suite.Equal("", sub.AuthorizationHeader)
}

func (suite *SubscriptionsHandlersTestSuite) TestVerifyPushEndpoint() {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("vhash-1"))
	}))

	defer ts.Close()

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/push-sub-v1:verifyPushEndpoint", nil)
	if err != nil {
		log.Fatal(err)
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")

	// add a temporary subscription
	q1 := stores.QSub{
		Name:             "push-sub-v1",
		ProjectUUID:      "argo_uuid",
		PushEndpoint:     ts.URL,
		PushType:         "http_endpoint",
		VerificationHash: "vhash-1",
		Verified:         false,
	}

	str.SubList = append(str.SubList, q1)
	str.SubsACL["push-sub-v1"] = stores.QAcl{}

	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:verifyPushEndpoint", WrapMockAuthConfig(SubVerifyPushEndpoint, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal("", w.Body.String())
	// check to see that the push worker user has been added to the subscription's acl
	a1, _ := str.QueryACL("argo_uuid", "subscriptions", "push-sub-v1")
	suite.Equal([]string{"uuid7"}, a1.ACL)
}

func (suite *SubscriptionsHandlersTestSuite) TestVerifyPushEndpointHashMisMatch() {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("unknown_hash"))
	}))

	defer ts.Close()

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/push-sub-v1:verifyPushEndpoint", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 401,
      "message": "Endpoint verification failed.Wrong verification hash",
      "status": "UNAUTHORIZED"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")

	// add a temporary subscription
	q1 := stores.QSub{
		Name:             "push-sub-v1",
		ProjectUUID:      "argo_uuid",
		PushEndpoint:     ts.URL,
		PushType:         "http_endpoint",
		VerificationHash: "vhash-1",
		Verified:         false,
	}

	str.SubList = append(str.SubList, q1)
	str.SubsACL["push-sub-v1"] = stores.QAcl{}

	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:verifyPushEndpoint", WrapMockAuthConfig(SubVerifyPushEndpoint, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(401, w.Code)
	suite.Equal(expResp, w.Body.String())
	// check to see that the push worker user has NOT been added to the subscription's acl
	a1, _ := str.QueryACL("argo_uuid", "subscriptions", "push-sub-v1")
	suite.Equal(0, len(a1.ACL))
}

func (suite *SubscriptionsHandlersTestSuite) TestVerifyPushEndpointUnknownResponse() {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unknown_hash"))
	}))

	defer ts.Close()

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/push-sub-v1:verifyPushEndpoint", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 401,
      "message": "Endpoint verification failed.Wrong response status code",
      "status": "UNAUTHORIZED"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")

	// add a temporary subscription
	q1 := stores.QSub{
		Name:             "push-sub-v1",
		ProjectUUID:      "argo_uuid",
		PushEndpoint:     ts.URL,
		PushType:         "http_endpoint",
		VerificationHash: "vhash-1",
		Verified:         false,
	}

	str.SubList = append(str.SubList, q1)
	str.SubsACL["push-sub-v1"] = stores.QAcl{}

	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:verifyPushEndpoint", WrapMockAuthConfig(SubVerifyPushEndpoint, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(401, w.Code)
	suite.Equal(expResp, w.Body.String())
	// check to see that the push worker user has NOT been added to the subscription's acl
	a1, _ := str.QueryACL("argo_uuid", "subscriptions", "push-sub-v1")
	suite.Equal(0, len(a1.ACL))
}

// TestVerifyPushEndpointPushServerError tests the case where the endpoint is verified, the push worker is moved to
// the sub's acl despite the push server being unavailable for now
func (suite *SubscriptionsHandlersTestSuite) TestVerifyPushEndpointPushServerError() {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("vhash-1"))
	}))

	defer ts.Close()

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/errorSub:verifyPushEndpoint", nil)
	if err != nil {
		log.Fatal(err)
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")

	// add a temporary subscription
	q1 := stores.QSub{
		Name:             "errorSub",
		ProjectUUID:      "argo_uuid",
		PushEndpoint:     ts.URL,
		PushType:         "http_endpoint",
		VerificationHash: "vhash-1",
		Verified:         false,
	}

	str.SubList = append(str.SubList, q1)
	str.SubsACL["errorSub"] = stores.QAcl{}

	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:verifyPushEndpoint", WrapMockAuthConfig(SubVerifyPushEndpoint, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal("", w.Body.String())
	// check to see that the push worker user has been added to the subscription's acl
	a1, _ := str.QueryACL("argo_uuid", "subscriptions", "errorSub")
	suite.Equal([]string{"uuid7"}, a1.ACL)
}

func (suite *SubscriptionsHandlersTestSuite) TestVerifyPushEndpointAlreadyVerified() {

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/push-sub-v1:verifyPushEndpoint", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 409,
      "message": "Push endpoint is already verified",
      "status": "CONFLICT"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")

	// add a temporary subscription
	q1 := stores.QSub{
		Name:             "push-sub-v1",
		ProjectUUID:      "argo_uuid",
		PushEndpoint:     "https://example.com/receive_here",
		PushType:         "http_endpoint",
		VerificationHash: "vhash-1",
		Verified:         true,
	}

	str.SubList = append(str.SubList, q1)

	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:verifyPushEndpoint", WrapMockAuthConfig(SubVerifyPushEndpoint, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(409, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestVerifyPushEndpointNotPushEnabled() {

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/push-sub-v1:verifyPushEndpoint", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 409,
      "message": "Subscription is not in http push mode",
      "status": "CONFLICT"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")

	// add a temporary subscription
	q1 := stores.QSub{
		Name:        "push-sub-v1",
		ProjectUUID: "argo_uuid",
		PushType:    "mattermost",
	}

	str.SubList = append(str.SubList, q1)

	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:verifyPushEndpoint", WrapMockAuthConfig(SubVerifyPushEndpoint, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(409, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubCreatePushConfig() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1",
	"pushConfig": {
         "type": "http_endpoint",
		 "pushEndpoint": "https://www.example.com",
		 "retryPolicy": {}
	}
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/projects/ARGO/subscriptions/subNew", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "name": "/projects/ARGO/subscriptions/subNew",
   "topic": "/projects/ARGO/topics/topic1",
   "pushConfig": {
      "type": "http_endpoint",
      "pushEndpoint": "https://www.example.com",
      "maxMessages": 1,
      "authorizationHeader": {
         "type": "autogen",
         "value": "{{AUTHZV}}"
      },
      "retryPolicy": {
         "type": "linear",
         "period": 3000
      },
      "verificationHash": "{{VHASH}}",
      "verified": false,
      "mattermostUrl": "",
      "mattermostUsername": "",
      "mattermostChannel": ""
   },
   "ackDeadlineSeconds": 10,
   "createdOn": "{{CON}}"
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	sub, _ := str.QueryOneSub("argo_uuid", "subNew")
	expResp = strings.Replace(expResp, "{{VHASH}}", sub.VerificationHash, 1)
	expResp = strings.Replace(expResp, "{{AUTHZV}}", sub.AuthorizationHeader, 1)
	expResp = strings.Replace(expResp, "{{CON}}", sub.CreatedOn.Format("2006-01-02T15:04:05Z"), 1)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubCreatePushConfigMattermost() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1",
	"pushConfig": {
         "type": "mattermost",
		 "pushEndpoint": "https://www.example.com",
         "maxMessages": 100,
		 "retryPolicy": {},
         "mattermostUrl": "mywebhook.com",
         "mattermostUsername": "willy",
         "mattermostChannel": "operations"
	}
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/projects/ARGO/subscriptions/subNew", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "name": "/projects/ARGO/subscriptions/subNew",
   "topic": "/projects/ARGO/topics/topic1",
   "pushConfig": {
      "type": "mattermost",
      "pushEndpoint": "",
      "maxMessages": 1,
      "authorizationHeader": {},
      "retryPolicy": {
         "type": "linear",
         "period": 3000
      },
      "verificationHash": "",
      "verified": true,
      "mattermostUrl": "mywebhook.com",
      "mattermostUsername": "willy",
      "mattermostChannel": "operations"
   },
   "ackDeadlineSeconds": 10,
   "createdOn": "{{CON}}"
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	sub, _ := str.QueryOneSub("argo_uuid", "subNew")
	expResp = strings.Replace(expResp, "{{CON}}", sub.CreatedOn.Format("2006-01-02T15:04:05Z"), 1)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubCreatePushConfigMattermostEmptyUrl() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1",
	"pushConfig": {
         "type": "mattermost",
		 "pushEndpoint": "https://www.example.com",
         "maxMessages": 100,
		 "retryPolicy": {},
         "mattermostUrl": "",
         "mattermostUsername": "willy",
         "mattermostChannel": "operations"
	}
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/projects/ARGO/subscriptions/subNew", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 400,
      "message": "Field mattermostUrl cannot be empty",
      "status": "INVALID_ARGUMENT"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubCreatePushConfigInvalidType() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1",
	"pushConfig": {
         "type": "",
		 "pushEndpoint": "https://www.example.com",
         "maxMessages": 100,
		 "retryPolicy": {},
         "mattermostUrl": "",
         "mattermostUsername": "willy",
         "mattermostChannel": "operations"
	}
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/projects/ARGO/subscriptions/subNew", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 400,
      "message": "Push configuration type can only be of 'http_endpoint' or 'mattermost'",
      "status": "INVALID_ARGUMENT"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubCreatePushConfigSlowStart() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1",
	"pushConfig": {
         "type": "http_endpoint",
		 "pushEndpoint": "https://www.example.com",
		 "authorizationHeader": {
         	"type": "disabled"
		 },
		 "retryPolicy": {
			"type": "slowstart"
		 }
	}
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/projects/ARGO/subscriptions/subNew", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "name": "/projects/ARGO/subscriptions/subNew",
   "topic": "/projects/ARGO/topics/topic1",
   "pushConfig": {
      "type": "http_endpoint",
      "pushEndpoint": "https://www.example.com",
      "maxMessages": 1,
      "authorizationHeader": {
         "type": "disabled"
      },
      "retryPolicy": {
         "type": "slowstart"
      },
      "verificationHash": "{{VHASH}}",
      "verified": false,
      "mattermostUrl": "",
      "mattermostUsername": "",
      "mattermostChannel": ""
   },
   "ackDeadlineSeconds": 10,
   "createdOn": "{{CON}}"
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	sub, _ := str.QueryOneSub("argo_uuid", "subNew")
	expResp = strings.Replace(expResp, "{{VHASH}}", sub.VerificationHash, 1)
	expResp = strings.Replace(expResp, "{{CON}}", sub.CreatedOn.Format("2006-01-02T15:04:05Z"), 1)
	suite.Equal(0, sub.RetPeriod)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubCreatePushConfigMissingPushWorker() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1",
	"pushConfig": {
		 "pushEndpoint": "https://www.example.com",
		 "retryPolicy": {}
	}
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/projects/ARGO/subscriptions/subNew", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 500,
      "message": "Push functionality is currently unavailable",
      "status": "INTERNAL_SERVER_ERROR"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushWorkerToken = "missing"
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	// subscription should not have been inserted to the store if it has push configuration
	// but we can't retrieve the push worker
	_, errSub := str.QueryOneSub("argo_uuid", "subNew")
	suite.Equal(500, w.Code)
	suite.Equal(expResp, w.Body.String())
	suite.Equal("empty", errSub.Error())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubCreatePushConfigPushDisabled() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1",
	"pushConfig": {
		 "pushEndpoint": "https://www.example.com",
		 "retryPolicy": {}
	}
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/projects/ARGO/subscriptions/subNew", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 409,
      "message": "Push functionality is currently disabled",
      "status": "CONFLICT"
   }
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
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	// subscription should not have been inserted to the store if it has push configuration
	// but push enables is false
	_, errSub := str.QueryOneSub("argo_uuid", "subNew")
	suite.Equal(409, w.Code)
	suite.Equal(expResp, w.Body.String())
	suite.Equal("empty", errSub.Error())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubCreateInvalidRetPol() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1",
	"pushConfig": {
         "type": "http_endpoint",
		 "pushEndpoint": "https://www.example.com",
		 "retryPolicy": {
			"type": "unknown"
		}
	}
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/projects/ARGO/subscriptions/subNew", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 400,
      "message": "Retry policy can only be of 'linear' or 'slowstart' type",
      "status": "INVALID_ARGUMENT"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubCreatePushConfigError() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1",
	"pushConfig": {
         "type": "http_endpoint",
		 "pushEndpoint": "http://www.example.com",
		 "retryPolicy": {}
	}
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/projects/ARGO/subscriptions/subNew", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 400,
      "message": "Push endpoint should be addressed by a valid https url",
      "status": "INVALID_ARGUMENT"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubCreate() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1"
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/projects/ARGO/subscriptions/subNew", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "name": "/projects/ARGO/subscriptions/subNew",
   "topic": "/projects/ARGO/topics/topic1",
   "pushConfig": {
      "type": "",
      "pushEndpoint": "",
      "maxMessages": 0,
      "authorizationHeader": {},
      "retryPolicy": {},
      "verificationHash": "",
      "verified": false,
      "mattermostUrl": "",
      "mattermostUsername": "",
      "mattermostChannel": ""
   },
   "ackDeadlineSeconds": 10,
   "createdOn": "{{CON}}"
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	sub, _ := str.QueryOneSub("argo_uuid", "subNew")
	fmt.Println(sub)
	expResp = strings.Replace(expResp, "{{CON}}", sub.CreatedOn.Format("2006-01-02T15:04:05Z"), 1)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubCreateExists() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1"
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 409,
      "message": "Subscription already exists",
      "status": "ALREADY_EXISTS"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(409, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubCreateErrorTopic() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topicFoo"
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 404,
      "message": "Topic doesn't exist",
      "status": "NOT_FOUND"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	mgr := oldPush.Manager{}
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubDelete() {

	req, err := http.NewRequest("DELETE", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := ""
	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	mgr := oldPush.Manager{}
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubDelete, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubWithPushConfigDelete() {

	req, err := http.NewRequest("DELETE", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub4", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{"message":"Subscription /projects/ARGO/subscriptions/sub4 deactivated"}`
	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	mgr := oldPush.Manager{}
	router := mux.NewRouter().StrictSlash(true)
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubDelete, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubWithPushConfigDeletePushServerError() {

	req, err := http.NewRequest("DELETE", "http://localhost:8080/v1/projects/ARGO/subscriptions/errorSub", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{"message":"Subscription /projects/ARGO/subscriptions/errorSub is not active"}`
	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	str.SubList = append(str.SubList, stores.QSub{
		Name:         "errorSub",
		ProjectUUID:  "argo_uuid",
		PushEndpoint: "example.com",
		// sub needs to be verified in order to perform the call to the push server
		Verified: true,
		PushType: "mattermost",
	})
	mgr := oldPush.Manager{}
	router := mux.NewRouter().StrictSlash(true)
	pc := new(push.MockClient)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubDelete, cfgKafka, &brk, str, &mgr, pc))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubGetOffsets() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub2:offsets", nil)
	if err != nil {
		log.Fatal(err)
	}
	expResp := `{
   "max": 2,
   "min": 1,
   "current": 1
}`
	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	// append a msg to the broker to cause the min topic from the offset to be at 1 while the sub's current is at 0
	brk.MsgList = append(brk.MsgList, "msg1")
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:offsets", WrapMockAuthConfig(SubGetOffsets, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubListOne() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "name": "/projects/ARGO/subscriptions/sub1",
   "topic": "/projects/ARGO/topics/topic1",
   "pushConfig": {
      "type": "",
      "pushEndpoint": "",
      "maxMessages": 0,
      "authorizationHeader": {},
      "retryPolicy": {},
      "verificationHash": "",
      "verified": false,
      "mattermostUrl": "",
      "mattermostUsername": "",
      "mattermostChannel": ""
   },
   "ackDeadlineSeconds": 10,
   "createdOn": "2020-11-19T00:00:00Z"
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubListOne, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubListAll() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "subscriptions": [
      {
         "name": "/projects/ARGO/subscriptions/sub4",
         "topic": "/projects/ARGO/topics/topic4",
         "pushConfig": {
            "type": "http_endpoint",
            "pushEndpoint": "endpoint.foo",
            "maxMessages": 1,
            "authorizationHeader": {
               "type": "autogen",
               "value": "auth-header-1"
            },
            "retryPolicy": {
               "type": "linear",
               "period": 300
            },
            "verificationHash": "push-id-1",
            "verified": true,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-22T00:00:00Z"
      },
      {
         "name": "/projects/ARGO/subscriptions/sub3",
         "topic": "/projects/ARGO/topics/topic3",
         "pushConfig": {
            "type": "",
            "pushEndpoint": "",
            "maxMessages": 0,
            "authorizationHeader": {},
            "retryPolicy": {},
            "verificationHash": "",
            "verified": false,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-21T00:00:00Z"
      },
      {
         "name": "/projects/ARGO/subscriptions/sub2",
         "topic": "/projects/ARGO/topics/topic2",
         "pushConfig": {
            "type": "",
            "pushEndpoint": "",
            "maxMessages": 0,
            "authorizationHeader": {},
            "retryPolicy": {},
            "verificationHash": "",
            "verified": false,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-20T00:00:00Z"
      },
      {
         "name": "/projects/ARGO/subscriptions/sub1",
         "topic": "/projects/ARGO/topics/topic1",
         "pushConfig": {
            "type": "",
            "pushEndpoint": "",
            "maxMessages": 0,
            "authorizationHeader": {},
            "retryPolicy": {},
            "verificationHash": "",
            "verified": false,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-19T00:00:00Z"
      }
   ],
   "nextPageToken": "",
   "totalSize": 4
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions", WrapMockAuthConfig(SubListAll, cfgKafka, &brk, str, &mgr, nil, "project_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubListAllFirstPage() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions?pageSize=2", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "subscriptions": [
      {
         "name": "/projects/ARGO/subscriptions/sub4",
         "topic": "/projects/ARGO/topics/topic4",
         "pushConfig": {
            "type": "http_endpoint",
            "pushEndpoint": "endpoint.foo",
            "maxMessages": 1,
            "authorizationHeader": {
               "type": "autogen",
               "value": "auth-header-1"
            },
            "retryPolicy": {
               "type": "linear",
               "period": 300
            },
            "verificationHash": "push-id-1",
            "verified": true,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-22T00:00:00Z"
      },
      {
         "name": "/projects/ARGO/subscriptions/sub3",
         "topic": "/projects/ARGO/topics/topic3",
         "pushConfig": {
            "type": "",
            "pushEndpoint": "",
            "maxMessages": 0,
            "authorizationHeader": {},
            "retryPolicy": {},
            "verificationHash": "",
            "verified": false,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-21T00:00:00Z"
      }
   ],
   "nextPageToken": "MQ==",
   "totalSize": 4
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions", WrapMockAuthConfig(SubListAll, cfgKafka, &brk, str, &mgr, nil, "project_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubListAllNextPage() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions?pageSize=2&pageToken=MQ==", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "subscriptions": [
      {
         "name": "/projects/ARGO/subscriptions/sub2",
         "topic": "/projects/ARGO/topics/topic2",
         "pushConfig": {
            "type": "",
            "pushEndpoint": "",
            "maxMessages": 0,
            "authorizationHeader": {},
            "retryPolicy": {},
            "verificationHash": "",
            "verified": false,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-20T00:00:00Z"
      },
      {
         "name": "/projects/ARGO/subscriptions/sub1",
         "topic": "/projects/ARGO/topics/topic1",
         "pushConfig": {
            "type": "",
            "pushEndpoint": "",
            "maxMessages": 0,
            "authorizationHeader": {},
            "retryPolicy": {},
            "verificationHash": "",
            "verified": false,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-19T00:00:00Z"
      }
   ],
   "nextPageToken": "",
   "totalSize": 4
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions", WrapMockAuthConfig(SubListAll, cfgKafka, &brk, str, &mgr, nil, "project_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubListAllEmpty() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "subscriptions": [],
   "nextPageToken": "",
   "totalSize": 0
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	// empty the store
	str.SubList = []stores.QSub{}
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions", WrapMockAuthConfig(SubListAll, cfgKafka, &brk, str, &mgr, nil, "project_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubListAllConsumer() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "subscriptions": [
      {
         "name": "/projects/ARGO/subscriptions/sub4",
         "topic": "/projects/ARGO/topics/topic4",
         "pushConfig": {
            "type": "http_endpoint",
            "pushEndpoint": "endpoint.foo",
            "maxMessages": 1,
            "authorizationHeader": {
               "type": "autogen",
               "value": "auth-header-1"
            },
            "retryPolicy": {
               "type": "linear",
               "period": 300
            },
            "verificationHash": "push-id-1",
            "verified": true,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-22T00:00:00Z"
      },
      {
         "name": "/projects/ARGO/subscriptions/sub3",
         "topic": "/projects/ARGO/topics/topic3",
         "pushConfig": {
            "type": "",
            "pushEndpoint": "",
            "maxMessages": 0,
            "authorizationHeader": {},
            "retryPolicy": {},
            "verificationHash": "",
            "verified": false,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-21T00:00:00Z"
      },
      {
         "name": "/projects/ARGO/subscriptions/sub2",
         "topic": "/projects/ARGO/topics/topic2",
         "pushConfig": {
            "type": "",
            "pushEndpoint": "",
            "maxMessages": 0,
            "authorizationHeader": {},
            "retryPolicy": {},
            "verificationHash": "",
            "verified": false,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-20T00:00:00Z"
      }
   ],
   "nextPageToken": "",
   "totalSize": 3
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions", WrapMockAuthConfig(SubListAll, cfgKafka, &brk, str, &mgr, nil, "consumer"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubListAllConsumerWithPagination() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions?pageSize=2", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "subscriptions": [
      {
         "name": "/projects/ARGO/subscriptions/sub4",
         "topic": "/projects/ARGO/topics/topic4",
         "pushConfig": {
            "type": "http_endpoint",
            "pushEndpoint": "endpoint.foo",
            "maxMessages": 1,
            "authorizationHeader": {
               "type": "autogen",
               "value": "auth-header-1"
            },
            "retryPolicy": {
               "type": "linear",
               "period": 300
            },
            "verificationHash": "push-id-1",
            "verified": true,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-22T00:00:00Z"
      },
      {
         "name": "/projects/ARGO/subscriptions/sub3",
         "topic": "/projects/ARGO/topics/topic3",
         "pushConfig": {
            "type": "",
            "pushEndpoint": "",
            "maxMessages": 0,
            "authorizationHeader": {},
            "retryPolicy": {},
            "verificationHash": "",
            "verified": false,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-21T00:00:00Z"
      }
   ],
   "nextPageToken": "MQ==",
   "totalSize": 3
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions", WrapMockAuthConfig(SubListAll, cfgKafka, &brk, str, &mgr, nil, "consumer"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubListAllInvalidPageSize() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions?pageSize=invalid", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 400,
      "message": "Invalid page size",
      "status": "INVALID_ARGUMENT"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions", WrapMockAuthConfig(SubListAll, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubListAllInvalidPageToken() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions?pageToken=invalid", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 400,
      "message": "Invalid page token",
      "status": "INVALID_ARGUMENT"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions", WrapMockAuthConfig(SubListAll, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestTopicDelete() {

	req, err := http.NewRequest("DELETE", "http://localhost:8080/v1/projects/ARGO/topics/topic1", nil)

	if err != nil {
		log.Fatal(err)
	}

	expResp := ""
	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Topics = map[string]string{}
	brk.Topics["argo_uuid.topic1"] = ""
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapMockAuthConfig(TopicDelete, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
	// make sure the topic got deleted
	suite.Equal(0, len(brk.Topics))
}

func (suite *SubscriptionsHandlersTestSuite) TestSubTimeToOffset() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1?time=2019-06-10T9:38:30.500Z", nil)

	if err != nil {
		log.Fatal(err)
	}

	expResp := `{"offset":93204}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.TopicTimeIndices = map[string][]brokers.TimeToOffset{}

	brk.TopicTimeIndices["argo_uuid.topic1"] = []brokers.TimeToOffset{
		{Timestamp: time.Date(2019, 6, 11, 0, 0, 0, 0, time.UTC), Offset: 93204},
		{Timestamp: time.Date(2019, 6, 12, 0, 0, 0, 0, time.UTC), Offset: 94000},
	}

	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubTimeToOffset, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubTimeToOffsetOutOfBounds() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1?time=2020-06-10T9:38:30.500Z", nil)

	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 409,
      "message": "Timestamp is out of bounds for the subscription's topic/partition",
      "status": "CONFLICT"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.TopicTimeIndices = map[string][]brokers.TimeToOffset{}
	brk.TopicTimeIndices["argo_uuid.topic1"] = []brokers.TimeToOffset{
		{Timestamp: time.Date(2019, 6, 11, 0, 0, 0, 0, time.UTC), Offset: 93204},
		{Timestamp: time.Date(2019, 6, 12, 0, 0, 0, 0, time.UTC), Offset: 94000},
	}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubTimeToOffset, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(409, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubDeleteNotFound() {

	req, err := http.NewRequest("DELETE", "http://localhost:8080/v1/projects/ARGO/subscriptions/subFoo", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 404,
      "message": "Subscription doesn't exist",
      "status": "NOT_FOUND"
   }
}`
	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubDelete, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestModSubACLWrong() {

	postExp := `{"authorized_users":["UserX","UserFoo"]}`

	expRes := `{
   "error": {
      "code": 404,
      "message": "User(s): UserFoo do not exist",
      "status": "NOT_FOUND"
   }
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub101:modAcl", bytes.NewBuffer([]byte(postExp)))
	if err != nil {
		log.Fatal(err)
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modAcl", WrapMockAuthConfig(SubModACL, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expRes, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestModSubACL01() {

	postExp := `{"authorized_users":["UserX","UserZ"]}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscription/sub1:modAcl", bytes.NewBuffer([]byte(postExp)))
	if err != nil {
		log.Fatal(err)
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscription/{subscription}:modAcl", WrapMockAuthConfig(SubModACL, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal("", w.Body.String())

	req2, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscription/sub1:acl", nil)
	if err != nil {
		log.Fatal(err)
	}
	router.HandleFunc("/v1/projects/{project}/subscription/{subscription}:acl", WrapMockAuthConfig(SubACL, cfgKafka, &brk, str, &mgr, nil))
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	suite.Equal(200, w2.Code)

	expResp := `{
   "authorized_users": [
      "UserX",
      "UserZ"
   ]
}`

	suite.Equal(expResp, w2.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubACL01() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscription/sub1:acl", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "authorized_users": [
      "UserA",
      "UserB"
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscription/{subscription}:acl", WrapMockAuthConfig(SubACL, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubACL02() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub3:acl", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "authorized_users": [
      "UserZ",
      "UserB",
      "UserA"
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:acl", WrapMockAuthConfig(SubACL, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubPullOne() {

	postJSON := `{
  "maxMessages":"1"
}`
	url := "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1:pull"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "receivedMessages": [
      {
         "ackId": "projects/ARGO/subscriptions/sub1:0",
         "message": {
            "messageId": "0",
            "attributes": {
               "foo": "bar"
            },
            "data": "YmFzZTY0ZW5jb2RlZA==",
            "publishTime": "2016-02-24T11:55:09.786127994Z"
         }
      }
   ]
}`
	tn := time.Now().UTC()

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Initialize([]string{"localhost"})
	brk.PopulateThree() // Add three messages to the broker queue
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:pull", WrapMockAuthConfig(SubPull, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expJSON, w.Body.String())
	spc, _, _, _ := str.QuerySubs("argo_uuid", "", "sub1", "", 0)
	suite.True(tn.Before(spc[0].LatestConsume))
	suite.NotEqual(spc[0].ConsumeRate, 10)

}

func (suite *SubscriptionsHandlersTestSuite) TestSubPullFromPushEnabledAsPushWorker() {

	postJSON := `{
  "maxMessages":"1"
}`
	url := "http://localhost:8080/v1/projects/ARGO/subscriptions/sub4:pull"
	req, err := http.NewRequest("POST", url, strings.NewReader(postJSON))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "receivedMessages": [
      {
         "ackId": "projects/ARGO/subscriptions/sub4:0",
         "message": {
            "messageId": "0",
            "attributes": {
               "foo": "bar"
            },
            "data": "YmFzZTY0ZW5jb2RlZA==",
            "publishTime": "2016-02-24T11:55:09.786127994Z"
         }
      }
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Initialize([]string{"localhost"})
	brk.PopulateThree() // Add three messages to the broker queue
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:pull", WrapMockAuthConfig(SubPull, cfgKafka, &brk, str, &mgr, nil, "push_worker"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expJSON, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubPullFromPushEnabledAsPushWorkerDISABLED() {

	postJSON := `{
  "maxMessages":"1"
}`
	url := "http://localhost:8080/v1/projects/ARGO/subscriptions/sub4:pull"
	req, err := http.NewRequest("POST", url, strings.NewReader(postJSON))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "error": {
      "code": 409,
      "message": "Push functionality is currently disabled",
      "status": "CONFLICT"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	// disable push functionality
	cfgKafka.PushEnabled = false
	brk := brokers.MockBroker{}
	brk.Initialize([]string{"localhost"})
	brk.PopulateThree() // Add three messages to the broker queue
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:pull", WrapMockAuthConfig(SubPull, cfgKafka, &brk, str, &mgr, nil, "push_worker"))
	router.ServeHTTP(w, req)
	suite.Equal(409, w.Code)
	suite.Equal(expJSON, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubPullFromPushEnabledAsServiceAdmin() {

	postJSON := `{
  "maxMessages":"1"
}`
	url := "http://localhost:8080/v1/projects/ARGO/subscriptions/sub4:pull"
	req, err := http.NewRequest("POST", url, strings.NewReader(postJSON))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "receivedMessages": [
      {
         "ackId": "projects/ARGO/subscriptions/sub4:0",
         "message": {
            "messageId": "0",
            "attributes": {
               "foo": "bar"
            },
            "data": "YmFzZTY0ZW5jb2RlZA==",
            "publishTime": "2016-02-24T11:55:09.786127994Z"
         }
      }
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Initialize([]string{"localhost"})
	brk.PopulateThree() // Add three messages to the broker queue
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:pull", WrapMockAuthConfig(SubPull, cfgKafka, &brk, str, &mgr, nil, "service_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expJSON, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubPullFromPushEnabledNoPushWorker() {

	postJSON := `{
  "maxMessages":"1"
}`
	url := "http://localhost:8080/v1/projects/ARGO/subscriptions/sub4:pull"
	req, err := http.NewRequest("POST", url, strings.NewReader(postJSON))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "error": {
      "code": 403,
      "message": "Access to this resource is forbidden",
      "status": "FORBIDDEN"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Initialize([]string{"localhost"})
	brk.PopulateThree() // Add three messages to the broker queue
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:pull", WrapMockAuthConfig(SubPull, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(403, w.Code)
	suite.Equal(expJSON, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubModAck() {

	postJSON := `{
  "ackDeadlineSeconds":33
}`

	postJSON2 := `{
  "ackDeadlineSeconds":700
}`

	postJSON3 := `{
  "ackDeadlineSeconds":-22
}`

	url := "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1:modifyAckDeadline"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON1 := ``

	expJSON2 := `{
   "error": {
      "code": 400,
      "message": "Invalid ackDeadlineSeconds(needs value between 0 and 600) Arguments",
      "status": "INVALID_ARGUMENT"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Initialize([]string{"localhost"})
	brk.PopulateThree() // Add three messages to the broker queue
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modifyAckDeadline", WrapMockAuthConfig(SubModAck, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expJSON1, w.Body.String())

	subRes, err := str.QueryOneSub("argo_uuid", "sub1")
	suite.Equal(33, subRes.Ack)

	req2, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON2)))
	router2 := mux.NewRouter().StrictSlash(true)
	w2 := httptest.NewRecorder()
	mgr = oldPush.Manager{}
	router2.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modifyAckDeadline", WrapMockAuthConfig(SubModAck, cfgKafka, &brk, str, &mgr, nil))
	router2.ServeHTTP(w2, req2)
	suite.Equal(400, w2.Code)
	suite.Equal(expJSON2, w2.Body.String())

	req3, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON3)))
	router3 := mux.NewRouter().StrictSlash(true)
	w3 := httptest.NewRecorder()
	mgr = oldPush.Manager{}
	router3.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modifyAckDeadline", WrapMockAuthConfig(SubModAck, cfgKafka, &brk, str, &mgr, nil))
	router3.ServeHTTP(w3, req3)
	suite.Equal(400, w3.Code)
	suite.Equal(expJSON2, w3.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubAck() {

	postJSON := `{
  "ackIds":["projects/ARGO/subscriptions/sub2:1"]
}`

	postJSON2 := `{
"ackIds":["projects/ARGO/subscriptions/sub1:2"]
}`

	postJSON3 := `{
"ackIds":["projects/ARGO/subscriptions/sub1:2"]
}`

	url := "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1:acknowledge"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON1 := `{
   "error": {
      "code": 400,
      "message": "Invalid ack id",
      "status": "INVALID_ARGUMENT"
   }
}`

	expJSON2 := `{
   "error": {
      "code": 408,
      "message": "ack timeout",
      "status": "TIMEOUT"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Initialize([]string{"localhost"})
	brk.PopulateThree() // Add three messages to the broker queue
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:acknowledge", WrapMockAuthConfig(SubAck, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expJSON1, w.Body.String())

	// grab sub1
	zSec := "2006-01-02T15:04:05Z"
	t := time.Now().UTC()
	ts := t.Format(zSec)
	str.SubList[0].PendingAck = ts
	str.SubList[0].NextOffset = 3

	req2, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON2)))
	router2 := mux.NewRouter().StrictSlash(true)
	w2 := httptest.NewRecorder()
	mgr = oldPush.Manager{}
	router2.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:acknowledge", WrapMockAuthConfig(SubAck, cfgKafka, &brk, str, &mgr, nil))
	router2.ServeHTTP(w2, req2)
	suite.Equal(200, w2.Code)
	suite.Equal("{}", w2.Body.String())

	// mess with the timeout
	t2 := time.Now().UTC().Add(-11 * time.Second)
	ts2 := t2.Format(zSec)
	str.SubList[0].PendingAck = ts2
	str.SubList[0].NextOffset = 4

	req3, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON3)))
	router3 := mux.NewRouter().StrictSlash(true)
	w3 := httptest.NewRecorder()
	mgr = oldPush.Manager{}
	router3.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:acknowledge", WrapMockAuthConfig(SubAck, cfgKafka, &brk, str, &mgr, nil))
	router3.ServeHTTP(w3, req3)
	suite.Equal(408, w3.Code)
	suite.Equal(expJSON2, w3.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubError() {

	postJSON := `{

}`
	url := "http://localhost:8080/v1/projects/ARGO/subscriptions/foo:pull"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "error": {
      "code": 404,
      "message": "Subscription doesn't exist",
      "status": "NOT_FOUND"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Initialize([]string{"localhost"})
	brk.PopulateThree() // Add three messages to the broker queue
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:pull", WrapMockAuthConfig(SubPull, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *SubscriptionsHandlersTestSuite) TestSubNoTopic() {

	postJSON := `{

}`
	url := "http://localhost:8080/v1/projects/ARGO/subscriptions/no_topic_sub:pull"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "error": {
      "code": 409,
      "message": "Subscription's topic doesn't exist",
      "status": "CONFLICT"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	// add a mock sub that is linked to a non existent topic
	str.SubList = append(str.SubList, stores.QSub{
		Name:        "no_topic_sub",
		ProjectUUID: "argo_uuid",
		Topic:       "unknown_topic"},
	)
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:pull", WrapMockAuthConfig(SubPull, cfgKafka, &brk, str, &mgr, nil, "project_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(409, w.Code)
	suite.Equal(expJSON, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestSubPullAll() {

	postJSON := `{

}`
	url := "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1:pull"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "receivedMessages": [
      {
         "ackId": "projects/ARGO/subscriptions/sub1:0",
         "message": {
            "messageId": "0",
            "attributes": {
               "foo": "bar"
            },
            "data": "YmFzZTY0ZW5jb2RlZA==",
            "publishTime": "2016-02-24T11:55:09.786127994Z"
         }
      },
      {
         "ackId": "projects/ARGO/subscriptions/sub1:1",
         "message": {
            "messageId": "1",
            "attributes": {
               "foo2": "bar2"
            },
            "data": "YmFzZTY0ZW5jb2RlZA==",
            "publishTime": "2016-02-24T11:55:09.827678754Z"
         }
      },
      {
         "ackId": "projects/ARGO/subscriptions/sub1:2",
         "message": {
            "messageId": "2",
            "attributes": {
               "foo2": "bar2"
            },
            "data": "YmFzZTY0ZW5jb2RlZA==",
            "publishTime": "2016-02-24T11:55:09.830417467Z"
         }
      }
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Initialize([]string{"localhost"})
	brk.PopulateThree() // Add three messages to the broker queue
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:pull", WrapMockAuthConfig(SubPull, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expJSON, w.Body.String())
}

func (suite *SubscriptionsHandlersTestSuite) TestValidationInSubs() {

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Initialize([]string{"localhost"})
	brk.PopulateThree() // Add three messages to the broker queue
	str := stores.NewMockStore("whatever", "argo_mgs")

	okResp := `{
   "name": "/projects/ARGO/subscriptions/sub1",
   "topic": "/projects/ARGO/topics/topic1",
   "pushConfig": {
      "type": "",
      "pushEndpoint": "",
      "maxMessages": 0,
      "authorizationHeader": {},
      "retryPolicy": {},
      "verificationHash": "",
      "verified": false,
      "mattermostUrl": "",
      "mattermostUsername": "",
      "mattermostChannel": ""
   },
   "ackDeadlineSeconds": 10,
   "createdOn": "2020-11-19T00:00:00Z"
}`

	invProject := `{
   "error": {
      "code": 400,
      "message": "Invalid project name",
      "status": "INVALID_ARGUMENT"
   }
}`

	invSub := `{
   "error": {
      "code": 400,
      "message": "Invalid subscription name",
      "status": "INVALID_ARGUMENT"
   }
}`

	urls := []string{
		"http://localhost:8080/v1/projects/ARGO/subscriptions/sub1",
		"http://localhost:8080/v1/projects/AR:GO/subscriptions/sub1",
		"http://localhost:8080/v1/projects/ARGO/subscriptions/s,ub1",
		"http://localhost:8080/v1/projects/AR,GO/subscriptions/s:ub1",
	}

	codes := []int(nil)
	responses := []string(nil)

	for _, url := range urls {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte("")))
		router := mux.NewRouter().StrictSlash(true)
		mgr := oldPush.Manager{}
		router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapValidate(WrapMockAuthConfig(SubListOne, cfgKafka, &brk, str, &mgr, nil)))

		if err != nil {
			log.Fatal(err)
		}

		router.ServeHTTP(w, req)
		codes = append(codes, w.Code)
		responses = append(responses, w.Body.String())

	}

	// First request is valid so response is ok
	suite.Equal(200, codes[0])
	suite.Equal(okResp, responses[0])

	// Second request has invalid project name
	suite.Equal(400, codes[1])
	suite.Equal(invProject, responses[1])

	// Third  request has invalid subscription name
	suite.Equal(400, codes[2])
	suite.Equal(invSub, responses[2])

	// Fourth request has invalid project and subscription name, but project is caught first
	suite.Equal(400, codes[3])
	suite.Equal(invProject, responses[3])

}

func TestSubscriptionsHandlersTestSuite(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	suite.Run(t, new(SubscriptionsHandlersTestSuite))
}
