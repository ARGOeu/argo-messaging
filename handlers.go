package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ARGOeu/argo-messaging/auth"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/messages"
	"github.com/ARGOeu/argo-messaging/push"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/ARGOeu/argo-messaging/subscriptions"
	"github.com/ARGOeu/argo-messaging/topics"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

// HandlerWrappers
//////////////////

// WrapValidate handles validation
func WrapValidate(hfn http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		urlVars := mux.Vars(r)

		// sort keys
		keys := []string(nil)
		for key := range urlVars {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		// Iterate alphabetically
		for _, key := range keys {
			if validName(urlVars[key]) == false {
				respondErr(w, 400, "Invalid "+key+" name", "INVALID_ARGUMENT")
				return
			}
		}
		hfn.ServeHTTP(w, r)

	})
}

// WrapConfig handle wrapper to retrieve kafka configuration
func WrapConfig(hfn http.HandlerFunc, cfg *config.APICfg, brk brokers.Broker, str stores.Store, mgr *push.Manager) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		nStr := str.Clone()
		defer nStr.Close()
		context.Set(r, "brk", brk)
		context.Set(r, "str", nStr)
		context.Set(r, "mgr", mgr)
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

// WrapAuthenticate handle wrapper to apply authentication
func WrapAuthenticate(hfn http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		urlVars := mux.Vars(r)
		urlValues := r.URL.Query()

		refStr := context.Get(r, "str").(stores.Store)

		roles := auth.Authenticate(urlVars["project"], urlValues.Get("key"), refStr)

		if len(roles) > 0 {
			context.Set(r, "auth_roles", roles)
			hfn.ServeHTTP(w, r)
		} else {
			respondErr(w, 401, "Unauthorized", "UNAUTHORIZED")
		}

	})
}

// WrapAuthorize handle wrapper to apply authentication
func WrapAuthorize(hfn http.Handler, routeName string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		refStr := context.Get(r, "str").(stores.Store)
		refRoles := context.Get(r, "auth_roles").([]string)

		if auth.Authorize(routeName, refRoles, refStr) {
			hfn.ServeHTTP(w, r)
		} else {
			respondErr(w, 403, "Access to this resource is forbidden", "FORBIDDEN")
		}

	})
}

// HandlerFunctions
///////////////////

// SubAck (GET) one subscription
func SubAck(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)

	// Initialize Subscription
	sub := subscriptions.Subscriptions{}
	sub.LoadFromStore(refStr)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 500, "Bad request body", "BAD_REQUEST")
		return
	}

	// Parse pull options
	postBody, err := subscriptions.GetAckFromJSON(body)
	if err != nil {
		respondErr(w, 400, "Invalid ack parameter", "INVALID_ARGUMENT")
		return
	}

	// Check if sub exists

	if sub.HasSub(urlVars["project"], urlVars["subscription"]) == false {
		respondErr(w, 404, "Subscription does not exist", "NOT_FOUND")
	}

	// Get Ack
	if postBody.IDs == nil {
		respondErr(w, 400, "Invalid ack id", "INVALID_ARGUMENT")
		return
	}

	ack := postBody.IDs[0]

	items := strings.Split(ack, "/")
	if len(items) != 4 || items[0] != "projects" || items[1] != urlVars["project"] || items[2] != "subscriptions" {
		respondErr(w, 400, "Invalid ack id", "INVALID_ARGUMENT")
		return
	}

	subItems := strings.Split(items[3], ":")
	if len(subItems) != 2 || subItems[0] != urlVars["subscription"] {
		respondErr(w, 400, "Invalid ack id", "INVALID_ARGUMENT")
		return
	}

	off, err := strconv.Atoi(subItems[1])
	if err != nil {
		respondErr(w, 400, "Invalid ack id", "INVALID_ARGUMENT")
		return
	}

	zSec := "2006-01-02T15:04:05Z"
	t := time.Now()
	ts := t.Format(zSec)

	err = refStr.UpdateSubOffsetAck(urlVars["subscription"], int64(off+1), ts)
	if err != nil {

		if err.Error() == "ack timeout" {
			respondErr(w, 408, err.Error(), "TIMEOUT")
			return
		}

		respondErr(w, 400, err.Error(), "INTERNAL")
		return
	}

	// Output result to JSON
	resJSON := "{}"

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// SubListOne (GET) one subscription
func SubListOne(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)

	// Initialize Subscription
	sub := subscriptions.Subscriptions{}
	sub.LoadFromStore(refStr)

	// Get Result Object
	res := subscriptions.Subscription{}
	res = sub.GetSubByName(urlVars["project"], urlVars["subscription"])

	// If not found
	if res.Name == "" {
		respondErr(w, 404, "Subscription does not exist", "NOT_FOUND")
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()

	if err != nil {
		respondErr(w, 500, "Error exporting data", "INTERNAL")
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// TopicDelete (DEL) deletes an existing topic
func TopicDelete(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)
	// Initialize Topics
	tp := topics.Topics{}
	tp.LoadFromStore(refStr)

	// Get Result Object
	err := tp.RemoveTopic(urlVars["project"], urlVars["topic"], refStr)
	if err != nil {
		if err.Error() == "not found" {
			respondErr(w, 404, "Topic doesn't exist", "NOT_FOUND")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL")
	}

	// Write empty response if anything ok
	respondOK(w, output)

}

// SubDelete (DEL) deletes an existing subscription
func SubDelete(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)
	// Initialize subs
	sb := subscriptions.Subscriptions{}
	sb.LoadFromStore(refStr)

	// Get Result Object
	err := sb.RemoveSub(urlVars["project"], urlVars["subscription"], refStr)

	if err != nil {

		if err.Error() == "not found" {
			respondErr(w, 404, "Subscription doesn't exist", "NOT_FOUND")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL")
		return
	}

	// Write empty response if anything ok
	respondOK(w, output)

}

// SubModPush (PUT) modifies the push configuration
func SubModPush(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	refMgr := context.Get(r, "mgr").(*push.Manager)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 500, "Bad Request body", "BAD_REQUEST")
		return
	}

	// Parse pull options
	postBody, err := subscriptions.GetFromJSON(body)
	if err != nil {
		respondErr(w, 400, "Invalid Subscription Arguments", "INVALID_ARGUMENT")
		return
	}

	pushEnd := ""
	rPolicy := ""
	rPeriod := 0
	if postBody.PushCfg != (subscriptions.PushConfig{}) {
		pushEnd = postBody.PushCfg.Pend
		rPolicy = postBody.PushCfg.RetPol.PolicyType
		rPeriod = postBody.PushCfg.RetPol.Period
		if rPolicy == "" {
			rPolicy = "linear"
		}
		if rPeriod <= 0 {
			rPeriod = 3000
		}
	}

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)
	// Initialize subs
	sb := subscriptions.Subscriptions{}
	sb.LoadFromStore(refStr)
	// Get Result Object
	subName := urlVars["subscription"]
	project := urlVars["project"]

	// Get Result Object
	old := subscriptions.Subscription{}
	old = sb.GetSubByName(project, subName)

	err = sb.ModSubPush(project, subName, pushEnd, rPolicy, rPeriod, refStr)

	if err != nil {

		if err.Error() == "not found" {
			respondErr(w, 404, "Subscription doesn't exist", "NOT_FOUND")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL")
		return
	}

	// According to push cfg set start/stop pushing
	if pushEnd != "" {
		if old.PushCfg.Pend == "" {
			refMgr.Add(project, subName)
			refMgr.Launch(project, subName)
		} else if old.PushCfg.Pend != pushEnd {
			refMgr.Restart(project, subName)
		} else if old.PushCfg.RetPol.PolicyType != rPolicy || old.PushCfg.RetPol.Period != rPeriod {
			refMgr.Restart(project, subName)
		}
	} else {
		refMgr.Stop(project, subName)
	}

	// Write empty response if anything ok
	respondOK(w, output)

}

// SubCreate (PUT) creates a new subscription
func SubCreate(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)
	refBrk := context.Get(r, "brk").(brokers.Broker)
	refMgr := context.Get(r, "mgr").(*push.Manager)

	// Initialize Subscriptions
	sb := subscriptions.Subscriptions{}
	sb.LoadFromStore(refStr)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 500, "Bad Request body", "BAD_REQUEST")
		return
	}

	// Parse pull options
	postBody, err := subscriptions.GetFromJSON(body)
	if err != nil {
		respondErr(w, 400, "Invalid Subscription Arguments", "INVALID_ARGUMENT")
		return
	}

	tProject, tName, err := subscriptions.ExtractFullTopicRef(postBody.FullTopic)

	if err != nil {
		respondErr(w, 400, "Invalid Topic Name", "INVALID_ARGUMENT")
		return
	}

	// Initialize Topics
	tp := topics.Topics{}
	tp.LoadFromStore(refStr)

	if tp.HasTopic(tProject, tName) == false {
		respondErr(w, 404, "Topic doesn't exist", "NOT_FOUND")
		return
	}

	// Get current topic offset

	fullTopic := tProject + "." + tName
	curOff := refBrk.GetOffset(fullTopic)

	pushEnd := ""
	rPolicy := ""
	rPeriod := 0
	if postBody.PushCfg != (subscriptions.PushConfig{}) {
		pushEnd = postBody.PushCfg.Pend
		rPolicy = postBody.PushCfg.RetPol.PolicyType
		rPeriod = postBody.PushCfg.RetPol.Period
		if rPolicy == "" {
			rPolicy = "linear"
		}
		if rPeriod <= 0 {
			rPeriod = 3000
		}
	}

	// Get Result Object
	res, err := sb.CreateSub(urlVars["project"], urlVars["subscription"], tName, pushEnd, curOff, postBody.Ack, rPolicy, rPeriod, refStr)

	if err != nil {
		if err.Error() == "exists" {
			respondErr(w, 409, "Subscription already exists", "ALREADY_EXISTS")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL")
	}

	// Enable pushManager if subscription has pushConfiguration
	if pushEnd != "" {
		refMgr.Add(res.Project, res.Name)
		refMgr.Launch(res.Project, res.Name)
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		respondErr(w, 500, "Error exporting data to JSON", "INTERNAL")
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// TopicCreate (PUT) creates a new  topic
func TopicCreate(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)
	// Initialize Topics
	tp := topics.Topics{}
	tp.LoadFromStore(refStr)

	// Get Result Object
	res, err := tp.CreateTopic(urlVars["project"], urlVars["topic"], refStr)
	if err != nil {
		if err.Error() == "exists" {
			respondErr(w, 409, "Topic already exists", "ALREADY_EXISTS")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL")
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		respondErr(w, 500, "Error Exporting Retrieved Data to JSON", "INTERNAL")
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

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
	refStr := context.Get(r, "str").(stores.Store)
	// Initialize Topics
	tp := topics.Topics{}
	tp.LoadFromStore(refStr)

	// Get Result Object
	res := topics.Topic{}
	res = tp.GetTopicByName(urlVars["project"], urlVars["topic"])

	// If not found
	if res.Name == "" {
		respondErr(w, 404, "Topic does not exist", "NOT_FOUND")
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		respondErr(w, 500, "Error exporting data to JSON", "INTERNAL")
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

//SubListAll (GET) all subscriptions
func SubListAll(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)

	// Initialize Subscriptions
	sb := subscriptions.Subscriptions{}
	sb.LoadFromStore(refStr)

	// Get result object
	res := subscriptions.Subscriptions{}
	res = sb.GetSubsByProject(urlVars["project"])

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		respondErr(w, 500, "Error exporting data to JSON", "INTERNAL")
		return
	}

	// Write Response
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
	refStr := context.Get(r, "str").(stores.Store)

	// Initialize Topics
	tp := topics.Topics{}
	tp.LoadFromStore(refStr)

	// Get result object
	res := topics.Topics{}
	res = tp.GetTopicsByProject(urlVars["project"])

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		respondErr(w, 500, "Error exporting data to JSON", "INTERNAL")
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
	refBrk := context.Get(r, "brk").(brokers.Broker)
	refStr := context.Get(r, "str").(stores.Store)

	// Create Topics Object
	tp := topics.Topics{}
	tp.LoadFromStore(refStr)

	// Check if Project/Topic exist
	if tp.HasTopic(urlVars["project"], urlVars["topic"]) == false {
		respondErr(w, 404, "Topic doesn't exist", "NOT_FOUND")
		return
	}

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 500, "Bad Request Body", "BAD REQUEST")
		return
	}

	// Create Message List from Post JSON
	msgList, err := messages.LoadMsgListJSON(body)
	if err != nil {
		respondErr(w, 400, "Invalid Message Arguments", "INVALID_ARGUMENT")
		return
	}

	// Init message ids list
	msgIDs := messages.MsgIDs{}

	// For each message in message list
	for _, msg := range msgList.Msgs {
		// Get offset and set it as msg
		fullTopic := urlVars["project"] + "." + urlVars["topic"]
		off := refBrk.GetOffset(fullTopic)
		msg.ID = strconv.FormatInt(off, 10)
		// Stamp time to UTC Z to nanoseconds
		zNano := "2006-01-02T15:04:05.999999999Z"
		t := time.Now()
		msg.PubTime = t.Format(zNano)

		// Publish the message
		payload, err := msg.ExportJSON()
		if err != nil {
			respondErr(w, 500, "Error during data export to JSON", "INTERNAL")
			return
		}

		rTop, _, rOff := refBrk.Publish(fullTopic, payload)

		// Assertions for Succesfull Publish
		if rTop != fullTopic {
			respondErr(w, 500, "Broker reports wrong topic", "INTERNAL")
			return
		}

		// if rPart != 0 {
		// 	respondErr(w, 500, "Broker reports wrong partition", "INTERNAL")
		// 	return
		// }

		if rOff != off {
			respondErr(w, 500, "Broker reports wrong offset", "INTERNAL")
			return
		}

		// Append the MsgID of the successful published message to the msgIds list
		msgIDs.IDs = append(msgIDs.IDs, msg.ID)
	}

	// Export the msgIDs
	resJSON, err := msgIDs.ExportJSON()
	if err != nil {
		respondErr(w, 500, "Error during export data to JSON", "INTERNAL")
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)
}

// SubPull (POST) publish a new topic
func SubPull(w http.ResponseWriter, r *http.Request) {
	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Get url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refBrk := context.Get(r, "brk").(brokers.Broker)
	refStr := context.Get(r, "str").(stores.Store)

	// Create Subscriptions Object
	sub := subscriptions.Subscriptions{}
	sub.LoadFromStore(refStr)

	// Check if Project/Topic exist
	if sub.HasSub(urlVars["project"], urlVars["subscription"]) == false {
		respondErr(w, 404, "Subscription doesn't exist", "NOT_FOUND")
		return
	}

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 500, "Bad Request Body", "BAD_REQUEST")
		return
	}

	// Parse pull options
	pullInfo, err := subscriptions.GetPullOptionsJSON(body)
	if err != nil {
		respondErr(w, 400, "Pull Parameters Invalid", "INVALID_ARGUMENT")
		return
	}

	// Init Received Message List
	recList := messages.RecList{}

	// Get the subscription info
	targSub := sub.GetSubByName(urlVars["project"], urlVars["subscription"])

	fullTopic := targSub.Project + "." + targSub.Topic
	retImm := false
	if pullInfo.RetImm == "true" {
		retImm = true
	}
	msgs := refBrk.Consume(fullTopic, targSub.Offset, retImm)

	var limit int
	limit, err = strconv.Atoi(pullInfo.MaxMsg)
	if err != nil {
		limit = 0
	}

	ackPrefix := "projects/" + urlVars["project"] + "/subscriptions/" + urlVars["subscription"] + ":"

	for i, msg := range msgs {
		if limit > 0 && i >= limit {
			break // max messages left
		}
		curMsg, err := messages.LoadMsgJSON([]byte(msg))
		if err != nil {
			respondErr(w, 500, "Message retrieved from broker network has invalid JSON Structure", "INTERNAL")
			return
		}

		curRec := messages.RecMsg{AckID: ackPrefix + curMsg.ID, Msg: curMsg}
		recList.RecMsgs = append(recList.RecMsgs, curRec)
	}

	resJSON, err := recList.ExportJSON()

	if err != nil {
		respondErr(w, 500, "Error during exporting message to JSON", "INTERNAL")
		return
	}

	// Stamp time to UTC Z to seconds
	zSec := "2006-01-02T15:04:05Z"
	t := time.Now()
	ts := t.Format(zSec)
	refStr.UpdateSubPull(targSub.Name, int64(len(recList.RecMsgs))+targSub.Offset, ts)

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
func respondErr(w http.ResponseWriter, errCode int, errMsg string, status string) {
	log.Printf("ERROR\t%d\t%s", errCode, errMsg)
	w.WriteHeader(errCode)
	rt := APIErrorRoot{}
	bd := APIErrorBody{}
	//em := APIError{}
	//em.Message = errMsg
	//em.Domain = "global"
	//em.Reason = "backend"
	bd.Code = errCode
	bd.Message = errMsg
	//bd.ErrList = append(bd.ErrList, em)
	bd.Status = status
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
	ErrList []APIError `json:"errors,omitempty"`
	Status  string     `json:"status"`
}

// APIError represents array items for error list array
type APIError struct {
	Message string `json:"message"`
	Domain  string `json:"domain"`
	Reason  string `json:"reason"`
}
