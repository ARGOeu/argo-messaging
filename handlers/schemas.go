package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ARGOeu/argo-messaging/messages"
	"github.com/ARGOeu/argo-messaging/schemas"
	"github.com/ARGOeu/argo-messaging/stores"
	gorillaContext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/twinj/uuid"
	"net/http"
)

// SchemaCreate (POST) handles the creation of a new schema
func SchemaCreate(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Get url path variables
	urlVars := mux.Vars(r)
	schemaName := urlVars["schema"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	schemaUUID := uuid.NewV4().String()

	schema := schemas.Schema{}

	err := json.NewDecoder(r.Body).Decode(&schema)
	if err != nil {
		err := APIErrorInvalidArgument("Schema")
		respondErr(rCTX, w, err)
		return
	}

	schema, err = schemas.Create(rCTX, projectUUID, schemaUUID, schemaName, schema.Type, schema.RawSchema, refStr)
	if err != nil {
		if err.Error() == "exists" {
			err := APIErrorConflict("Schema")
			respondErr(rCTX, w, err)
			return

		}

		if err.Error() == "unsupported" {
			err := APIErrorInvalidData(schemas.UnsupportedSchemaError)
			respondErr(rCTX, w, err)
			return

		}

		err := APIErrorInvalidData(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	output, _ := json.MarshalIndent(schema, "", " ")
	respondOK(w, output)
}

// SchemaListOne (GET) retrieves information about the requested schema
func SchemaListOne(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Get url path variables
	urlVars := mux.Vars(r)
	schemaName := urlVars["schema"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)
	schemasList, err := schemas.Find(rCTX, projectUUID, "", schemaName, refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	if schemasList.Empty() {
		err := APIErrorNotFound("Schema")
		respondErr(rCTX, w, err)
		return
	}

	output, _ := json.MarshalIndent(schemasList.Schemas[0], "", " ")
	respondOK(w, output)
}

// SchemaLisAll (GET) retrieves all the schemas under the given project
func SchemaListAll(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)
	schemasList, err := schemas.Find(rCTX, projectUUID, "", "", refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	output, _ := json.MarshalIndent(schemasList, "", " ")
	respondOK(w, output)
}

// SchemaUpdate (PUT) updates the given schema
func SchemaUpdate(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Get url path variables
	urlVars := mux.Vars(r)
	schemaName := urlVars["schema"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)
	schemasList, err := schemas.Find(rCTX, projectUUID, "", schemaName, refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	if schemasList.Empty() {
		err := APIErrorNotFound("Schema")
		respondErr(rCTX, w, err)
		return
	}

	updatedSchema := schemas.Schema{}

	err = json.NewDecoder(r.Body).Decode(&updatedSchema)
	if err != nil {
		err := APIErrorInvalidArgument("Schema")
		respondErr(rCTX, w, err)
		return
	}

	if updatedSchema.FullName != "" {
		_, schemaName, err := schemas.ExtractSchema(updatedSchema.FullName)
		if err != nil {
			err := APIErrorInvalidData(err.Error())
			respondErr(rCTX, w, err)
			return
		}
		updatedSchema.Name = schemaName
	}

	schema, err := schemas.Update(rCTX, schemasList.Schemas[0], updatedSchema.Name, updatedSchema.Type, updatedSchema.RawSchema, refStr)
	if err != nil {
		if err.Error() == "exists" {
			err := APIErrorConflict("Schema")
			respondErr(rCTX, w, err)
			return

		}

		if err.Error() == "unsupported" {
			err := APIErrorInvalidData(schemas.UnsupportedSchemaError)
			respondErr(rCTX, w, err)
			return

		}

		err := APIErrorInvalidData(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	output, _ := json.MarshalIndent(schema, "", " ")
	respondOK(w, output)
}

func SchemaDelete(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Get url path variables
	urlVars := mux.Vars(r)
	schemaName := urlVars["schema"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	schemasList, err := schemas.Find(rCTX, projectUUID, "", schemaName, refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	if schemasList.Empty() {
		err := APIErrorNotFound("Schema")
		respondErr(rCTX, w, err)
		return
	}

	err = schemas.Delete(rCTX, schemasList.Schemas[0].UUID, refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	respondOK(w, nil)
}

// SchemaValidateMessage (POST) validates the given message against the schema
func SchemaValidateMessage(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Get url path variables
	urlVars := mux.Vars(r)
	schemaName := urlVars["schema"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)
	schemasList, err := schemas.Find(rCTX, projectUUID, "", schemaName, refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	if schemasList.Empty() {
		err := APIErrorNotFound("Schema")
		respondErr(rCTX, w, err)
		return
	}

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		err := APIErrorInvalidData(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	msgList := messages.MsgList{}

	switch schemasList.Schemas[0].Type {
	case schemas.JSON:
		msg := messages.Message{
			Data: base64.StdEncoding.EncodeToString(buf.Bytes()),
		}

		msgList.Msgs = append(msgList.Msgs, msg)

	case schemas.AVRO:

		body := map[string]string{}
		err := json.Unmarshal(buf.Bytes(), &body)
		if err != nil {
			err := APIErrorInvalidRequestBody()
			respondErr(rCTX, w, err)
			return
		}

		// check to find the payload field
		if val, ok := body["data"]; ok {

			msg := messages.Message{
				Data: val,
			}

			msgList.Msgs = append(msgList.Msgs, msg)

		} else {

			err := APIErrorInvalidArgument("Schema Payload")
			respondErr(rCTX, w, err)
			return
		}
	}

	err = schemas.ValidateMessages(schemasList.Schemas[0], msgList)
	if err != nil {
		if err.Error() == "500" {
			err := APIErrGenericInternal(schemas.GenericError)
			respondErr(rCTX, w, err)
			return
		} else {
			err := APIErrorInvalidData(err.Error())
			respondErr(rCTX, w, err)
			return
		}
	}

	res, _ := json.MarshalIndent(map[string]string{"message": "Message validated successfully"}, "", " ")

	respondOK(w, res)
}
