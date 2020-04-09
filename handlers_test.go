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

	log "github.com/sirupsen/logrus"

	"encoding/json"
	"fmt"
	"github.com/ARGOeu/argo-messaging/auth"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/metrics"
	"github.com/ARGOeu/argo-messaging/projects"
	oldPush "github.com/ARGOeu/argo-messaging/push"
	push "github.com/ARGOeu/argo-messaging/push/grpc/client"
	"github.com/ARGOeu/argo-messaging/schemas"
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

func (suite *HandlerTestSuite) TestUserProfile() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users/profile?key=S3CR3T1", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "uuid": "uuid1",
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
   "first_name": "FirstA",
   "last_name": "LastA",
   "organization": "OrgA",
   "description": "DescA",
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/users/profile", WrapMockAuthConfig(UserProfile, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestUserProfileUnauthorized() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users/profile?key=unknonwn", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 401,
      "message": "Unauthorized",
      "status": "UNAUTHORIZED"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}

	// unknown key
	router.HandleFunc("/v1/users/profile", WrapMockAuthConfig(UserProfile, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(401, w.Code)
	suite.Equal(expResp, w.Body.String())

	// empty key
	w2 := httptest.NewRecorder()
	req2, err2 := http.NewRequest("GET", "http://localhost:8080/v1/users/profile", nil)
	if err2 != nil {
		log.Fatal(err2)
	}
	router.HandleFunc("/v1/users/profile", WrapMockAuthConfig(UserProfile, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w2, req2)
	suite.Equal(401, w2.Code)
	suite.Equal(expResp, w2.Body.String())

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
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserCreate, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	usrOut, _ := auth.GetUserFromJSON([]byte(w.Body.String()))

	suite.Equal("USERNEW", usrOut.Name)
	// Check if the mock authenticated userA has been marked as the creator
	suite.Equal("email@foo.com", usrOut.Email)
	//suite.Equal([]string{"admin", "viewer"}, usrOut.Projects[0].Role)
}

func (suite *HandlerTestSuite) TestUserCreateDuplicateRef() {

	postJSON := `{
	"email":"email@foo.com",
	"projects":[{"project":"ARGO","roles":["admin","viewer"]},{"project":"ARGO","roles":["admin","viewer"]}]
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/users/USERNEW", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "error": {
      "code": 400,
      "message": "duplicate reference of project ARGO",
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
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserCreate, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expJSON, w.Body.String())
}

func (suite *HandlerTestSuite) TestUserCreateInvalidServiceRole() {

	postJSON := `{
	"email":"email@foo.com",
	"projects":[{"project":"ARGO","roles":["admin","viewer"]}],
	"service_roles": ["unknown"]
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/users/USERNEW", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "error": {
      "code": 400,
      "message": "invalid role: unknown",
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
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserCreate, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expJSON, w.Body.String())
}

func (suite *HandlerTestSuite) TestUserCreateInvalidProjectName() {

	postJSON := `{
	"email":"email@foo.com",
	"projects":[{"project":"unknown","roles":["admin","viewer"]},{"project":"ARGO","roles":["admin","viewer"]}]
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/users/USERNEW", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "error": {
      "code": 400,
      "message": "invalid project: unknown",
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
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserCreate, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expJSON, w.Body.String())
}

func (suite *HandlerTestSuite) TestUserCreateInvalidRoles() {

	postJSON := `{
	"email":"email@foo.com",
	"projects":[{"project":"ARGO2","roles":["unknown","viewer"]},{"project":"ARGO","roles":["admin","viewer"]}]
}`

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/users/USERNEW", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "error": {
      "code": 400,
      "message": "invalid role: unknown",
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
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserCreate, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expJSON, w.Body.String())
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
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/users/{user}:refreshToken", WrapMockAuthConfig(RefreshToken, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserUpdate, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	userOut, _ := auth.GetUserFromJSON([]byte(w.Body.String()))
	suite.Equal("UPDATED_NAME", userOut.Name)
	suite.Equal([]string{"service_admin"}, userOut.ServiceRoles)
	suite.Equal("UserA", userOut.CreatedBy)

}

func (suite *HandlerTestSuite) TestUserUpdateInvalidProjectName() {

	postJSON := `{
	"name":"UPDATED_NAME",
	"projects": [{"project": "unknown"}],
	"service_roles":["service_admin"]
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/users/UserZ", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "error": {
      "code": 400,
      "message": "invalid project: unknown",
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
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserUpdate, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expJSON, w.Body.String())
}

func (suite *HandlerTestSuite) TestUserUpdateInvalidRoles() {

	postJSON := `{
	"name":"UPDATED_NAME",
	"projects": [{"project": "ARGO2", "roles": ["unknown"]}],
	"service_roles":["service_admin"]
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/users/UserZ", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "error": {
      "code": 400,
      "message": "invalid role: unknown",
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
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserUpdate, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expJSON, w.Body.String())
}

func (suite *HandlerTestSuite) TestUserUpdateInvalidServiceRoles() {

	postJSON := `{
	"name":"UPDATED_NAME",
	"projects": [{"project": "ARGO2", "roles": ["consumer"]}],
	"service_roles":["unknown"]
}`

	req, err := http.NewRequest("PUT", "http://localhost:8080/v1/users/UserZ", bytes.NewBuffer([]byte(postJSON)))
	if err != nil {
		log.Fatal(err)
	}

	expJSON := `{
   "error": {
      "code": 400,
      "message": "invalid role: unknown",
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
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserUpdate, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expJSON, w.Body.String())
}

func (suite *HandlerTestSuite) TestUserUpdateDuplicate() {
	postJSON := `{
		"email":"email@foo.com",
		"projects":[{"project":"ARGO","roles":["admin","viewer"]},{"project":"ARGO2","roles":["admin","viewer"]},{"project":"ARGO2","roles":["admin","viewer"]}]
	}`

	expJSON := `{
   "error": {
      "code": 400,
      "message": "duplicate reference of project ARGO2",
      "status": "INVALID_ARGUMENT"
   }
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
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserUpdate, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *HandlerTestSuite) TestUserListByToken() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users:byToken/S3CR3T1", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "uuid": "uuid1",
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
   "first_name": "FirstA",
   "last_name": "LastA",
   "organization": "OrgA",
   "description": "DescA",
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/users:byToken/{token}", WrapMockAuthConfig(UserListByToken, cfgKafka, &brk, str, &mgr, nil))
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
   "uuid": "uuid4",
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/users:byUUID/{uuid}", WrapMockAuthConfig(UserListByUUID, cfgKafka, &brk, str, &mgr, nil))
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
      "message": "User doesn't exist",
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
	router.HandleFunc("/v1/users:byUUID/{uuid}", WrapMockAuthConfig(UserListByUUID, cfgKafka, &brk, str, &mgr, nil))
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
      "status": "INTERNAL_SERVER_ERROR"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/users:byUUID/{uuid}", WrapMockAuthConfig(UserListByUUID, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(500, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestProjectUserListOne() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/members/UserZ", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "uuid": "uuid4",
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/members/{user}", WrapMockAuthConfig(ProjectUserListOne, cfgKafka, &brk, str, &mgr, nil, "service_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestProjectUserListOneUnpriv() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/members/UserZ", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "uuid": "uuid4",
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/members/{user}", WrapMockAuthConfig(ProjectUserListOne, cfgKafka, &brk, str, &mgr, nil, "project_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestUserListOne() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users/UserA", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "uuid": "uuid1",
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
   "first_name": "FirstA",
   "last_name": "LastA",
   "organization": "OrgA",
   "description": "DescA",
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserListOne, cfgKafka, &brk, str, &mgr, nil))
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
         "uuid": "uuid8",
         "projects": [
            {
               "project": "ARGO2",
               "roles": [
                  "consumer",
                  "publisher"
               ],
               "topics": [],
               "subscriptions": []
            }
         ],
         "name": "UserZ",
         "token": "S3CR3T1",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "uuid": "uuid7",
         "projects": [],
         "name": "push_worker_0",
         "token": "push_token",
         "email": "foo-email",
         "service_roles": [
            "push_worker"
         ],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "uuid": "same_uuid",
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
      },
      {
         "uuid": "same_uuid",
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
         "uuid": "uuid4",
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
         "uuid": "uuid3",
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
         "uuid": "uuid2",
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
         "uuid": "uuid1",
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
         "first_name": "FirstA",
         "last_name": "LastA",
         "organization": "OrgA",
         "description": "DescA",
         "token": "S3CR3T1",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "uuid": "uuid0",
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
      }
   ],
   "nextPageToken": "",
   "totalSize": 9
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/users", WrapMockAuthConfig(UserListAll, cfgKafka, &brk, str, &mgr, nil, "service_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestUserListAllStartingPage() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users?pageSize=2", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "users": [
      {
         "uuid": "uuid8",
         "projects": [
            {
               "project": "ARGO2",
               "roles": [
                  "consumer",
                  "publisher"
               ],
               "topics": [],
               "subscriptions": []
            }
         ],
         "name": "UserZ",
         "token": "S3CR3T1",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "uuid": "uuid7",
         "projects": [],
         "name": "push_worker_0",
         "token": "push_token",
         "email": "foo-email",
         "service_roles": [
            "push_worker"
         ],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      }
   ],
   "nextPageToken": "Ng==",
   "totalSize": 9
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/users", WrapMockAuthConfig(UserListAll, cfgKafka, &brk, str, &mgr, nil, "service_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestUserListAllProjectARGO() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users?project=ARGO", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "users": [
      {
         "uuid": "same_uuid",
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
      },
      {
         "uuid": "same_uuid",
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
         "uuid": "uuid4",
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
         "uuid": "uuid3",
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
         "uuid": "uuid2",
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
         "uuid": "uuid1",
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
         "first_name": "FirstA",
         "last_name": "LastA",
         "organization": "OrgA",
         "description": "DescA",
         "token": "S3CR3T1",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "uuid": "uuid0",
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
      }
   ],
   "nextPageToken": "",
   "totalSize": 7
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/users", WrapMockAuthConfig(UserListAll, cfgKafka, &brk, str, &mgr, nil, "service_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestProjectUserListARGO() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/users", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "users": [
      {
         "uuid": "same_uuid",
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
      },
      {
         "uuid": "same_uuid",
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
         "uuid": "uuid4",
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
         "uuid": "uuid3",
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
         "uuid": "uuid2",
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
         "uuid": "uuid1",
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
         "first_name": "FirstA",
         "last_name": "LastA",
         "organization": "OrgA",
         "description": "DescA",
         "token": "S3CR3T1",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "uuid": "uuid0",
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
      }
   ],
   "nextPageToken": "",
   "totalSize": 7
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/users", WrapMockAuthConfig(ProjectListUsers, cfgKafka, &brk, str, &mgr, nil, "service_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestProjectUserListUnprivARGO() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/members", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "users": [
      {
         "uuid": "same_uuid",
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
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "uuid": "same_uuid",
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
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "uuid": "uuid4",
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
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "uuid": "uuid3",
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
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "uuid": "uuid2",
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
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "uuid": "uuid1",
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
         "first_name": "FirstA",
         "last_name": "LastA",
         "organization": "OrgA",
         "description": "DescA",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "uuid": "uuid0",
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
         "email": "Test@test.com",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      }
   ],
   "nextPageToken": "",
   "totalSize": 7
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/members", WrapMockAuthConfig(ProjectListUsers, cfgKafka, &brk, str, &mgr, nil, "project_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestRegisterUser() {

	type td struct {
		user               string
		postBody           string
		expectedResponse   string
		expectedStatusCode int
		msg                string
	}

	testData := []td{
		{
			user: "new-register-user",
			postBody: `{
							"first_name": "first-name",
							"last_name": "last-name",
							"email": "test@example.com",
							"organization": "org1",
							"description": "desc1"
					   }`,
			expectedResponse: `{
   "uuid": "{{UUID}}",
   "name": "new-register-user",
   "first_name": "first-name",
   "last_name": "last-name",
   "organization": "org1",
   "description": "desc1",
   "email": "test@example.com",
   "status": "pending",
   "activation_token": "{{ATKN}}",
   "registered_at": "{{REAT}}"
}`,
			expectedStatusCode: 200,
			msg:                "User registration successful",
		},
		{
			user: "UserA",
			postBody: `{
							"first_name": "new-name",
							"last_name": "last-name",
							"email": "test@example.com",
							"organization": "org1",
							"description": "desc1"
					   }`,
			expectedResponse: `{
   "error": {
      "code": 409,
      "message": "User already exists",
      "status": "ALREADY_EXISTS"
   }
}`,
			expectedStatusCode: 409,
			msg:                "user already exists",
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
		url := fmt.Sprintf("http://localhost:8080/v1/registrations/%v", t.user)
		req, err := http.NewRequest("POST", url, strings.NewReader(t.postBody))
		if err != nil {
			log.Fatal(err)
		}
		router.HandleFunc("/v1/registrations/{user}", WrapMockAuthConfig(RegisterUser, cfgKafka, &brk, str, &mgr, pc))
		router.ServeHTTP(w, req)
		if t.expectedStatusCode == 200 {
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{UUID}}", str.UserRegistrations[1].UUID, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{REAT}}", str.UserRegistrations[1].RegisteredAt, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{ATKN}}", str.UserRegistrations[1].ActivationToken, 1)

		}
		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}

}

func (suite *HandlerTestSuite) TestAcceptRegisterUser() {

	type td struct {
		registerToken      string
		uname              string
		expectedResponse   string
		expectedStatusCode int
		msg                string
	}

	testData := []td{{
		registerToken: "uratkn-1",
		uname:         "urname",
		expectedResponse: `{
   "uuid": "{{UUID}}",
   "projects": [],
   "name": "urname",
   "first_name": "urfname",
   "last_name": "urlname",
   "organization": "urorg",
   "description": "urdesc",
   "token": "{{TOKEN}}",
   "email": "uremail",
   "service_roles": [],
   "created_on": "{{CON}}",
   "modified_on": "{{MON}}",
   "created_by": "UserA"
}`,
		expectedStatusCode: 200,
		msg:                "Successfully accepted a user's registration",
	},
		{
			registerToken: "unknown",
			uname:         "urname",
			expectedResponse: `{
   "error": {
      "code": 404,
      "message": "User registration doesn't exist",
      "status": "NOT_FOUND"
   }
}`,
			expectedStatusCode: 404,
			msg:                "User registration doesn't exist",
		}}

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
		url := fmt.Sprintf("http://localhost:8080/v1/registrations/%v:accept", t.registerToken)
		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		router.HandleFunc("/v1/registrations/{activation_token}:accept", WrapMockAuthConfig(AcceptRegisterUser, cfgKafka, &brk, str, &mgr, pc))
		router.ServeHTTP(w, req)
		if t.expectedStatusCode == 200 {
			u, _ := auth.FindUsers("", "", t.uname, true, str)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{UUID}}", u.List[0].UUID, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{TOKEN}}", u.List[0].Token, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{CON}}", u.List[0].CreatedOn, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{MON}}", u.List[0].ModifiedOn, 1)
		}
		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}

}

func (suite *HandlerTestSuite) TestProjectUserCreate() {

	type td struct {
		user               string
		postBody           string
		expectedResponse   string
		expectedStatusCode int
		msg                string
	}

	testData := []td{
		{
			user: "member-user",
			postBody: `{
							"email": "test@example.com",
							"service_roles": ["service_admin"],
							"projects": [
											{
												"project": "ARGO",
												"roles": ["project_admin", "publisher", "consumer"]
											},
											{
												"project": "unknown"
											}
										]
					   }`,
			expectedResponse: `{
   "uuid": "{{UUID}}",
   "projects": [
      {
         "project": "ARGO",
         "roles": [
            "project_admin",
            "publisher",
            "consumer"
         ],
         "topics": [],
         "subscriptions": []
      }
   ],
   "name": "member-user",
   "token": "{{TOKEN}}",
   "email": "test@example.com",
   "service_roles": [],
   "created_on": "{{CON}}",
   "modified_on": "{{MON}}",
   "created_by": "UserA"
}`,
			expectedStatusCode: 200,
			msg:                "Create a member of a project(ignore other projects & service roles)",
		},
		{
			user: "member-user-2",
			postBody: `{
							"email": "test@example.com",
							"service_roles": ["service_admin"],
							"projects": []
					   }`,
			expectedResponse: `{
   "uuid": "{{UUID}}",
   "projects": [
      {
         "project": "ARGO",
         "roles": [],
         "topics": [],
         "subscriptions": []
      }
   ],
   "name": "member-user-2",
   "token": "{{TOKEN}}",
   "email": "test@example.com",
   "service_roles": [],
   "created_on": "{{CON}}",
   "modified_on": "{{MON}}",
   "created_by": "UserA"
}`,
			expectedStatusCode: 200,
			msg:                "Create a member/user that automatically gets assigned to the respective project",
		},
		{
			user: "member-user-unknown",
			postBody: `{
							"email": "test@example.com",
							"service_roles": ["service_admin"],
							"projects": [
											{
												"project": "ARGO",
												"roles": ["unknown"]
											},
											{
												"project": "unknown"
											}
										]
					   }`,
			expectedResponse: `{
   "error": {
      "code": 400,
      "message": "invalid role: unknown",
      "status": "INVALID_ARGUMENT"
   }
}`,
			expectedStatusCode: 400,
			msg:                "Invalid user role",
		},
		{
			user: "member-user",
			postBody: `{
							"email": "test@example.com",
							"service_roles": ["service_admin"],
							"projects": [
											{
												"project": "ARGO",
												"roles": ["unknown"]
											},
											{
												"project": "unknown"
											}
										]
					   }`,
			expectedResponse: `{
   "error": {
      "code": 409,
      "message": "User already exists",
      "status": "ALREADY_EXISTS"
   }
}`,
			expectedStatusCode: 409,
			msg:                "user already exists",
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
		url := fmt.Sprintf("http://localhost:8080/v1/projects/ARGO/members/%v", t.user)
		req, err := http.NewRequest("POST", url, strings.NewReader(t.postBody))
		if err != nil {
			log.Fatal(err)
		}
		router.HandleFunc("/v1/projects/{project}/members/{user}", WrapMockAuthConfig(ProjectUserCreate, cfgKafka, &brk, str, &mgr, pc))
		router.ServeHTTP(w, req)
		if t.expectedStatusCode == 200 {
			u, _ := auth.FindUsers("argo_uuid", "", t.user, true, str)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{UUID}}", u.List[0].UUID, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{TOKEN}}", u.List[0].Token, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{CON}}", u.List[0].CreatedOn, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{MON}}", u.List[0].ModifiedOn, 1)
		}
		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}

}

func (suite *HandlerTestSuite) TestProjectUserUpdate() {

	type td struct {
		user               string
		postBody           string
		authRole           string
		expectedResponse   string
		expectedStatusCode int
		msg                string
	}

	testData := []td{
		{
			user: "UserA",
			postBody: `{
							"email": "test@example.com",
							"name": "new-name",
							"service_roles": ["service_admin"],
							"projects": [
											{
												"project": "ARGO",
												"roles": ["project_admin", "publisher"]
											},
											{
												"project": "unknown"
											}
										]
					   }`,
			authRole: "project_admin",
			expectedResponse: `{
   "uuid": "{{UUID}}",
   "projects": [
      {
         "project": "ARGO",
         "roles": [
            "project_admin",
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
   "first_name": "FirstA",
   "last_name": "LastA",
   "organization": "OrgA",
   "description": "DescA",
   "email": "foo-email",
   "service_roles": [],
   "created_on": "{{CON}}",
   "modified_on": "{{MON}}"
}`,
			expectedStatusCode: 200,
			msg:                "Update a member of a project(ignore other projects & service roles & email & name)(project_admin)",
		},
		{
			user: "UserA",
			postBody: `{
							"email": "test@example.com",
							"name": "new-name",
							"service_roles": ["service_admin"],
							"projects": [
											{
												"project": "ARGO",
												"roles": ["project_admin", "publisher"]
											},
											{
												"project": "unknown"
											}
										]
					   }`,
			authRole: "service_admin",
			expectedResponse: `{
   "uuid": "{{UUID}}",
   "projects": [
      {
         "project": "ARGO",
         "roles": [
            "project_admin",
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
   "first_name": "FirstA",
   "last_name": "LastA",
   "organization": "OrgA",
   "description": "DescA",
   "token": "{{TOKEN}}",
   "email": "foo-email",
   "service_roles": [],
   "created_on": "{{CON}}",
   "modified_on": "{{MON}}"
}`,
			expectedStatusCode: 200,
			msg:                "Update a member of a project(ignore other projects & service roles & email & name)(service_admin)",
		},
		{
			user: "UserA",
			postBody: `{
							"email": "test@example.com",
							"service_roles": ["service_admin"],
							"projects": [
											{
												"project": "ARGO",
												"roles": ["unknown"]
											}
										]
					   }`,
			authRole: "project_admin",
			expectedResponse: `{
   "error": {
      "code": 400,
      "message": "invalid role: unknown",
      "status": "INVALID_ARGUMENT"
   }
}`,
			expectedStatusCode: 400,
			msg:                "Invalid user role",
		},
		{
			user: "UserA",
			postBody: `{
							"email": "test@example.com",
							"service_roles": ["service_admin"],
							"projects": [
											{
												"project": "ARGO2",
												"roles": ["publisher"]
											}
										]
					   }`,
			authRole: "project_admin",
			expectedResponse: `{
   "error": {
      "code": 403,
      "message": "Access to this resource is forbidden. User is not a member of the project",
      "status": "FORBIDDEN"
   }
}`,
			expectedStatusCode: 403,
			msg:                "user is not a member of the project",
		},
		{
			user: "unknown",
			postBody: `{
							"email": "test@example.com",
							"service_roles": ["service_admin"],
							"projects": [
											{
												"project": "ARGO",
												"roles": ["publisher"]
											}
										]
					   }`,
			authRole: "project_admin",
			expectedResponse: `{
   "error": {
      "code": 404,
      "message": "User doesn't exist",
      "status": "NOT_FOUND"
   }
}`,
			expectedStatusCode: 404,
			msg: "user doesn't exist" +
				"",
		},
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushEnabled = true
	cfgKafka.PushWorkerToken = "push_token"
	cfgKafka.ResAuth = true
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)

	for _, t := range testData {

		w := httptest.NewRecorder()
		url := fmt.Sprintf("http://localhost:8080/v1/projects/ARGO/members/%v", t.user)
		req, err := http.NewRequest("PUT", url, strings.NewReader(t.postBody))
		if err != nil {
			log.Fatal(err)
		}
		router := mux.NewRouter().StrictSlash(true)
		router.HandleFunc("/v1/projects/{project}/members/{user}", WrapMockAuthConfig(ProjectUserUpdate, cfgKafka, &brk, str, &mgr, pc, t.authRole))
		router.ServeHTTP(w, req)
		if t.expectedStatusCode == 200 {
			u, _ := auth.FindUsers("argo_uuid", "", t.user, true, str)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{UUID}}", u.List[0].UUID, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{TOKEN}}", u.List[0].Token, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{CON}}", u.List[0].CreatedOn, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{MON}}", u.List[0].ModifiedOn, 1)
		}
		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}

}

func (suite *HandlerTestSuite) TestProjectUserRemove() {

	type td struct {
		user               string
		expectedResponse   string
		expectedStatusCode int
		msg                string
	}

	testData := []td{
		{
			user:               "UserA",
			expectedResponse:   `{}`,
			expectedStatusCode: 200,
			msg:                "Remove a member from the project",
		},
		{
			user: "UserA",
			expectedResponse: `{
   "error": {
      "code": 403,
      "message": "Access to this resource is forbidden. User is not a member of the project",
      "status": "FORBIDDEN"
   }
}`,
			expectedStatusCode: 403,
			msg:                "user is not a member of the project",
		},
		{
			user: "unknown",
			expectedResponse: `{
   "error": {
      "code": 404,
      "message": "User doesn't exist",
      "status": "NOT_FOUND"
   }
}`,
			expectedStatusCode: 404,
			msg: "user doesn't exist" +
				"",
		},
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushEnabled = true
	cfgKafka.PushWorkerToken = "push_token"
	cfgKafka.ResAuth = true
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)

	for _, t := range testData {

		w := httptest.NewRecorder()
		url := fmt.Sprintf("http://localhost:8080/v1/projects/ARGO/members/%v:remove", t.user)
		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		router.HandleFunc("/v1/projects/{project}/members/{user}:remove", WrapMockAuthConfig(ProjectUserRemove, cfgKafka, &brk, str, &mgr, pc))
		router.ServeHTTP(w, req)
		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}
}

func (suite *HandlerTestSuite) TestProjectUserAdd() {

	type td struct {
		user               string
		project            string
		authRole           string
		postBody           string
		expectedResponse   string
		expectedStatusCode int
		msg                string
	}

	testData := []td{
		{
			user:    "UserA",
			project: "ARGO2",
			postBody: `{
							"roles": ["unknown"]
						}`,
			expectedResponse: `{
   "error": {
      "code": 400,
      "message": "invalid role: unknown",
      "status": "INVALID_ARGUMENT"
   }
}`,
			expectedStatusCode: 400,
			msg:                "Invalid user role",
		},
		{
			user:    "UserA",
			project: "ARGO2",
			postBody: `{
						"roles": ["project_admin", "publisher", "consumer"]
					   }`,
			authRole: "project_admin",
			expectedResponse: `{
   "uuid": "{{UUID}}",
   "projects": [
      {
         "project": "ARGO2",
         "roles": [
            "project_admin",
            "publisher",
            "consumer"
         ],
         "topics": [],
         "subscriptions": []
      }
   ],
   "name": "UserA",
   "first_name": "FirstA",
   "last_name": "LastA",
   "organization": "OrgA",
   "description": "DescA",
   "email": "foo-email",
   "service_roles": [],
   "created_on": "{{CON}}",
   "modified_on": "{{MON}}"
}`,
			expectedStatusCode: 200,
			msg:                "Add user to project(project_admin)",
		},
		{
			user:    "UserA",
			project: "ARGO2",
			postBody: `{
						"roles": ["project_admin", "consumer", "publisher"]
					   }`,
			authRole: "service_admin",
			expectedResponse: `{
   "uuid": "{{UUID}}",
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
      },
      {
         "project": "ARGO2",
         "roles": [
            "project_admin",
            "consumer",
            "publisher"
         ],
         "topics": [],
         "subscriptions": []
      }
   ],
   "name": "UserA",
   "first_name": "FirstA",
   "last_name": "LastA",
   "organization": "OrgA",
   "description": "DescA",
   "token": "{{TOKEN}}",
   "email": "foo-email",
   "service_roles": [],
   "created_on": "{{CON}}",
   "modified_on": "{{MON}}"
}`,
			expectedStatusCode: 200,
			msg:                "Add user to project(service_admin)",
		},
		{
			user:    "UserA",
			project: "ARGO",
			postBody: `{
							"roles": ["project_admin"]
					   }`,
			expectedResponse: `{
   "error": {
      "code": 409,
      "message": "User is already a member of the project",
      "status": "CONFLICT"
   }
}`,
			expectedStatusCode: 409,
			msg:                "user already member of the project",
		},
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushEnabled = true
	cfgKafka.PushWorkerToken = "push_token"
	cfgKafka.ResAuth = false
	brk := brokers.MockBroker{}
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)

	for _, t := range testData {
		str := stores.NewMockStore("whatever", "argo_mgs")
		w := httptest.NewRecorder()
		url := fmt.Sprintf("http://localhost:8080/v1/projects/%v/members/%v:add", t.project, t.user)
		req, err := http.NewRequest("POST", url, strings.NewReader(t.postBody))
		if err != nil {
			log.Fatal(err)
		}
		router := mux.NewRouter().StrictSlash(true)
		router.HandleFunc("/v1/projects/{project}/members/{user}:add", WrapMockAuthConfig(ProjectUserAdd, cfgKafka, &brk, str, &mgr, pc, t.authRole))
		router.ServeHTTP(w, req)
		if t.expectedStatusCode == 200 {
			u, _ := auth.FindUsers("argo_uuid", "", t.user, true, str)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{UUID}}", u.List[0].UUID, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{TOKEN}}", u.List[0].Token, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{CON}}", u.List[0].CreatedOn, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{MON}}", u.List[0].ModifiedOn, 1)
		}
		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}

}

func (suite *HandlerTestSuite) TestUserListAllProjectARGO2() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users?project=ARGO2", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "users": [
      {
         "uuid": "uuid8",
         "projects": [
            {
               "project": "ARGO2",
               "roles": [
                  "consumer",
                  "publisher"
               ],
               "topics": [],
               "subscriptions": []
            }
         ],
         "name": "UserZ",
         "token": "S3CR3T1",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      }
   ],
   "nextPageToken": "",
   "totalSize": 1
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/users", WrapMockAuthConfig(UserListAll, cfgKafka, &brk, str, &mgr, nil, "service_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestUserListAllProjectUNKNOWN() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users?project=UNKNOWN", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 404,
      "message": "ProjectUUID doesn't exist",
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
	router.HandleFunc("/v1/users", WrapMockAuthConfig(UserListAll, cfgKafka, &brk, str, &mgr, nil, "service_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestUserListAllStartingAtSecond() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users?pageSize=2&pageToken=Nw==", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "users": [
      {
         "uuid": "uuid7",
         "projects": [],
         "name": "push_worker_0",
         "token": "push_token",
         "email": "foo-email",
         "service_roles": [
            "push_worker"
         ],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "uuid": "same_uuid",
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
   ],
   "nextPageToken": "NQ==",
   "totalSize": 9
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/users", WrapMockAuthConfig(UserListAll, cfgKafka, &brk, str, &mgr, nil, "service_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestUserListAllEmptyCollection() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "users": [],
   "nextPageToken": "",
   "totalSize": 0
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	// empty the store
	str.UserList = []stores.QUser{}
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/users", WrapMockAuthConfig(UserListAll, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestUserListAllIntermediatePage() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users?pageToken=NA==&pageSize=2", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "users": [
      {
         "uuid": "uuid4",
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
         "uuid": "uuid3",
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
      }
   ],
   "nextPageToken": "Mg==",
   "totalSize": 9
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)

	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/users", WrapMockAuthConfig(UserListAll, cfgKafka, &brk, str, &mgr, nil, "service_admin"))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestUserListAllInvalidPageSize() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users?pageSize=invalid", nil)
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
	router.HandleFunc("/v1/users", WrapMockAuthConfig(UserListAll, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestUserListAllInvalidPageToken() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/users?pageToken=invalid", nil)
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
	router.HandleFunc("/v1/users", WrapMockAuthConfig(UserListAll, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserDelete, cfgKafka, &brk, str, &mgr, nil))
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
      "message": "User doesn't exist",
      "status": "NOT_FOUND"
   }
}`

	router = mux.NewRouter().StrictSlash(true)
	w = httptest.NewRecorder()
	router.HandleFunc("/v1/users/{user}", WrapMockAuthConfig(UserListOne, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}", WrapMockAuthConfig(ProjectDelete, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}", WrapMockAuthConfig(ProjectUpdate, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}", WrapMockAuthConfig(ProjectCreate, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}

	router.HandleFunc("/v1/projects", WrapMockAuthConfig(ProjectListAll, cfgKafka, &brk, str, &mgr, nil))

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
      "message": "ProjectUUID doesn't exist",
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
	router.HandleFunc("/v1/projects/{project}", WrapMockAuthConfig(ProjectListOne, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}", WrapMockAuthConfig(ProjectListOne, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modifyPushConfig", WrapMockAuthConfig(SubModPush, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(400, w.Code)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestSubModPushInvalidRetPol() {

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
func (suite *HandlerTestSuite) TestSubModPushConfigToActive() {

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
}

// TestSubModPushConfigToInactive tests the use case where the user modifies the push configuration
// in order to deactivate the subscription on the push server
// the push configuration has values before the call and turns into an empty one by the end of the call
func (suite *HandlerTestSuite) TestSubModPushConfigToInactive() {

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
func (suite *HandlerTestSuite) TestSubModPushConfigToInactivePushDisabled() {

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
func (suite *HandlerTestSuite) TestSubModPushConfigToInactiveMissingPushWorker() {

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
func (suite *HandlerTestSuite) TestSubModPushConfigUpdate() {

	postJSON := `{
	"pushConfig": {
		 "pushEndpoint": "https://www.example2.com",
         "maxMessages": 5,
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
}

// TestSubModPushConfigToActiveORUpdatePushDisabled tests the case where the user modifies the push configuration,
// in order to activate the subscription on the push server
// the push enabled config option is set to false
func (suite *HandlerTestSuite) TestSubModPushConfigToActiveORUpdatePushDisabled() {

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
func (suite *HandlerTestSuite) TestSubModPushConfigToActiveORUpdateMissingWorker() {

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

func (suite *HandlerTestSuite) TestVerifyPushEndpoint() {

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

func (suite *HandlerTestSuite) TestVerifyPushEndpointHashMisMatch() {

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

func (suite *HandlerTestSuite) TestVerifyPushEndpointUnknownResponse() {

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
func (suite *HandlerTestSuite) TestVerifyPushEndpointPushServerError() {

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

func (suite *HandlerTestSuite) TestVerifyPushEndpointAlreadyVerified() {

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

func (suite *HandlerTestSuite) TestVerifyPushEndpointNotPushEnabled() {

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/projects/ARGO/subscriptions/push-sub-v1:verifyPushEndpoint", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "error": {
      "code": 409,
      "message": "Subscription is not in push mode",
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
      "maxMessages": 1,
      "retryPolicy": {
         "type": "linear",
         "period": 3000
      },
      "verification_hash": "{{VHASH}}",
      "verified": false
   },
   "ackDeadlineSeconds": 10
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
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestSubCreatePushConfigSlowStart() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1",
	"pushConfig": {
		 "pushEndpoint": "https://www.example.com",
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
      "pushEndpoint": "https://www.example.com",
      "maxMessages": 1,
      "retryPolicy": {
         "type": "slowstart"
      },
      "verification_hash": "{{VHASH}}",
      "verified": false
   },
   "ackDeadlineSeconds": 10
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
	suite.Equal(0, sub.RetPeriod)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestSubCreatePushConfigMissingPushWorker() {

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

func (suite *HandlerTestSuite) TestSubCreatePushConfigPushDisabled() {

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

func (suite *HandlerTestSuite) TestSubCreateInvalidRetPol() {

	postJSON := `{
	"topic":"projects/ARGO/topics/topic1",
	"pushConfig": {
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
	mgr := oldPush.Manager{}
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr, nil))
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
      "maxMessages": 0,
      "retryPolicy": {},
      "verification_hash": "",
      "verified": false
   },
   "ackDeadlineSeconds": 10
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubCreate, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", WrapMockAuthConfig(SubDelete, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestSubWithPushConfigDelete() {

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

func (suite *HandlerTestSuite) TestSubWithPushConfigDeletePushServerError() {

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
      "maxMessages": 0,
      "retryPolicy": {},
      "verification_hash": "",
      "verified": false
   },
   "ackDeadlineSeconds": 10
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

func (suite *HandlerTestSuite) TestSubListAll() {

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
            "pushEndpoint": "endpoint.foo",
            "maxMessages": 1,
            "retryPolicy": {
               "type": "linear",
               "period": 300
            },
            "verification_hash": "push-id-1",
            "verified": true
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub3",
         "topic": "/projects/ARGO/topics/topic3",
         "pushConfig": {
            "pushEndpoint": "",
            "maxMessages": 0,
            "retryPolicy": {},
            "verification_hash": "",
            "verified": false
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub2",
         "topic": "/projects/ARGO/topics/topic2",
         "pushConfig": {
            "pushEndpoint": "",
            "maxMessages": 0,
            "retryPolicy": {},
            "verification_hash": "",
            "verified": false
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub1",
         "topic": "/projects/ARGO/topics/topic1",
         "pushConfig": {
            "pushEndpoint": "",
            "maxMessages": 0,
            "retryPolicy": {},
            "verification_hash": "",
            "verified": false
         },
         "ackDeadlineSeconds": 10
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

func (suite *HandlerTestSuite) TestSubListAllFirstPage() {

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
            "pushEndpoint": "endpoint.foo",
            "maxMessages": 1,
            "retryPolicy": {
               "type": "linear",
               "period": 300
            },
            "verification_hash": "push-id-1",
            "verified": true
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub3",
         "topic": "/projects/ARGO/topics/topic3",
         "pushConfig": {
            "pushEndpoint": "",
            "maxMessages": 0,
            "retryPolicy": {},
            "verification_hash": "",
            "verified": false
         },
         "ackDeadlineSeconds": 10
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

func (suite *HandlerTestSuite) TestSubListAllNextPage() {

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
            "pushEndpoint": "",
            "maxMessages": 0,
            "retryPolicy": {},
            "verification_hash": "",
            "verified": false
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub1",
         "topic": "/projects/ARGO/topics/topic1",
         "pushConfig": {
            "pushEndpoint": "",
            "maxMessages": 0,
            "retryPolicy": {},
            "verification_hash": "",
            "verified": false
         },
         "ackDeadlineSeconds": 10
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

func (suite *HandlerTestSuite) TestSubListAllEmpty() {

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

func (suite *HandlerTestSuite) TestSubListAllConsumer() {

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
            "pushEndpoint": "endpoint.foo",
            "maxMessages": 1,
            "retryPolicy": {
               "type": "linear",
               "period": 300
            },
            "verification_hash": "push-id-1",
            "verified": true
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub3",
         "topic": "/projects/ARGO/topics/topic3",
         "pushConfig": {
            "pushEndpoint": "",
            "maxMessages": 0,
            "retryPolicy": {},
            "verification_hash": "",
            "verified": false
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub2",
         "topic": "/projects/ARGO/topics/topic2",
         "pushConfig": {
            "pushEndpoint": "",
            "maxMessages": 0,
            "retryPolicy": {},
            "verification_hash": "",
            "verified": false
         },
         "ackDeadlineSeconds": 10
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

func (suite *HandlerTestSuite) TestSubListAllConsumerWithPagination() {

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
            "pushEndpoint": "endpoint.foo",
            "maxMessages": 1,
            "retryPolicy": {
               "type": "linear",
               "period": 300
            },
            "verification_hash": "push-id-1",
            "verified": true
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub3",
         "topic": "/projects/ARGO/topics/topic3",
         "pushConfig": {
            "pushEndpoint": "",
            "maxMessages": 0,
            "retryPolicy": {},
            "verification_hash": "",
            "verified": false
         },
         "ackDeadlineSeconds": 10
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

func (suite *HandlerTestSuite) TestSubListAllInvalidPageSize() {

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

func (suite *HandlerTestSuite) TestSubListAllInvalidPageToken() {

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

func (suite *HandlerTestSuite) TestTopicDelete() {

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
func (suite *HandlerTestSuite) TestSubTimeToOffset() {

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

func (suite *HandlerTestSuite) TestSubTimeToOffsetOutOfBounds() {

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

func (suite *HandlerTestSuite) TestSubDeleteNotFound() {

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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapMockAuthConfig(TopicDelete, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}

	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapMockAuthConfig(TopicCreate, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapMockAuthConfig(TopicCreate, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}", WrapMockAuthConfig(TopicListOne, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestTopicListSubscriptions() {

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

func (suite *HandlerTestSuite) TestTopicListSubscriptionsEmpty() {

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

func (suite *HandlerTestSuite) TestProjectMessageCount() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/metrics/daily-message-average?start_date=2018-10-01&end_date=2018-10-04", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
 "projects": [
  {
   "project": "ARGO",
   "message_count": 30,
   "average_daily_messages": 10
  }
 ],
 "total_message_count": 30,
 "average_daily_messages": 10
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/metrics/daily-message-average", WrapMockAuthConfig(DailyMessageAverage, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestProjectMessageCountErrors() {

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects-message-count", WrapMockAuthConfig(DailyMessageAverage, cfgKafka, &brk, str, &mgr, nil))

	// wrong start date
	expResp1 := `{
   "error": {
      "code": 400,
      "message": "Start date is not in valid format",
      "status": "INVALID_ARGUMENT"
   }
}`
	req1, err := http.NewRequest("GET", "http://localhost:8080/v1/projects-message-count?start_date=ffff", nil)
	if err != nil {
		log.Fatal(err)
	}
	router.ServeHTTP(w, req1)
	suite.Equal(400, w.Code)
	suite.Equal(expResp1, w.Body.String())
	w.Body.Reset()

	// wrong end date
	expResp2 := `{
   "error": {
      "code": 400,
      "message": "End date is not in valid format",
      "status": "INVALID_ARGUMENT"
   }
}`
	req2, err := http.NewRequest("GET", "http://localhost:8080/v1/projects-message-count?end_date=ffff", nil)
	if err != nil {
		log.Fatal(err)
	}
	router.ServeHTTP(w, req2)
	suite.Equal(400, w.Code)
	suite.Equal(expResp2, w.Body.String())
	w.Body.Reset()

	// one of the projects doesn't exist end date
	expResp3 := `{
   "error": {
      "code": 404,
      "message": "Project ffff doesn't exist",
      "status": "NOT_FOUND"
   }
}`
	req3, err := http.NewRequest("GET", "http://localhost:8080/v1/projects-message-count?projects=ARGO,ffff", nil)
	if err != nil {
		log.Fatal(err)
	}
	router.ServeHTTP(w, req3)
	suite.Equal(400, w.Code)
	suite.Equal(expResp3, w.Body.String())
	w.Body.Reset()

	// start date is off
	expResp4 := `{
   "error": {
      "code": 400,
      "message": "Start date cannot be after the end date",
      "status": "INVALID_ARGUMENT"
   }
}`
	req4, err := http.NewRequest("GET", "http://localhost:8080/v1/projects-message-count?start_date=2019-04-04&end_date=2018-01-01", nil)
	if err != nil {
		log.Fatal(err)
	}
	router.ServeHTTP(w, req4)
	suite.Equal(400, w.Code)
	suite.Equal(expResp4, w.Body.String())
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
      },
      {
         "metric": "subscription.consumption_rate",
         "metric_type": "rate",
         "value_type": "float64",
         "resource_type": "subscription",
         "resource_name": "sub1",
         "timeseries": [
            {
               "timestamp": "2019-05-06T00:00:00Z",
               "value": 10
            }
         ],
         "description": "A rate that displays how many messages were consumed per second between the last two consume events"
      }
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:metrics", WrapMockAuthConfig(SubMetrics, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)

	metricOut, _ := metrics.GetMetricsFromJSON([]byte(w.Body.String()))
	ts1 := metricOut.Metrics[0].Timeseries[0].Timestamp
	ts2 := metricOut.Metrics[1].Timeseries[0].Timestamp
	expResp = strings.Replace(expResp, "{{TS1}}", ts1, -1)
	expResp = strings.Replace(expResp, "{{TS2}}", ts2, -1)
	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestSubMetricsNotFound() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions/unknown_sub:metrics", nil)
	if err != nil {
		log.Fatal(err)
	}

	expRes := `{
   "error": {
      "code": 404,
      "message": "Subscription doesn't exist",
      "status": "NOT_FOUND"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	// temporarily disable auth for this test case
	cfgKafka.ResAuth = false
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:metrics", WrapMockAuthConfig(SubMetrics, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expRes, w.Body.String())

}

func (suite *HandlerTestSuite) TestProjectMetrics() {

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
               "value": 4
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
      },
      {
         "metric": "project.number_of_daily_messages",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project",
         "resource_name": "ARGO",
         "timeseries": [
            {
               "timestamp": "{{TS11}}",
               "value": 30
            },
            {
               "timestamp": "{{TS12}}",
               "value": 110
            }
         ],
         "description": "A collection of counters that represents the total number of messages published each day to all of the project's topics"
      }
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}:metrics", WrapMockAuthConfig(ProjectMetrics, cfgKafka, &brk, str, &mgr, nil))
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
	ts11 := metricOut.Metrics[10].Timeseries[0].Timestamp
	ts12 := metricOut.Metrics[10].Timeseries[1].Timestamp
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
	expResp = strings.Replace(expResp, "{{TS11}}", ts11, -1)
	expResp = strings.Replace(expResp, "{{TS12}}", ts12, -1)
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/metrics", WrapMockAuthConfig(OpMetrics, cfgKafka, &brk, str, &mgr, nil))
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
      },
      {
         "metric": "topic.number_of_daily_messages",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "topic",
         "resource_name": "topic1",
         "timeseries": [
            {
               "timestamp": "{{TIMESTAMP4}}",
               "value": 30
            },
            {
               "timestamp": "{{TIMESTAMP5}}",
               "value": 40
            }
         ],
         "description": "A collection of counters that represents the total number of messages published each day to a specific topic"
      },
      {
         "metric": "topic.publishing_rate",
         "metric_type": "rate",
         "value_type": "float64",
         "resource_type": "topic",
         "resource_name": "topic1",
         "timeseries": [
            {
               "timestamp": "2019-05-06T00:00:00Z",
               "value": 10
            }
         ],
         "description": "A rate that displays how many messages were published per second between the last two publish events"
      }
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:metrics", WrapMockAuthConfig(TopicMetrics, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	metricOut, _ := metrics.GetMetricsFromJSON([]byte(w.Body.String()))
	ts1 := metricOut.Metrics[0].Timeseries[0].Timestamp
	ts2 := metricOut.Metrics[1].Timeseries[0].Timestamp
	ts3 := metricOut.Metrics[2].Timeseries[0].Timestamp
	ts4 := metricOut.Metrics[3].Timeseries[0].Timestamp
	ts5 := metricOut.Metrics[3].Timeseries[1].Timestamp
	expResp = strings.Replace(expResp, "{{TIMESTAMP1}}", ts1, -1)
	expResp = strings.Replace(expResp, "{{TIMESTAMP2}}", ts2, -1)
	expResp = strings.Replace(expResp, "{{TIMESTAMP3}}", ts3, -1)
	expResp = strings.Replace(expResp, "{{TIMESTAMP4}}", ts4, -1)
	expResp = strings.Replace(expResp, "{{TIMESTAMP5}}", ts5, -1)

	suite.Equal(expResp, w.Body.String())

}

func (suite *HandlerTestSuite) TestTopicMetricsNotFound() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics/topic_not_found:metrics", nil)
	if err != nil {
		log.Fatal(err)
	}

	expRes := `{
   "error": {
      "code": 404,
      "message": "Topic doesn't exist",
      "status": "NOT_FOUND"
   }
}`
	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	// deactivate auth for this specific test case
	cfgKafka.ResAuth = false
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:metrics", WrapMockAuthConfig(TopicMetrics, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expRes, w.Body.String())

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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:acl", WrapMockAuthConfig(TopicACL, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:acl", WrapMockAuthConfig(TopicACL, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:modAcl", WrapMockAuthConfig(TopicModACL, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:modAcl", WrapMockAuthConfig(SubModACL, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscription/{subscription}:acl", WrapMockAuthConfig(SubACL, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:acl", WrapMockAuthConfig(SubACL, cfgKafka, &brk, str, &mgr, nil))
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

func (suite *HandlerTestSuite) TestTopicListAllPublisher() {

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

func (suite *HandlerTestSuite) TestTopicListAllPublisherWithPagination() {

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

func (suite *HandlerTestSuite) TestPublishWithSchema() {

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

func (suite *HandlerTestSuite) TestTopicListAllFirstPage() {

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

func (suite *HandlerTestSuite) TestTopicListAllNextPage() {

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

func (suite *HandlerTestSuite) TestTopicListAllEmpty() {

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

func (suite *HandlerTestSuite) TestTopicListAllInvalidPageSize() {

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

func (suite *HandlerTestSuite) TestTopicListAllInvalidPageToken() {

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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapMockAuthConfig(TopicPublish, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapMockAuthConfig(TopicPublish, cfgKafka, &brk, str, &mgr, nil))
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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:publish", WrapMockAuthConfig(TopicPublish, cfgKafka, &brk, str, &mgr, nil))
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

func (suite *HandlerTestSuite) TestSubPullFromPushEnabledAsPushWorker() {

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

func (suite *HandlerTestSuite) TestSubPullFromPushEnabledAsPushWorkerDISABLED() {

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

func (suite *HandlerTestSuite) TestSubPullFromPushEnabledAsServiceAdmin() {

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

func (suite *HandlerTestSuite) TestSubPullFromPushEnabledNoPushWorker() {

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

func (suite *HandlerTestSuite) TestSubModAck() {

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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:pull", WrapMockAuthConfig(SubPull, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expJSON, w.Body.String())

}

func (suite *HandlerTestSuite) TestSubNoTopic() {

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
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:pull", WrapMockAuthConfig(SubPull, cfgKafka, &brk, str, &mgr, nil))
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
      "maxMessages": 0,
      "retryPolicy": {},
      "verification_hash": "",
      "verified": false
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

func (suite *HandlerTestSuite) TestSchemaCreate() {

	type td struct {
		postBody           string
		expectedResponse   string
		schemaName         string
		expectedStatusCode int
		msg                string
	}

	testData := []td{
		{
			postBody: `{
	"type": "json",
	"schema":{
  			"type": "string"
	}
}`,
			schemaName:         "new-schema",
			expectedStatusCode: 200,
			expectedResponse: `{
 "uuid": "{{UUID}}",
 "name": "projects/ARGO/schemas/new-schema",
 "type": "json",
 "schema": {
  "type": "string"
 }
}`,
			msg: "Case where the schema is valid and successfully created(JSON)",
		},
		{
			postBody: `{
	"type": "avro",
	"schema":{
  			"type": "record",
 			"namespace": "user.avro",
			"name":"User",
			"fields": [
						{"name": "username", "type": "string"},
						{"name": "phone", "type": "int"}
			]
	}
}`,
			schemaName:         "new-schema-avro",
			expectedStatusCode: 200,
			expectedResponse: `{
 "uuid": "{{UUID}}",
 "name": "projects/ARGO/schemas/new-schema-avro",
 "type": "avro",
 "schema": {
  "fields": [
   {
    "name": "username",
    "type": "string"
   },
   {
    "name": "phone",
    "type": "int"
   }
  ],
  "name": "User",
  "namespace": "user.avro",
  "type": "record"
 }
}`,
			msg: "Case where the schema is valid and successfully created(AVRO)",
		},
		{
			postBody: `{
	"type": "unknown",
	"schema":{
  			"type": "string"
	}
}`,
			schemaName:         "new-schema-2",
			expectedStatusCode: 400,
			expectedResponse: `{
   "error": {
      "code": 400,
      "message": "Schema type can only be 'json' or 'avro'",
      "status": "INVALID_ARGUMENT"
   }
}`,
			msg: "Case where the schema type is unsupported",
		},
		{
			postBody: `{
	"type": "json",
	"schema":{
  			"type": "unknown"
	}
}`,
			schemaName:         "new-schema-2",
			expectedStatusCode: 400,
			expectedResponse: `{
   "error": {
      "code": 400,
      "message": "has a primitive type that is NOT VALID -- given: /unknown/ Expected valid values are:[array boolean integer number null object string]",
      "status": "INVALID_ARGUMENT"
   }
}`,
			msg: "Case where the json schema is not valid",
		},
		{
			postBody: `{
	"type": "avro",
	"schema":{
  			"type": "unknown"
	}
}`,
			schemaName:         "new-schema-2",
			expectedStatusCode: 400,
			expectedResponse: `{
   "error": {
      "code": 400,
      "message": "unknown type name: \"unknown\"",
      "status": "INVALID_ARGUMENT"
   }
}`,
			msg: "Case where the avro schema is not valid",
		},
		{
			postBody: `{
	"type": "json",
	"schema":{
  			"type": "string"
	}
}`,
			schemaName:         "schema-1",
			expectedStatusCode: 409,
			expectedResponse: `{
   "error": {
      "code": 409,
      "message": "Schema already exists",
      "status": "ALREADY_EXISTS"
   }
}`,
			msg: "Case where the json schema name already exists",
		},
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushEnabled = true
	cfgKafka.PushWorkerToken = "push_token"
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)

	for _, t := range testData {

		w := httptest.NewRecorder()
		url := fmt.Sprintf("http://localhost:8080/v1/projects/ARGO/schemas/%v", t.schemaName)
		req, err := http.NewRequest("POST", url, strings.NewReader(t.postBody))
		if err != nil {
			log.Fatal(err)
		}
		router.HandleFunc("/v1/projects/{project}/schemas/{schema}", WrapMockAuthConfig(SchemaCreate, cfgKafka, &brk, str, &mgr, pc))
		router.ServeHTTP(w, req)

		if t.expectedStatusCode == 200 {
			s := schemas.Schema{}
			json.Unmarshal(w.Body.Bytes(), &s)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{UUID}}", s.UUID, 1)
		}

		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}

}

func (suite *HandlerTestSuite) TestSchemaListOne() {

	type td struct {
		expectedResponse   string
		schemaName         string
		expectedStatusCode int
		msg                string
	}

	testData := []td{
		{
			schemaName:         "schema-1",
			expectedStatusCode: 200,
			expectedResponse: `{
 "uuid": "schema_uuid_1",
 "name": "projects/ARGO/schemas/schema-1",
 "type": "json",
 "schema": {
  "properties": {
   "address": {
    "type": "string"
   },
   "email": {
    "type": "string"
   },
   "name": {
    "type": "string"
   },
   "telephone": {
    "type": "string"
   }
  },
  "required": [
   "name",
   "email"
  ],
  "type": "object"
 }
}`,
			msg: "Case where a specific schema is retrieved successfully",
		},
		{
			schemaName:         "unknown",
			expectedStatusCode: 404,
			expectedResponse: `{
   "error": {
      "code": 404,
      "message": "Schema doesn't exist",
      "status": "NOT_FOUND"
   }
}`,
			msg: "Case where the requested schema doesn't exist",
		},
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushEnabled = true
	cfgKafka.PushWorkerToken = "push_token"
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)

	for _, t := range testData {

		w := httptest.NewRecorder()
		url := fmt.Sprintf("http://localhost:8080/v1/projects/ARGO/schemas/%v", t.schemaName)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		router.HandleFunc("/v1/projects/{project}/schemas/{schema}", WrapMockAuthConfig(SchemaListOne, cfgKafka, &brk, str, &mgr, pc))
		router.ServeHTTP(w, req)

		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}
}

func (suite *HandlerTestSuite) TestSchemaListAll() {

	type td struct {
		expectedResponse   string
		projectName        string
		expectedStatusCode int
		msg                string
	}

	testData := []td{
		{
			projectName:        "ARGO",
			expectedStatusCode: 200,
			expectedResponse: `{
 "schemas": [
  {
   "uuid": "schema_uuid_1",
   "name": "projects/ARGO/schemas/schema-1",
   "type": "json",
   "schema": {
    "properties": {
     "address": {
      "type": "string"
     },
     "email": {
      "type": "string"
     },
     "name": {
      "type": "string"
     },
     "telephone": {
      "type": "string"
     }
    },
    "required": [
     "name",
     "email"
    ],
    "type": "object"
   }
  },
  {
   "uuid": "schema_uuid_2",
   "name": "projects/ARGO/schemas/schema-2",
   "type": "json",
   "schema": {
    "properties": {
     "address": {
      "type": "string"
     },
     "email": {
      "type": "string"
     },
     "name": {
      "type": "string"
     },
     "telephone": {
      "type": "string"
     }
    },
    "required": [
     "name",
     "email"
    ],
    "type": "object"
   }
  },
  {
   "uuid": "schema_uuid_3",
   "name": "projects/ARGO/schemas/schema-3",
   "type": "avro",
   "schema": {
    "fields": [
     {
      "name": "username",
      "type": "string"
     },
     {
      "name": "phone",
      "type": "int"
     }
    ],
    "name": "User",
    "namespace": "user.avro",
    "type": "record"
   }
  }
 ]
}`,
			msg: "Case where the schemas under a project are successfully retrieved",
		},
		{
			projectName:        "ARGO2",
			expectedStatusCode: 200,
			expectedResponse: `{
 "schemas": []
}`,
			msg: "Case where the given project has no schemas",
		},
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushEnabled = true
	cfgKafka.PushWorkerToken = "push_token"
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)

	for _, t := range testData {

		w := httptest.NewRecorder()
		url := fmt.Sprintf("http://localhost:8080/v1/projects/%s/schemas", t.projectName)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		router.HandleFunc("/v1/projects/{project}/schemas", WrapMockAuthConfig(SchemaListAll, cfgKafka, &brk, str, &mgr, pc))
		router.ServeHTTP(w, req)

		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}
}

func (suite *HandlerTestSuite) TestSchemaUpdate() {

	type td struct {
		postBody           string
		expectedResponse   string
		schemaName         string
		expectedStatusCode int
		msg                string
	}

	testData := []td{
		{
			schemaName:         "schema-2",
			postBody:           `{"name": "projects/ARGO/schemas/schema-1"}`,
			expectedStatusCode: 409,
			expectedResponse: `{
   "error": {
      "code": 409,
      "message": "Schema already exists",
      "status": "ALREADY_EXISTS"
   }
}`,
			msg: "Case where the requested schema wants to update the name field to an already existing one",
		},
		{
			schemaName:         "schema-1",
			postBody:           `{"type":"unsupported"}`,
			expectedStatusCode: 400,
			expectedResponse: `{
   "error": {
      "code": 400,
      "message": "Schema type can only be 'json' or 'avro'",
      "status": "INVALID_ARGUMENT"
   }
}`,
			msg: "Case where the requested schema wants to update its type field to an unsupported option",
		},
		{
			schemaName:         "schema-1",
			postBody:           `{"schema":{"type":"unknown"}}`,
			expectedStatusCode: 400,
			expectedResponse: `{
   "error": {
      "code": 400,
      "message": "has a primitive type that is NOT VALID -- given: /unknown/ Expected valid values are:[array boolean integer number null object string]",
      "status": "INVALID_ARGUMENT"
   }
}`,
			msg: "Case where the requested schema wants to update its schema with invalid contents",
		},
		{
			schemaName:         "schema-1",
			expectedStatusCode: 200,
			expectedResponse: `{
 "uuid": "schema_uuid_1",
 "name": "projects/ARGO/schemas/new-name",
 "type": "json",
 "schema": {
  "properties": {
   "address": {
    "type": "string"
   },
   "email": {
    "type": "string"
   },
   "name": {
    "type": "string"
   },
   "telephone": {
    "type": "string"
   }
  },
  "required": [
   "name",
   "email",
   "address"
  ],
  "type": "object"
 }
}`,
			postBody: `{
 "name": "projects/ARGO/schemas/new-name",
 "type": "json",
 "schema": {
  "properties": {
   "address": {
    "type": "string"
   },
   "email": {
    "type": "string"
   },
   "name": {
    "type": "string"
   },
   "telephone": {
    "type": "string"
   }
  },
  "required": [
   "name",
   "email",
   "address"
  ],
  "type": "object"
 }
}`,

			msg: "Case where a specific schema has all its fields updated successfully",
		},
		{
			schemaName:         "unknown",
			postBody:           "",
			expectedStatusCode: 404,
			expectedResponse: `{
   "error": {
      "code": 404,
      "message": "Schema doesn't exist",
      "status": "NOT_FOUND"
   }
}`,
			msg: "Case where the requested schema doesn't exist",
		},
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushEnabled = true
	cfgKafka.PushWorkerToken = "push_token"
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)

	for _, t := range testData {

		w := httptest.NewRecorder()
		url := fmt.Sprintf("http://localhost:8080/v1/projects/ARGO/schemas/%v", t.schemaName)
		req, err := http.NewRequest("PUT", url, strings.NewReader(t.postBody))
		if err != nil {
			log.Fatal(err)
		}
		router.HandleFunc("/v1/projects/{project}/schemas/{schema}", WrapMockAuthConfig(SchemaUpdate, cfgKafka, &brk, str, &mgr, pc))
		router.ServeHTTP(w, req)

		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}
}

func (suite *HandlerTestSuite) TestSchemaDelete() {

	type td struct {
		expectedResponse   string
		schemaName         string
		expectedStatusCode int
		msg                string
	}

	testData := []td{
		{
			expectedResponse:   "",
			schemaName:         "schema-1",
			expectedStatusCode: 200,
			msg:                "Case where the schema is successfully deleted",
		},
		{
			schemaName:         "unknown",
			expectedStatusCode: 404,
			expectedResponse: `{
   "error": {
      "code": 404,
      "message": "Schema doesn't exist",
      "status": "NOT_FOUND"
   }
}`,
			msg: "Case where the requested schema doesn't exist",
		},
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushEnabled = true
	cfgKafka.PushWorkerToken = "push_token"
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)

	for _, t := range testData {

		w := httptest.NewRecorder()
		url := fmt.Sprintf("http://localhost:8080/v1/projects/ARGO/schemas/%v", t.schemaName)
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		router.HandleFunc("/v1/projects/{project}/schemas/{schema}", WrapMockAuthConfig(SchemaDelete, cfgKafka, &brk, str, &mgr, pc))
		router.ServeHTTP(w, req)

		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}
}

func (suite *HandlerTestSuite) TestSchemaValidateMessage() {

	type td struct {
		expectedResponse   string
		postBody           map[string]interface{}
		schemaName         string
		expectedStatusCode int
		msg                string
	}

	testData := []td{
		{
			expectedResponse: `{
 "message": "Message validated successfully"
}`,
			postBody: map[string]interface{}{
				"name":  "name1",
				"email": "email1",
			},
			schemaName:         "schema-1",
			expectedStatusCode: 200,
			msg:                "Case where the message is successfully validated(JSON)",
		},
		{
			expectedResponse: `{
 "message": "Message validated successfully"
}`,
			postBody: map[string]interface{}{
				"data": "DGFnZWxvc8T8Cg==",
			},
			schemaName:         "schema-3",
			expectedStatusCode: 200,
			msg:                "Case where the message is successfully validated(AVRO)",
		},
		{
			postBody: map[string]interface{}{
				"name": "name1",
			},
			schemaName:         "schema-1",
			expectedStatusCode: 400,
			msg:                "Case where the message is not valid(omit required email field)(JSON)",
			expectedResponse: `{
   "error": {
      "code": 400,
      "message": "Message 0 data is not valid,(root): email is required",
      "status": "INVALID_ARGUMENT"
   }
}`,
		},
		{
			postBody: map[string]interface{}{
				"data": "T2JqAQQWYXZyby5zY2hlbWGYAnsidHlwZSI6InJlY29yZCIsIm5hbWUiOiJQbGFjZSIsIm5hbWVzcGFjZSI6InBsYWNlLmF2cm8iLCJmaWVsZHMiOlt7Im5hbWUiOiJwbGFjZW5hbWUiLCJ0eXBlIjoic3RyaW5nIn0seyJuYW1lIjoiYWRkcmVzcyIsInR5cGUiOiJzdHJpbmcifV19FGF2cm8uY29kZWMIbnVsbABM1P4b0GpYaCg9tqxa+YDZAiQSc3RyZWV0IDIyDnBsYWNlIGFM1P4b0GpYaCg9tqxa+YDZ",
			},
			schemaName:         "schema-3",
			expectedStatusCode: 400,
			msg:                "Case where the message is not valid(AVRO)",
			expectedResponse: `{
   "error": {
      "code": 400,
      "message": "Message 0 is not valid.cannot decode binary record \"user.avro.User\" field \"username\": cannot decode binary string: cannot decode binary bytes: negative size: -40",
      "status": "INVALID_ARGUMENT"
   }
}`,
		},
		{
			postBody: map[string]interface{}{
				"data": "DGFnZWxvc8T8Cg",
			},
			schemaName:         "schema-3",
			expectedStatusCode: 400,
			msg:                "Case where the message is not in valid base64(AVRO)",
			expectedResponse: `{
   "error": {
      "code": 400,
      "message": "Message 0 is not in valid base64 enocding,illegal base64 data at input byte 12",
      "status": "INVALID_ARGUMENT"
   }
}`,
		},
		{
			postBody: map[string]interface{}{
				"unknown": "unknown",
			},
			schemaName:         "schema-3",
			expectedStatusCode: 400,
			msg:                "Case where the request arguments are missing the required data field(AVRO)",
			expectedResponse: `{
   "error": {
      "code": 400,
      "message": "Invalid Schema Payload Arguments",
      "status": "INVALID_ARGUMENT"
   }
}`,
		},
		{
			schemaName:         "unknown",
			expectedStatusCode: 404,
			expectedResponse: `{
   "error": {
      "code": 404,
      "message": "Schema doesn't exist",
      "status": "NOT_FOUND"
   }
}`,
			msg: "Case where the schema doesn't exist",
		},
	}

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	cfgKafka.PushEnabled = true
	cfgKafka.PushWorkerToken = "push_token"
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	mgr := oldPush.Manager{}
	pc := new(push.MockClient)

	for _, t := range testData {

		w := httptest.NewRecorder()

		url := fmt.Sprintf("http://localhost:8080/v1/projects/ARGO/schemas/%v:validate", t.schemaName)

		body, _ := json.MarshalIndent(t.postBody, "", "")

		req, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			log.Fatal(err)
		}
		router.HandleFunc("/v1/projects/{project}/schemas/{schema}:validate", WrapMockAuthConfig(SchemaValidateMessage, cfgKafka, &brk, str, &mgr, pc))
		router.ServeHTTP(w, req)

		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
