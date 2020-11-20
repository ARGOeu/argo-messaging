package handlers

import (
	"bytes"
	"github.com/ARGOeu/argo-messaging/auth"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	oldPush "github.com/ARGOeu/argo-messaging/push"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type UsersHandlersTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *UsersHandlersTestSuite) SetupTest() {
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

func (suite *UsersHandlersTestSuite) TestUserProfile() {

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

func (suite *UsersHandlersTestSuite) TestUserProfileUnauthorized() {

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

func (suite *UsersHandlersTestSuite) TestUserCreate() {

	postJSON := `{
	"email":"email@foo.com",
	"first_name": "fname-1",
	"last_name": "lname-1",
	"organization": "org-1",
	"description": "desc-1",
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
	suite.Equal("fname-1", usrOut.FirstName)
	suite.Equal("lname-1", usrOut.LastName)
	suite.Equal("org-1", usrOut.Organization)
	suite.Equal("desc-1", usrOut.Description)
}

func (suite *UsersHandlersTestSuite) TestUserCreateDuplicateRef() {

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

func (suite *UsersHandlersTestSuite) TestUserCreateInvalidServiceRole() {

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

func (suite *UsersHandlersTestSuite) TestUserCreateInvalidProjectName() {

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

func (suite *UsersHandlersTestSuite) TestUserCreateInvalidRoles() {

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

func (suite *UsersHandlersTestSuite) TestRefreshToken() {

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

func (suite *UsersHandlersTestSuite) TestUserUpdate() {

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

func (suite *UsersHandlersTestSuite) TestUserUpdateInvalidProjectName() {

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

func (suite *UsersHandlersTestSuite) TestUserUpdateInvalidRoles() {

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

func (suite *UsersHandlersTestSuite) TestUserUpdateInvalidServiceRoles() {

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

func (suite *UsersHandlersTestSuite) TestUserUpdateDuplicate() {
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

func (suite *UsersHandlersTestSuite) TestUserListByToken() {

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

func (suite *UsersHandlersTestSuite) TestUserListByUUID() {

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

func (suite *UsersHandlersTestSuite) TestUserListByUUIDNotFound() {

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

func (suite *UsersHandlersTestSuite) TestUserListByUUIDConflict() {

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

func (suite *UsersHandlersTestSuite) TestUserListOne() {

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

func (suite *UsersHandlersTestSuite) TestUserListAll() {

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

func (suite *UsersHandlersTestSuite) TestUserListAllStartingPage() {

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

func (suite *UsersHandlersTestSuite) TestUserListAllProjectARGO() {

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

func (suite *UsersHandlersTestSuite) TestUserListAllProjectARGO2() {

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

func (suite *UsersHandlersTestSuite) TestUserListAllProjectUNKNOWN() {

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

func (suite *UsersHandlersTestSuite) TestUserListAllStartingAtSecond() {

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

func (suite *UsersHandlersTestSuite) TestUserListAllEmptyCollection() {

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

func (suite *UsersHandlersTestSuite) TestUserListAllIntermediatePage() {

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

func (suite *UsersHandlersTestSuite) TestUserListAllInvalidPageSize() {

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

func (suite *UsersHandlersTestSuite) TestUserListAllInvalidPageToken() {

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

func (suite *UsersHandlersTestSuite) TestUserDelete() {

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

func TestUsersHandlersTestSuite(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	suite.Run(t, new(UsersHandlersTestSuite))
}
