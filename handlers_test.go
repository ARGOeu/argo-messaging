package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/ARGOeu/argo-messaging/auth"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/metrics"
	"github.com/ARGOeu/argo-messaging/projects"
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
	"kafka_znode":"",
	"store_host":"localhost",
	"store_db":"argo_msg",
	"certificate":"/etc/pki/tls/certs/localhost.crt",
	"certificate_key":"/etc/pki/tls/private/localhost.key",
	"per_resource_auth":"true"
	}`

	log.SetOutput(ioutil.Discard)
}

func (suite *HandlerTestSuite) TestValidHTTPS() {
	suite.Equal(false, isValidHTTPS("ht"))
	suite.Equal(false, isValidHTTPS("www.example.com"))
	suite.Equal(false, isValidHTTPS("https:www.example.com"))
	suite.Equal(false, isValidHTTPS("http://www.example.com"))
	suite.Equal(true, isValidHTTPS("https://www.example.com"))

}

func (suite *HandlerTestSuite) TestValidation() {
	// nameValidations
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

	// ackID validations
	suite.Equal(true, validAckID("ARGO", "sub101", "projects/ARGO/subscriptions/sub101:5"))
	suite.Equal(false, validAckID("ARGO", "sub101", "projects/ARGO/subscriptions/sub101:aaa"))
	suite.Equal(false, validAckID("ARGO", "sub101", "projects/FARGO/subscriptions/sub101:5"))
	suite.Equal(false, validAckID("ARGO", "sub101", "projects/ARGO/subscriptions/subF00:5"))
	suite.Equal(false, validAckID("ARGO", "sub101", "falsepath/ARGO/subscriptions/sub101:5"))
	suite.Equal(true, validAckID("FOO", "BAR", "projects/FOO/subscriptions/BAR:11155"))
	suite.Equal(false, validAckID("FOO", "BAR", "projects/FOO//subscriptions/BAR:11155"))

}

func (suite *HandlerTestSuite) TestUserCreate() {

	postJSON := `{
	"email":"email@foo.com",
	"projects":[{"project_uuid":"argo_uuid","roles":["admin","viewer"]}]
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/users/USERNEW", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := push.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserCreate, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	usrOut, _ := auth.GetUserFromJSON([]byte(w.Body.String()))

	suite.Equal("USERNEW", usrOut.Name)
	// Check if the mock authenticated userA has been marked as the creator
	suite.Equal("email@foo.com", usrOut.Email)
	//suite.Equal([]string{"admin", "viewer"}, usrOut.Projects[0].Role)
}

func (suite *HandlerTestSuite) TestRefreshToken() {

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/users/UserZ:refreshToken", nil)
	if err != nil {
		log.Fatal(err)
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := push.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/users/{user}:refreshToken", WrapMockAuthConfig(RefreshToken, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	userOut, _ := auth.GetUserFromJSON([]byte(w.Body.String()))
	suite.NotEqual("S3CR3T", userOut.Token)
}

func (suite *HandlerTestSuite) TestUserUpdate() {

	postJSON := `{
	"name":"UPDATED_NAME",
	"service_roles":["service_admin"]
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/users/UserZ", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := push.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserUpdate, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	userOut, _ := auth.GetUserFromJSON([]byte(w.Body.String()))
	suite.Equal("UPDATED_NAME", userOut.Name)
	suite.Equal([]string{"service_admin"}, userOut.ServiceRoles)
	suite.Equal("UserA", userOut.CreatedBy)

}

func (suite *HandlerTestSuite) TestUserListByToken() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users:byToken/S3CR3T1", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "projects": [
      {
         "project": "ARGO",
         "roles": [
            "consumer",
            "publisher"
         ],
         "topics": [
            "topic1",
            "topic2"
         ],
         "subscriptions": [
            "sub1",
            "sub2",
            "sub3"
         ]
      }
   ],
   "name": "UserA",
   "token": "S3CR3T1",
   "email": "foo-email",
   "service_roles": [],
   "created_on": "2009-11-10T23:00:00Z",
   "modified_on": "2009-11-10T23:00:00Z"
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}
	router.HandleFunc("/v1/users:byToken/{token}", WrapMockAuthConfig(UserListByToken, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestUserListByUUID() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users:byUUID/uuid4", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "projects": [
      {
         "project": "ARGO",
         "roles": [
            "publisher",
            "consumer"
         ],
         "topics": [
            "topic2"
         ],
         "subscriptions": [
            "sub3",
            "sub4"
         ]
      }
   ],
   "name": "UserZ",
   "token": "S3CR3T4",
   "email": "foo-email",
   "service_roles": [],
   "created_on": "2009-11-10T23:00:00Z",
   "modified_on": "2009-11-10T23:00:00Z",
   "created_by": "UserA"
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}
	router.HandleFunc("/v1/users:byUUID/{uuid}", WrapMockAuthConfig(UserListByUUID, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestUserListByUUIDNotFound() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users:byUUID/uuid10", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 404,
      "message": "User does not exist",
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
	router.HandleFunc("/v1/users:byUUID/{uuid}", WrapMockAuthConfig(UserListByUUID, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestUserListByUUIDConflict() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users:byUUID/same_uuid", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 500,
      "message": "Multiple users found with the same uuid",
      "status": "INTERNAL"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}
	router.HandleFunc("/v1/users:byUUID/{uuid}", WrapMockAuthConfig(UserListByUUID, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(500, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestUserListOne() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users/UserA", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "projects": [
      {
         "project": "ARGO",
         "roles": [
            "consumer",
            "publisher"
         ],
         "topics": [
            "topic1",
            "topic2"
         ],
         "subscriptions": [
            "sub1",
            "sub2",
            "sub3"
         ]
      }
   ],
   "name": "UserA",
   "token": "S3CR3T1",
   "email": "foo-email",
   "service_roles": [],
   "created_on": "2009-11-10T23:00:00Z",
   "modified_on": "2009-11-10T23:00:00Z"
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserListOne, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestUserListAll() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "users": [
      {
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "consumer",
                  "publisher"
               ],
               "topics": [],
               "subscriptions": []
            }
         ],
         "name": "Test",
         "token": "S3CR3T",
         "email": "Test@test.com",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "consumer",
                  "publisher"
               ],
               "topics": [
                  "topic1",
                  "topic2"
               ],
               "subscriptions": [
                  "sub1",
                  "sub2",
                  "sub3"
               ]
            }
         ],
         "name": "UserA",
         "token": "S3CR3T1",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "consumer",
                  "publisher"
               ],
               "topics": [
                  "topic1",
                  "topic2"
               ],
               "subscriptions": [
                  "sub1",
                  "sub3",
                  "sub4"
               ]
            }
         ],
         "name": "UserB",
         "token": "S3CR3T2",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA"
      },
      {
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "publisher",
                  "consumer"
               ],
               "topics": [
                  "topic3"
               ],
               "subscriptions": [
                  "sub2"
               ]
            }
         ],
         "name": "UserX",
         "token": "S3CR3T3",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA"
      },
      {
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "publisher",
                  "consumer"
               ],
               "topics": [
                  "topic2"
               ],
               "subscriptions": [
                  "sub3",
                  "sub4"
               ]
            }
         ],
         "name": "UserZ",
         "token": "S3CR3T4",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA"
      },
      {
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "publisher",
                  "consumer"
               ],
               "topics": [],
               "subscriptions": []
            }
         ],
         "name": "UserSame1",
         "token": "S3CR3T41",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA"
      },
      {
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "publisher",
                  "consumer"
               ],
               "topics": [],
               "subscriptions": []
            }
         ],
         "name": "UserSame2",
         "token": "S3CR3T42",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA"
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
	router.HandleFunc("/v1/users", WrapMockAuthConfig(UserListAll, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestUserDelete() {

	req, err := http.NewRequest("DELETE", "http://localhost:8080/v1/users/UserA", nil)

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
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserDelete, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

	// Search the deleted user

	req, err = http.NewRequest("GET", "http://localhost:8080/v1/users/UserA", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp2 := `{
   "error": {
      "code": 404,
      "message": "User does not exist",
      "status": "NOT_FOUND"
   }
}`

	router = mux.NewRouter().StrictSlash(true)
	w = httptest.NewRecorder()
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserListOne, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expResp2, w.Body.String())

}

func (suite *HandlerTestSuite) TestProjectDelete() {

	req, err := http.NewRequest("DELETE", "http://localhost:8080/v1/projects/ARGO", nil)

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
	router.HandleFunc("/v1/projects/{project}", WrapMockAuthConfig(ProjectDelete, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestProjectUpdate() {

	postJSON := `{
	"name":"NEWARGO",
	"description":"time to change the description mates and the name"
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/projects/ARGO", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := push.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}", WrapMockAuthConfig(ProjectUpdate, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	projOut, _ := projects.GetFromJSON([]byte(w.Body.String()))
	suite.Equal("NEWARGO", projOut.Name)
	// Check if the mock authenticated userA has been marked as the creator
	suite.Equal("UserA", projOut.CreatedBy)
	suite.Equal("time to change the description mates and the name", projOut.Description)
}

func (suite *HandlerTestSuite) TestProjectCreate() {

	postJSON := `{
	"description":"This is a newly created project"
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGONEW", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := push.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}", WrapMockAuthConfig(ProjectCreate, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	projOut, _ := projects.GetFromJSON([]byte(w.Body.String()))
	suite.Equal("ARGONEW", projOut.Name)
	// Check if the mock authenticated userA has been marked as the creator
	suite.Equal("UserA", projOut.CreatedBy)
	suite.Equal("This is a newly created project", projOut.Description)
}

func (suite *HandlerTestSuite) TestProjectListAll() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "projects": [
      {
         "name": "ARGO",
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA",
         "description": "simple project"
      },
      {
         "name": "ARGO2",
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA",
         "description": "simple project"
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

	router.HandleFunc("/v1/projects", WrapMockAuthConfig(ProjectListAll, cfgKafka, &brk, str, &mgr))

	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestProjectListOneNotFound() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGONAUFTS", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 404,
      "message": "Project does not exist",
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
	router.HandleFunc("/v1/projects/{project}", WrapMockAuthConfig(ProjectListOne, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestProjectListOne() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "name": "ARGO",
   "created_on": "2009-11-10T23:00:00Z",
   "modified_on": "2009-11-10T23:00:00Z",
   "created_by": "UserA",
   "description": "simple project"
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}", WrapMockAuthConfig(ProjectListOne, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestSubModPushConfigError() {

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
	mgr := push.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modifyPushConfig", WrapMockAuthConfig(SubModPush, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestSubModPushConfig() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1",
	"pushConfig": {
		 "pushEndpoint": "https://www.example.com",
		 "retryPolicy": {}
	}
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1:modifyPushConfig", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "name": "/projects/ARGO/subscriptions/sub1",
   "topic": "/projects/ARGO/topics/topic1",
   "pushConfig": {
      "pushEndpoint": "https://www.example.com",
      "retryPolicy": {
         "type": "linear",
         "period": 3000
      }
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
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modifyPushConfig", WrapMockAuthConfig(SubModPush, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal("", w.Body.String())

	// Retrieve the subscription
	req2, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1", nil)
	if err != nil {
		log.Fatal(err)
	}
	router2 := mux.NewRouter().StrictSlash(true)
	w2 := httptest.NewRecorder()
	router2.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubListOne, cfgKafka, &brk, str, &mgr))
	router2.ServeHTTP(w2, req2)
	suite.Equal(200, w2.Code)
	suite.Equal(expResp, w2.Body.String())

}

func (suite *HandlerTestSuite) TestSubCreatePushConfig() {

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
   "name": "/projects/ARGO/subscriptions/subNew",
   "topic": "/projects/ARGO/topics/topic1",
   "pushConfig": {
      "pushEndpoint": "https://www.example.com",
      "retryPolicy": {
         "type": "linear",
         "period": 3000
      }
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
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestSubCreatePushConfigError() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1",
	"pushConfig": {
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
	mgr := push.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expResp, w.Body.String())

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
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr))
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
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr))
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
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr))
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
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubDelete, cfgKafka, &brk, str, &mgr))
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
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubListOne, cfgKafka, &brk, str, &mgr))
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
	router.HandleFunc("/v1/projects/{project}/subscriptions", WrapMockAuthConfig(SubListAll, cfgKafka, &brk, str, &mgr))
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
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapMockAuthConfig(TopicDelete, cfgKafka, &brk, str, &mgr))
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
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubDelete, cfgKafka, &brk, str, &mgr))
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
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapMockAuthConfig(TopicDelete, cfgKafka, &brk, str, &mgr))
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

	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapMockAuthConfig(TopicCreate, cfgKafka, &brk, str, &mgr))
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
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapMockAuthConfig(TopicCreate, cfgKafka, &brk, str, &mgr))
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
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapMockAuthConfig(TopicListOne, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestSubMetrics() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1:metrics", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "metrics": [
      {
         "metric": "subscription.number_of_messages",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "subscription",
         "resource_name": "sub1",
         "timeseries": [
            {
               "timestamp": "{{TS1}}",
               "value": 0
            }
         ],
         "description": "Counter that displays the number of messages consumed from the specific subscription"
      },
      {
         "metric": "subscription.number_of_bytes",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "subscription",
         "resource_name": "sub1",
         "timeseries": [
            {
               "timestamp": "{{TS2}}",
               "value": 0
            }
         ],
         "description": "Counter that displays the total size of data (in bytes) consumed from the specific subscription"
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
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:metrics", WrapMockAuthConfig(SubMetrics, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)

	metricOut, _ := metrics.GetMetricsFromJSON([]byte(w.Body.String()))
	ts1 := metricOut.Metrics[0].Timeseries[0].Timestamp
	ts2 := metricOut.Metrics[1].Timeseries[0].Timestamp
	expResp = strings.Replace(expResp, "{{TS1}}", ts1, -1)
	expResp = strings.Replace(expResp, "{{TS2}}", ts2, -1)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestProjectcMetrics() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO:metrics", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "metrics": [
      {
         "metric": "project.number_of_topics",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project",
         "resource_name": "ARGO",
         "timeseries": [
            {
               "timestamp": "{{TS1}}",
               "value": 3
            }
         ],
         "description": "Counter that displays the number of topics belonging to the specific project"
      },
      {
         "metric": "project.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project",
         "resource_name": "ARGO",
         "timeseries": [
            {
               "timestamp": "{{TS2}}",
               "value": 4
            }
         ],
         "description": "Counter that displays the number of subscriptions belonging to the specific project"
      },
      {
         "metric": "project.user.number_of_topics",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserA",
         "timeseries": [
            {
               "timestamp": "{{TS3}}",
               "value": 2
            }
         ],
         "description": "Counter that displays the number of topics that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_topics",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserB",
         "timeseries": [
            {
               "timestamp": "{{TS4}}",
               "value": 2
            }
         ],
         "description": "Counter that displays the number of topics that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_topics",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserX",
         "timeseries": [
            {
               "timestamp": "{{TS5}}",
               "value": 1
            }
         ],
         "description": "Counter that displays the number of topics that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_topics",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserZ",
         "timeseries": [
            {
               "timestamp": "{{TS6}}",
               "value": 1
            }
         ],
         "description": "Counter that displays the number of topics that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserA",
         "timeseries": [
            {
               "timestamp": "{{TS7}}",
               "value": 3
            }
         ],
         "description": "Counter that displays the number of subscriptions that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserB",
         "timeseries": [
            {
               "timestamp": "{{TS8}}",
               "value": 3
            }
         ],
         "description": "Counter that displays the number of subscriptions that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserX",
         "timeseries": [
            {
               "timestamp": "{{TS9}}",
               "value": 1
            }
         ],
         "description": "Counter that displays the number of subscriptions that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserZ",
         "timeseries": [
            {
               "timestamp": "{{TS10}}",
               "value": 2
            }
         ],
         "description": "Counter that displays the number of subscriptions that a user has access to the specific project"
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
	router.HandleFunc("/v1/projects/{project}:metrics", WrapMockAuthConfig(ProjectMetrics, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	metricOut, _ := metrics.GetMetricsFromJSON([]byte(w.Body.String()))
	ts1 := metricOut.Metrics[0].Timeseries[0].Timestamp
	ts2 := metricOut.Metrics[1].Timeseries[0].Timestamp
	ts3 := metricOut.Metrics[2].Timeseries[0].Timestamp
	ts4 := metricOut.Metrics[3].Timeseries[0].Timestamp
	ts5 := metricOut.Metrics[4].Timeseries[0].Timestamp
	ts6 := metricOut.Metrics[5].Timeseries[0].Timestamp
	ts7 := metricOut.Metrics[6].Timeseries[0].Timestamp
	ts8 := metricOut.Metrics[7].Timeseries[0].Timestamp
	ts9 := metricOut.Metrics[8].Timeseries[0].Timestamp
	ts10 := metricOut.Metrics[9].Timeseries[0].Timestamp
	expResp = strings.Replace(expResp, "{{TS1}}", ts1, -1)
	expResp = strings.Replace(expResp, "{{TS2}}", ts2, -1)
	expResp = strings.Replace(expResp, "{{TS3}}", ts3, -1)
	expResp = strings.Replace(expResp, "{{TS4}}", ts4, -1)
	expResp = strings.Replace(expResp, "{{TS5}}", ts5, -1)
	expResp = strings.Replace(expResp, "{{TS6}}", ts6, -1)
	expResp = strings.Replace(expResp, "{{TS7}}", ts7, -1)
	expResp = strings.Replace(expResp, "{{TS8}}", ts8, -1)
	expResp = strings.Replace(expResp, "{{TS9}}", ts9, -1)
	expResp = strings.Replace(expResp, "{{TS10}}", ts10, -1)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestOpMetrics() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/metrics", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "metrics": [
      {
         "metric": "ams_node.cpu_usage",
         "metric_type": "percentage",
         "value_type": "float64",
         "resource_type": "ams_node",
         "resource_name": "{{HOST}}",
         "timeseries": [
            {
               "timestamp": "{{TS1}}",
               "value": {{VAL1}}
            }
         ],
         "description": "Percentage value that displays the CPU usage of ams service in the specific node"
      },
      {
         "metric": "ams_node.memory_usage",
         "metric_type": "percentage",
         "value_type": "float64",
         "resource_type": "ams_node",
         "resource_name": "{{HOST}}",
         "timeseries": [
            {
               "timestamp": "{{TS1}}",
               "value": {{VAL2}}
            }
         ],
         "description": "Percentage value that displays the Memory usage of ams service in the specific node"
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
	router.HandleFunc("/v1/metrics", WrapMockAuthConfig(OpMetrics, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	metricOut, _ := metrics.GetMetricsFromJSON([]byte(w.Body.String()))
	ts1 := metricOut.Metrics[0].Timeseries[0].Timestamp
	val1 := metricOut.Metrics[0].Timeseries[0].Value.(float64)
	ts2 := metricOut.Metrics[1].Timeseries[0].Timestamp
	val2 := metricOut.Metrics[1].Timeseries[0].Value.(float64)
	host := metricOut.Metrics[0].Resource
	expResp = strings.Replace(expResp, "{{TS1}}", ts1, -1)
	expResp = strings.Replace(expResp, "{{TS2}}", ts2, -1)
	expResp = strings.Replace(expResp, "{{VAL1}}", strconv.FormatFloat(val1, 'g', 1, 64), -1)
	expResp = strings.Replace(expResp, "{{VAL2}}", strconv.FormatFloat(val2, 'g', 1, 64), -1)
	expResp = strings.Replace(expResp, "{{HOST}}", host, -1)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestTopicMetrics() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics/topic1:metrics", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "metrics": [
      {
         "metric": "topic.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "topic",
         "resource_name": "topic1",
         "timeseries": [
            {
               "timestamp": "{{TIMESTAMP1}}",
               "value": 1
            }
         ],
         "description": "Counter that displays the number of subscriptions belonging to a specific topic"
      },
      {
         "metric": "topic.number_of_messages",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "topic",
         "resource_name": "topic1",
         "timeseries": [
            {
               "timestamp": "{{TIMESTAMP2}}",
               "value": 0
            }
         ],
         "description": "Counter that displays the number of messages published to the specific topic"
      },
      {
         "metric": "topic.number_of_bytes",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "topic",
         "resource_name": "topic1",
         "timeseries": [
            {
               "timestamp": "{{TIMESTAMP3}}",
               "value": 0
            }
         ],
         "description": "Counter that displays the total size of data (in bytes) published to the specific topic"
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
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:metrics", WrapMockAuthConfig(TopicMetrics, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	metricOut, _ := metrics.GetMetricsFromJSON([]byte(w.Body.String()))
	ts1 := metricOut.Metrics[0].Timeseries[0].Timestamp
	ts2 := metricOut.Metrics[1].Timeseries[0].Timestamp
	ts3 := metricOut.Metrics[2].Timeseries[0].Timestamp
	expResp = strings.Replace(expResp, "{{TIMESTAMP1}}", ts1, -1)
	expResp = strings.Replace(expResp, "{{TIMESTAMP2}}", ts2, -1)
	expResp = strings.Replace(expResp, "{{TIMESTAMP3}}", ts3, -1)
	suite.Equal(expResp, w.Body.String())

}
func (suite *HandlerTestSuite) TestTopicACL01() {

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
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:acl", WrapMockAuthConfig(TopicACL, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestTopicACL02() {

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
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:acl", WrapMockAuthConfig(TopicACL, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestModTopicACLWrong() {

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
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:modAcl", WrapMockAuthConfig(TopicModACL, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expRes, w.Body.String())

}

func (suite *HandlerTestSuite) TestModSubACLWrong() {

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
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modAcl", WrapMockAuthConfig(SubModACL, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expRes, w.Body.String())

}

func (suite *HandlerTestSuite) TestModTopicACL01() {

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
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:modAcl", WrapMockAuthConfig(TopicModACL, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal("", w.Body.String())

	req2, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics/topic1:acl", nil)
	if err != nil {
		log.Fatal(err)
	}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:acl", WrapMockAuthConfig(TopicACL, cfgKafka, &brk, str, &mgr))
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

func (suite *HandlerTestSuite) TestModSubACL01() {

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
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscription/{subscription}:modAcl", WrapMockAuthConfig(SubModACL, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal("", w.Body.String())

	req2, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscription/sub1:acl", nil)
	if err != nil {
		log.Fatal(err)
	}
	router.HandleFunc("/v1/projects/{project}/subscription/{subscription}:acl", WrapMockAuthConfig(SubACL, cfgKafka, &brk, str, &mgr))
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

func (suite *HandlerTestSuite) TestSubACL01() {

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
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscription/{subscription}:acl", WrapMockAuthConfig(SubACL, cfgKafka, &brk, str, &mgr))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestSubACL02() {

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
	mgr := push.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:acl", WrapMockAuthConfig(SubACL, cfgKafka, &brk, str, &mgr))
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
	router.HandleFunc("/v1/projects/{project}/topics", WrapMockAuthConfig(TopicListAll, cfgKafka, &brk, str, &mgr))
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
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:acknowledge", WrapMockAuthConfig(SubAck, cfgKafka, &brk, str, &mgr))
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
	router2.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:acknowledge", WrapMockAuthConfig(SubAck, cfgKafka, &brk, str, &mgr))
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
	router3.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:acknowledge", WrapMockAuthConfig(SubAck, cfgKafka, &brk, str, &mgr))
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
		router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapValidate(WrapMockAuthConfig(SubListOne, cfgKafka, &brk, str, &mgr)))

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
		router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapValidate(WrapMockAuthConfig(TopicListOne, cfgKafka, &brk, str, &mgr)))

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
