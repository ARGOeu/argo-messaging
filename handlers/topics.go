package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ARGOeu/argo-messaging/auth"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/messages"
	"github.com/ARGOeu/argo-messaging/schemas"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/ARGOeu/argo-messaging/subscriptions"
	"github.com/ARGOeu/argo-messaging/topics"
	gorillaContext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// TopicDelete (DEL) deletes an existing topic
func TopicDelete(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refBrk := gorillaContext.Get(r, "brk").(brokers.Broker)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Get Result Object

	err := topics.RemoveTopic(rCTX, projectUUID, urlVars["topic"], refStr)
	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("Topic")
			respondErr(rCTX, w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	fullTopic := projectUUID + "." + urlVars["topic"]
	err = refBrk.DeleteTopic(rCTX, fullTopic)
	if err != nil {
		log.WithFields(
			log.Fields{
				"trace_id": rCTX.Value("trace_id"),
				"type":     "service_log",
				"topic":    fullTopic,
				"error":    err.Error(),
			},
		).Error("Couldn't delete topic from broker")
	}

	// Write empty response if anything ok
	respondOK(w, output)
}

// TopicModACL (PUT) modifies the ACL
func TopicModACL(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	// Get Result Object
	urlTopic := urlVars["topic"]

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(rCTX, w, err)
		return
	}

	// Parse pull options
	postBody, err := auth.GetACLFromJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("Topic ACL")
		respondErr(rCTX, w, err)
		return
	}

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// check if user list contain valid users for the given project
	_, err = auth.AreValidUsers(rCTX, projectUUID, postBody.AuthUsers, refStr)
	if err != nil {
		err := APIErrorRoot{Body: APIErrorBody{Code: http.StatusNotFound, Message: err.Error(), Status: "NOT_FOUND"}}
		respondErr(rCTX, w, err)
		return
	}

	err = auth.ModACL(rCTX, projectUUID, "topics", urlTopic, postBody.AuthUsers, refStr)

	if err != nil {

		if err.Error() == "not found" {
			err := APIErrorNotFound("Topic")
			respondErr(rCTX, w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	respondOK(w, output)

}

// TopicCreate (PUT) creates a new  topic
func TopicCreate(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	postBody := map[string]string{}
	schemaUUID := ""

	// check if there's a request body provided before trying to decode
	if r.Body != nil {

		b, err := ioutil.ReadAll(r.Body)

		if err != nil {
			err := APIErrorInvalidRequestBody()
			respondErr(rCTX, w, err)
			return
		}
		defer r.Body.Close()

		if len(b) > 0 {
			err = json.Unmarshal(b, &postBody)
			if err != nil {
				err := APIErrorInvalidRequestBody()
				respondErr(rCTX, w, err)
				return
			}

			schemaRef := postBody["schema"]

			// if there was a schema name provided, check its existence
			if schemaRef != "" {
				_, schemaName, err := schemas.ExtractSchema(schemaRef)
				if err != nil {
					err := APIErrorInvalidData(err.Error())
					respondErr(rCTX, w, err)
					return
				}
				sl, err := schemas.Find(rCTX, projectUUID, "", schemaName, refStr)
				if err != nil {
					err := APIErrGenericInternal(err.Error())
					respondErr(rCTX, w, err)
					return
				}

				if sl.Empty() {
					err := APIErrorNotFound("Schema")
					respondErr(rCTX, w, err)
					return
				}

				schemaUUID = sl.Schemas[0].UUID
			}
		}
	}

	created := time.Now().UTC()

	// Get Result Object
	res, err := topics.CreateTopic(rCTX, projectUUID, urlVars["topic"], schemaUUID, created, refStr)
	if err != nil {
		if err.Error() == "exists" {
			err := APIErrorConflict("Topic")
			respondErr(rCTX, w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(rCTX, w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)
}

// TopicAttachSchema (POST) attaches an already created schema to the given topic
func TopicAttachSchema(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	postBody := map[string]string{}
	schemaUUID := ""

	results, err := topics.Find(rCTX, projectUUID, "", urlVars["topic"], "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(rCTX, w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Topic")
		respondErr(rCTX, w, err)
		return
	}

	// check if there's a request body provided before trying to decode
	if r.Body != nil {

		b, err := ioutil.ReadAll(r.Body)

		if err != nil {
			err := APIErrorInvalidRequestBody()
			respondErr(rCTX, w, err)
			return
		}
		defer r.Body.Close()

		if len(b) > 0 {
			err = json.Unmarshal(b, &postBody)
			if err != nil {
				err := APIErrorInvalidRequestBody()
				respondErr(rCTX, w, err)
				return
			}

			schemaRef := postBody["schema"]

			// if there was a schema name provided, check its existence
			if schemaRef != "" {
				_, schemaName, err := schemas.ExtractSchema(schemaRef)
				if err != nil {
					err := APIErrorInvalidData(err.Error())
					respondErr(rCTX, w, err)
					return
				}
				sl, err := schemas.Find(rCTX, projectUUID, "", schemaName, refStr)
				if err != nil {
					err := APIErrGenericInternal(err.Error())
					respondErr(rCTX, w, err)
					return
				}

				if sl.Empty() {
					err := APIErrorNotFound("Schema")
					respondErr(rCTX, w, err)
					return
				}

				schemaUUID = sl.Schemas[0].UUID
			}
		}
	}

	err = topics.AttachSchemaToTopic(rCTX, projectUUID, urlVars["topic"], schemaUUID, refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
	}

	// Write response
	output = []byte("")
	respondOK(w, output)
}

// TopicDetachSchema (POST) removes the schema from the given topic
func TopicDetachSchema(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	results, err := topics.Find(rCTX, projectUUID, "", urlVars["topic"], "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(rCTX, w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Topic")
		respondErr(rCTX, w, err)
		return
	}

	err = topics.DetachSchemaFromTopic(rCTX, projectUUID, urlVars["topic"], refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
	}

	// Write response
	output = []byte("")
	respondOK(w, output)
}

// TopicListOne (GET) one topic
func TopicListOne(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	results, err := topics.Find(rCTX, projectUUID, "", urlVars["topic"], "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(rCTX, w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Topic")
		respondErr(rCTX, w, err)
		return
	}

	res := results.Topics[0]

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(rCTX, w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)
}

// ListSubsByTopic (GET) lists all subscriptions associated with the given topic
func ListSubsByTopic(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	results, err := topics.Find(rCTX, projectUUID, "", urlVars["topic"], "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(rCTX, w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Topic")
		respondErr(rCTX, w, err)
		return
	}

	subs, err := subscriptions.FindByTopic(rCTX, projectUUID, results.Topics[0].Name, refStr)
	if err != nil {
		err := APIErrGenericBackend()
		respondErr(rCTX, w, err)
		return

	}

	resJSON, err := json.Marshal(subs)
	if err != nil {
		err := APIErrExportJSON()
		respondErr(rCTX, w, err)
		return
	}

	respondOK(w, resJSON)
}

// TopicACL (GET) one topic's authorized users
func TopicACL(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	urlTopic := urlVars["topic"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)
	res, err := auth.GetACL(rCTX, projectUUID, "topics", urlTopic, refStr)

	// If not found
	if err != nil {
		err := APIErrorNotFound("Topic")
		respondErr(rCTX, w, err)
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(rCTX, w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)
}

// TopicListAll (GET) all topics
func TopicListAll(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

	var err error
	var strPageSize string
	var pageSize int
	var res topics.PaginatedTopics

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)
	roles := gorillaContext.Get(r, "auth_roles").([]string)

	urlValues := r.URL.Query()
	pageToken := urlValues.Get("pageToken")
	strPageSize = urlValues.Get("pageSize")

	// if this route is used by a user who only  has a publisher role
	// return all topics that he has access to
	userUUID := ""
	if !auth.IsProjectAdmin(roles) && !auth.IsServiceAdmin(roles) && auth.IsPublisher(roles) {
		userUUID = gorillaContext.Get(r, "auth_user_uuid").(string)
	}

	if strPageSize != "" {
		if pageSize, err = strconv.Atoi(strPageSize); err != nil {
			log.WithFields(
				log.Fields{
					"trace_id":  rCTX.Value("trace_id"),
					"type":      "request_log",
					"page_size": pageSize,
					"error":     err.Error(),
				},
			).Error("error while converting page size to int")
			err := APIErrorInvalidData("Invalid page size")
			respondErr(rCTX, w, err)
			return
		}
	}

	if res, err = topics.Find(rCTX, projectUUID, userUUID, "", pageToken, int64(pageSize), refStr); err != nil {
		err := APIErrorInvalidData("Invalid page token")
		respondErr(rCTX, w, err)
		return
	}
	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(rCTX, w, err)
		return
	}

	// Write Response
	output = []byte(resJSON)
	respondOK(w, output)
}

// TopicPublish (POST) publish messages to a topic
func TopicPublish(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)
	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Get url path variables
	urlVars := mux.Vars(r)
	urlTopic := urlVars["topic"]

	// Grab context references

	refBrk := gorillaContext.Get(r, "brk").(brokers.Broker)
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refUserUUID := gorillaContext.Get(r, "auth_user_uuid").(string)
	refRoles := gorillaContext.Get(r, "auth_roles").([]string)
	refAuthResource := gorillaContext.Get(r, "auth_resource").(bool)
	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	results, err := topics.Find(rCTX, projectUUID, "", urlVars["topic"], "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(rCTX, w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Topic")
		respondErr(rCTX, w, err)
		return
	}

	res := results.Topics[0]

	// Check Authorization per topic
	// - if enabled in config
	// - if user has only publisher role

	if refAuthResource && auth.IsPublisher(refRoles) {

		if auth.PerResource(rCTX, projectUUID, "topics", urlTopic, refUserUUID, refStr) == false {
			err := APIErrorForbidden()
			respondErr(rCTX, w, err)
			return
		}
	}

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(rCTX, w, err)
		return
	}

	// Create Message List from Post JSON
	msgList, err := messages.LoadMsgListJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("Message")
		respondErr(rCTX, w, err)
		return
	}

	// check if the topic has a schema associated with it
	if res.Schema != "" {

		// retrieve the schema
		_, schemaName, err := schemas.ExtractSchema(res.Schema)
		if err != nil {
			log.WithFields(
				log.Fields{
					"trace_id":    rCTX.Value("trace_id"),
					"type":        "service_log",
					"schema_name": res.Schema,
					"topic_name":  res.Name,
					"error":       err.Error(),
				},
			).Error("Could not extract schema name")
			err := APIErrGenericInternal(schemas.GenericError)
			respondErr(rCTX, w, err)
			return
		}

		sl, err := schemas.Find(rCTX, projectUUID, "", schemaName, refStr)

		if err != nil {
			log.WithFields(
				log.Fields{
					"trace_id":    rCTX.Value("trace_id"),
					"type":        "service_log",
					"schema_name": schemaName,
					"topic_name":  res.Name,
					"error":       err.Error(),
				},
			).Error("Could not retrieve schema from the store")
			err := APIErrGenericInternal(schemas.GenericError)
			respondErr(rCTX, w, err)
			return
		}

		if !sl.Empty() {
			err := schemas.ValidateMessages(sl.Schemas[0], msgList)
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
		} else {
			log.WithFields(
				log.Fields{
					"trace_id":    rCTX.Value("trace_id"),
					"type":        "service_log",
					"schema_name": res.Schema,
					"topic_name":  res.Name,
				},
			).Error("List of schemas was empty")
			err := APIErrGenericInternal(schemas.GenericError)
			respondErr(rCTX, w, err)
			return
		}
	}

	// Init message ids list
	msgIDs := messages.MsgIDs{IDs: []string{}}

	// For each message in message list
	for _, msg := range msgList.Msgs {
		// Get offset and set it as msg
		fullTopic := projectUUID + "." + urlTopic

		msgID, rTop, _, _, err := refBrk.Publish(rCTX, fullTopic, msg)

		if err != nil {
			if err.Error() == "kafka server: Message was too large, server rejected it to avoid allocation error." {
				err := APIErrTooLargeMessage("Message size too large")
				respondErr(rCTX, w, err)
				return
			}

			err := APIErrGenericBackend()
			respondErr(rCTX, w, err)
			return
		}

		msg.ID = msgID
		// Assertions for Successful Publish
		if rTop != fullTopic {
			err := APIErrGenericInternal("Broker reports wrong topic")
			respondErr(rCTX, w, err)
			return
		}

		// Append the MsgID of the successful published message to the msgIds list
		msgIDs.IDs = append(msgIDs.IDs, msg.ID)
	}

	// timestamp of the publish event
	publishTime := time.Now().UTC()

	// amount of messages published
	msgCount := int64(len(msgList.Msgs))

	// increment topic number of message metric
	refStr.IncrementTopicMsgNum(rCTX, projectUUID, urlTopic, msgCount)

	// increment daily count of topic messages
	year, month, day := publishTime.Date()
	refStr.IncrementDailyTopicMsgCount(rCTX, projectUUID, urlTopic, msgCount, time.Date(year, month, day, 0, 0, 0, 0, time.UTC))

	// increment topic total bytes published
	refStr.IncrementTopicBytes(rCTX, projectUUID, urlTopic, msgList.TotalSize())

	// update latest publish date for the given topic
	refStr.UpdateTopicLatestPublish(rCTX, projectUUID, urlTopic, publishTime)

	// count the rate of published messages per sec between the last two publish events
	var dt float64 = 1
	// if its the first publish to the topic
	// skip the subtraction that computes the DT between the last two publish events
	if !res.LatestPublish.IsZero() {
		dt = publishTime.Sub(res.LatestPublish).Seconds()
	}
	refStr.UpdateTopicPublishRate(rCTX, projectUUID, urlTopic, float64(msgCount)/dt)

	// Export the msgIDs
	resJSON, err := msgIDs.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(rCTX , w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)
}
