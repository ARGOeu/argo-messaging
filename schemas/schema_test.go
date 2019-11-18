package schemas

import (
	"errors"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SchemasTestSuite struct {
	suite.Suite
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
					{UUID: "schema_uuid_2",
						ProjectUUID: "argo_uuid",
						Name:        "schema-2",
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
			newName:        "schema-3",
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
			newName:        "schema-3",
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
			msg: "Case where the provided schema type is supported and the format of the schema is correct",
		},
		{
			schemaType: JSON,
			schema: map[string]interface{}{
				"type": "unknown",
			},
			err: errors.New("has a primitive type that is NOT VALID -- given: /unknown/ Expected valid values are:[array boolean integer number null object string]"),
			msg: "Case where the provided schema type is supported but the format of the schema is incorrect",
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
