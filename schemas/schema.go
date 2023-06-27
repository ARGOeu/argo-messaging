package schemas

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ARGOeu/argo-messaging/messages"
	"github.com/ARGOeu/argo-messaging/projects"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/linkedin/goavro"
	log "github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
	"strings"
)

const (
	JSON                   = "json"
	AVRO                   = "avro"
	UnsupportedSchemaError = `Schema type can only be 'json' or 'avro'`
	GenericError           = "Could not load schema for topic"
)

// Schema holds information regarding a schema that will be used to validate a topic's published messages
type Schema struct {
	ProjectUUID string                 `json:"-"`
	UUID        string                 `json:"uuid"`
	Name        string                 `json:"-"`
	FullName    string                 `json:"name"`
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

// ValidateMessages validates a list of messages against the provided schema
func ValidateMessages(schema Schema, msgList messages.MsgList) error {

	switch schema.Type {
	case JSON:
		// load the schema
		s, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(schema.RawSchema))

		if err != nil {
			log.WithFields(
				log.Fields{
					"type":        "service_log",
					"schema_name": schema.FullName,
					"error":       err.Error(),
				},
			).Error("Could not load json schema")
			return errors.New("500")
		}

		for idx, msg := range msgList.Msgs {

			// decode the message payload from base64
			messageBytes, err := base64.StdEncoding.DecodeString(msg.Data)

			if err != nil {
				return fmt.Errorf("Message %v is not in valid base64 enocding,%s", idx, err.Error())
			}

			documentLoader := gojsonschema.NewBytesLoader(messageBytes)

			result, err := s.Validate(documentLoader)
			if err != nil {
				return fmt.Errorf("Message %v data is not valid JSON format,%s", idx, err.Error())
			}

			if !result.Valid() {
				if len(result.Errors()) > 1 {
					sb := strings.Builder{}

					for idx, e := range result.Errors() {
						sb.WriteString(fmt.Sprintf("%v)%s.", idx+1, e.String()))
					}

					return fmt.Errorf("Message %v data is not valid.%s", idx, sb.String())
				} else {
					return fmt.Errorf("Message %v data is not valid,%v", idx, result.Errors()[0].String())
				}
			}
		}
	case AVRO:
		// convert the schema to a json string representation
		b, err := json.Marshal(schema.RawSchema)
		if err != nil {
			log.WithFields(
				log.Fields{
					"type":        "service_log",
					"schema_name": schema.FullName,
					"error":       err.Error(),
				},
			).Error("Could not convert to json bytes representation")
			return errors.New("500")
		}

		c, err := goavro.NewCodec(string(b))
		if err != nil {
			log.WithFields(
				log.Fields{
					"type":        "service_log",
					"schema_name": schema.FullName,
					"error":       err.Error(),
				},
			).Error("Could not load avro schema")
			return errors.New("500")
		}

		for idx, msg := range msgList.Msgs {

			// decode the message payload from base64
			messageBytes, err := base64.StdEncoding.DecodeString(msg.Data)
			if err != nil {
				return fmt.Errorf("Message %v is not in valid base64 enocding,%s", idx, err.Error())
			}

			_, _, err = c.NativeFromBinary(messageBytes)
			if err != nil {
				return fmt.Errorf("Message %v is not valid.%s", idx, err.Error())
			}
		}

	default:
		log.WithFields(
			log.Fields{
				"type":        "service_log",
				"schema_name": schema.FullName,
				"schema_type": schema.Type,
			},
		).Error("Schema with unsupported type")
		return errors.New("500")
	}

	return nil
}

// ExtractSchema gets a full schema ref and extracts project and schema
// the format of the schema ref should follow the pattern projects/{project}/schemas/schema
func ExtractSchema(schemaRef string) (string, string, error) {
	items := strings.Split(schemaRef, "/")

	if len(items) != 4 {
		return "", "", errors.New("wrong schema name declaration")
	}

	if items[0] != "projects" || items[2] != "schemas" {
		return "", "", errors.New("wrong schema name declaration")
	}

	return items[1], items[3], nil
}

// FormatSchemaRef formats the full resource reference for a schema
// format is projects/{project}/schemas/{schema}
func FormatSchemaRef(projectName, schemaName string) string {
	return fmt.Sprintf("projects/%s/schemas/%s", projectName, schemaName)
}

// Find retrieves a specific schema or all the schemas under a project
func Find(ctx context.Context, projectUUID, schemaUUID, schemaName string, str stores.Store) (SchemaList, error) {

	schemaList := SchemaList{
		Schemas: []Schema{},
	}

	qSchemas, err := str.QuerySchemas(ctx, projectUUID, schemaUUID, schemaName)
	if err != nil {
		return schemaList, err
	}

	for _, s := range qSchemas {
		_schema := Schema{}
		_schema.UUID = s.UUID
		_schema.Name = s.Name
		_schema.Type = s.Type
		_schema.ProjectUUID = s.ProjectUUID

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

		projectName := projects.GetNameByUUID(ctx, projectUUID, str)

		_schema.FullName = FormatSchemaRef(projectName, s.Name)

		schemaList.Schemas = append(schemaList.Schemas, _schema)
	}

	return schemaList, nil
}

// Delete wraps the store's method for removing a schema
func Delete(ctx context.Context, schemaUUID string, str stores.Store) error {
	return str.DeleteSchema(ctx, schemaUUID)
}

// Update updates the provided schema , validates its content and saves it to the store
func Update(ctx context.Context, existingSchema Schema, newSchemaName, newSchemaType string, newRawSchema map[string]interface{}, str stores.Store) (Schema, error) {

	newSchema := Schema{}

	if newSchemaName != "" {
		// if the name has changed check that is not already taken by another schema under the given project
		if existingSchema.Name != newSchemaName {
			exists, err := ExistsWithName(ctx, existingSchema.ProjectUUID, newSchemaName, str)
			if err != nil {
				return Schema{}, err
			}

			if exists {
				return Schema{}, errors.New("exists")
			}

			existingSchema.Name = newSchemaName
		}
	}

	newSchema.Name = newSchemaName

	if newSchemaType != "" {
		newSchemaType = strings.ToLower(newSchemaType)
		newSchema.Type = newSchemaType
		existingSchema.Type = newSchemaType
	} else {
		newSchema.Type = existingSchema.Type
	}

	rawSchemaString := ""

	// if there is a new schema check the validity
	if len(newRawSchema) > 0 {
		err := checkSchema(newSchema.Type, newRawSchema)
		if err != nil {
			return Schema{}, err
		}

		schemaBytes, err := json.Marshal(newRawSchema)
		if err != nil {
			return Schema{}, err
		}

		rawSchemaString = base64.StdEncoding.EncodeToString(schemaBytes)

		newSchema.RawSchema = newRawSchema

		existingSchema.RawSchema = newRawSchema

	}

	// if there is a new type for the already existing schema
	if len(newRawSchema) == 0 && newSchemaType != "" {
		err := checkSchema(newSchema.Type, existingSchema.RawSchema)
		if err != nil {
			return Schema{}, err
		}

	}

	err := str.UpdateSchema(ctx, existingSchema.UUID, newSchema.Name, newSchema.Type, rawSchemaString)
	if err != nil {
		return Schema{}, err
	}

	projectName := projects.GetNameByUUID(ctx, existingSchema.ProjectUUID, str)

	existingSchema.FullName = FormatSchemaRef(projectName, existingSchema.Name)

	return existingSchema, nil
}

// Create checks the validity of the schema to be created and then saves it to the store
func Create(ctx context.Context, projectUUID, schemaUUID, name, schemaType string, rawSchema map[string]interface{}, str stores.Store) (Schema, error) {

	exists, err := ExistsWithName(ctx, projectUUID, name, str)
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

	b64SchemaString := base64.StdEncoding.EncodeToString(schemaBytes)

	schemaType = strings.ToLower(schemaType)

	err = checkSchema(schemaType, rawSchema)
	if err != nil {
		return Schema{}, err
	}

	err = str.InsertSchema(ctx, projectUUID, schemaUUID, name, schemaType, b64SchemaString)
	if err != nil {
		return Schema{}, err
	}

	projectName := projects.GetNameByUUID(ctx, projectUUID, str)

	schema := Schema{
		UUID:      schemaUUID,
		Name:      name,
		Type:      schemaType,
		RawSchema: rawSchema,
		FullName:  FormatSchemaRef(projectName, name),
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

	case AVRO:

		// convert the schema to a json string representation
		b, err := json.Marshal(schemaContent)
		if err != nil {
			return err
		}

		_, err = goavro.NewCodec(string(b))
		if err != nil {
			return err
		}

	default:
		return errors.New("unsupported")
	}
	return nil
}

// ExistsWithName checks if a schema with the given name exists under the given project
func ExistsWithName(ctx context.Context, projectUUID string, schemaName string, str stores.Store) (bool, error) {
	qSchemas, err := str.QuerySchemas(ctx, projectUUID, "", schemaName)

	if err != nil {
		return false, errors.New("backend error")
	}

	if len(qSchemas) == 0 {
		return false, nil
	}

	return true, nil
}
