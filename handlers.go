package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/gorilla/context"
	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/gorilla/mux"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/messages"
	"github.com/ARGOeu/argo-messaging/topics"
)

// HandlerWrappers
//////////////////

// WrapConfig handle wrapper to retrieve kafka configuration
func WrapConfig(hfn http.HandlerFunc, cfg *config.KafkaCfg, brk brokers.Broker) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context.Set(r, "cfg", cfg)
		context.Set(r, "brk", brk)
		hfn.ServeHTTP(w, r)

	})
}

// WrapLog handle wrapper to apply Logging
func WrapLog(hfn http.Handler, name string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		hfn.ServeHTTP(w, r)

		log.Printf(
			"ACCESS\t%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}

// HandlerFunctions
///////////////////

// TopicListOne (GET) one topic
func TopicListOne(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refCfg := context.Get(r, "cfg").(*config.KafkaCfg)

	// Initialize Topics
	tp := topics.Topics{}
	tp.LoadFromCfg(refCfg)

	// Get Result Object
	res := topics.Topic{}
	res = tp.GetTopicByName(urlVars["project"], urlVars["topic"])

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		respondErr(w, 500, "Error Exporting Retrieved Data to JSON")
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// TopicListAll (GET) all topics
func TopicListAll(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refCfg := context.Get(r, "cfg").(*config.KafkaCfg)

	// Initialize Topics
	tp := topics.Topics{}
	tp.LoadFromCfg(refCfg)

	// Get result object
	res := topics.Topics{}
	res = tp.GetTopicsByProject(urlVars["project"])

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		respondErr(w, 500, "RESPONSE: Error Exporting Retrieved Data to JSON")
		return
	}

	// Write Response
	output = []byte(resJSON)
	respondOK(w, output)

}

// TopicPublish (POST) publish a new topic
func TopicPublish(w http.ResponseWriter, r *http.Request) {
	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Get url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refCfg := context.Get(r, "cfg").(*config.KafkaCfg)
	refBrk := context.Get(r, "brk").(brokers.Broker)

	// Create Topics Object
	tp := topics.Topics{}
	tp.LoadFromCfg(refCfg)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 500, "POST: Could not read POST Body")
		return
	}

	// Create Message List from Post JSON
	msgList, err := messages.LoadMsgListJSON(body)
	if err != nil {
		respondErr(w, 500, "POST: Input JSON schema is not valid")
		return
	}

	// Init message ids list
	msgIDs := messages.MsgIDs{}

	// For each message in message list
	for _, msg := range msgList.Msgs {
		// Get offset and set it as msg
		off := refBrk.GetOffset(urlVars["topic"])
		msg.ID = strconv.FormatInt(off, 10)
		// Stamp time to UTC Z to nanoseconds
		zNano := "2006-01-02T15:04:05.999999999Z"
		t := time.Now()
		msg.PubTime = t.Format(zNano)

		// Publish the message
		payload, err := msg.ExportJSON()
		if err != nil {
			respondErr(w, 500, "PUBLISH: Error during exporting message to JSON")
			return
		}
		rTop, rPart, rOff := refBrk.Publish(urlVars["topic"], payload)

		// Assertions for Succesfull Publish
		if rTop != urlVars["topic"] {
			respondErr(w, 500, "PUBLISH: Broker reports wrong topic")
			return
		}

		if rPart != 0 {
			respondErr(w, 500, "PUBLISH: Broker reports wrong partition")
			return
		}

		if rOff != off {
			respondErr(w, 500, "PUBLISH: Broker reports wrong offset")
			return
		}

		// Append the MsgID of the successful published message to the msgIds list
		msgIDs.IDs = append(msgIDs.IDs, msg.ID)
	}

	// Export the msgIDs
	resJSON, err := msgIDs.ExportJSON()
	if err != nil {
		respondErr(w, 500, "RESPONSE: Error during exporting message to JSON")
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)
}

// Respond utility functions
///////////////////////////////

// respondOK is used to finalize response writer with proper code and output
func respondOK(w http.ResponseWriter, output []byte) {
	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

// respondErr is used to finalize response writer with proper error codes and error output
func respondErr(w http.ResponseWriter, errCode int, errMsg string) {
	log.Printf("ERROR\t%d\t%s", errCode, errMsg)
	w.WriteHeader(errCode)
	rt := APIErrorRoot{}
	bd := APIErrorBody{}
	em := APIError{}
	em.Message = errMsg
	em.Domain = "global"
	em.Reason = "backend"
	bd.Code = errCode
	bd.Message = errMsg
	bd.ErrList = append(bd.ErrList, em)
	bd.Status = "INTERNAL"
	rt.Body = bd
	// Output API Erorr object to JSON
	output, _ := json.MarshalIndent(rt, "", "   ")
	w.Write(output)
}

// APIErrorRoot holds the root json object of an error response
type APIErrorRoot struct {
	Body APIErrorBody `json:"error"`
}

// APIErrorBody represents the inner json body of the error response
type APIErrorBody struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	ErrList []APIError `json:"errors"`
	Status  string     `json:"status"`
}

// APIError represents array items for error list array
type APIError struct {
	Message string `json:"message"`
	Domain  string `json:"domain"`
	Reason  string `json:"reason"`
}
