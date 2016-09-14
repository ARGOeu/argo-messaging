package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/ARGOeu/argo-messaging/auth"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/messages"
	"github.com/ARGOeu/argo-messaging/projects"
	"github.com/ARGOeu/argo-messaging/push"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/ARGOeu/argo-messaging/subscriptions"
	"github.com/ARGOeu/argo-messaging/topics"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/twinj/uuid"
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

// WrapMockAuthConfig handle wrapper is used in tests were some auth context is needed
func WrapMockAuthConfig(hfn http.HandlerFunc, cfg *config.APICfg, brk brokers.Broker, str stores.Store, mgr *push.Manager) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		nStr := str.Clone()
		defer nStr.Close()
		context.Set(r, "brk", brk)
		context.Set(r, "str", nStr)
		context.Set(r, "mgr", mgr)
		context.Set(r, "auth_resource", cfg.ResAuth)
		context.Set(r, "auth_user", "userA")
		context.Set(r, "auth_roles", []string{"publisher", "consumer"})
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
		context.Set(r, "auth_resource", cfg.ResAuth)
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

		roles, user := auth.Authenticate(urlVars["project"], urlValues.Get("key"), refStr)

		if len(roles) > 0 {
			context.Set(r, "auth_roles", roles)
			context.Set(r, "auth_user", user)
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

// ProjectCreate (POST) creates a new project
func ProjectCreate(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	urlProject := urlVars["project"]

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)
	refUser := context.Get(r, "auth_user").(string)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 400, "Invalid Request body", "INVALID_ARGUMENT")
		return
	}

	// Parse pull options
	postBody, err := projects.GetFromJSON(body)
	if err != nil {
		respondErr(w, 400, "Invalid Project Arguments", "INVALID_ARGUMENT")
		return
	}

	uuid := uuid.NewV4().String() // generate a new uuid to attach to the new project
	user := refUser               // log current authenticated user as the creator
	created := time.Now()
	// Get Result Object
	res, err := projects.CreateProject(uuid, urlProject, created, user, postBody.Description, refStr)

	if err != nil {
		if err.Error() == "exists" {
			respondErr(w, 409, "Project already exists", "ALREADY_EXISTS")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL_SERVER_ERROR")
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		respondErr(w, 500, "Error exporting data to JSON", "INTERNAL_SERVER_ERROR")
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// ProjectListAll (GET) all projects
func ProjectListAll(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)

	// Get Results Object
	res, err := projects.Find("", refStr)

	if err != nil && err.Error() != "not found" {
		respondErr(w, 500, "Internal error while querying datastore", "INTERNAL_SERVER_ERROR")
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()

	if err != nil {
		respondErr(w, 500, "Error exporting data", "INTERNAL_SERVER_ERROR")
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// ProjectListOne (GET) one project
func ProjectListOne(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	urlProject := urlVars["project"]

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)

	// Get Results Object
	res, err := projects.Find(urlProject, refStr)

	if err != nil {
		if err.Error() == "not found" {
			respondErr(w, 404, "Project does not exist", "NOT_FOUND")
			return
		}

		respondErr(w, 500, "Internal error while querying datastore", "INTERNAL_SERVER_ERROR")
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()

	if err != nil {
		respondErr(w, 500, "Error exporting data", "INTERNAL_SERVER_ERROR")
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

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

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 400, "Invalid request body", "INVALID_ARGUMENT")
		return
	}

	// Parse pull options
	postBody, err := subscriptions.GetAckFromJSON(body)
	if err != nil {
		respondErr(w, 400, "Invalid ack parameter", "INVALID_ARGUMENT")
		return
	}

	// Get urlParams
	projectName := urlVars["project"]
	subName := urlVars["subscription"]

	// Check if sub exists

	projectUUID := projects.GetUUIDByName(urlVars["project"], refStr)
	if subscriptions.HasSub(projectUUID, urlVars["subscription"], refStr) == false {
		respondErr(w, 404, "Subscription does not exist", "NOT_FOUND")
		return
	}

	// Get list of AckIDs
	if postBody.IDs == nil {
		respondErr(w, 400, "Invalid ack id", "INVALID_ARGUMENT")
		return
	}

	// Check if each AckID is valid
	for _, ackID := range postBody.IDs {
		if validAckID(projectName, subName, ackID) == false {
			respondErr(w, 400, "Invalid ack id", "INVALID_ARGUMENT")
			return
		}
	}

	// Get Max ackID
	maxAckID, err := subscriptions.GetMaxAckID(postBody.IDs)
	if err != nil {
		respondErr(w, 500, "Error handling acknowledgement", "INTERNAL_SERVER_ERROR")
		return
	}
	// Extract offset from max ackID
	off, err := subscriptions.GetOffsetFromAckID(maxAckID)

	if err != nil {
		respondErr(w, 400, "Invalid ack id", "INVALID_ARGUMENT")
		return
	}

	zSec := "2006-01-02T15:04:05Z"
	t := time.Now()
	ts := t.Format(zSec)

	projectUUID = projects.GetUUIDByName(urlVars["project"], refStr)
	err = refStr.UpdateSubOffsetAck(projectUUID, urlVars["subscription"], int64(off+1), ts)
	if err != nil {

		if err.Error() == "ack timeout" {
			respondErr(w, 408, err.Error(), "TIMEOUT")
			return
		}

		respondErr(w, 400, err.Error(), "INTERNAL_SERVER_ERROR")
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

	projectUUID := projects.GetUUIDByName(urlVars["project"], refStr)
	results, err := subscriptions.Find(projectUUID, urlVars["subscription"], refStr)

	if err != nil {
		respondErr(w, 500, "Backend Error", "INTERNAL_SERVER_ERROR")
	}

	// If not found
	if results.Empty() {
		respondErr(w, 404, "Subscription does not exist", "NOT_FOUND")
		return
	}

	// Output result to JSON
	resJSON, err := results.List[0].ExportJSON()

	if err != nil {
		respondErr(w, 500, "Error exporting data", "INTERNAL_SERVER_ERROR")
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

	// Get Result Object
	// Get project UUID First to use as reference
	projectUUID := projects.GetUUIDByName(urlVars["project"], refStr)
	err := topics.RemoveTopic(projectUUID, urlVars["topic"], refStr)
	if err != nil {
		if err.Error() == "not found" {
			respondErr(w, 404, "Topic doesn't exist", "NOT_FOUND")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL_SERVER_ERROR")
		return
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
	refMgr := context.Get(r, "mgr").(*push.Manager)

	// Get Result Object
	projectUUID := projects.GetUUIDByName(urlVars["project"], refStr)
	err := subscriptions.RemoveSub(projectUUID, urlVars["subscription"], refStr)

	if err != nil {

		if err.Error() == "not found" {
			respondErr(w, 404, "Subscription doesn't exist", "NOT_FOUND")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL_SERVER_ERROR")
		return
	}

	refMgr.Stop(urlVars["project"], urlVars["subscription"])

	// Write empty response if anything ok
	respondOK(w, output)

}

// TopicModACL (PUT) modifies the ACL
func TopicModACL(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 400, "Invalid Request body", "INVALID_ARGUMENT")
		return
	}

	// Parse pull options
	postBody, err := topics.GetACLFromJSON(body)
	if err != nil {
		respondErr(w, 400, "Invalid Topic ACL Arguments", "INVALID_ARGUMENT")
		return
	}

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)

	// Get Result Object
	subName := urlVars["topic"]
	project := urlVars["project"]

	// check if user list contain valid users for the given project
	_, err = auth.AreValidUsers(project, postBody.AuthUsers, refStr)
	if err != nil {
		respondErr(w, 404, err.Error(), "NOT_FOUND")
		return
	}

	err = topics.ModACL(project, subName, postBody.AuthUsers, refStr)

	if err != nil {

		if err.Error() == "not found" {
			respondErr(w, 404, "Topic doesn't exist", "NOT_FOUND")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL_SERVER_ERROR")
		return
	}

	respondOK(w, output)

}

// SubModACL (PUT) modifies the ACL
func SubModACL(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 400, "Invalid Request body", "INVALID_ARGUMENT")
		return
	}

	// Parse pull options
	postBody, err := subscriptions.GetACLFromJSON(body)
	if err != nil {
		respondErr(w, 400, "Invalid Subscription ACL Arguments", "INVALID_ARGUMENT")
		return
	}

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)

	// Get Result Object
	subName := urlVars["subscription"]
	project := urlVars["project"]

	// check if user list contain valid users for the given project
	_, err = auth.AreValidUsers(project, postBody.AuthUsers, refStr)
	if err != nil {
		respondErr(w, 404, err.Error(), "NOT_FOUND")
		return
	}

	err = subscriptions.ModACL(project, subName, postBody.AuthUsers, refStr)

	if err != nil {

		if err.Error() == "not found" {
			respondErr(w, 404, "Subscription doesn't exist", "NOT_FOUND")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL_SERVER_ERROR")
		return
	}

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
		respondErr(w, 400, "Invalid Request body", "INVALID_ARGUMENT")
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

	// Get Result Object
	subName := urlVars["subscription"]
	project := urlVars["project"]

	projectUUID := projects.GetUUIDByName(urlVars["project"], refStr)
	res, err := subscriptions.Find(projectUUID, subName, refStr)

	if err != nil {
		respondErr(w, 500, "Backend Error", "INTERNAL_SERVER_ERROR")
		return
	}

	if res.Empty() {
		respondErr(w, 404, "Subscription doesn't exist", "NOT_FOUND")
		return
	}
	old := res.List[0]

	err = subscriptions.ModSubPush(projectUUID, subName, pushEnd, rPolicy, rPeriod, refStr)

	if err != nil {

		if err.Error() == "not found" {
			respondErr(w, 404, "Subscription doesn't exist", "NOT_FOUND")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL_SERVER_ERROR")
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

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 400, "Invalid Request body", "INVALID_ARGUMENT")
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

	// Get project UUID First to use as reference
	projectUUID := projects.GetUUIDByName(urlVars["project"], refStr)
	if topics.HasTopic(projectUUID, tName, refStr) == false {
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
	res, err := subscriptions.CreateSub(projectUUID, urlVars["subscription"], tName, pushEnd, curOff, postBody.Ack, rPolicy, rPeriod, refStr)

	if err != nil {
		if err.Error() == "exists" {
			respondErr(w, 409, "Subscription already exists", "ALREADY_EXISTS")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL_SERVER_ERROR")
		return
	}

	// Enable pushManager if subscription has pushConfiguration
	if pushEnd != "" {
		refMgr.Add(res.ProjectUUID, res.Name)
		refMgr.Launch(res.ProjectUUID, res.Name)
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		respondErr(w, 500, "Error exporting data to JSON", "INTERNAL_SERVER_ERROR")
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

	// Get Result Object
	// Get project UUID First to use as reference
	projectUUID := projects.GetUUIDByName(urlVars["project"], refStr)
	res, err := topics.CreateTopic(projectUUID, urlVars["topic"], refStr)
	if err != nil {
		if err.Error() == "exists" {
			respondErr(w, 409, "Topic already exists", "ALREADY_EXISTS")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL_SERVER_ERROR")
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		respondErr(w, 500, "Error Exporting Retrieved Data to JSON", "INTERNAL_SERVER_ERROR")
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

	// Get project UUID First to use as reference
	projectUUID := projects.GetUUIDByName(urlVars["project"], refStr)
	results, err := topics.Find(projectUUID, urlVars["topic"], refStr)

	if err != nil {
		respondErr(w, 500, "Backend error", "INTERNAL_SERVER_ERROR")
	}

	// If not found
	if results.Empty() {
		respondErr(w, 404, "Topic does not exist", "NOT_FOUND")
		return
	}

	res := results.List[0]

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		respondErr(w, 500, "Error exporting data to JSON", "INTERNAL_SERVER_ERROR")
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// TopicACL (GET) one topic's authorized users
func TopicACL(w http.ResponseWriter, r *http.Request) {

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

	// Get Result Object
	res, err := topics.GetTopicACL(urlVars["project"], urlVars["topic"], refStr)

	// If not found
	if err != nil {
		respondErr(w, 404, "Topic does not exist", "NOT_FOUND")
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		respondErr(w, 500, "Error exporting data to JSON", "INTERNAL_SERVER_ERROR")
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// SubACL (GET) one topic's authorized users
func SubACL(w http.ResponseWriter, r *http.Request) {

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

	// Get Result Object
	res, err := subscriptions.GetSubACL(urlVars["project"], urlVars["subscription"], refStr)

	// If not found
	if err != nil {
		respondErr(w, 404, "Subscription does not exist", "NOT_FOUND")
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		respondErr(w, 500, "Error exporting data to JSON", "INTERNAL_SERVER_ERROR")
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

	projectUUID := projects.GetUUIDByName(urlVars["project"], refStr)
	res, err := subscriptions.Find(projectUUID, "", refStr)
	if err != nil {
		respondErr(w, 500, "Backend error", "INTERNAL_SERVER_ERROR")
		return
	}
	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		respondErr(w, 500, "Error exporting data to JSON", "INTERNAL_SERVER_ERROR")
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

	// Get project UUID First to use as reference
	projectUUID := projects.GetUUIDByName(urlVars["project"], refStr)
	res, err := topics.Find(projectUUID, "", refStr)
	if err != nil {
		respondErr(w, 500, "Backend error", "INTERNAL_SERVER_ERROR")
		return
	}
	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		respondErr(w, 500, "Error exporting data to JSON", "INTERNAL_SERVER_ERROR")
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
	refUser := context.Get(r, "auth_user").(string)
	refRoles := context.Get(r, "auth_roles").([]string)
	refAuthResource := context.Get(r, "auth_resource").(bool)

	// Get project UUID First to use as reference
	projectUUID := projects.GetUUIDByName(urlVars["project"], refStr)
	// Check if Project/Topic exist
	if topics.HasTopic(projectUUID, urlVars["topic"], refStr) == false {
		respondErr(w, 404, "Topic doesn't exist", "NOT_FOUND")
		return
	}

	// Check Authorization per topic
	// - if enabled in config
	// - if user has only publisher role

	if refAuthResource && auth.IsPublisher(refRoles) {
		if auth.PerResource(urlVars["project"], "topic", urlVars["topic"], refUser, refStr) == false {
			respondErr(w, 403, "Access to this resource is forbidden", "FORBIDDEN")
			return
		}
	}

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 400, "Bad Request Body", "BAD REQUEST")
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
		fullTopic := projectUUID + "." + urlVars["topic"]

		msgID, rTop, _, _, err := refBrk.Publish(fullTopic, msg)

		if err != nil {
			if err.Error() == "kafka server: Message was too large, server rejected it to avoid allocation error." {
				respondErr(w, 413, "Message size too large", "INVALID_ARGUMENT")
				return
			}
			respondErr(w, 500, err.Error(), "INTERNAL_SERVER_ERROR")
			return
		}
		msg.ID = msgID
		// Assertions for Succesfull Publish
		if rTop != fullTopic {
			respondErr(w, 500, "Broker reports wrong topic", "INTERNAL_SERVER_ERROR")
			return
		}

		// if rPart != 0 {
		// 	respondErr(w, 500, "Broker reports wrong partition", "INTERNAL_SERVER_ERROR")
		// 	return
		// }
		//
		// if rOff != off {
		// 	respondErr(w, 500, "Broker reports wrong offset", "INTERNAL_SERVER_ERROR")
		// 	return
		// }

		// Append the MsgID of the successful published message to the msgIds list
		msgIDs.IDs = append(msgIDs.IDs, msg.ID)
	}

	// Export the msgIDs
	resJSON, err := msgIDs.ExportJSON()
	if err != nil {
		respondErr(w, 500, "Error during export data to JSON", "INTERNAL_SERVER_ERROR")
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
	refUser := context.Get(r, "auth_user").(string)
	refRoles := context.Get(r, "auth_roles").([]string)
	refAuthResource := context.Get(r, "auth_resource").(bool)

	// Get project UUID First to use as reference
	projectUUID := projects.GetUUIDByName(urlVars["project"], refStr)
	// Check if sub exists
	if subscriptions.HasSub(projectUUID, urlVars["subscription"], refStr) == false {
		respondErr(w, 404, "Subscription doesn't exist", "NOT_FOUND")
		return
	}

	// Check Authorization per subscription
	// - if enabled in config
	// - if user has only consumer role
	if refAuthResource && auth.IsConsumer(refRoles) {
		if auth.PerResource(urlVars["project"], "subscription", urlVars["subscription"], refUser, refStr) == false {
			respondErr(w, 403, "Access to this resource is forbidden", "FORBIDDEN")
			return
		}
	}

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 400, "Invalid Request Body", "INVALID_ARGUMENT")
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
	results, err := subscriptions.Find(projectUUID, urlVars["subscription"], refStr)
	if err != nil {
		respondErr(w, 500, "Backend error", "INTERNAL_SERVER_ERROR")
		return
	}
	if results.Empty() {
		respondErr(w, 404, "Subscription doesn't exist", "NOT_FOUND")
		return
	}
	targSub := results.List[0]

	fullTopic := targSub.ProjectUUID + "." + targSub.Topic
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
			respondErr(w, 500, "Message retrieved from broker network has invalid JSON Structure", "INTERNAL_SERVER_ERROR")
			return
		}

		curRec := messages.RecMsg{AckID: ackPrefix + curMsg.ID, Msg: curMsg}
		recList.RecMsgs = append(recList.RecMsgs, curRec)
	}

	resJSON, err := recList.ExportJSON()

	if err != nil {
		respondErr(w, 500, "Error during exporting message to JSON", "INTERNAL_SERVER_ERROR")
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
