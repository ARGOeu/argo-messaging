package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	oldPush "github.com/ARGOeu/argo-messaging/push"
	push "github.com/ARGOeu/argo-messaging/push/grpc/client"
	"github.com/ARGOeu/argo-messaging/schemas"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type SchemasHandlersTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *SchemasHandlersTestSuite) SetupTest() {
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

func (suite *SchemasHandlersTestSuite) TestSchemaCreate() {

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

func (suite *SchemasHandlersTestSuite) TestSchemaListOne() {

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

func (suite *SchemasHandlersTestSuite) TestSchemaListAll() {

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

func (suite *SchemasHandlersTestSuite) TestSchemaUpdate() {

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

func (suite *SchemasHandlersTestSuite) TestSchemaDelete() {

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

func (suite *SchemasHandlersTestSuite) TestSchemaValidateMessage() {

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

func TestSchemasHandlersTestSuite(t *testing.T) {
	log.SetOutput(io.Discard)
	suite.Run(t, new(SchemasHandlersTestSuite))
}
