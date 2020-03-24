package schemas

import (
	"errors"
	"github.com/ARGOeu/argo-messaging/messages"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SchemasTestSuite struct {
	suite.Suite
}

func (suite *SchemasTestSuite) TestExtractSchema() {

	type td struct {
		schemaRef           string
		expectedProjectName string
		expectedSchemaName  string
		expectedErr         error
		msg                 string
	}

	testdata := []td{
		{
			schemaRef:           "projects/ARGO/schemas/s1",
			expectedProjectName: "ARGO",
			expectedSchemaName:  "s1",
			expectedErr:         nil,
			msg:                 "Case where both the Project and Schema names were extracted successfully",
		},
		{
			schemaRef:           "projectsARGO/schemas/s1",
			expectedProjectName: "",
			expectedSchemaName:  "",
			expectedErr:         errors.New("wrong schema name declaration"),
			msg:                 "Case where the schema ref doesn't contain the 4 needed \\/",
		},
		{
			schemaRef:           "proj/ARGO/schemas/s1",
			expectedProjectName: "",
			expectedSchemaName:  "",
			expectedErr:         errors.New("wrong schema name declaration"),
			msg:                 "Case where the schema ref doesn't contain the keyword projects",
		},
		{
			schemaRef:           "projects/ARGO/sch/s1",
			expectedProjectName: "",
			expectedSchemaName:  "",
			expectedErr:         errors.New("wrong schema name declaration"),
			msg:                 "Case where the schema ref doesn't contain the keyword schemas",
		},
		{
			schemaRef:           "projects/ARGO/schemas/s1/s2",
			expectedProjectName: "",
			expectedSchemaName:  "",
			expectedErr:         errors.New("wrong schema name declaration"),
			msg:                 "Case where the schema ref contains more than 4 \\/",
		},
	}
	for _, t := range testdata {
		p, s, e := ExtractSchema(t.schemaRef)
		suite.Equal(t.expectedProjectName, p, t.msg)
		suite.Equal(t.expectedSchemaName, s, t.msg)
		suite.Equal(t.expectedErr, e, t.msg)
	}
}

func (suite *SchemasTestSuite) TestFormatSchemaRef() {

	suite.Equal("projects/ARGO/schemas/s1", FormatSchemaRef("ARGO", "s1"))
}

func (suite *SchemasTestSuite) TestFind() {

	store := stores.NewMockStore("", "")

	type td struct {
		projectUUID string
		schemaUUID  string
		schemaName  string
		schemaList  SchemaList
		err         error
		msg         string
	}

	testData := []td{
		{projectUUID: "argo_uuid",
			schemaUUID: "",
			schemaName: "schema-1",
			schemaList: SchemaList{
				Schemas: []Schema{
					{UUID: "schema_uuid_1",
						ProjectUUID: "argo_uuid",
						Name:        "schema-1",
						FullName:    "projects/ARGO/schemas/schema-1",
						Type:        JSON,
						RawSchema: map[string]interface{}{
							"properties": map[string]interface{}{
								"address":   map[string]interface{}{"type": "string"},
								"email":     map[string]interface{}{"type": "string"},
								"name":      map[string]interface{}{"type": "string"},
								"telephone": map[string]interface{}{"type": "string"},
							},
							"required": []interface{}{"name", "email"},
							"type":     "object",
						},
					},
				},
			},
			err: nil,
			msg: "Case where we request for a single schema under a project and it is successfully retrieved",
		},
		{
			projectUUID: "argo_uuid",
			schemaUUID:  "",
			schemaName:  "",
			schemaList: SchemaList{
				Schemas: []Schema{
					{
						ProjectUUID: "argo_uuid",
						UUID:        "schema_uuid_1",
						Name:        "schema-1",
						FullName:    "projects/ARGO/schemas/schema-1",
						Type:        JSON,
						RawSchema: map[string]interface{}{
							"properties": map[string]interface{}{
								"address":   map[string]interface{}{"type": "string"},
								"email":     map[string]interface{}{"type": "string"},
								"name":      map[string]interface{}{"type": "string"},
								"telephone": map[string]interface{}{"type": "string"},
							},
							"required": []interface{}{"name", "email"},
							"type":     "object",
						},
					},
					{
						ProjectUUID: "argo_uuid",
						UUID:        "schema_uuid_2",
						Name:        "schema-2",
						FullName:    "projects/ARGO/schemas/schema-2",
						Type:        JSON,
						RawSchema: map[string]interface{}{
							"properties": map[string]interface{}{
								"address":   map[string]interface{}{"type": "string"},
								"email":     map[string]interface{}{"type": "string"},
								"name":      map[string]interface{}{"type": "string"},
								"telephone": map[string]interface{}{"type": "string"},
							},
							"required": []interface{}{"name", "email"},
							"type":     "object",
						},
					},
					{
						ProjectUUID: "argo_uuid",
						UUID:        "schema_uuid_3",
						Name:        "schema-3",
						FullName:    "projects/ARGO/schemas/schema-3",
						Type:        AVRO,
						RawSchema: map[string]interface{}{
							"namespace": "user.avro",
							"type":      "record",
							"name":      "User",
							"fields": []interface{}{
								map[string]interface{}{"name": "username", "type": "string"},
								map[string]interface{}{"name": "phone", "type": "int"},
							},
						},
					},
				},
			},
			err: nil,
			msg: "Case where we request for all schemas under a project and they are successfully retrieved",
		},
	}

	for _, t := range testData {
		s, e := Find(t.projectUUID, t.schemaUUID, t.schemaName, store)
		suite.Equal(t.err, e, t.msg)
		suite.Equal(t.schemaList, s, t.msg)
	}
}

func (suite *SchemasTestSuite) TestUpdate() {

	store := stores.NewMockStore("", "")

	type td struct {
		existingSchema Schema
		newName        string
		newType        string
		newSchema      map[string]interface{}
		expectedSchema Schema
		err            error
		queryFunc      func() interface{}
		returnQuery    interface{}
		msg            string
	}

	testData := []td{
		{
			existingSchema: Schema{
				ProjectUUID: "argo_uuid",
				UUID:        "schema_uuid_1",
				Name:        "schema-1",
				Type:        JSON,
				RawSchema: map[string]interface{}{
					"properties": map[string]interface{}{
						"address":   map[string]interface{}{"type": "string"},
						"email":     map[string]interface{}{"type": "string"},
						"name":      map[string]interface{}{"type": "string"},
						"telephone": map[string]interface{}{"type": "string"},
					},
					"required": []interface{}{"name", "email"},
					"type":     "object",
				},
			},
			newName:   "new-schema-name",
			newType:   JSON,
			newSchema: map[string]interface{}{"type": "string"},
			expectedSchema: Schema{
				ProjectUUID: "argo_uuid",
				UUID:        "schema_uuid_1",
				Name:        "new-schema-name",
				FullName:    "projects/ARGO/schemas/new-schema-name",
				Type:        JSON,
				RawSchema:   map[string]interface{}{"type": "string"},
			},
			err: nil,
			queryFunc: func() interface{} {
				qs, _ := store.QuerySchemas("argo_uuid", "schema_uuid_1", "new-schema-name")
				return qs[0]
			},
			returnQuery: stores.QSchema{
				ProjectUUID: "argo_uuid",
				UUID:        "schema_uuid_1",
				Name:        "new-schema-name",
				Type:        JSON,
				RawSchema:   "eyJ0eXBlIjoic3RyaW5nIn0=",
			},
			msg: "Case where a schema has all its fields successfully updated",
		},
		{
			existingSchema: Schema{
				ProjectUUID: "argo_uuid",
				UUID:        "schema_uuid_1",
				Name:        "schema-1",
				Type:        JSON,
				RawSchema: map[string]interface{}{
					"properties": map[string]interface{}{
						"address":   map[string]interface{}{"type": "string"},
						"email":     map[string]interface{}{"type": "string"},
						"name":      map[string]interface{}{"type": "string"},
						"telephone": map[string]interface{}{"type": "string"},
					},
					"required": []interface{}{"name", "email"},
					"type":     "object",
				},
			},
			newName:        "schema-2",
			newType:        JSON,
			newSchema:      map[string]interface{}{"type": "string"},
			expectedSchema: Schema{},
			err:            errors.New("exists"),
			queryFunc: func() interface{} {
				return nil
			},
			returnQuery: nil,
			msg:         "Case where a schema has been updated with a name that already exists",
		},
		{
			existingSchema: Schema{
				ProjectUUID: "argo_uuid",
				UUID:        "schema_uuid_1",
				Name:        "schema-1",
				Type:        JSON,
				RawSchema: map[string]interface{}{
					"properties": map[string]interface{}{
						"address":   map[string]interface{}{"type": "string"},
						"email":     map[string]interface{}{"type": "string"},
						"name":      map[string]interface{}{"type": "string"},
						"telephone": map[string]interface{}{"type": "string"},
					},
					"required": []interface{}{"name", "email"},
					"type":     "object",
				},
			},
			newName:        "schema-new-type",
			newType:        "new-type",
			newSchema:      map[string]interface{}{"type": "string"},
			expectedSchema: Schema{},
			err:            errors.New("unsupported"),
			queryFunc: func() interface{} {
				return nil
			},
			returnQuery: nil,
			msg:         "Case where a schema has been updated with an unsupported type",
		},
		{
			existingSchema: Schema{
				ProjectUUID: "argo_uuid",
				UUID:        "schema_uuid_1",
				Name:        "schema-1",
				Type:        JSON,
				RawSchema: map[string]interface{}{
					"properties": map[string]interface{}{
						"address":   map[string]interface{}{"type": "string"},
						"email":     map[string]interface{}{"type": "string"},
						"name":      map[string]interface{}{"type": "string"},
						"telephone": map[string]interface{}{"type": "string"},
					},
					"required": []interface{}{"name", "email"},
					"type":     "object",
				},
			},
			newName:        "schema-error-type",
			newType:        JSON,
			newSchema:      map[string]interface{}{"type": "unknown"},
			expectedSchema: Schema{},
			err:            errors.New("has a primitive type that is NOT VALID -- given: /unknown/ Expected valid values are:[array boolean integer number null object string]"),
			queryFunc: func() interface{} {
				return nil
			},
			returnQuery: nil,
			msg:         "Case where a schema has been updated with a erroneous schema",
		},
	}

	for _, t := range testData {
		s, e := Update(t.existingSchema, t.newName, t.newType, t.newSchema, store)
		suite.Equal(t.expectedSchema, s, t.msg)
		suite.Equal(t.err, e, t.msg)
		suite.Equal(t.returnQuery, t.queryFunc(), t.msg)
	}
}

func (suite *SchemasTestSuite) TestValidateMessages() {

	type td struct {
		schema      Schema
		messageList messages.MsgList
		err         error
		msg         string
	}

	schema := Schema{
		ProjectUUID: "argo_uuid",
		UUID:        "schema_uuid_1",
		Name:        "schema-1",
		Type:        JSON,
		RawSchema: map[string]interface{}{
			"properties": map[string]interface{}{
				"address":   map[string]interface{}{"type": "string"},
				"email":     map[string]interface{}{"type": "string"},
				"name":      map[string]interface{}{"type": "string"},
				"telephone": map[string]interface{}{"type": "string"},
			},
			"required": []interface{}{"name", "email"},
			"type":     "object",
		},
	}

	avroschema := Schema{
		ProjectUUID: "argo_uuid",
		UUID:        "schema_uuid_3",
		Name:        "schema-3",
		Type:        AVRO,
		RawSchema: map[string]interface{}{
			"namespace": "user.avro",
			"type":      "record",
			"name":      "User",
			"fields": []interface{}{
				map[string]interface{}{"name": "username", "type": "string"},
				map[string]interface{}{"name": "phone", "type": "int"},
			},
		},
	}

	testdata := []td{
		{
			schema: schema,
			messageList: messages.MsgList{
				Msgs: []messages.Message{
					{
						// {"name":"name-1", "email": "test@example.com"}
						Data: "eyJuYW1lIjoibmFtZS0xIiwgImVtYWlsIjogInRlc3RAZXhhbXBsZS5jb20ifQ==",
					},
					{
						//{"name":"name-1", "email": "test@example.com", "address":"Street 13","telephone":"6948567889"}
						Data: "eyJuYW1lIjoibmFtZS0xIiwgImVtYWlsIjogInRlc3RAZXhhbXBsZS5jb20iLCAiYWRkcmVzcyI6IlN0cmVldCAxMyIsInRlbGVwaG9uZSI6IjY5NDg1Njc4ODkifQ==",
					},
				},
			},
			err: nil,
			msg: "Case where the provided messages are successfully validated(JSON)",
		},
		{
			schema: avroschema,
			messageList: messages.MsgList{
				Msgs: []messages.Message{
					{
						// {"username":"agelos", "email": "698090"}
						Data: "DGFnZWxvc8T8Cg==",
					},
					{
						//{"placename":"street 22", "address": "place a"}
						Data: "T2JqAQQWYXZyby5zY2hlbWGYAnsidHlwZSI6InJlY29yZCIsIm5hbWUiOiJQbGFjZSIsIm5hbWVzcGFjZSI6InBsYWNlLmF2cm8iLCJmaWVsZHMiOlt7Im5hbWUiOiJwbGFjZW5hbWUiLCJ0eXBlIjoic3RyaW5nIn0seyJuYW1lIjoiYWRkcmVzcyIsInR5cGUiOiJzdHJpbmcifV19FGF2cm8uY29kZWMIbnVsbABM1P4b0GpYaCg9tqxa+YDZAiQSc3RyZWV0IDIyDnBsYWNlIGFM1P4b0GpYaCg9tqxa+YDZ",
					},
				},
			},
			err: errors.New("Message 1 is not valid.cannot decode binary record \"user.avro.User\" field \"username\": cannot decode binary string: cannot decode binary bytes: negative size: -40"),
			msg: "Case where one of the messages is not successfully validated(1 errors)(AVRO)",
		},
		{
			schema: schema,
			messageList: messages.MsgList{
				Msgs: []messages.Message{
					{
						// {"name":"name-1","address":"Street 13","telephone":6948567889}
						Data: "eyJuYW1lIjoibmFtZS0xIiwiYWRkcmVzcyI6IlN0cmVldCAxMyIsInRlbGVwaG9uZSI6Njk0ODU2Nzg4OX0=",
					},
					{
						//{"name":"name-1", "email": "test@example.com", "address":"Street 13","telephone":"6948567889"}
						Data: "eyJuYW1lIjoibmFtZS0xIiwgImVtYWlsIjogInRlc3RAZXhhbXBsZS5jb20iLCAiYWRkcmVzcyI6IlN0cmVldCAxMyIsInRlbGVwaG9uZSI6IjY5NDg1Njc4ODkifQ==",
					},
				},
			},
			err: errors.New("Message 0 data is not valid.1)(root): email is required.2)telephone: Invalid type. Expected: string, given: integer."),
			msg: "Case where one of the messages is not successfully validated(2 errors)(JSON)",
		},
		{
			schema: schema,
			messageList: messages.MsgList{
				Msgs: []messages.Message{
					{
						// {"name":"name-1","address":"Street 13","telephone":"6948567889"}
						Data: "eyJuYW1lIjoibmFtZS0xIiwiYWRkcmVzcyI6IlN0cmVldCAxMyIsInRlbGVwaG9uZSI6IjY5NDg1Njc4ODkifQo=",
					},
					{
						//{"name":"name-1", "email": "test@example.com", "address":"Street 13","telephone":"6948567889"}
						Data: "eyJuYW1lIjoibmFtZS0xIiwgImVtYWlsIjogInRlc3RAZXhhbXBsZS5jb20iLCAiYWRkcmVzcyI6IlN0cmVldCAxMyIsInRlbGVwaG9uZSI6IjY5NDg1Njc4ODkifQ==",
					},
				},
			},
			err: errors.New("Message 0 data is not valid,(root): email is required"),
			msg: "Case where the one of the messages is not successfully validated(1 error)(JSON)",
		},
		{
			schema: schema,
			messageList: messages.MsgList{
				Msgs: []messages.Message{
					{
						//{"name":"name-1", "email": "test@example.com", "address":"Street 13","telephone":"6948567889"}
						Data: "eyJuYW1lIjoibmFtZS0xIiwgImVtYWlsIjogInRlc3RAZXhhbXBsZS5jb20iLCAiYWRkcmVzcyI6IlN0cmVldCAxMyIsInRlbGVwaG9uZSI6IjY5NDg1Njc4ODkifQ==",
					},
					{
						//{"name":"name-1","address":"Street 13","telephone":"6948567889"
						Data: "eyJuYW1lIjoibmFtZS0xIiwiYWRkcmVzcyI6IlN0cmVldCAxMyIsInRlbGVwaG9uZSI6IjY5NDg1Njc4ODkiCg==",
					},
				},
			},
			err: errors.New("Message 1 data is not valid JSON format,unexpected EOF"),
			msg: "Case where the one of the messages is not in valid json format",
		},
	}

	for _, t := range testdata {
		suite.Equal(t.err, ValidateMessages(t.schema, t.messageList), t.msg)
	}
}

func (suite *SchemasTestSuite) TestDelete() {

	store := stores.NewMockStore("", "")

	e1 := Delete("schema_uuid_1", store)
	sl, _ := Find("argo_uuid", "schema_uuid_1", "", store)
	qtd, _, _, _ := store.QueryTopics("argo_uuid", "", "topic2", "", 1)
	suite.Equal([]Schema{}, sl.Schemas)
	suite.Equal("", qtd[0].SchemaUUID)
	suite.Nil(e1)
}

func (suite *SchemasTestSuite) TestCreate() {

	store := stores.NewMockStore("", "")

	type td struct {
		projectUUID    string
		uuid           string
		name           string
		schemaType     string
		rawSchema      map[string]interface{}
		err            error
		returnedSchema Schema
		queryFunc      func() interface{}
		returnQuery    interface{}
		msg            string
	}

	testData := []td{
		{
			projectUUID: "argo_uuid",
			uuid:        "suuid",
			name:        "s1",
			schemaType:  "jSOn",
			rawSchema:   map[string]interface{}{"type": "string"},
			err:         nil,
			returnedSchema: Schema{
				UUID:      "suuid",
				Name:      "s1",
				FullName:  "projects/ARGO/schemas/s1",
				Type:      JSON,
				RawSchema: map[string]interface{}{"type": "string"},
			},
			queryFunc: func() interface{} {
				qs, _ := store.QuerySchemas("argo_uuid", "suuid", "s1")
				return qs[0]
			},
			returnQuery: stores.QSchema{
				ProjectUUID: "argo_uuid",
				UUID:        "suuid",
				Name:        "s1",
				Type:        JSON,
				RawSchema:   "eyJ0eXBlIjoic3RyaW5nIn0=",
			},
			msg: "Case where the given schema has been validated and saved to the store successfully",
		},
		{
			projectUUID:    "argo_uuid",
			uuid:           "suuid",
			name:           "schema-1",
			schemaType:     "jSOn",
			rawSchema:      map[string]interface{}{"type": "string"},
			err:            errors.New("exists"),
			returnedSchema: Schema{},
			queryFunc: func() interface{} {
				return nil
			},
			returnQuery: nil,
			msg:         "Case where the given schema name is already taken by another schema",
		},
	}

	for _, t := range testData {
		s, e := Create(t.projectUUID, t.uuid, t.name, t.schemaType, t.rawSchema, store)
		suite.Equal(t.err, e, t.msg)
		suite.Equal(t.returnedSchema, s, t.msg)
		suite.Equal(t.returnQuery, t.queryFunc())
	}

}

func (suite *SchemasTestSuite) TestExistsWithName() {

	type td struct {
		schemaName string
		exists     bool
		msg        string
	}

	testData := []td{
		{
			schemaName: "schema-1",
			exists:     true,
			msg:        "Case where the given schema exists",
		},
		{
			schemaName: "schema-unknown",
			exists:     false,
			msg:        "Case where the given schema doesn't exist",
		},
	}

	store := stores.NewMockStore("", "")

	for _, t := range testData {
		b, _ := ExistsWithName("argo_uuid", t.schemaName, store)
		suite.Equal(t.exists, b, t.msg)
	}
}

func (suite *SchemasTestSuite) TestCheckSchema() {

	type td struct {
		schemaType string
		schema     map[string]interface{}
		err        error
		msg        string
	}

	testData := []td{
		{
			schemaType: JSON,
			schema: map[string]interface{}{
				"type": "string",
			},
			err: nil,
			msg: "Case where the provided schema type is supported and the format of the schema is correct(JSON)",
		},
		{
			schemaType: AVRO,
			schema: map[string]interface{}{
				"namespace": "user.avro",
				"type":      "record",
				"name":      "User",
				"fields": []interface{}{
					map[string]interface{}{"name": "username", "type": "string"},
					map[string]interface{}{"name": "phone", "type": "int"},
				},
			},
			err: nil,
			msg: "Case where the provided schema type is supported and the format of the schema is correct(AVRO)",
		},
		{
			schemaType: AVRO,
			schema: map[string]interface{}{
				"namespace": "user.avro",
				"type":      "unknown",
				"name":      "User",
			},
			err: errors.New("unknown type name: \"unknown\""),
			msg: "Case where the provided schema type is supported but the format of the schema is incorrect(AVRO)",
		},
		{
			schemaType: JSON,
			schema: map[string]interface{}{
				"type": "unknown",
			},
			err: errors.New("has a primitive type that is NOT VALID -- given: /unknown/ Expected valid values are:[array boolean integer number null object string]"),
			msg: "Case where the provided schema type is supported but the format of the schema is incorrect(JSON)",
		},
		{
			schemaType: "unknown",
			schema: map[string]interface{}{
				"type": "unknown",
			},
			err: errors.New("unsupported"),
			msg: "Case where the provided schema type is unsupported",
		},
	}

	for _, t := range testData {
		e := checkSchema(t.schemaType, t.schema)
		suite.Equal(t.err, e, t.msg)
	}
}

func TestSchemasTestSuite(t *testing.T) {
	suite.Run(t, new(SchemasTestSuite))
}
