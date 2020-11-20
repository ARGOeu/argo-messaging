package handlers

import (
	"bytes"
	"fmt"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	oldPush "github.com/ARGOeu/argo-messaging/push"
	push "github.com/ARGOeu/argo-messaging/push/grpc/client"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type TopicsHandlersTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *TopicsHandlersTestSuite) SetupTest() {
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

func (suite *TopicsHandlersTestSuite) TestTopicDeleteNotfound() {

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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapMockAuthConfig(TopicDelete, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *TopicsHandlersTestSuite) TestTopicCreate() {

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
	mgr := oldPush.Manager{}

	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapMockAuthConfig(TopicCreate, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *TopicsHandlersTestSuite) TestTopicCreateExists() {

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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapMockAuthConfig(TopicCreate, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(409, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *TopicsHandlersTestSuite) TestTopicListOne() {

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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapMockAuthConfig(TopicListOne, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *TopicsHandlersTestSuite) TestTopicListSubscriptions() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics/topic1/subscriptions", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{"subscriptions":["/projects/ARGO/subscriptions/sub1"]}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}/subscriptions", WrapMockAuthConfig(ListSubsByTopic, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *TopicsHandlersTestSuite) TestTopicListSubscriptionsEmpty() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics/topic1/subscriptions", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{"subscriptions":[]}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	str.SubList = nil
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}/subscriptions", WrapMockAuthConfig(ListSubsByTopic, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *TopicsHandlersTestSuite) TestModTopicACL01() {

	postExp := `{"authorized_users":["UserX","UserZ"]}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/topics/topic1:modAcl", bytes.NewBuffer([]byte(postExp)))
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
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:modAcl", WrapMockAuthConfig(TopicModACL, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal("", w.Body.String())

	req2, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics/topic1:acl", nil)
	if err != nil {
		log.Fatal(err)
	}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:acl", WrapMockAuthConfig(TopicACL, cfgKafka, &brk, str, &mgr, nil))
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

func (suite *TopicsHandlersTestSuite) TestTopicACL01() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics/topic1:acl", nil)
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
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:acl", WrapMockAuthConfig(TopicACL, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *TopicsHandlersTestSuite) TestTopicACL02() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics/topic3:acl", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "authorized_users": [
      "UserX"
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:acl", WrapMockAuthConfig(TopicACL, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *TopicsHandlersTestSuite) TestModTopicACLWrong() {

	postExp := `{"authorized_users":["UserX","UserFoo"]}`

	expRes := `{
   "error": {
      "code": 404,
      "message": "User(s): UserFoo do not exist",
      "status": "NOT_FOUND"
   }
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/topics/topic1:modAcl", bytes.NewBuffer([]byte(postExp)))
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
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:modAcl", WrapMockAuthConfig(TopicModACL, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expRes, w.Body.String())

}

func (suite *TopicsHandlersTestSuite) TestTopicListAll() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "topics": [
      {
         "name": "/projects/ARGO/topics/topic4"
      },
      {
         "name": "/projects/ARGO/topics/topic3",
         "schema": "projects/ARGO/schemas/schema-3"
      },
      {
         "name": "/projects/ARGO/topics/topic2",
         "schema": "projects/ARGO/schemas/schema-1"
      },
      {
         "name": "/projects/ARGO/topics/topic1"
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
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics", WrapMockAuthConfig(TopicListAll, cfgKafka, &brk, str, &mgr, nil, "project_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *TopicsHandlersTestSuite) TestTopicListAllPublisher() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "topics": [
      {
         "name": "/projects/ARGO/topics/topic2",
         "schema": "projects/ARGO/schemas/schema-1"
      },
      {
         "name": "/projects/ARGO/topics/topic1"
      }
   ],
   "nextPageToken": "",
   "totalSize": 2
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics", WrapMockAuthConfig(TopicListAll, cfgKafka, &brk, str, &mgr, nil, "publisher"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *TopicsHandlersTestSuite) TestTopicListAllPublisherWithPagination() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics?pageSize=1", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "topics": [
      {
         "name": "/projects/ARGO/topics/topic2",
         "schema": "projects/ARGO/schemas/schema-1"
      }
   ],
   "nextPageToken": "MA==",
   "totalSize": 2
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics", WrapMockAuthConfig(TopicListAll, cfgKafka, &brk, str, &mgr, nil, "publisher"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *TopicsHandlersTestSuite) TestPublishWithSchema() {

	type td struct {
		topic              string
		postBody           string
		expectedResponse   string
		expectedStatusCode int
		msg                string
	}

	testData := []td{
		{
			topic: "topic2",
			postBody: `{
	"messages" : [
		
		{
			"attributes": {},
			"data": "eyJuYW1lIjoibmFtZS0xIiwgImVtYWlsIjogInRlc3RAZXhhbXBsZS5jb20ifQ=="
		},
		
		{
			"attributes": {},
			"data": "eyJuYW1lIjoibmFtZS0xIiwgImVtYWlsIjogInRlc3RAZXhhbXBsZS5jb20iLCAiYWRkcmVzcyI6IlN0cmVldCAxMyIsInRlbGVwaG9uZSI6IjY5NDg1Njc4ODkifQ=="
		}
	]
}`,
			expectedStatusCode: 200,
			expectedResponse: `{
   "messageIds": [
      "1",
      "2"
   ]
}`,
			msg: "Case where the messages are validated successfully(JSON)",
		},
		{
			topic: "topic3",
			postBody: `{
	"messages" : [
		
		{
			"attributes": {},
			"data": "DGFnZWxvc8T8Cg=="
		},	
		{
			"attributes": {},
			"data": "DGFnZWxvc8T8Cg=="
		}
	]
}`,
			expectedStatusCode: 200,
			expectedResponse: `{
   "messageIds": [
      "3",
      "4"
   ]
}`,
			msg: "Case where the messages are validated successfully(AVRO)",
		},
		{
			topic: "topic2",
			postBody: `{
	"messages" : [
		
		{
			"attributes": {},
			"data": "eyJuYW1lIjoibmFtZS0xIiwiYWRkcmVzcyI6IlN0cmVldCAxMyIsInRlbGVwaG9uZSI6Njk0ODU2Nzg4OX0="
		},
		
		{
			"attributes": {},
			"data": "eyJuYW1lIjoibmFtZS0xIiwgImVtYWlsIjogInRlc3RAZXhhbXBsZS5jb20iLCAiYWRkcmVzcyI6IlN0cmVldCAxMyIsInRlbGVwaG9uZSI6IjY5NDg1Njc4ODkifQ=="
		}
	]
}`,
			expectedStatusCode: 400,
			expectedResponse: `{
   "error": {
      "code": 400,
      "message": "Message 0 data is not valid.1)(root): email is required.2)telephone: Invalid type. Expected: string, given: integer.",
      "status": "INVALID_ARGUMENT"
   }
}`,
			msg: "Case where one of the messages is not successfully validated(2 errors)",
		},
		{
			topic: "topic3",
			postBody: `{
	"messages" : [
		
		{
			"attributes": {},
			"data": "T2JqAQQWYXZyby5zY2hlbWGYAnsidHlwZSI6InJlY29yZCIsIm5hbWUiOiJQbGFjZSIsIm5hbWVzcGFjZSI6InBsYWNlLmF2cm8iLCJmaWVsZHMiOlt7Im5hbWUiOiJwbGFjZW5hbWUiLCJ0eXBlIjoic3RyaW5nIn0seyJuYW1lIjoiYWRkcmVzcyIsInR5cGUiOiJzdHJpbmcifV19FGF2cm8uY29kZWMIbnVsbABM1P4b0GpYaCg9tqxa+YDZAiQSc3RyZWV0IDIyDnBsYWNlIGFM1P4b0GpYaCg9tqxa+YDZ"
		},
		
		{
			"attributes": {},
			"data": "DGFnZWxvc8T8Cg=="
		}
	]
}`,
			expectedStatusCode: 400,
			expectedResponse: `{
   "error": {
      "code": 400,
      "message": "Message 0 is not valid.cannot decode binary record \"user.avro.User\" field \"username\": cannot decode binary string: cannot decode binary bytes: negative size: -40",
      "status": "INVALID_ARGUMENT"
   }
}`,
			msg: "Case where one of the messages is not successfully validated(1 error)(AVRO)",
		},

		{
			topic: "topic2",
			postBody: `{
	"messages" : [
		
		{
			"attributes": {},
			"data": "eyJuYW1lIjoibmFtZS0xIiwiYWRkcmVzcyI6IlN0cmVldCAxMyIsInRlbGVwaG9uZSI6IjY5NDg1Njc4ODkifQo="
		},
		
		{
			"attributes": {},
			"data": "eyJuYW1lIjoibmFtZS0xIiwgImVtYWlsIjogInRlc3RAZXhhbXBsZS5jb20iLCAiYWRkcmVzcyI6IlN0cmVldCAxMyIsInRlbGVwaG9uZSI6IjY5NDg1Njc4ODkifQ=="
		}
	]
}`,
			expectedStatusCode: 400,
			expectedResponse: `{
   "error": {
      "code": 400,
      "message": "Message 0 data is not valid,(root): email is required",
      "status": "INVALID_ARGUMENT"
   }
}`,
			msg: "Case where the one of the messages is not successfully validated(1 error)",
		},
		{
			topic: "topic2",
			postBody: `{
	"messages" : [
		
		{
			"attributes": {},
			"data": "eyJuYW1lIjoibmFtZS0xIiwgImVtYWlsIjogInRlc3RAZXhhbXBsZS5jb20iLCAiYWRkcmVzcyI6IlN0cmVldCAxMyIsInRlbGVwaG9uZSI6IjY5NDg1Njc4ODkifQ=="
		},
		
		{
			"attributes": {},
			"data": "eyJuYW1lIjoibmFtZS0xIiwiYWRkcmVzcyI6IlN0cmVldCAxMyIsInRlbGVwaG9uZSI6IjY5NDg1Njc4ODkiCg=="
		}
	]
}`,
			expectedStatusCode: 400,
			expectedResponse: `{
   "error": {
      "code": 400,
      "message": "Message 1 data is not valid JSON format,unexpected EOF",
      "status": "INVALID_ARGUMENT"
   }
}`,
			msg: "Case where the one of the messages is not in valid json format",
		},
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushEnabled = true
	cfgKafka.PushWorkerToken = "push_token"
	cfgKafka.ResAuth = false
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)

	for _, t := range testData {

		w := httptest.NewRecorder()
		url := fmt.Sprintf("http://localhost:8080/v1/projects/ARGO/topics/%v", t.topic)
		req, err := http.NewRequest("POST", url, strings.NewReader(t.postBody))
		if err != nil {
			log.Fatal(err)
		}
		router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapMockAuthConfig(TopicPublish, cfgKafka, &brk, str, &mgr, pc))
		router.ServeHTTP(w, req)

		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}
}

func (suite *TopicsHandlersTestSuite) TestTopicListAllFirstPage() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics?pageSize=2", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "topics": [
      {
         "name": "/projects/ARGO/topics/topic4"
      },
      {
         "name": "/projects/ARGO/topics/topic3",
         "schema": "projects/ARGO/schemas/schema-3"
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
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics", WrapMockAuthConfig(TopicListAll, cfgKafka, &brk, str, &mgr, nil, "project_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *TopicsHandlersTestSuite) TestTopicListAllNextPage() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics?pageSize=2&pageToken=MA==", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "topics": [
      {
         "name": "/projects/ARGO/topics/topic1"
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
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics", WrapMockAuthConfig(TopicListAll, cfgKafka, &brk, str, &mgr, nil, "project_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *TopicsHandlersTestSuite) TestTopicListAllEmpty() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics?pageSize=2", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "topics": [],
   "nextPageToken": "",
   "totalSize": 0
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	// empty the store
	str.TopicList = []stores.QTopic{}
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics", WrapMockAuthConfig(TopicListAll, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *TopicsHandlersTestSuite) TestTopicListAllInvalidPageSize() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics?pageSize=invalid", nil)
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
	router.HandleFunc("/v1/projects/{project}/topics", WrapMockAuthConfig(TopicListAll, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *TopicsHandlersTestSuite) TestTopicListAllInvalidPageToken() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics?pageToken=invalid", nil)
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
	router.HandleFunc("/v1/projects/{project}/topics", WrapMockAuthConfig(TopicListAll, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *TopicsHandlersTestSuite) TestPublish() {

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
	tn := time.Now().UTC()

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapMockAuthConfig(TopicPublish, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expJSON, w.Body.String())
	tpc, _, _, _ := str.QueryTopics("argo_uuid", "", "topic1", "", 0)
	suite.True(tn.Before(tpc[0].LatestPublish))
	suite.NotEqual(tpc[0].PublishRate, 10)

}

func (suite *TopicsHandlersTestSuite) TestPublishMultiple() {

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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapMockAuthConfig(TopicPublish, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *TopicsHandlersTestSuite) TestPublishError() {

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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapMockAuthConfig(TopicPublish, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *TopicsHandlersTestSuite) TestPublishNoTopic() {

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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapMockAuthConfig(TopicPublish, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expJSON, w.Body.String())
}

func (suite *TopicsHandlersTestSuite) TestValidationInTopics() {

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
		mgr := oldPush.Manager{}
		router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapValidate(WrapMockAuthConfig(TopicListOne, cfgKafka, &brk, str, &mgr, nil)))

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

func TestTopicsHandlersTestSuite(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	suite.Run(t, new(TopicsHandlersTestSuite))
}
