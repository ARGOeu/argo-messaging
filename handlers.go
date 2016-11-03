package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

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

		urlVars := mux.Vars(r)

		nStr := str.Clone()
		defer nStr.Close()

		projectUUID := projects.GetUUIDByName(urlVars["project"], nStr)
		context.Set(r, "auth_project_uuid", projectUUID)
		context.Set(r, "brk", brk)
		context.Set(r, "str", nStr)
		context.Set(r, "mgr", mgr)
		context.Set(r, "auth_resource", cfg.ResAuth)
		context.Set(r, "auth_user", "UserA")
		context.Set(r, "auth_user_uuid", "uuid1")
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
		context.Set(r, "auth_service_token", cfg.ServiceToken)
		hfn.ServeHTTP(w, r)

	})
}

// WrapLog handle wrapper to apply Logging
func WrapLog(hfn http.Handler, name string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		hfn.ServeHTTP(w, r)

		log.Info(
			"ACCESS", "\t",
			r.Method, "\t",
			r.RequestURI, "\t",
			name, "\t",
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
		serviceToken := context.Get(r, "auth_service_token").(string)

		projectUUID := projects.GetUUIDByName(urlVars["project"], refStr)

		// Check first if service token is used
		if serviceToken != "" && serviceToken == urlValues.Get("key") {
			context.Set(r, "auth_roles", []string{})
			context.Set(r, "auth_user", "")
			context.Set(r, "auth_user_uuid", "")
			context.Set(r, "auth_project_uuid", projectUUID)
			hfn.ServeHTTP(w, r)
			return
		}

		roles, user := auth.Authenticate(projectUUID, urlValues.Get("key"), refStr)

		if len(roles) > 0 {
			userUUID := auth.GetUUIDByName(user, refStr)
			context.Set(r, "auth_roles", roles)
			context.Set(r, "auth_user", user)
			context.Set(r, "auth_user_uuid", userUUID)
			context.Set(r, "auth_project_uuid", projectUUID)
			hfn.ServeHTTP(w, r)
		} else {
			respondErr(w, 401, "Unauthorized", "UNAUTHORIZED")
		}

	})
}

// WrapAuthorize handle wrapper to apply authentication
func WrapAuthorize(hfn http.Handler, routeName string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		urlValues := r.URL.Query()

		refStr := context.Get(r, "str").(stores.Store)
		refRoles := context.Get(r, "auth_roles").([]string)
		serviceToken := context.Get(r, "auth_service_token").(string)

		// Check first if service token is used
		if serviceToken != "" && serviceToken == urlValues.Get("key") {
			hfn.ServeHTTP(w, r)
			return
		}

		if auth.Authorize(routeName, refRoles, refStr) {
			hfn.ServeHTTP(w, r)
		} else {
			respondErr(w, 403, "Access to this resource is forbidden", "FORBIDDEN")
		}

	})
}

// HandlerFunctions
///////////////////

// ProjectDelete (DEL) deletes an existing project (also removes it's topics and subscriptions)
func ProjectDelete(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)
	refMgr := context.Get(r, "mgr").(*push.Manager)
	// Get Result Object
	// Get project UUID First to use as reference
	projectUUID := context.Get(r, "auth_project_uuid").(string)
	// RemoveProject removes also attached subs and topics from the datastore
	err := projects.RemoveProject(projectUUID, refStr)
	if err != nil {
		if err.Error() == "not found" {
			respondErr(w, 404, "Project doesn't exist", "NOT_FOUND")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL")
		return
	}
	// Stop any relevant push subscriptions
	if err := refMgr.RemoveProjectAll(projectUUID); err != nil {
		respondErr(w, 500, err.Error(), "INTERNAL")
	}
	// Write empty response if anything ok
	respondOK(w, output)

}

// ProjectUpdate (PUT) updates the name or the description of an existing project
func ProjectUpdate(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)
	projectUUID := context.Get(r, "auth_project_uuid").(string)
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
		log.Error(string(body[:]))
		return
	}

	modified := time.Now()
	// Get Result Object

	res, err := projects.UpdateProject(projectUUID, postBody.Name, postBody.Description, modified, refStr)

	if err != nil {
		if err.Error() == "not found" {
			respondErr(w, 403, "Project not found", "NOT_FOUND")
			return
		}

		if strings.HasPrefix(err.Error(), "invalid") {
			respondErr(w, 400, err.Error(), "INVALID_ARGUMENT")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL")
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
	refUserUUID := context.Get(r, "auth_user_uuid").(string)

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
		log.Error(string(body[:]))
		return
	}

	uuid := uuid.NewV4().String() // generate a new uuid to attach to the new project
	created := time.Now()
	// Get Result Object
	res, err := projects.CreateProject(uuid, urlProject, created, refUserUUID, postBody.Description, refStr)

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

	res, err := projects.Find("", "", refStr)

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
	results, err := projects.Find("", urlProject, refStr)

	if err != nil {

		if err.Error() == "not found" {
			respondErr(w, 404, "Project does not exist", "NOT_FOUND")
			return
		}
		respondErr(w, 500, "Internal error while querying datastore", "INTERNAL_SERVER_ERROR")
		return
	}

	// Output result to JSON
	res := results.One()
	resJSON, err := res.ExportJSON()

	if err != nil {
		respondErr(w, 500, "Error exporting data", "INTERNAL")
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// RefreshToken (POST) refreshes user's token
func RefreshToken(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	urlUser := urlVars["user"]

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)

	// Get Result Object
	userUUID := auth.GetUUIDByName(urlUser, refStr)
	token, err := auth.GenToken() // generate a new user token

	res, err := auth.UpdateUserToken(userUUID, token, refStr)

	if err != nil {
		if err.Error() == "not found" {
			respondErr(w, 403, "User not found", "NOT_FOUND")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL")
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

// UserUpdate (PUT) updates the user information
func UserUpdate(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	urlUser := urlVars["user"]

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 400, "Invalid Request body", "INVALID_ARGUMENT")
		return
	}

	// Parse pull options
	postBody, err := auth.GetUserFromJSON(body)
	if err != nil {
		respondErr(w, 400, "Invalid User Arguments", "INVALID_ARGUMENT")
		log.Error(string(body[:]))
		return
	}

	// Get Result Object
	userUUID := auth.GetUUIDByName(urlUser, refStr)
	modified := time.Now()
	res, err := auth.UpdateUser(userUUID, postBody.Name, postBody.Projects, postBody.Email, postBody.ServiceRoles, modified, refStr)

	if err != nil {

		// In case of invalid project or role in post body

		if err.Error() == "not found" {
			respondErr(w, 403, "User not found", "NOT_FOUND")
			return
		}

		if strings.HasPrefix(err.Error(), "invalid") {
			respondErr(w, 400, err.Error(), "INVALID_ARGUMENT")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL")
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

// UserCreate (POST) creates a new user inside a project
func UserCreate(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	urlUser := urlVars["user"]

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)
	refUserUUID := context.Get(r, "auth_user_uuid").(string)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 400, "Invalid Request body", "INVALID_ARGUMENT")
		return
	}

	// Parse pull options
	postBody, err := auth.GetUserFromJSON(body)
	if err != nil {
		respondErr(w, 400, "Invalid User  Arguments", "INVALID_ARGUMENT")
		log.Error(string(body[:]))
		return
	}

	uuid := uuid.NewV4().String() // generate a new uuid to attach to the new project
	token, err := auth.GenToken() // generate a new user token
	created := time.Now()
	// Get Result Object
	res, err := auth.CreateUser(uuid, urlUser, postBody.Projects, token, postBody.Email, postBody.ServiceRoles, created, refUserUUID, refStr)

	if err != nil {
		if err.Error() == "exists" {
			respondErr(w, 409, "User already exists", "ALREADY_EXISTS")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL")
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

// UserListOne (GET) one user
func UserListOne(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	urlUser := urlVars["user"]

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)

	// Get Results Object
	results, err := auth.FindUsers("", "", urlUser, refStr)

	if err != nil {
		if err.Error() == "not found" {
			respondErr(w, 404, "User does not exist", "NOT_FOUND")
			return
		}

		respondErr(w, 500, "Internal error while querying datastore", "INTERNAL")
		return
	}

	res := results.One()

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

// UserListAll (GET) all users belonging to a project
func UserListAll(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)

	// Get Results Object
	res, err := auth.FindUsers("", "", "", refStr)

	if err != nil && err.Error() != "not found" {
		respondErr(w, 500, "Internal error while querying datastore", "INTERNAL")
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

// UserDelete (DEL) deletes an existing user
func UserDelete(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)
	// Grab url path variables
	urlVars := mux.Vars(r)
	urlUser := urlVars["user"]

	userUUID := auth.GetUUIDByName(urlUser, refStr)

	err := auth.RemoveUser(userUUID, refStr)
	if err != nil {
		if err.Error() == "not found" {
			respondErr(w, 404, "User doesn't exist", "NOT_FOUND")
			return
		}

		respondErr(w, 500, err.Error(), "INTERNAL")
		return
	}

	// Write empty response if anything ok
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
	projectUUID := context.Get(r, "auth_project_uuid").(string)
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
	projectUUID := context.Get(r, "auth_project_uuid").(string)

	results, err := subscriptions.Find(projectUUID, urlVars["subscription"], refStr)

	if err != nil {
		respondErr(w, 500, "Backend Error", "INTERNAL_SERVER_ERROR")
		return
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
	projectUUID := context.Get(r, "auth_project_uuid").(string)

	// Get Result Object

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
	projectUUID := context.Get(r, "auth_project_uuid").(string)

	// Get Result Object
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
	// Get Result Object
	urlTopic := urlVars["topic"]

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 400, "Invalid Request body", "INVALID_ARGUMENT")
		return
	}

	// Parse pull options
	postBody, err := auth.GetACLFromJSON(body)
	if err != nil {
		respondErr(w, 400, "Invalid Topic ACL Arguments", "INVALID_ARGUMENT")
		log.Error(string(body[:]))
		return
	}

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)
	// Get project UUID First to use as reference
	projectUUID := context.Get(r, "auth_project_uuid").(string)

	// check if user list contain valid users for the given project
	_, err = auth.AreValidUsers(projectUUID, postBody.AuthUsers, refStr)
	if err != nil {
		respondErr(w, 404, err.Error(), "NOT_FOUND")
		return
	}

	err = auth.ModACL(projectUUID, "topics", urlTopic, postBody.AuthUsers, refStr)

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
	// Get Result Object
	urlSub := urlVars["subscription"]

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, 400, "Invalid Request body", "INVALID_ARGUMENT")
		return
	}

	// Parse pull options
	postBody, err := auth.GetACLFromJSON(body)
	if err != nil {
		respondErr(w, 400, "Invalid Subscription ACL Arguments", "INVALID_ARGUMENT")
		log.Error(string(body[:]))
		return
	}

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)
	// Get project UUID First to use as reference
	projectUUID := context.Get(r, "auth_project_uuid").(string)

	// check if user list contain valid users for the given project
	_, err = auth.AreValidUsers(projectUUID, postBody.AuthUsers, refStr)
	if err != nil {
		respondErr(w, 404, err.Error(), "NOT_FOUND")
		return
	}

	err = auth.ModACL(projectUUID, "subscriptions", urlSub, postBody.AuthUsers, refStr)

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
	subName := urlVars["subscription"]
	project := urlVars["project"]

	refMgr := context.Get(r, "mgr").(*push.Manager)
	// Get project UUID First to use as reference
	projectUUID := context.Get(r, "auth_project_uuid").(string)

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
		log.Error(string(body[:]))
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
	projectUUID := context.Get(r, "auth_project_uuid").(string)

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
		log.Error(string(body[:]))
		return
	}

	tProject, tName, err := subscriptions.ExtractFullTopicRef(postBody.FullTopic)

	if err != nil {
		respondErr(w, 400, "Invalid Topic Name", "INVALID_ARGUMENT")
		return
	}

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
	projectUUID := context.Get(r, "auth_project_uuid").(string)

	// Get Result Object
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
	projectUUID := context.Get(r, "auth_project_uuid").(string)

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
	urlTopic := urlVars["topic"]

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)

	// Get project UUID First to use as reference
	projectUUID := context.Get(r, "auth_project_uuid").(string)
	res, err := auth.GetACL(projectUUID, "topics", urlTopic, refStr)

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
	urlSub := urlVars["subscription"]

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)

	// Get project UUID First to use as reference
	projectUUID := context.Get(r, "auth_project_uuid").(string)
	res, err := auth.GetACL(projectUUID, "subscriptions", urlSub, refStr)

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

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)
	projectUUID := context.Get(r, "auth_project_uuid").(string)

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

	// Grab context references
	refStr := context.Get(r, "str").(stores.Store)
	projectUUID := context.Get(r, "auth_project_uuid").(string)

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
	urlTopic := urlVars["topic"]

	// Grab context references

	refBrk := context.Get(r, "brk").(brokers.Broker)
	refStr := context.Get(r, "str").(stores.Store)
	refUser := context.Get(r, "auth_user").(string)
	refRoles := context.Get(r, "auth_roles").([]string)
	refAuthResource := context.Get(r, "auth_resource").(bool)
	// Get project UUID First to use as reference
	projectUUID := context.Get(r, "auth_project_uuid").(string)

	// Check if Project/Topic exist
	if topics.HasTopic(projectUUID, urlVars["topic"], refStr) == false {
		respondErr(w, 404, "Topic doesn't exist", "NOT_FOUND")
		return
	}

	// Check Authorization per topic
	// - if enabled in config
	// - if user has only publisher role

	if refAuthResource && auth.IsPublisher(refRoles) {

		if auth.PerResource(projectUUID, "topics", urlTopic, refUser, refStr) == false {
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
		log.Error(string(body[:]))
		return
	}

	// Init message ids list
	msgIDs := messages.MsgIDs{}

	// For each message in message list
	for _, msg := range msgList.Msgs {
		// Get offset and set it as msg
		fullTopic := projectUUID + "." + urlTopic

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
	urlProject := urlVars["project"]
	urlSub := urlVars["subscription"]

	// Grab context references
	refBrk := context.Get(r, "brk").(brokers.Broker)
	refStr := context.Get(r, "str").(stores.Store)
	refUser := context.Get(r, "auth_user").(string)
	refRoles := context.Get(r, "auth_roles").([]string)
	refAuthResource := context.Get(r, "auth_resource").(bool)

	// Get project UUID First to use as reference
	projectUUID := context.Get(r, "auth_project_uuid").(string)
	// Check if sub exists
	if subscriptions.HasSub(projectUUID, urlSub, refStr) == false {
		respondErr(w, 404, "Subscription doesn't exist", "NOT_FOUND")
		return
	}

	// Check Authorization per subscription
	// - if enabled in config
	// - if user has only consumer role
	if refAuthResource && auth.IsConsumer(refRoles) {
		if auth.PerResource(urlProject, "subscriptions", urlSub, refUser, refStr) == false {
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
		log.Error(string(body[:]))
		return
	}

	// Init Received Message List
	recList := messages.RecList{}

	// Get the subscription info
	results, err := subscriptions.Find(projectUUID, urlSub, refStr)
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

	ackPrefix := "projects/" + urlProject + "/subscriptions/" + urlSub + ":"

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
	log.Error(errCode, "\t", errMsg)
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
