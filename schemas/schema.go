package schemas

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/ARGOeu/argo-messaging/stores"
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
