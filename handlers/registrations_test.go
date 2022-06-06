package handlers

import (
	"fmt"
	"github.com/ARGOeu/argo-messaging/auth"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	oldPush "github.com/ARGOeu/argo-messaging/push"
	push "github.com/ARGOeu/argo-messaging/push/grpc/client"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type RegistrationsHandlersTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *RegistrationsHandlersTestSuite) SetupTest() {
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

func (suite *RegistrationsHandlersTestSuite) TestRegisterUser() {

	type td struct {
		postBody           string
		expectedResponse   string
		expectedStatusCode int
		msg                string
	}

	testData := []td{
		{
			postBody: `{
							"name": "new-register-user",
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
			postBody: `{
							"name": "UserA",
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
		req, err := http.NewRequest("POST", "http://localhost:8080/v1/registrations", strings.NewReader(t.postBody))
		if err != nil {
			log.Fatal(err)
		}
		router.HandleFunc("/v1/registrations", WrapMockAuthConfig(RegisterUser, cfgKafka, &brk, str, &mgr, pc))
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

func (suite *RegistrationsHandlersTestSuite) TestAcceptRegisterUser() {

	type td struct {
		ruuid              string
		uname              string
		expectedResponse   string
		expectedStatusCode int
		msg                string
	}

	testData := []td{{
		ruuid: "ur-uuid1",
		uname: "urname",
		expectedResponse: `{
   "uuid": "{{UUID}}",
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
			ruuid: "ur-uuid1",
			uname: "urname",
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
		url := fmt.Sprintf("http://localhost:8080/v1/registrations/%v:accept", t.ruuid)
		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		router.HandleFunc("/v1/registrations/{uuid}:accept", WrapMockAuthConfig(AcceptRegisterUser, cfgKafka, &brk, str, &mgr, pc))
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

func (suite *RegistrationsHandlersTestSuite) TestDeclineRegisterUser() {

	type td struct {
		postBody           string
		regUUID            string
		expectedResponse   string
		declineComment     string
		expectedStatusCode int
		msg                string
	}

	testData := []td{{
		postBody: `{
						"comment": "decline comment"
				   }`,
		regUUID:            "ur-uuid1",
		declineComment:     "decline comment",
		expectedResponse:   `{}`,
		expectedStatusCode: 200,
		msg:                "Successfully declined a user's registration",
	},
		{
			regUUID: "unknown",
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
		url := fmt.Sprintf("http://localhost:8080/v1/registrations/%v:decline", t.regUUID)
		req, err := http.NewRequest("POST", url, strings.NewReader(t.postBody))
		if err != nil {
			log.Fatal(err)
		}
		router.HandleFunc("/v1/registrations/{uuid}:decline", WrapMockAuthConfig(DeclineRegisterUser, cfgKafka, &brk, str, &mgr, pc))
		router.ServeHTTP(w, req)
		if t.expectedStatusCode == 200 {
			suite.Equal(auth.DeclinedRegistrationStatus, str.UserRegistrations[0].Status)
			suite.Equal(t.declineComment, str.UserRegistrations[0].DeclineComment)
		}
		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}

}

func (suite *RegistrationsHandlersTestSuite) TestListOneRegistration() {

	type td struct {
		regUUID            string
		expectedResponse   string
		expectedStatusCode int
		msg                string
	}

	testData := []td{
		{
			regUUID: "ur-uuid1",
			expectedResponse: `{
   "uuid": "ur-uuid1",
   "name": "urname",
   "first_name": "urfname",
   "last_name": "urlname",
   "organization": "urorg",
   "description": "urdesc",
   "email": "uremail",
   "status": "pending",
   "activation_token": "uratkn-1",
   "registered_at": "2019-05-12T22:26:58Z",
   "modified_by": "UserA",
   "modified_at": "2020-05-15T22:26:58Z"
}`,
			expectedStatusCode: 200,
			msg:                "User registration retrieved successfully",
		},
		{
			regUUID: "unknown",
			expectedResponse: `{
   "error": {
      "code": 404,
      "message": "User registration doesn't exist",
      "status": "NOT_FOUND"
   }
}`,
			expectedStatusCode: 404,
			msg:                "User registration doesn't exist",
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
		url := fmt.Sprintf("http://localhost:8080/v1/registrations/%v", t.regUUID)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		router.HandleFunc("/v1/registrations/{uuid}", WrapMockAuthConfig(ListOneRegistration, cfgKafka, &brk, str, &mgr, pc))
		router.ServeHTTP(w, req)
		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}

}

func (suite *RegistrationsHandlersTestSuite) TestListManyRegistrations() {

	type td struct {
		urlPath            string
		expectedResponse   string
		expectedStatusCode int
		msg                string
	}

	testData := []td{
		{
			urlPath: "registrations",
			expectedResponse: `{
   "user_registrations": [
      {
         "uuid": "ur-uuid1",
         "name": "urname",
         "first_name": "urfname",
         "last_name": "urlname",
         "organization": "urorg",
         "description": "urdesc",
         "email": "uremail",
         "status": "pending",
         "activation_token": "uratkn-1",
         "registered_at": "2019-05-12T22:26:58Z",
         "modified_by": "UserA",
         "modified_at": "2020-05-15T22:26:58Z"
      }
   ]
}`,
			expectedStatusCode: 200,
			msg:                "Retrieve all available user registrations without any filters",
		},
		{
			urlPath: "registrations?status=pending&name=urname&activation_token=uratkn-1&email=uremail&organization=urorg",
			expectedResponse: `{
   "user_registrations": [
      {
         "uuid": "ur-uuid1",
         "name": "urname",
         "first_name": "urfname",
         "last_name": "urlname",
         "organization": "urorg",
         "description": "urdesc",
         "email": "uremail",
         "status": "pending",
         "activation_token": "uratkn-1",
         "registered_at": "2019-05-12T22:26:58Z",
         "modified_by": "UserA",
         "modified_at": "2020-05-15T22:26:58Z"
      }
   ]
}`,
			expectedStatusCode: 200,
			msg:                "Retrieve all available user registrations with filters",
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
		url := fmt.Sprintf("http://localhost:8080/v1/%v", t.urlPath)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		router.HandleFunc("/v1/registrations", WrapMockAuthConfig(ListAllRegistrations, cfgKafka, &brk, str, &mgr, pc))
		router.ServeHTTP(w, req)
		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}

}

func TestRegistrationsHandlersTestSuite(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	suite.Run(t, new(RegistrationsHandlersTestSuite))
}
