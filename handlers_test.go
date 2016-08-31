package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/push"
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
	"store_host":"localhost",
	"store_db":"argo_msg",
	"certificate":"/etc/pki/tls/certs/localhost.crt",
	"certificate_key":"/etc/pki/tls/private/localhost.key",
	"per_resource_auth":"true"
	}`

	log.SetOutput(ioutil.Discard)
}

func (suite *HandlerTestSuite) TestValidation() {
	suite.Equal(true, validName("topic101"))
	suite.Equal(true, validName("topic_101"))
	suite.Equal(true, validName("topic_101_another_thing"))
	suite.Equal(true, validName("topic___343_random"))
	suite.Equal(true, validName("topic_dc1cc538-1361-4317-a235-0bf383d4a69f"))
	suite.Equal(false, validName("topic_dc1cc538.1361-4317-a235-0bf383d4a69f"))
	suite.Equal(false, validName("topic.not.valid"))
	suite.Equal(false, validName("spaces are not valid"))
	suite.Equal(false, validName("topic/A"))
	suite.Equal(false, validName("topic/B"))

}

func (suite *HandlerTestSuite) TestSubCreate() {

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
      "pushEndpoint": "",
      "retryPolicy": {}
   },
   "ackDeadlineSeconds": 10
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := push.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapConfig(SubCreate, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestSubCreateExists() {

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
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapConfig(SubCreate, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(409, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestSubCreateErrorTopic() {

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
	mgr := push.Manager{}
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapConfig(SubCreate, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestSubDelete() {

	req, err := http.NewRequest("DELETE", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := ""
	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	mgr := push.Manager{}
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapConfig(SubDelete, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestSubListOne() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "name": "/projects/ARGO/subscriptions/sub1",
   "topic": "/projects/ARGO/topics/topic1",
   "pushConfig": {
      "pushEndpoint": "",
      "retryPolicy": {}
   },
   "ackDeadlineSeconds": 10
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapConfig(SubListOne, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestSubListAll() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "subscriptions": [
      {
         "name": "/projects/ARGO/subscriptions/sub1",
         "topic": "/projects/ARGO/topics/topic1",
         "pushConfig": {
            "pushEndpoint": "",
            "retryPolicy": {}
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub2",
         "topic": "/projects/ARGO/topics/topic2",
         "pushConfig": {
            "pushEndpoint": "",
            "retryPolicy": {}
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub3",
         "topic": "/projects/ARGO/topics/topic3",
         "pushConfig": {
            "pushEndpoint": "",
            "retryPolicy": {}
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub4",
         "topic": "/projects/ARGO/topics/topic4",
         "pushConfig": {
            "pushEndpoint": "endpoint.foo",
            "retryPolicy": {
               "type": "linear",
               "period": 300
            }
         },
         "ackDeadlineSeconds": 10
      }
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := push.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions", WrapConfig(SubListAll, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestTopicDelete() {

	req, err := http.NewRequest("DELETE", "http://localhost:8080/v1/projects/ARGO/topics/topic1", nil)

	if err != nil {
		log.Fatal(err)
	}

	expResp := ""
	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapConfig(TopicDelete, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestSubDeleteNotfound() {

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
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapConfig(SubDelete, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestTopicDeleteNotfound() {

	req, err := http.NewRequest("DELETE", "http://localhost:8080/v1/projects/ARGO/topics/topicFoo", nil)

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
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapConfig(TopicDelete, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestTopicCreate() {

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/projects/ARGO/topics/topicNew", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "name": "/projects/ARGO/topics/topicNew"
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}

	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapConfig(TopicCreate, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestTopicCreateExists() {

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/projects/ARGO/topics/topic1", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 409,
      "message": "Topic already exists",
      "status": "ALREADY_EXISTS"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapConfig(TopicCreate, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(409, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestTopicListOne() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics/topic1", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "name": "/projects/ARGO/topics/topic1"
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapConfig(TopicListOne, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestTopicListAll() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "topics": [
      {
         "name": "/projects/ARGO/topics/topic1"
      },
      {
         "name": "/projects/ARGO/topics/topic2"
      },
      {
         "name": "/projects/ARGO/topics/topic3"
      }
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics", WrapConfig(TopicListAll, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestPublish() {

	postJSON := `{
  "messages": [
    {
      "attributes":
        {
         "foo":"bar"
        }
      ,
      "data": "YmFzZTY0ZW5jb2RlZA=="
    }
  ]
}`
	url := "http://localhost:8080/v1/projects/ARGO/topics/topic1:publish"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "messageIds": [
      "1"
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapMockAuthConfig(TopicPublish, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *HandlerTestSuite) TestPublishMultiple() {

	postJSON := `{
  "messages": [
    {
      "attributes":
        {
          "foo":"bar"
        }
      ,
      "data": "YmFzZTY0ZW5jb2RlZA=="
    },
    {
      "attributes":
        {
      		"foo2":"bar2"
        }
      ,
      "data": "YmFzZTY0ZW5jb2RlZA=="
    },
    {
      "attributes":
        {
          "foo2":"bar2"
        }
      ,
      "data": "YmFzZTY0ZW5jb2RlZA=="
    }
  ]
}`
	url := "http://localhost:8080/v1/projects/ARGO/topics/topic1:publish"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "messageIds": [
      "1",
      "2",
      "3"
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Initialize([]string{"localhost"})
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapMockAuthConfig(TopicPublish, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *HandlerTestSuite) TestPublishError() {

	postJSON := `{
  "messages": [
    {
      "attributes": [
        {
          "key": "foo",
          "valu2RlZA=="
    },
    {
      "attributes": [
        {
          "key": "foo2",
          "value": "bar2"
        }
      ],
      "data": "YmFzZTY0ZW5jb2RlZA=="
    }
  ]
}`
	url := "http://localhost:8080/v1/projects/ARGO/topics/topic1:publish"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "error": {
      "code": 400,
      "message": "Invalid Message Arguments",
      "status": "INVALID_ARGUMENT"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapMockAuthConfig(TopicPublish, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *HandlerTestSuite) TestPublishNoTopic() {

	postJSON := `{
  "messages": [
		{
			"attributes": [
				{
					"key": "foo",
					"value": "bar"
				}
			],
			"data": "YmFzZTY0ZW5jb2RlZA=="
		},
    {
      "attributes": [
        {
          "key": "foo2",
          "value": "bar2"
        }
      ],
      "data": "YmFzZTY0ZW5jb2RlZA=="
    }
  ]
}`
	url := "http://localhost:8080/v1/projects/ARGO/topics/FOO:publish"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
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
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapMockAuthConfig(TopicPublish, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *HandlerTestSuite) TestSubPullOne() {

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

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Initialize([]string{"localhost"})
	brk.PopulateThree() // Add three messages to the broker queue
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:pull", WrapMockAuthConfig(SubPull, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *HandlerTestSuite) TestSubAck() {

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
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:acknowledge", WrapConfig(SubAck, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expJSON1, w.Body.String())

	// grab sub1
	zSec := "2006-01-02T15:04:05Z"
	t := time.Now()
	ts := t.Format(zSec)
	str.SubList[0].PendingAck = ts
	str.SubList[0].NextOffset = 3

	req2, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON2)))
	router2 := mux.NewRouter().StrictSlash(true)
	w2 := httptest.NewRecorder()
	mgr = push.Manager{}
	router2.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:acknowledge", WrapConfig(SubAck, cfgKafka, &brk, str, &mgr))
	router2.ServeHTTP(w2, req2)
	suite.Equal(200, w2.Code)
	suite.Equal("{}", w2.Body.String())

	// mess with the timeout
	t2 := time.Now().Add(-11 * time.Second)
	ts2 := t2.Format(zSec)
	str.SubList[0].PendingAck = ts2
	str.SubList[0].NextOffset = 4

	req3, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON3)))
	router3 := mux.NewRouter().StrictSlash(true)
	w3 := httptest.NewRecorder()
	mgr = push.Manager{}
	router3.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:acknowledge", WrapConfig(SubAck, cfgKafka, &brk, str, &mgr))
	router3.ServeHTTP(w3, req3)
	suite.Equal(408, w3.Code)
	suite.Equal(expJSON2, w3.Body.String())

}

func (suite *HandlerTestSuite) TestSubError() {

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
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:pull", WrapMockAuthConfig(SubPull, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *HandlerTestSuite) TestSubPullAll() {

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
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:pull", WrapMockAuthConfig(SubPull, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *HandlerTestSuite) TestValidationInSubs() {

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
      "pushEndpoint": "",
      "retryPolicy": {}
   },
   "ackDeadlineSeconds": 10
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
		mgr := push.Manager{}
		router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapValidate(WrapConfig(SubListOne, cfgKafka, &brk, str, &mgr)))

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

func (suite *HandlerTestSuite) TestValidationInTopics() {

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Initialize([]string{"localhost"})
	brk.PopulateThree() // Add three messages to the broker queue
	str := stores.NewMockStore("whatever", "argo_mgs")

	okResp := `{
   "name": "/projects/ARGO/topics/topic1"
}`
	invProject := `{
   "error": {
      "code": 400,
      "message": "Invalid project name",
      "status": "INVALID_ARGUMENT"
   }
}`

	invTopic := `{
   "error": {
      "code": 400,
      "message": "Invalid topic name",
      "status": "INVALID_ARGUMENT"
   }
}`

	urls := []string{
		"http://localhost:8080/v1/projects/ARGO/topics/topic1",
		"http://localhost:8080/v1/projects/AR:GO/topics/topic1",
		"http://localhost:8080/v1/projects/ARGO/topics/top,ic1",
		"http://localhost:8080/v1/projects/AR,GO/topics/top:ic1",
	}

	codes := []int(nil)
	responses := []string(nil)

	for _, url := range urls {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte("")))
		router := mux.NewRouter().StrictSlash(true)
		mgr := push.Manager{}
		router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapValidate(WrapConfig(TopicListOne, cfgKafka, &brk, str, &mgr)))

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

	// Third  request has invalid topic name
	suite.Equal(400, codes[2])
	suite.Equal(invTopic, responses[2])

	// Fourth request has invalid project and topic names, but project is caught first
	suite.Equal(400, codes[3])
	suite.Equal(invProject, responses[3])

}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
