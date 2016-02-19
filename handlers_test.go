package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/gorilla/mux"
	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/stretchr/testify/suite"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
)

type HandlerTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.cfgStr = `
	{
	  "server":"localhost:9092",
	  "topics":["topic1","topic2"],
		"subscriptions":{"sub1":"topic1","sub2":"topic2"}
	}
	`
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
      "pushEndpoint": ""
   },
   "ackDeadlineSeconds": 10
}`

	cfgKafka := config.NewKafkaCfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapConfig(SubListOne, cfgKafka, &brk))
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
            "pushEndpoint": ""
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub2",
         "topic": "/projects/ARGO/topics/topic2",
         "pushConfig": {
            "pushEndpoint": ""
         },
         "ackDeadlineSeconds": 10
      }
   ]
}`

	cfgKafka := config.NewKafkaCfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions", WrapConfig(SubListAll, cfgKafka, &brk))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
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

	cfgKafka := config.NewKafkaCfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapConfig(TopicListOne, cfgKafka, &brk))
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
      }
   ]
}`

	cfgKafka := config.NewKafkaCfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/topics", WrapConfig(TopicListAll, cfgKafka, &brk))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestPublish() {
	//broker.Initialize(kafkaCfg.Server)
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
    }
  ]
}`
	url := "http://localhost:8080/v1/projects/ARGO/topics/mocktopic:publish"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "messageIDs": [
      "1"
   ]
}`

	cfgKafka := config.NewKafkaCfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapConfig(TopicPublish, cfgKafka, &brk))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *HandlerTestSuite) TestPublishMultiple() {
	//broker.Initialize(kafkaCfg.Server)
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
	url := "http://localhost:8080/v1/projects/ARGO/topics/mocktopic:publish"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "messageIDs": [
      "1",
      "2",
      "3"
   ]
}`

	cfgKafka := config.NewKafkaCfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapConfig(TopicPublish, cfgKafka, &brk))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *HandlerTestSuite) TestPublishError() {
	//broker.Initialize(kafkaCfg.Server)
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
	url := "http://localhost:8080/v1/projects/ARGO/topics/mocktopic:publish"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "error": {
      "code": 500,
      "message": "POST: Input JSON schema is not valid",
      "errors": [
         {
            "message": "POST: Input JSON schema is not valid",
            "domain": "global",
            "reason": "backend"
         }
      ],
      "status": "INTERNAL"
   }
}`

	cfgKafka := config.NewKafkaCfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapConfig(TopicPublish, cfgKafka, &brk))
	router.ServeHTTP(w, req)
	suite.Equal(500, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func TestFactorsTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
