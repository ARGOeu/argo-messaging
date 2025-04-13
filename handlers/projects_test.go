package handlers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ARGOeu/argo-messaging/auth"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/projects"
	oldPush "github.com/ARGOeu/argo-messaging/push"
	push "github.com/ARGOeu/argo-messaging/push/grpc/client"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type ProjectsHandlersTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *ProjectsHandlersTestSuite) SetupTest() {
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

func (suite *ProjectsHandlersTestSuite) TestProjectUserListOne() {

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

func (suite *ProjectsHandlersTestSuite) TestProjectUserListOneUnpriv() {

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

func (suite *ProjectsHandlersTestSuite) TestProjectUserListARGO() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/users?details=true", nil)
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

func (suite *ProjectsHandlersTestSuite) TestProjectUserListARGONoUserDetails() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/users?details=false&pageSize=1", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "users": [
      {
         "uuid": "same_uuid",
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

func (suite *ProjectsHandlersTestSuite) TestProjectUserListUnprivARGO() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/members?details=true", nil)
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

func (suite *ProjectsHandlersTestSuite) TestProjectUserCreate() {

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
			u, _ := auth.FindUsers(context.Background(), "argo_uuid", "", t.user, true, str)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{UUID}}", u.List[0].UUID, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{TOKEN}}", u.List[0].Token, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{CON}}", u.List[0].CreatedOn, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{MON}}", u.List[0].ModifiedOn, 1)
		}
		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}

}

func (suite *ProjectsHandlersTestSuite) TestProjectUserUpdate() {

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
			u, _ := auth.FindUsers(context.Background(), "argo_uuid", "", t.user, true, str)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{UUID}}", u.List[0].UUID, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{TOKEN}}", u.List[0].Token, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{CON}}", u.List[0].CreatedOn, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{MON}}", u.List[0].ModifiedOn, 1)
		}
		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}

}

func (suite *ProjectsHandlersTestSuite) TestProjectUserRemove() {

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

func (suite *ProjectsHandlersTestSuite) TestProjectUserAdd() {

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
			u, _ := auth.FindUsers(context.Background(), "argo_uuid", "", t.user, true, str)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{UUID}}", u.List[0].UUID, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{TOKEN}}", u.List[0].Token, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{CON}}", u.List[0].CreatedOn, 1)
			t.expectedResponse = strings.Replace(t.expectedResponse, "{{MON}}", u.List[0].ModifiedOn, 1)
		}
		suite.Equal(t.expectedStatusCode, w.Code, t.msg)
		suite.Equal(t.expectedResponse, w.Body.String(), t.msg)
	}

}

func (suite *ProjectsHandlersTestSuite) TestProjectDelete() {

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

func (suite *ProjectsHandlersTestSuite) TestProjectUpdate() {

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

func (suite *ProjectsHandlersTestSuite) TestProjectCreate() {

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

func (suite *ProjectsHandlersTestSuite) TestProjectListAll() {

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

func (suite *ProjectsHandlersTestSuite) TestProjectListOneNotFound() {

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

func (suite *ProjectsHandlersTestSuite) TestProjectListOne() {

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

func TestProjectsHandlersTestSuite(t *testing.T) {
	log.SetOutput(io.Discard)
	suite.Run(t, new(ProjectsHandlersTestSuite))
}
