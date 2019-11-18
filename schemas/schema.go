package schemas

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/ARGOeu/argo-messaging/stores"
	log "github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
	"strings"
)

const (
	JSON                   = "json"
	UnsupportedSchemaError = `Schema type can only be 'json'`
)

// schema holds information regarding a schema that will be used to validate a topic's published messages
type Schema struct {
	ProjectUUID string                 `json:"-"`
	UUID        string                 `json:"uuid"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	RawSchema   map[string]interface{} `json:"schema"`
}

// SchemaList is a wrapper for a slice of schemas
type SchemaList struct {
	Schemas []Schema `json:"schemas"`
}

// Empty returns weather or not there are any schemas inside the schema list
func (sl *SchemaList) Empty() bool {
	return len(sl.Schemas) <= 0
}

// Find retrieves a specific schema or all the schemas under a project
func Find(projectUUID, schemaUUID, schemaName string, str stores.Store) (SchemaList, error) {

	schemaList := SchemaList{}

	qSchemas, err := str.QuerySchemas(projectUUID, schemaUUID, schemaName)
	if err != nil {
		return schemaList, err
	}

	for _, s := range qSchemas {
		_schema := Schema{}
		_schema.UUID = s.UUID
		_schema.Name = s.Name
		_schema.Type = s.Type

		decodedSchemaBytes, err := base64.StdEncoding.DecodeString(s.RawSchema)
		if err != nil {
			log.WithFields(
				log.Fields{
					"type":         "service_log",
					"schema_name":  schemaName,
					"project_uuid": projectUUID,
					"error":        err.Error(),
				},
			).Error("Could not decode the base64 encoded schema")
			return SchemaList{}, errors.New("Could not load the schema")
		}

		err = json.Unmarshal(decodedSchemaBytes, &_schema.RawSchema)
		if err != nil {
			log.WithFields(
				log.Fields{
					"type":         "service_log",
					"schema_name":  schemaName,
					"project_uuid": projectUUID,
					"error":        err.Error(),
				},
			).Error("Could not marshal the schema bytes")
			return SchemaList{}, errors.New("Could not load the schema")
		}

		schemaList.Schemas = append(schemaList.Schemas, _schema)
	}

	return schemaList, nil
}

// Create checks the validity of the schema to be created and then saves it to the store
func Create(projectUUID, schemaUUID, name, schemaType string, rawSchema map[string]interface{}, str stores.Store) (Schema, error) {

	exists, err := ExistsWithName(projectUUID, name, str)
	if err != nil {
		return Schema{}, err
	}

	if exists {
		return Schema{}, errors.New("exists")
	}

	schemaBytes, err := json.Marshal(rawSchema)
	if err != nil {
		return Schema{}, err
	}

	schemaType = strings.ToLower(schemaType)

	b64SchemaString := base64.StdEncoding.EncodeToString(schemaBytes)

	err = checkSchema(schemaType, rawSchema)
	if err != nil {
		return Schema{}, err
	}

	err = str.InsertSchema(projectUUID, schemaUUID, name, schemaType, b64SchemaString)
	if err != nil {
		return Schema{}, err
	}

	schema := Schema{
		UUID:      schemaUUID,
		Name:      name,
		Type:      schemaType,
		RawSchema: rawSchema,
	}

	return schema, nil
}

// checkSchema checks that the schema content is indeed of its provided schema type
func checkSchema(schemaType string, schemaContent map[string]interface{}) error {

	switch strings.ToLower(schemaType) {
	case JSON:

		jsonLoader := gojsonschema.NewGoLoader(schemaContent)
		_, err := gojsonschema.NewSchemaLoader().Compile(jsonLoader)
		if err != nil {
			return err
		}
	default:
		return errors.New("unsupported")
	}
	return nil
}

// ExistsWithName checks if a schema with the given name exists under the given project
func ExistsWithName(projectUUID string, schemaName string, str stores.Store) (bool, error) {
	qSchemas, err := str.QuerySchemas(projectUUID, "", schemaName)

	if err != nil {
		return false, errors.New("backend error")
	}

	if len(qSchemas) == 0 {
		return false, nil
	}

	return true, nil
}
