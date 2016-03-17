package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
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
		"broker_host":"localhost:9092",
		"store_host":"localhost",
		"store_db":"argo_msg"
	}`
	log.SetOutput(ioutil.Discard)
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

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapConfig(SubListOne, cfgKafka, &brk, str))
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
      },
      {
         "name": "/projects/ARGO/subscriptions/sub3",
         "topic": "/projects/ARGO/topics/topic3",
         "pushConfig": {
            "pushEndpoint": ""
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
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions", WrapConfig(SubListAll, cfgKafka, &brk, str))
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
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapConfig(TopicDelete, cfgKafka, &brk, str))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
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
      "message": "Topic Not Found",
      "errors": [
         {
            "message": "Topic Not Found",
            "domain": "global",
            "reason": "backend"
         }
      ],
      "status": "INTERNAL"
   }
}`
	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapConfig(TopicDelete, cfgKafka, &brk, str))
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
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapConfig(TopicCreate, cfgKafka, &brk, str))
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
      "message": "Topic Already Exists",
      "errors": [
         {
            "message": "Topic Already Exists",
            "domain": "global",
            "reason": "backend"
         }
      ],
      "status": "INTERNAL"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapConfig(TopicCreate, cfgKafka, &brk, str))
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
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapConfig(TopicListOne, cfgKafka, &brk, str))
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
	router.HandleFunc("/v1/projects/{project}/topics", WrapConfig(TopicListAll, cfgKafka, &brk, str))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestPublish() {

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
	url := "http://localhost:8080/v1/projects/ARGO/topics/topic1:publish"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "messageIDs": [
      "1"
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapConfig(TopicPublish, cfgKafka, &brk, str))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *HandlerTestSuite) TestPublishMultiple() {

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
	url := "http://localhost:8080/v1/projects/ARGO/topics/topic1:publish"
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

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Initialize(cfgKafka.BrokerHost)
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapConfig(TopicPublish, cfgKafka, &brk, str))
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

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapConfig(TopicPublish, cfgKafka, &brk, str))
	router.ServeHTTP(w, req)
	suite.Equal(500, w.Code)
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
      "message": "POST: Project/Topic combination: ARGO/FOO doesnt exist",
      "errors": [
         {
            "message": "POST: Project/Topic combination: ARGO/FOO doesnt exist",
            "domain": "global",
            "reason": "backend"
         }
      ],
      "status": "INTERNAL"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapConfig(TopicPublish, cfgKafka, &brk, str))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *HandlerTestSuite) TestSubPullAll() {

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
         "message": {
            "messageId": "0",
            "attributes": [
               {
                  "key": "foo",
                  "value": "bar"
               }
            ],
            "data": "YmFzZTY0ZW5jb2RlZA==",
            "publishTime": "2016-02-24T11:55:09.786127994Z"
         }
      }
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Initialize(cfgKafka.BrokerHost)
	brk.PopulateThree() // Add three messages to the broker queue
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:pull", WrapConfig(SubPull, cfgKafka, &brk, str))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expJSON, w.Body.String())

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
      "message": "POST: Project/subscription combination: ARGO/foo doesnt exist",
      "errors": [
         {
            "message": "POST: Project/subscription combination: ARGO/foo doesnt exist",
            "domain": "global",
            "reason": "backend"
         }
      ],
      "status": "INTERNAL"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Initialize(cfgKafka.BrokerHost)
	brk.PopulateThree() // Add three messages to the broker queue
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:pull", WrapConfig(SubPull, cfgKafka, &brk, str))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *HandlerTestSuite) TestSubPullOne() {

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
         "message": {
            "messageId": "0",
            "attributes": [
               {
                  "key": "foo",
                  "value": "bar"
               }
            ],
            "data": "YmFzZTY0ZW5jb2RlZA==",
            "publishTime": "2016-02-24T11:55:09.786127994Z"
         }
      },
      {
         "message": {
            "messageId": "1",
            "attributes": [
               {
                  "key": "foo2",
                  "value": "bar2"
               }
            ],
            "data": "YmFzZTY0ZW5jb2RlZA==",
            "publishTime": "2016-02-24T11:55:09.827678754Z"
         }
      },
      {
         "message": {
            "messageId": "2",
            "attributes": [
               {
                  "key": "foo2",
                  "value": "bar2"
               }
            ],
            "data": "YmFzZTY0ZW5jb2RlZA==",
            "publishTime": "2016-02-24T11:55:09.830417467Z"
         }
      }
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	brk.Initialize(cfgKafka.BrokerHost)
	brk.PopulateThree() // Add three messages to the broker queue
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:pull", WrapConfig(SubPull, cfgKafka, &brk, str))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
