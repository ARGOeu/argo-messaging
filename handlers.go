package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"context"

	"github.com/ARGOeu/argo-messaging/auth"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/messages"
	"github.com/ARGOeu/argo-messaging/metrics"
	"github.com/ARGOeu/argo-messaging/projects"
	oldPush "github.com/ARGOeu/argo-messaging/push"
	"github.com/ARGOeu/argo-messaging/push/grpc/client"
	"github.com/ARGOeu/argo-messaging/schemas"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/ARGOeu/argo-messaging/subscriptions"
	"github.com/ARGOeu/argo-messaging/topics"
	"github.com/ARGOeu/argo-messaging/version"

	"bytes"
	"encoding/base64"
	gorillaContext "github.com/gorilla/context"
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
				err := APIErrorInvalidName(key)
				respondErr(w, err)
				return
			}
		}
		hfn.ServeHTTP(w, r)

	})
}

// WrapMockAuthConfig handle wrapper is used in tests were some auth context is needed
func WrapMockAuthConfig(hfn http.HandlerFunc, cfg *config.APICfg, brk brokers.Broker, str stores.Store, mgr *oldPush.Manager, c push.Client, roles ...string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		urlVars := mux.Vars(r)

		userRoles := []string{"publisher", "consumer"}
		if len(roles) > 0 {
			userRoles = roles
		}

		nStr := str.Clone()
		defer nStr.Close()

		projectUUID := projects.GetUUIDByName(urlVars["project"], nStr)
		gorillaContext.Set(r, "auth_project_uuid", projectUUID)
		gorillaContext.Set(r, "brk", brk)
		gorillaContext.Set(r, "str", nStr)
		gorillaContext.Set(r, "mgr", mgr)
		gorillaContext.Set(r, "apsc", c)
		gorillaContext.Set(r, "auth_resource", cfg.ResAuth)
		gorillaContext.Set(r, "auth_user", "UserA")
		gorillaContext.Set(r, "auth_user_uuid", "uuid1")
		gorillaContext.Set(r, "auth_roles", userRoles)
		gorillaContext.Set(r, "push_worker_token", cfg.PushWorkerToken)
		gorillaContext.Set(r, "push_enabled", cfg.PushEnabled)
		hfn.ServeHTTP(w, r)

	})
}

// WrapConfig handle wrapper to retrieve kafka configuration
func WrapConfig(hfn http.HandlerFunc, cfg *config.APICfg, brk brokers.Broker, str stores.Store, mgr *oldPush.Manager, c push.Client) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		nStr := str.Clone()
		defer nStr.Close()
		gorillaContext.Set(r, "brk", brk)
		gorillaContext.Set(r, "str", nStr)
		gorillaContext.Set(r, "mgr", mgr)
		gorillaContext.Set(r, "apsc", c)
		gorillaContext.Set(r, "auth_resource", cfg.ResAuth)
		gorillaContext.Set(r, "auth_service_token", cfg.ServiceToken)
		gorillaContext.Set(r, "push_worker_token", cfg.PushWorkerToken)
		gorillaContext.Set(r, "push_enabled", cfg.PushEnabled)
		hfn.ServeHTTP(w, r)

	})
}

// WrapLog handle wrapper to apply Logging
func WrapLog(hfn http.Handler, name string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		hfn.ServeHTTP(w, r)

		log.WithFields(
			log.Fields{
				"type":            "request_log",
				"method":          r.Method,
				"path":            r.RequestURI,
				"action":          name,
				"requester":       gorillaContext.Get(r, "auth_user_uuid"),
				"processing_time": time.Since(start).String(),
			},
		).Info("")
	})
}

// WrapAuthenticate handle wrapper to apply authentication
func WrapAuthenticate(hfn http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		urlVars := mux.Vars(r)
		urlValues := r.URL.Query()

		// if the url parameter 'key' is empty or absent, end the request with an unauthorized response
		if urlValues.Get("key") == "" {
			err := APIErrorUnauthorized()
			respondErr(w, err)
			return
		}

		refStr := gorillaContext.Get(r, "str").(stores.Store)
		serviceToken := gorillaContext.Get(r, "auth_service_token").(string)

		projectName := urlVars["project"]
		projectUUID := projects.GetUUIDByName(urlVars["project"], refStr)

		// In all cases instead of project create
		if "projects:create" != mux.CurrentRoute(r).GetName() {
			// Check if given a project name the project wasn't found
			if projectName != "" && projectUUID == "" {
				apiErr := APIErrorNotFound("project")
				respondErr(w, apiErr)
				return
			}
		}

		// Check first if service token is used
		if serviceToken != "" && serviceToken == urlValues.Get("key") {
			gorillaContext.Set(r, "auth_roles", []string{"service_admin"})
			gorillaContext.Set(r, "auth_user", "")
			gorillaContext.Set(r, "auth_user_uuid", "")
			gorillaContext.Set(r, "auth_project_uuid", projectUUID)
			hfn.ServeHTTP(w, r)
			return
		}

		roles, user := auth.Authenticate(projectUUID, urlValues.Get("key"), refStr)

		if len(roles) > 0 {
			userUUID := auth.GetUUIDByName(user, refStr)
			gorillaContext.Set(r, "auth_roles", roles)
			gorillaContext.Set(r, "auth_user", user)
			gorillaContext.Set(r, "auth_user_uuid", userUUID)
			gorillaContext.Set(r, "auth_project_uuid", projectUUID)
			hfn.ServeHTTP(w, r)
		} else {
			err := APIErrorUnauthorized()
			respondErr(w, err)
		}

	})
}

// WrapAuthorize handle wrapper to apply authorization
func WrapAuthorize(hfn http.Handler, routeName string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		urlValues := r.URL.Query()

		refStr := gorillaContext.Get(r, "str").(stores.Store)
		refRoles := gorillaContext.Get(r, "auth_roles").([]string)
		serviceToken := gorillaContext.Get(r, "auth_service_token").(string)

		// Check first if service token is used
		if serviceToken != "" && serviceToken == urlValues.Get("key") {
			hfn.ServeHTTP(w, r)
			return
		}

		if auth.Authorize(routeName, refRoles, refStr) {
			hfn.ServeHTTP(w, r)
		} else {
			err := APIErrorForbidden()
			respondErr(w, err)
		}
	})
}

// HandlerFunctions
///////////////////

// UserProfile returns a user's profile based on the provided url parameter(key)
func UserProfile(w http.ResponseWriter, r *http.Request) {

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	urlValues := r.URL.Query()

	// if the url parameter 'key' is empty or absent, end the request with an unauthorized response
	if urlValues.Get("key") == "" {
		err := APIErrorUnauthorized()
		respondErr(w, err)
		return
	}

	result, err := auth.GetUserByToken(urlValues.Get("key"), refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorUnauthorized()
			respondErr(w, err)
			return
		}
		err := APIErrQueryDatastore()
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := result.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	respondOK(w, []byte(resJSON))

}

// ProjectDelete (DEL) deletes an existing project (also removes it's topics and subscriptions)
func ProjectDelete(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get Result Object
	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)
	// RemoveProject removes also attached subs and topics from the datastore
	err := projects.RemoveProject(projectUUID, refStr)
	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("ProjectUUID")
			respondErr(w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	// TODO Stop any relevant push subscriptions when deleting a project

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
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)
	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// Parse pull options
	postBody, err := projects.GetFromJSON(body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		log.Error(string(body[:]))
		return
	}

	modified := time.Now().UTC()
	// Get Result Object

	res, err := projects.UpdateProject(projectUUID, postBody.Name, postBody.Description, modified, refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("ProjectUUID")
			respondErr(w, err)
			return
		}

		if strings.HasPrefix(err.Error(), "invalid") {
			err := APIErrorInvalidData(err.Error())
			respondErr(w, err)
			return
		}

		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
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
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refUserUUID := gorillaContext.Get(r, "auth_user_uuid").(string)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// Parse pull options
	postBody, err := projects.GetFromJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("Project")
		respondErr(w, err)
		log.Error(string(body[:]))
		return
	}

	uuid := uuid.NewV4().String() // generate a new uuid to attach to the new project
	created := time.Now().UTC()
	// Get Result Object

	res, err := projects.CreateProject(uuid, urlProject, created, refUserUUID, postBody.Description, refStr)

	if err != nil {
		if err.Error() == "exists" {
			err := APIErrorConflict("Project")
			respondErr(w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
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
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get Results Object

	res, err := projects.Find("", "", refStr)

	if err != nil && err.Error() != "not found" {
		err := APIErrQueryDatastore()
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
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
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get Results Object
	results, err := projects.Find("", urlProject, refStr)

	if err != nil {

		if err.Error() == "not found" {
			err := APIErrorNotFound("ProjectUUID")
			respondErr(w, err)
			return
		}
		err := APIErrQueryDatastore()
		respondErr(w, err)
		return
	}

	// Output result to JSON
	res := results.One()
	resJSON, err := res.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
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
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get Result Object
	userUUID := auth.GetUUIDByName(urlUser, refStr)
	token, err := auth.GenToken() // generate a new user token

	res, err := auth.UpdateUserToken(userUUID, token, refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
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
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// Parse pull options
	postBody, err := auth.GetUserFromJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("User")
		respondErr(w, err)
		return
	}

	// Get Result Object
	userUUID := auth.GetUUIDByName(urlUser, refStr)
	modified := time.Now().UTC()
	res, err := auth.UpdateUser(userUUID, postBody.Name, postBody.Projects, postBody.Email, postBody.ServiceRoles, modified, true, refStr)

	if err != nil {

		// In case of invalid project or role in post body

		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(w, err)
			return
		}

		if strings.HasPrefix(err.Error(), "invalid") {
			err := APIErrorInvalidData(err.Error())
			respondErr(w, err)
			return
		}

		if strings.HasPrefix(err.Error(), "duplicate") {
			err := APIErrorInvalidData(err.Error())
			respondErr(w, err)
			return
		}

		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
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
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refUserUUID := gorillaContext.Get(r, "auth_user_uuid").(string)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// Parse pull options
	postBody, err := auth.GetUserFromJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("User")
		respondErr(w, err)
		log.Error(string(body[:]))
		return
	}

	uuid := uuid.NewV4().String() // generate a new uuid to attach to the new project
	token, err := auth.GenToken() // generate a new user token
	created := time.Now().UTC()
	// Get Result Object
	res, err := auth.CreateUser(uuid, urlUser, "", "", "", "", postBody.Projects, token, postBody.Email, postBody.ServiceRoles, created, refUserUUID, refStr)

	if err != nil {
		if err.Error() == "exists" {
			err := APIErrorConflict("User")
			respondErr(w, err)
			return
		}

		if strings.HasPrefix(err.Error(), "duplicate") {
			err := APIErrorInvalidData(err.Error())
			respondErr(w, err)
			return
		}

		if strings.HasPrefix(err.Error(), "invalid") {
			err := APIErrorInvalidData(err.Error())
			respondErr(w, err)
			return
		}

		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// OpMetrics (GET) all operational metrics
func OpMetrics(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get Results Object
	res, err := metrics.GetUsageCpuMem(refStr)

	if err != nil && err.Error() != "not found" {
		err := APIErrQueryDatastore()
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// DailyMessageAverage (GET) retrieves the average amount of published messages per day
func DailyMessageAverage(w http.ResponseWriter, r *http.Request) {

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	startDate := time.Time{}
	endDate := time.Time{}
	var err error

	// if no start date was provided, set it to the start of the unix time
	if r.URL.Query().Get("start_date") != "" {
		startDate, err = time.Parse("2006-01-02", r.URL.Query().Get("start_date"))
		if err != nil {
			err := APIErrorInvalidData("Start date is not in valid format")
			respondErr(w, err)
			return
		}
	} else {
		startDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	// if no end date was provided, set it to to today
	if r.URL.Query().Get("end_date") != "" {
		endDate, err = time.Parse("2006-01-02", r.URL.Query().Get("end_date"))
		if err != nil {
			err := APIErrorInvalidData("End date is not in valid format")
			respondErr(w, err)
			return
		}
	} else {
		endDate = time.Now().UTC()
	}

	if startDate.After(endDate) {
		err := APIErrorInvalidData("Start date cannot be after the end date")
		respondErr(w, err)
		return
	}

	projectsList := make([]string, 0)
	projectsUrlValue := r.URL.Query().Get("projects")
	if projectsUrlValue != "" {
		projectsList = strings.Split(projectsUrlValue, ",")
	}

	cc, err := projects.GetProjectsMessageCount(projectsList, startDate, endDate, refStr)
	if err != nil {
		err := APIErrorNotFound(err.Error())
		respondErr(w, err)
		return
	}

	output, err := json.MarshalIndent(cc, "", " ")
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	respondOK(w, output)
}

// UserListByToken (GET) one user by his token
func UserListByToken(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	urlToken := urlVars["token"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get Results Object
	result, err := auth.GetUserByToken(urlToken, refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(w, err)
			return
		}
		err := APIErrQueryDatastore()
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := result.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// ProjectUserListOne (GET) one user member of a specific project
func ProjectUserListOne(w http.ResponseWriter, r *http.Request) {

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
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refRoles := gorillaContext.Get(r, "auth_roles").([]string)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// check that user is indeed a service admin in order to be priviledged to see full user info
	priviledged := auth.IsServiceAdmin(refRoles)

	// Get Results Object
	results, err := auth.FindUsers(projectUUID, "", urlUser, priviledged, refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(w, err)
			return
		}

		err := APIErrQueryDatastore()
		respondErr(w, err)
		return
	}

	res := results.One()

	// Output result to JSON
	resJSON, err := res.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// ProjectUserCreate (POST) creates a user under the respective project by the project's admin
func ProjectUserCreate(w http.ResponseWriter, r *http.Request) {

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
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refUserUUID := gorillaContext.Get(r, "auth_user_uuid").(string)
	refProjUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// Parse pull options
	postBody, err := auth.GetUserFromJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("User")
		respondErr(w, err)
		log.Error(string(body[:]))
		return
	}

	// omit service wide roles
	postBody.ServiceRoles = []string{}

	// allow the user to be created to only have reference to the project under which is being created
	prName := projects.GetNameByUUID(refProjUUID, refStr)
	if prName == "" {
		err := APIErrGenericInternal("Internal Error")
		respondErr(w, err)
		return
	}

	projectRoles := auth.ProjectRoles{}

	for _, p := range postBody.Projects {
		if p.Project == prName {
			projectRoles.Project = prName
			projectRoles.Roles = p.Roles
			projectRoles.Topics = p.Topics
			projectRoles.Subs = p.Subs
			break
		}
	}

	// if the project was not mentioned in the creation, add it
	if projectRoles.Project == "" {
		projectRoles.Project = prName
	}

	postBody.Projects = []auth.ProjectRoles{projectRoles}

	uuid := uuid.NewV4().String() // generate a new uuid to attach to the new project
	token, err := auth.GenToken() // generate a new user token
	created := time.Now().UTC()

	// Get Result Object
	res, err := auth.CreateUser(uuid, urlUser, "", "", "", "", postBody.Projects, token, postBody.Email, postBody.ServiceRoles, created, refUserUUID, refStr)

	if err != nil {
		if err.Error() == "exists" {
			err := APIErrorConflict("User")
			respondErr(w, err)
			return
		}

		if strings.HasPrefix(err.Error(), "invalid") {
			err := APIErrorInvalidData(err.Error())
			respondErr(w, err)
			return
		}

		if strings.HasPrefix(err.Error(), "duplicate") {
			err := APIErrorInvalidData(err.Error())
			respondErr(w, err)
			return
		}

		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// ProjectUserUpdate (PUT) updates a user under the respective project by the project's admin
func ProjectUserUpdate(w http.ResponseWriter, r *http.Request) {

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
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refProjUUID := gorillaContext.Get(r, "auth_project_uuid").(string)
	refRoles := gorillaContext.Get(r, "auth_roles").([]string)

	// allow the user to be updated to only have reference to the project under which is being updated
	prName := projects.GetNameByUUID(refProjUUID, refStr)
	if prName == "" {
		err := APIErrGenericInternal("Internal Error")
		respondErr(w, err)
		return
	}

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// Parse pull options
	postBody, err := auth.GetUserFromJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("User")
		respondErr(w, err)
		return
	}

	u, err := auth.FindUsers("", "", urlUser, true, refStr)
	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(w, err)
			return
		}

		err := APIErrQueryDatastore()
		respondErr(w, err)
		return
	}

	// from the post request keep only the reference to the current project
	projectRoles := auth.ProjectRoles{}

	for _, p := range postBody.Projects {
		if p.Project == prName {
			projectRoles.Project = prName
			projectRoles.Roles = p.Roles
			projectRoles.Topics = p.Topics
			projectRoles.Subs = p.Subs
			break
		}
	}

	// if the user is already a member of the project, update it with the accepted contents of the post body
	found := false
	for idx, p := range u.One().Projects {
		if p.Project == projectRoles.Project {
			u.One().Projects[idx].Roles = projectRoles.Roles
			u.One().Projects[idx].Topics = projectRoles.Topics
			u.One().Projects[idx].Subs = projectRoles.Subs
			found = true
			break
		}
	}

	if !found {
		err := APIErrorForbiddenWithMsg("User is not a member of the project")
		respondErr(w, err)
		return
	}

	// check that user is indeed a service admin in order to be privileged to see full user info
	privileged := auth.IsServiceAdmin(refRoles)

	// Get Result Object
	userUUID := u.One().UUID
	modified := time.Now().UTC()
	userProjects := u.One().Projects
	userEmail := u.One().Email
	userSRoles := u.One().ServiceRoles
	userName := u.One().Name

	_, err = auth.UpdateUser(userUUID, userName, userProjects, userEmail, userSRoles, modified, false, refStr)

	if err != nil {

		// In case of invalid project or role in post body
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(w, err)
			return
		}

		if strings.HasPrefix(err.Error(), "invalid") {
			err := APIErrorInvalidData(err.Error())
			respondErr(w, err)
			return
		}

		if strings.HasPrefix(err.Error(), "duplicate") {
			err := APIErrorInvalidData(err.Error())
			respondErr(w, err)
			return
		}

		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	stored, err := auth.FindUsers(refProjUUID, userUUID, urlUser, privileged, refStr)

	if err != nil {

		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(w, err)
			return
		}

		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return

	}

	// Output result to JSON
	resJSON, err := json.MarshalIndent(stored.One(), "", "   ")
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// ProjectUserRemove (POST) removes a user from the respective project
func ProjectUserRemove(w http.ResponseWriter, r *http.Request) {

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	urlUser := urlVars["user"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refProjUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	projName := projects.GetNameByUUID(refProjUUID, refStr)

	u, err := auth.FindUsers("", "", urlUser, true, refStr)
	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(w, err)
			return
		}

		err := APIErrQueryDatastore()
		respondErr(w, err)
		return
	}

	userProjects := []auth.ProjectRoles{}

	// if the user is already a member of the project, update it with the accepted contents of the post body
	found := false
	for idx, p := range u.One().Projects {
		if p.Project == projName {
			userProjects = append(userProjects, u.One().Projects[:idx]...)
			userProjects = append(userProjects, u.One().Projects[idx+1:]...)
			found = true
			break
		}
	}

	if !found {
		err := APIErrorForbiddenWithMsg("User is not a member of the project")
		respondErr(w, err)
		return
	}

	// Get Result Object
	userUUID := u.One().UUID
	modified := time.Now().UTC()
	userEmail := u.One().Email
	userSRoles := u.One().ServiceRoles
	userName := u.One().Name

	_, err = auth.UpdateUser(userUUID, userName, userProjects, userEmail, userSRoles, modified, false, refStr)

	if err != nil {

		// In case of invalid project or role in post body
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(w, err)
			return
		}

		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	// Write response
	respondOK(w, []byte("{}"))

}

// ProjectUserAdd (POST) adds a user to the respective project
func ProjectUserAdd(w http.ResponseWriter, r *http.Request) {

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	urlUser := urlVars["user"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refProjUUID := gorillaContext.Get(r, "auth_project_uuid").(string)
	refRoles := gorillaContext.Get(r, "auth_roles").([]string)

	projName := projects.GetNameByUUID(refProjUUID, refStr)

	u, err := auth.FindUsers("", "", urlUser, true, refStr)
	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(w, err)
			return
		}

		err := APIErrQueryDatastore()
		respondErr(w, err)
		return
	}

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	data := auth.ProjectRoles{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// check if the user is already a user of the project
	found := false
	for _, p := range u.One().Projects {
		if p.Project == projName {
			found = true
			break
		}
	}

	if found {
		err := APIErrorGenericConflict("User is already a member of the project")
		respondErr(w, err)
		return
	}

	// Get Result Object
	userUUID := u.One().UUID
	modified := time.Now().UTC()
	userEmail := u.One().Email
	userSRoles := u.One().ServiceRoles
	userName := u.One().Name
	userProjects := u.One().Projects

	userProjects = append(userProjects, auth.ProjectRoles{
		Project: projName,
		Roles:   data.Roles,
		Subs:    data.Subs,
		Topics:  data.Topics,
	})

	_, err = auth.UpdateUser(userUUID, userName, userProjects, userEmail, userSRoles, modified, false, refStr)

	if err != nil {

		// In case of invalid project or role in post body
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(w, err)
			return
		}

		if strings.HasPrefix(err.Error(), "invalid") {
			err := APIErrorInvalidData(err.Error())
			respondErr(w, err)
			return
		}

		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	// Write response
	privileged := auth.IsServiceAdmin(refRoles)
	fmt.Println(privileged)
	results, err := auth.FindUsers(refProjUUID, "", urlUser, privileged, refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(w, err)
			return
		}

		err := APIErrQueryDatastore()
		respondErr(w, err)
		return
	}

	res := results.One()

	// Output result to JSON
	resJSON, err := res.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	respondOK(w, []byte(resJSON))
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
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get Results Object
	results, err := auth.FindUsers("", "", urlUser, true, refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(w, err)
			return
		}

		err := APIErrQueryDatastore()
		respondErr(w, err)
		return
	}

	res := results.One()

	// Output result to JSON
	resJSON, err := res.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

func UserListByUUID(w http.ResponseWriter, r *http.Request) {

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

	// Get Results Object
	result, err := auth.GetUserByUUID(urlVars["uuid"], refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(w, err)
			return
		}

		if err.Error() == "multiple uuids" {
			err := APIErrGenericInternal("Multiple users found with the same uuid")
			respondErr(w, err)
			return
		}

		err := APIErrQueryDatastore()
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := result.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)
}

// ProjectListUsers (GET) all users belonging to a project
func ProjectListUsers(w http.ResponseWriter, r *http.Request) {

	var err error
	var pageSize int
	var paginatedUsers auth.PaginatedUsers

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refRoles := gorillaContext.Get(r, "auth_roles").([]string)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Grab url path variables
	urlValues := r.URL.Query()
	pageToken := urlValues.Get("pageToken")
	strPageSize := urlValues.Get("pageSize")

	if strPageSize != "" {
		if pageSize, err = strconv.Atoi(strPageSize); err != nil {
			log.Errorf("Pagesize %v produced an error  while being converted to int: %v", strPageSize, err.Error())
			err := APIErrorInvalidData("Invalid page size")
			respondErr(w, err)
			return
		}
	}

	// check that user is indeed a service admin in order to be priviledged to see full user info
	priviledged := auth.IsServiceAdmin(refRoles)

	// Get Results Object - call is always priviledged because this handler is only accessible by service admins
	if paginatedUsers, err = auth.PaginatedFindUsers(pageToken, int32(pageSize), projectUUID, priviledged, refStr); err != nil {
		err := APIErrorInvalidData("Invalid page token")
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := paginatedUsers.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// UserListAll (GET) all users - or users belonging to a project
func UserListAll(w http.ResponseWriter, r *http.Request) {

	var err error
	var pageSize int
	var paginatedUsers auth.PaginatedUsers

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refRoles := gorillaContext.Get(r, "auth_roles").([]string)

	// Grab url path variables
	urlValues := r.URL.Query()
	pageToken := urlValues.Get("pageToken")
	strPageSize := urlValues.Get("pageSize")
	projectName := urlValues.Get("project")
	projectUUID := ""

	if projectName != "" {
		projectUUID = projects.GetUUIDByName(projectName, refStr)
		if projectUUID == "" {
			err := APIErrorNotFound("ProjectUUID")
			respondErr(w, err)
			return
		}
	}

	if strPageSize != "" {
		if pageSize, err = strconv.Atoi(strPageSize); err != nil {
			log.Errorf("Pagesize %v produced an error  while being converted to int: %v", strPageSize, err.Error())
			err := APIErrorInvalidData("Invalid page size")
			respondErr(w, err)
			return
		}
	}

	// check that user is indeed a service admin in order to be priviledged to see full user info
	priviledged := auth.IsServiceAdmin(refRoles)

	// Get Results Object - call is always priviledged because this handler is only accessible by service admins
	if paginatedUsers, err = auth.PaginatedFindUsers(pageToken, int32(pageSize), projectUUID, priviledged, refStr); err != nil {
		err := APIErrorInvalidData("Invalid page token")
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := paginatedUsers.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
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
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	// Grab url path variables
	urlVars := mux.Vars(r)
	urlUser := urlVars["user"]

	userUUID := auth.GetUUIDByName(urlUser, refStr)

	err := auth.RemoveUser(userUUID, refStr)
	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	// Write empty response if anything ok
	respondOK(w, output)

}

// RegisterUser(POST) registers a new user
func RegisterUser(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	username := urlVars["user"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// check if a user with that name already exists
	if auth.ExistsWithName(username, refStr) {
		err := APIErrorConflict("User")
		respondErr(w, err)
		return
	}

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// Parse pull options
	requestBody := auth.UserRegistration{}
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		err := APIErrorInvalidArgument("User")
		respondErr(w, err)
		return
	}

	uuid := uuid.NewV4().String()
	registered := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	tkn, err := auth.GenToken()
	if err != nil {
		err := APIErrGenericInternal("")
		respondErr(w, err)
		return
	}

	ur, err := auth.RegisterUser(uuid, username, requestBody.FirstName, requestBody.LastName, requestBody.Email,
		requestBody.Organization, requestBody.Description, registered, tkn, auth.PendingRegistrationStatus, refStr)

	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	output, err = json.MarshalIndent(ur, "", "   ")
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	respondOK(w, output)
}

// AcceptUserRegister (POST) accepts a user registration and creates the respective user
func AcceptRegisterUser(w http.ResponseWriter, r *http.Request) {

	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	activationToken := urlVars["activation_token"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refUserUUID := gorillaContext.Get(r, "auth_user_uuid").(string)

	ru, err := auth.FindUserRegistration(activationToken, auth.PendingRegistrationStatus, refStr)
	if err != nil {

		if err.Error() == "not found" {
			err := APIErrorNotFound("User registration")
			respondErr(w, err)
			return
		}

		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	userUUID := uuid.NewV4().String() // generate a new userUUID to attach to the new project
	token, err := auth.GenToken()     // generate a new user token
	created := time.Now().UTC()
	// Get Result Object
	res, err := auth.CreateUser(userUUID, ru.Name, ru.FirstName, ru.LastName, ru.Organization, ru.Description,
		[]auth.ProjectRoles{}, token, ru.Email, []string{}, created, refUserUUID, refStr)

	if err != nil {
		if err.Error() == "exists" {
			err := APIErrorConflict("User")
			respondErr(w, err)
			return
		}

		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	// update the registration
	err = auth.UpdateUserRegistration(activationToken, auth.AcceptedRegistrationStatus, refUserUUID, created, refStr)
	if err != nil {
		log.Errorf("Could not update registration, %v", err.Error())
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	respondOK(w, []byte(resJSON))
}

// DeclineRegisterUser(POST) declines a user's registration
func DeclineRegisterUser(w http.ResponseWriter, r *http.Request) {

	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	activationToken := urlVars["activation_token"]
	refUserUUID := gorillaContext.Get(r, "auth_user_uuid").(string)

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	_, err := auth.FindUserRegistration(activationToken, auth.PendingRegistrationStatus, refStr)
	if err != nil {

		if err.Error() == "not found" {
			err := APIErrorNotFound("User registration")
			respondErr(w, err)
			return
		}

		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	err = auth.UpdateUserRegistration(activationToken, auth.DeclineddRegistrationStatus, refUserUUID, time.Now().UTC(), refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	respondOK(w, []byte("{}"))

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
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)
	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// Parse pull options
	postBody, err := subscriptions.GetAckFromJSON(body)
	if err != nil {
		err := APIErrorInvalidData("Invalid ack parameter")
		respondErr(w, err)
		return
	}

	// Get urlParams
	projectName := urlVars["project"]
	subName := urlVars["subscription"]

	// Check if sub exists

	cur_sub, err := subscriptions.Find(projectUUID, "", subName, "", 0, refStr)
	if err != nil {
		err := APIErrHandlingAcknowledgement()
		respondErr(w, err)
		return
	}
	if len(cur_sub.Subscriptions) == 0 {
		err := APIErrorNotFound("Subscription")
		respondErr(w, err)
		return
	}

	// Get list of AckIDs
	if postBody.IDs == nil {
		err := APIErrorInvalidData("Invalid ack id")
		respondErr(w, err)
		return
	}

	// Check if each AckID is valid
	for _, ackID := range postBody.IDs {
		if validAckID(projectName, subName, ackID) == false {
			err := APIErrorInvalidData("Invalid ack id")
			respondErr(w, err)
			return
		}
	}

	// Get Max ackID
	maxAckID, err := subscriptions.GetMaxAckID(postBody.IDs)
	if err != nil {
		err := APIErrHandlingAcknowledgement()
		respondErr(w, err)
		return
	}
	// Extract offset from max ackID
	off, err := subscriptions.GetOffsetFromAckID(maxAckID)

	if err != nil {
		err := APIErrorInvalidData("Invalid ack id")
		respondErr(w, err)
		return
	}

	zSec := "2006-01-02T15:04:05Z"
	t := time.Now().UTC()
	ts := t.Format(zSec)

	err = refStr.UpdateSubOffsetAck(projectUUID, urlVars["subscription"], int64(off+1), ts)
	if err != nil {

		if err.Error() == "ack timeout" {
			err := APIErrorTimeout(err.Error())
			respondErr(w, err)
			return
		}

		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
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
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	results, err := subscriptions.Find(projectUUID, "", urlVars["subscription"], "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Subscription")
		respondErr(w, err)
		return
	}

	// if its a push enabled sub and it has a verified endpoint
	// call the push server to find its real time push status
	if results.Subscriptions[0].PushCfg != (subscriptions.PushConfig{}) {
		if results.Subscriptions[0].PushCfg.Verified {
			apsc := gorillaContext.Get(r, "apsc").(push.Client)
			results.Subscriptions[0].PushStatus = apsc.SubscriptionStatus(context.TODO(), results.Subscriptions[0].FullName).Result()
		}
	}

	// Output result to JSON
	resJSON, err := results.Subscriptions[0].ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// SubSetOffset (PUT) sets subscriptions current offset
func SubSetOffset(w http.ResponseWriter, r *http.Request) {

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
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// Parse pull options
	postBody, err := subscriptions.GetSetOffsetJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("Offset")
		respondErr(w, err)
		log.Error(string(body[:]))
		return
	}

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refBrk := gorillaContext.Get(r, "brk").(brokers.Broker)
	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Find Subscription
	results, err := subscriptions.Find(projectUUID, "", urlVars["subscription"], "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Subscription")
		respondErr(w, err)
		return
	}
	brk_topic := projectUUID + "." + results.Subscriptions[0].Topic
	min_offset := refBrk.GetMinOffset(brk_topic)
	max_offset := refBrk.GetMaxOffset(brk_topic)

	//Check if given offset is between min max
	if postBody.Offset < min_offset || postBody.Offset > max_offset {
		err := APIErrorInvalidData("Offset out of bounds")
		respondErr(w, err)
		log.Error(string(body[:]))
	}

	// Get subscription offsets

	refStr.UpdateSubOffset(projectUUID, urlSub, postBody.Offset)

	respondOK(w, output)

}

// SubGetOffsets (GET) gets offset metrics from a subscription
func SubGetOffsets(w http.ResponseWriter, r *http.Request) {

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

	results, err := subscriptions.Find(projectUUID, "", urlVars["subscription"], "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Subscription")
		respondErr(w, err)
		return
	}

	// Output result to JSON
	brk_topic := projectUUID + "." + results.Subscriptions[0].Topic
	cur_offset := results.Subscriptions[0].Offset
	min_offset := refBrk.GetMinOffset(brk_topic)
	max_offset := refBrk.GetMaxOffset(brk_topic)

	// Create offset struct
	offResult := subscriptions.Offsets{Current: cur_offset, Min: min_offset, Max: max_offset}
	resJSON, err := offResult.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

func SubTimeToOffset(w http.ResponseWriter, r *http.Request) {

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

	results, err := subscriptions.Find(projectUUID, "", urlVars["subscription"], "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Subscription")
		respondErr(w, err)
		return
	}

	t, err := time.Parse("2006-01-02T15:04:05.000Z", r.URL.Query().Get("time"))
	if err != nil {
		err := APIErrorInvalidData("Time is not in valid Zulu format.")
		respondErr(w, err)
		return
	}

	// Output result to JSON
	brk_topic := projectUUID + "." + results.Subscriptions[0].Topic
	off, err := refBrk.TimeToOffset(brk_topic, t.Local())

	if err != nil {
		log.Errorf(err.Error())
		err := APIErrGenericBackend()
		respondErr(w, err)
		return
	}

	if off < 0 {
		err := APIErrorGenericConflict("Timestamp is out of bounds for the subscription's topic/partition")
		respondErr(w, err)
		return
	}

	topicOffset := brokers.TopicOffset{Offset: off}
	output, err = json.Marshal(topicOffset)
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

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
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refBrk := gorillaContext.Get(r, "brk").(brokers.Broker)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Get Result Object

	err := topics.RemoveTopic(projectUUID, urlVars["topic"], refStr)
	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("Topic")
			respondErr(w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	fullTopic := projectUUID + "." + urlVars["topic"]
	err = refBrk.DeleteTopic(fullTopic)
	if err != nil {
		log.Errorf("Couldn't delete topic %v from broker, %v", fullTopic, err.Error())
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
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Get Result Object
	results, err := subscriptions.Find(projectUUID, "", urlVars["subscription"], "", 0, refStr)
	if err != nil {
		err := APIErrGenericBackend()
		respondErr(w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Subscription")
		respondErr(w, err)
		return
	}

	err = subscriptions.RemoveSub(projectUUID, urlVars["subscription"], refStr)
	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("Subscription")
			respondErr(w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	// if it is a push sub and it is also has a verified push endpoint, deactivate it
	if results.Subscriptions[0].PushCfg != (subscriptions.PushConfig{}) {
		if results.Subscriptions[0].PushCfg.Verified {
			pr := make(map[string]string)
			apsc := gorillaContext.Get(r, "apsc").(push.Client)
			pr["message"] = apsc.DeactivateSubscription(context.TODO(), results.Subscriptions[0].FullName).Result()
			b, _ := json.Marshal(pr)
			output = b
		}
	}
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
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// Parse pull options
	postBody, err := auth.GetACLFromJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("Topic ACL")
		respondErr(w, err)
		log.Error(string(body[:]))
		return
	}

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// check if user list contain valid users for the given project
	_, err = auth.AreValidUsers(projectUUID, postBody.AuthUsers, refStr)
	if err != nil {
		err := APIErrorRoot{Body: APIErrorBody{Code: http.StatusNotFound, Message: err.Error(), Status: "NOT_FOUND"}}
		respondErr(w, err)
		return
	}

	err = auth.ModACL(projectUUID, "topics", urlTopic, postBody.AuthUsers, refStr)

	if err != nil {

		if err.Error() == "not found" {
			err := APIErrorNotFound("Topic")
			respondErr(w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	respondOK(w, output)

}

// SubModACL (POST) modifies the ACL
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
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// Parse pull options
	postBody, err := auth.GetACLFromJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("Subscription ACL")
		respondErr(w, err)
		log.Error(string(body[:]))
		return
	}

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// check if user list contain valid users for the given project
	_, err = auth.AreValidUsers(projectUUID, postBody.AuthUsers, refStr)
	if err != nil {
		err := APIErrorRoot{Body: APIErrorBody{Code: http.StatusNotFound, Message: err.Error(), Status: "NOT_FOUND"}}
		respondErr(w, err)
		return
	}

	err = auth.ModACL(projectUUID, "subscriptions", urlSub, postBody.AuthUsers, refStr)

	if err != nil {

		if err.Error() == "not found" {
			err := APIErrorNotFound("Subscription")
			respondErr(w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	respondOK(w, output)

}

// SubModPush (POST) modifies the push configuration
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

	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// Parse pull options
	postBody, err := subscriptions.GetFromJSON(body)
	if err != nil {
		APIErrorInvalidArgument("Subscription")
		log.Error(string(body[:]))
		return
	}

	pushEnd := ""
	rPolicy := ""
	rPeriod := 0
	vhash := ""
	verified := false
	maxMessages := int64(0)
	pushWorker := auth.User{}
	pwToken := gorillaContext.Get(r, "push_worker_token").(string)

	if postBody.PushCfg != (subscriptions.PushConfig{}) {

		pushEnabled := gorillaContext.Get(r, "push_enabled").(bool)

		// check the state of the push functionality
		if !pushEnabled {
			err := APIErrorPushConflict()
			respondErr(w, err)
			return
		}

		pushWorker, err = auth.GetPushWorker(pwToken, refStr)
		if err != nil {
			err := APIErrInternalPush()
			respondErr(w, err)
			return
		}

		pushEnd = postBody.PushCfg.Pend
		// Check if push endpoint is not a valid https:// endpoint
		if !(isValidHTTPS(pushEnd)) {
			err := APIErrorInvalidData("Push endpoint should be addressed by a valid https url")
			respondErr(w, err)
			return
		}
		rPolicy = postBody.PushCfg.RetPol.PolicyType
		rPeriod = postBody.PushCfg.RetPol.Period
		maxMessages = postBody.PushCfg.MaxMessages

		if rPolicy == "" {
			rPolicy = subscriptions.LinearRetryPolicyType
		}
		if rPeriod <= 0 {
			rPeriod = 3000
		}

		if !subscriptions.IsRetryPolicySupported(rPolicy) {
			err := APIErrorInvalidData(subscriptions.UnSupportedRetryPolicyError)
			respondErr(w, err)
			return
		}
	}

	// Get Result Object
	res, err := subscriptions.Find(projectUUID, "", subName, "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(w, err)
		return
	}

	if res.Empty() {
		err := APIErrorNotFound("Subscription")
		respondErr(w, err)
		return
	}

	existingSub := res.Subscriptions[0]

	if maxMessages == 0 {
		if existingSub.PushCfg.MaxMessages == 0 {
			maxMessages = int64(1)
		} else {
			maxMessages = existingSub.PushCfg.MaxMessages
		}
	}

	// if the request wants to transform a pull subscription to a push one
	// we need to begin the verification process
	if postBody.PushCfg != (subscriptions.PushConfig{}) {

		// if the endpoint in not the same with the old one, we need to verify it again
		if postBody.PushCfg.Pend != existingSub.PushCfg.Pend {
			vhash, err = auth.GenToken()
			if err != nil {
				log.Errorf("Could not generate verification hash for subscription %v, %v", urlVars["subscription"], err.Error())
				err := APIErrGenericInternal("Could not generate verification hash")
				respondErr(w, err)
				return
			}
			// else keep the already existing data
		} else {
			vhash = existingSub.PushCfg.VerificationHash
			verified = existingSub.PushCfg.Verified
		}
	}

	err = subscriptions.ModSubPush(projectUUID, subName, pushEnd, maxMessages, rPolicy, rPeriod, vhash, verified, refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("Subscription")
			respondErr(w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	// if this is an deactivate request, try to retrieve the push worker in order to remove him from the sub's acl
	if existingSub.PushCfg != (subscriptions.PushConfig{}) && postBody.PushCfg == (subscriptions.PushConfig{}) {
		pushWorker, _ = auth.GetPushWorker(pwToken, refStr)
	}

	// if the sub, was push enabled before the update and the endpoint was verified
	// we need to deactivate it on the push server
	if existingSub.PushCfg != (subscriptions.PushConfig{}) {
		if existingSub.PushCfg.Verified {
			// deactivate the subscription on the push backend
			apsc := gorillaContext.Get(r, "apsc").(push.Client)
			apsc.DeactivateSubscription(context.TODO(), existingSub.FullName)

			// remove the push worker user from the sub's acl
			err = auth.RemoveFromACL(projectUUID, "subscriptions", existingSub.Name, []string{pushWorker.Name}, refStr)
			if err != nil {
				err := APIErrGenericInternal(err.Error())
				respondErr(w, err)
				return
			}
		}
	}
	// if the update on push configuration is not intended to stop the push functionality
	// activate the subscription with the new values
	if postBody.PushCfg != (subscriptions.PushConfig{}) {

		// reactivate only if the push endpoint hasn't changed and it wes already verified
		// otherwise we need to verify the ownership again before wee activate it
		if postBody.PushCfg.Pend == existingSub.PushCfg.Pend && existingSub.PushCfg.Verified {

			// activate the subscription on the push backend
			apsc := gorillaContext.Get(r, "apsc").(push.Client)
			apsc.ActivateSubscription(context.TODO(), existingSub.FullName, existingSub.FullTopic,
				pushEnd, rPolicy, uint32(rPeriod), maxMessages)

			// modify the sub's acl with the push worker's uuid
			err = auth.AppendToACL(projectUUID, "subscriptions", existingSub.Name, []string{pushWorker.Name}, refStr)
			if err != nil {
				err := APIErrGenericInternal(err.Error())
				respondErr(w, err)
				return
			}

			// link the sub's project with the push worker
			err = auth.AppendToUserProjects(pushWorker.UUID, projectUUID, refStr)
			if err != nil {
				err := APIErrGenericInternal(err.Error())
				respondErr(w, err)
				return
			}
		}
	}

	// Write empty response if everything's ok
	respondOK(w, output)
}

// SubVerifyPushEndpoint (POST) verifies the ownership of a push endpoint registered in a push enabled subscription
func SubVerifyPushEndpoint(w http.ResponseWriter, r *http.Request) {

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	subName := urlVars["subscription"]

	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	pwToken := gorillaContext.Get(r, "push_worker_token").(string)

	pushEnabled := gorillaContext.Get(r, "push_enabled").(bool)

	pushW := auth.User{}

	// check the state of the push functionality
	if !pushEnabled {
		err := APIErrorPushConflict()
		respondErr(w, err)
		return
	}

	pushW, err := auth.GetPushWorker(pwToken, refStr)
	if err != nil {
		err := APIErrInternalPush()
		respondErr(w, err)
		return
	}

	// Get Result Object
	res, err := subscriptions.Find(projectUUID, "", subName, "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(w, err)
		return
	}

	if res.Empty() {
		err := APIErrorNotFound("Subscription")
		respondErr(w, err)
		return
	}

	sub := res.Subscriptions[0]

	// check that the subscription is push enabled
	if sub.PushCfg == (subscriptions.PushConfig{}) {
		err := APIErrorGenericConflict("Subscription is not in push mode")
		respondErr(w, err)
		return
	}

	// check that the endpoint isn't already verified
	if sub.PushCfg.Verified {
		err := APIErrorGenericConflict("Push endpoint is already verified")
		respondErr(w, err)
		return
	}

	// verify the push endpoint
	c := new(http.Client)
	err = subscriptions.VerifyPushEndpoint(sub, c, refStr)
	if err != nil {
		err := APIErrPushVerification(err.Error())
		respondErr(w, err)
		return
	}

	// activate the subscription on the push backend
	apsc := gorillaContext.Get(r, "apsc").(push.Client)
	apsc.ActivateSubscription(context.TODO(), sub.FullName, sub.FullTopic, sub.PushCfg.Pend,
		sub.PushCfg.RetPol.PolicyType, uint32(sub.PushCfg.RetPol.Period), sub.PushCfg.MaxMessages)

	// modify the sub's acl with the push worker's uuid
	err = auth.AppendToACL(projectUUID, "subscriptions", sub.Name, []string{pushW.Name}, refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	// link the sub's project with the push worker
	err = auth.AppendToUserProjects(pushW.UUID, projectUUID, refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	respondOK(w, []byte{})
}

// SubModAck (POST) modifies the Ack deadline of the subscription
func SubModAck(w http.ResponseWriter, r *http.Request) {

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
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// Parse pull options
	postBody, err := subscriptions.GetAckDeadlineFromJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("ackDeadlineSeconds(needs value between 0 and 600)")
		respondErr(w, err)
		log.Error(string(body[:]))
		return
	}

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	err = subscriptions.ModAck(projectUUID, urlSub, postBody.AckDeadline, refStr)

	if err != nil {
		if err.Error() == "wrong value" {
			respondErr(w, APIErrorInvalidArgument("ackDeadlineSeconds(needs value between 0 and 600)"))
			return
		}
		if err.Error() == "not found" {
			err := APIErrorNotFound("Subscription")
			respondErr(w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

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
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refBrk := gorillaContext.Get(r, "brk").(brokers.Broker)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// Parse pull options
	postBody, err := subscriptions.GetFromJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("Subscription")
		respondErr(w, err)
		log.Error(string(body[:]))
		return
	}

	tProject, tName, err := subscriptions.ExtractFullTopicRef(postBody.FullTopic)

	if err != nil {
		err := APIErrorInvalidName("Topic")
		respondErr(w, err)
		return
	}

	if topics.HasTopic(projectUUID, tName, refStr) == false {
		err := APIErrorNotFound("Topic")
		respondErr(w, err)
		return
	}

	// Get current topic offset
	tProjectUUID := projects.GetUUIDByName(tProject, refStr)
	fullTopic := tProjectUUID + "." + tName
	curOff := refBrk.GetMaxOffset(fullTopic)

	pushEnd := ""
	rPolicy := ""
	rPeriod := 0
	maxMessages := int64(1)

	//pushWorker := auth.User{}
	verifyHash := ""

	if postBody.PushCfg != (subscriptions.PushConfig{}) {

		// check the state of the push functionality
		pwToken := gorillaContext.Get(r, "push_worker_token").(string)
		pushEnabled := gorillaContext.Get(r, "push_enabled").(bool)

		if !pushEnabled {
			err := APIErrorPushConflict()
			respondErr(w, err)
			return
		}

		_, err = auth.GetPushWorker(pwToken, refStr)
		if err != nil {
			err := APIErrInternalPush()
			respondErr(w, err)
			return
		}

		pushEnd = postBody.PushCfg.Pend
		// Check if push endpoint is not a valid https:// endpoint
		if !(isValidHTTPS(pushEnd)) {
			err := APIErrorInvalidData("Push endpoint should be addressed by a valid https url")
			respondErr(w, err)
			return
		}
		rPolicy = postBody.PushCfg.RetPol.PolicyType
		rPeriod = postBody.PushCfg.RetPol.Period
		maxMessages = postBody.PushCfg.MaxMessages

		if rPolicy == "" {
			rPolicy = subscriptions.LinearRetryPolicyType
		}

		if maxMessages == 0 {
			maxMessages = int64(1)
		}

		if rPeriod <= 0 {
			rPeriod = 3000
		}

		if !subscriptions.IsRetryPolicySupported(rPolicy) {
			err := APIErrorInvalidData(subscriptions.UnSupportedRetryPolicyError)
			respondErr(w, err)
			return
		}

		verifyHash, err = auth.GenToken()
		if err != nil {
			log.Errorf("Could not generate verification hash for subscription %v, %v", urlVars["subscription"], err.Error())
			err := APIErrGenericInternal("Could not generate verification hash")
			respondErr(w, err)
			return
		}

	}

	// Get Result Object
	res, err := subscriptions.CreateSub(projectUUID, urlVars["subscription"], tName, pushEnd, curOff, maxMessages, postBody.Ack, rPolicy, rPeriod, verifyHash, false, refStr)

	if err != nil {
		if err.Error() == "exists" {
			err := APIErrorConflict("Subscription")
			respondErr(w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
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
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	postBody := map[string]string{}
	schemaUUID := ""

	// check if there's a request body provided before trying to decode
	if r.Body != nil {

		b, err := ioutil.ReadAll(r.Body)

		if err != nil {
			err := APIErrorInvalidRequestBody()
			respondErr(w, err)
			return
		}
		defer r.Body.Close()

		if len(b) > 0 {
			err = json.Unmarshal(b, &postBody)
			if err != nil {
				err := APIErrorInvalidRequestBody()
				respondErr(w, err)
				return
			}

			schemaRef := postBody["schema"]

			// if there was a schema name provided, check its existence
			if schemaRef != "" {
				_, schemaName, err := schemas.ExtractSchema(schemaRef)
				if err != nil {
					err := APIErrorInvalidData(err.Error())
					respondErr(w, err)
					return
				}
				sl, err := schemas.Find(projectUUID, "", schemaName, refStr)
				if err != nil {
					err := APIErrGenericInternal(err.Error())
					respondErr(w, err)
					return
				}

				if sl.Empty() {
					err := APIErrorNotFound("Schema")
					respondErr(w, err)
					return
				}

				schemaUUID = sl.Schemas[0].UUID
			}
		}
	}
	// Get Result Object
	res, err := topics.CreateTopic(projectUUID, urlVars["topic"], schemaUUID, refStr)
	if err != nil {
		if err.Error() == "exists" {
			err := APIErrorConflict("Topic")
			respondErr(w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// ProjectMetrics (GET) metrics for one project (number of topics)
func ProjectMetrics(w http.ResponseWriter, r *http.Request) {

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
	//refRoles := gorillaContext.Get(r, "auth_roles").([]string)
	//refUser := gorillaContext.Get(r, "auth_user").(string)
	//refAuthResource := gorillaContext.Get(r, "auth_resource").(bool)

	urlProject := urlVars["project"]

	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Check Authorization per topic
	// - if enabled in config
	// - if user has only publisher role

	numTopics := int64(0)
	numSubs := int64(0)

	numTopics2, err2 := metrics.GetProjectTopics(projectUUID, refStr)
	if err2 != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}
	numTopics = numTopics2
	numSubs2, err2 := metrics.GetProjectSubs(projectUUID, refStr)
	if err2 != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}
	numSubs = numSubs2

	var timePoints []metrics.Timepoint
	var err error

	if timePoints, err = metrics.GetDailyProjectMsgCount(projectUUID, refStr); err != nil {
		err := APIErrGenericBackend()
		respondErr(w, err)
		return
	}

	m1 := metrics.NewProjectTopics(urlProject, numTopics, metrics.GetTimeNowZulu())
	m2 := metrics.NewProjectSubs(urlProject, numSubs, metrics.GetTimeNowZulu())
	res := metrics.NewMetricList(m1)
	res.Metrics = append(res.Metrics, m2)

	// ProjectUUID User topics aggregation
	m3, err := metrics.AggrProjectUserTopics(projectUUID, refStr)
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	for _, item := range m3.Metrics {
		res.Metrics = append(res.Metrics, item)
	}

	// ProjectUUID User subscriptions aggregation
	m4, err := metrics.AggrProjectUserSubs(projectUUID, refStr)
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	for _, item := range m4.Metrics {
		res.Metrics = append(res.Metrics, item)
	}

	m5 := metrics.NewDailyProjectMsgCount(urlProject, timePoints)
	res.Metrics = append(res.Metrics, m5)

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// TopicMetrics (GET) metrics for one topic
func TopicMetrics(w http.ResponseWriter, r *http.Request) {

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
	refRoles := gorillaContext.Get(r, "auth_roles").([]string)
	refUserUUID := gorillaContext.Get(r, "auth_user_uuid").(string)
	refAuthResource := gorillaContext.Get(r, "auth_resource").(bool)

	urlTopic := urlVars["topic"]

	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Check Authorization per topic
	// - if enabled in config
	// - if user has only publisher role

	if refAuthResource && auth.IsPublisher(refRoles) {

		if auth.PerResource(projectUUID, "topics", urlTopic, refUserUUID, refStr) == false {
			err := APIErrorForbidden()
			respondErr(w, err)
			return
		}
	}

	// Number of bytes and number of messages
	resultsMsg, err := topics.FindMetric(projectUUID, urlTopic, refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("Topic")
			respondErr(w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	numMsg := resultsMsg.MsgNum
	numBytes := resultsMsg.TotalBytes

	numSubs := int64(0)
	numSubs, err = metrics.GetProjectSubsByTopic(projectUUID, urlTopic, refStr)
	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("Topic")
			respondErr(w, err)
			return
		}
		err := APIErrGenericBackend()
		respondErr(w, err)
		return
	}

	var timePoints []metrics.Timepoint
	if timePoints, err = metrics.GetDailyTopicMsgCount(projectUUID, urlTopic, refStr); err != nil {
		err := APIErrGenericBackend()
		respondErr(w, err)
		return
	}

	m1 := metrics.NewTopicSubs(urlTopic, numSubs, metrics.GetTimeNowZulu())
	res := metrics.NewMetricList(m1)

	m2 := metrics.NewTopicMsgs(urlTopic, numMsg, metrics.GetTimeNowZulu())
	m3 := metrics.NewTopicBytes(urlTopic, numBytes, metrics.GetTimeNowZulu())
	m4 := metrics.NewDailyTopicMsgCount(urlTopic, timePoints)
	m5 := metrics.NewTopicRate(urlTopic, resultsMsg.PublishRate, resultsMsg.LatestPublish.Format("2006-01-02T15:04:05Z"))

	res.Metrics = append(res.Metrics, m2, m3, m4, m5)

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
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
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	results, err := topics.Find(projectUUID, "", urlVars["topic"], "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Topic")
		respondErr(w, err)
		return
	}

	res := results.Topics[0]

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// ListSubsByTopic (GET) lists all subscriptions associated with the given topic
func ListSubsByTopic(w http.ResponseWriter, r *http.Request) {

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	results, err := topics.Find(projectUUID, "", urlVars["topic"], "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Topic")
		respondErr(w, err)
		return
	}

	subs, err := subscriptions.FindByTopic(projectUUID, results.Topics[0].Name, refStr)
	if err != nil {
		err := APIErrGenericBackend()
		respondErr(w, err)
		return

	}

	resJSON, err := json.Marshal(subs)
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	respondOK(w, resJSON)
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
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)
	res, err := auth.GetACL(projectUUID, "topics", urlTopic, refStr)

	// If not found
	if err != nil {
		err := APIErrorNotFound("Topic")
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
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
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)
	res, err := auth.GetACL(projectUUID, "subscriptions", urlSub, refStr)

	// If not found
	if err != nil {
		err := APIErrorNotFound("Subscription")
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// SubMetrics (GET) metrics for one subscription
func SubMetrics(w http.ResponseWriter, r *http.Request) {

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
	refRoles := gorillaContext.Get(r, "auth_roles").([]string)
	refUserUUID := gorillaContext.Get(r, "auth_user_uuid").(string)
	refAuthResource := gorillaContext.Get(r, "auth_resource").(bool)

	urlSub := urlVars["subscription"]

	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Check Authorization per topic
	// - if enabled in config
	// - if user has only publisher role

	if refAuthResource && auth.IsConsumer(refRoles) {

		if auth.PerResource(projectUUID, "subscriptions", urlSub, refUserUUID, refStr) == false {
			err := APIErrorForbidden()
			respondErr(w, err)
			return
		}
	}

	resultMsg, err := subscriptions.FindMetric(projectUUID, urlSub, refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("Subscription")
			respondErr(w, err)
			return
		}
		err := APIErrGenericBackend()
		respondErr(w, err)
	}

	numMsg := resultMsg.MsgNum
	numBytes := resultMsg.TotalBytes

	m1 := metrics.NewSubMsgs(urlSub, numMsg, metrics.GetTimeNowZulu())
	res := metrics.NewMetricList(m1)
	m2 := metrics.NewSubBytes(urlSub, numBytes, metrics.GetTimeNowZulu())
	m3 := metrics.NewSubRate(urlSub, resultMsg.ConsumeRate, resultMsg.LatestConsume.Format("2006-01-02T15:04:05Z"))

	res.Metrics = append(res.Metrics, m2, m3)

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

//SubListAll (GET) all subscriptions
func SubListAll(w http.ResponseWriter, r *http.Request) {

	var err error
	var strPageSize string
	var pageSize int
	var res subscriptions.PaginatedSubscriptions

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

	// if this route is used by a user who only  has a consumer role
	// return all subscriptions that he has access to
	userUUID := ""
	if !auth.IsProjectAdmin(roles) && !auth.IsServiceAdmin(roles) && auth.IsConsumer(roles) {
		userUUID = gorillaContext.Get(r, "auth_user_uuid").(string)
	}

	if strPageSize != "" {
		if pageSize, err = strconv.Atoi(strPageSize); err != nil {
			log.Errorf("Pagesize %v produced an error  while being converted to int: %v", strPageSize, err.Error())
			err := APIErrorInvalidData("Invalid page size")
			respondErr(w, err)
			return
		}
	}

	if res, err = subscriptions.Find(projectUUID, userUUID, "", pageToken, int32(pageSize), refStr); err != nil {
		err := APIErrorInvalidData("Invalid page token")
		respondErr(w, err)
		return
	}

	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write Response
	output = []byte(resJSON)
	respondOK(w, output)

}

// TopicListAll (GET) all topics
func TopicListAll(w http.ResponseWriter, r *http.Request) {

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
			log.Errorf("Pagesize %v produced an error  while being converted to int: %v", strPageSize, err.Error())
			err := APIErrorInvalidData("Invalid page size")
			respondErr(w, err)
			return
		}
	}

	if res, err = topics.Find(projectUUID, userUUID, "", pageToken, int32(pageSize), refStr); err != nil {
		err := APIErrorInvalidData("Invalid page token")
		respondErr(w, err)
		return
	}
	// Output result to JSON
	resJSON, err := res.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
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

	refBrk := gorillaContext.Get(r, "brk").(brokers.Broker)
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refUserUUID := gorillaContext.Get(r, "auth_user_uuid").(string)
	refRoles := gorillaContext.Get(r, "auth_roles").([]string)
	refAuthResource := gorillaContext.Get(r, "auth_resource").(bool)
	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	results, err := topics.Find(projectUUID, "", urlVars["topic"], "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Topic")
		respondErr(w, err)
		return
	}

	res := results.Topics[0]

	// Check Authorization per topic
	// - if enabled in config
	// - if user has only publisher role

	if refAuthResource && auth.IsPublisher(refRoles) {

		if auth.PerResource(projectUUID, "topics", urlTopic, refUserUUID, refStr) == false {
			err := APIErrorForbidden()
			respondErr(w, err)
			return
		}
	}

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// Create Message List from Post JSON
	msgList, err := messages.LoadMsgListJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("Message")
		respondErr(w, err)
		log.Error(string(body[:]))
		return
	}

	// check if the topic has a schema associated with it
	if res.Schema != "" {

		// retrieve the schema
		_, schemaName, err := schemas.ExtractSchema(res.Schema)
		if err != nil {
			log.WithFields(
				log.Fields{
					"type":        "service_log",
					"schema_name": res.Schema,
					"topic_name":  res.Name,
					"error":       err.Error(),
				},
			).Error("Could not extract schema name")
			err := APIErrGenericInternal(schemas.GenericError)
			respondErr(w, err)
			return
		}

		sl, err := schemas.Find(projectUUID, "", schemaName, refStr)

		if err != nil {
			log.WithFields(
				log.Fields{
					"type":        "service_log",
					"schema_name": schemaName,
					"topic_name":  res.Name,
					"error":       err.Error(),
				},
			).Error("Could not retrieve schema from the store")
			err := APIErrGenericInternal(schemas.GenericError)
			respondErr(w, err)
			return
		}

		if !sl.Empty() {
			err := schemas.ValidateMessages(sl.Schemas[0], msgList)
			if err != nil {
				if err.Error() == "500" {
					err := APIErrGenericInternal(schemas.GenericError)
					respondErr(w, err)
					return
				} else {
					err := APIErrorInvalidData(err.Error())
					respondErr(w, err)
					return
				}
			}
		} else {
			log.WithFields(
				log.Fields{
					"type":        "service_log",
					"schema_name": res.Schema,
					"topic_name":  res.Name,
				},
			).Error("List of schemas was empty")
			err := APIErrGenericInternal(schemas.GenericError)
			respondErr(w, err)
			return
		}
	}

	// Init message ids list
	msgIDs := messages.MsgIDs{IDs: []string{}}

	// For each message in message list
	for _, msg := range msgList.Msgs {
		// Get offset and set it as msg
		fullTopic := projectUUID + "." + urlTopic

		msgID, rTop, _, _, err := refBrk.Publish(fullTopic, msg)

		if err != nil {
			if err.Error() == "kafka server: Message was too large, server rejected it to avoid allocation error." {
				err := APIErrTooLargeMessage("Message size too large")
				respondErr(w, err)
				return
			}

			err := APIErrGenericBackend()
			respondErr(w, err)
			return
		}

		msg.ID = msgID
		// Assertions for Succesfull Publish
		if rTop != fullTopic {
			err := APIErrGenericInternal("Broker reports wrong topic")
			respondErr(w, err)
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
	refStr.IncrementTopicMsgNum(projectUUID, urlTopic, msgCount)

	// increment daily count of topic messages
	year, month, day := publishTime.Date()
	refStr.IncrementDailyTopicMsgCount(projectUUID, urlTopic, msgCount, time.Date(year, month, day, 0, 0, 0, 0, time.UTC))

	// increment topic total bytes published
	refStr.IncrementTopicBytes(projectUUID, urlTopic, msgList.TotalSize())

	// update latest publish date for the given topic
	refStr.UpdateTopicLatestPublish(projectUUID, urlTopic, publishTime)

	// count the rate of published messages per sec between the last two publish events
	var dt float64 = 1
	// if its the first publish to the topic
	// skip the subtraction that computes the DT between the last two publish events
	if !res.LatestPublish.IsZero() {
		dt = publishTime.Sub(res.LatestPublish).Seconds()
	}
	refStr.UpdateTopicPublishRate(projectUUID, urlTopic, float64(msgCount)/dt)

	// Export the msgIDs
	resJSON, err := msgIDs.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
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
	refBrk := gorillaContext.Get(r, "brk").(brokers.Broker)
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refUserUUID := gorillaContext.Get(r, "auth_user_uuid").(string)
	refRoles := gorillaContext.Get(r, "auth_roles").([]string)
	refAuthResource := gorillaContext.Get(r, "auth_resource").(bool)
	pushEnabled := gorillaContext.Get(r, "push_enabled").(bool)

	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Get the subscription
	results, err := subscriptions.Find(projectUUID, "", urlSub, "", 0, refStr)
	if err != nil {
		err := APIErrGenericBackend()
		respondErr(w, err)
		return
	}

	if results.Empty() {
		err := APIErrorNotFound("Subscription")
		respondErr(w, err)
		return
	}

	targetSub := results.Subscriptions[0]
	fullTopic := targetSub.ProjectUUID + "." + targetSub.Topic
	retImm := true
	max := 1

	// if the subscription is push enabled but push enabled is false, don't allow push worker user to consume
	if targetSub.PushCfg != (subscriptions.PushConfig{}) && !pushEnabled && auth.IsPushWorker(refRoles) {
		err := APIErrorPushConflict()
		respondErr(w, err)
		return
	}

	// if the subscription is push enabled, allow only push worker and service_admin users to pull from it
	if targetSub.PushCfg != (subscriptions.PushConfig{}) && !auth.IsPushWorker(refRoles) && !auth.IsServiceAdmin(refRoles) {
		err := APIErrorForbidden()
		respondErr(w, err)
		return
	}

	// Check Authorization per subscription
	// - if enabled in config
	// - if user has only consumer role
	if refAuthResource && auth.IsConsumer(refRoles) {
		if auth.PerResource(projectUUID, "subscriptions", targetSub.Name, refUserUUID, refStr) == false {
			err := APIErrorForbidden()
			respondErr(w, err)
			return
		}
	}

	// check if the subscription's topic exists
	if !topics.HasTopic(projectUUID, targetSub.Topic, refStr) {
		err := APIErrorPullNoTopic()
		respondErr(w, err)
		return
	}

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(w, err)
		return
	}

	// Parse pull options
	pullInfo, err := subscriptions.GetPullOptionsJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("Pull Parameters")
		respondErr(w, err)
		log.Error(string(body[:]))
		return
	}

	if pullInfo.MaxMsg != "" {
		max, err = strconv.Atoi(pullInfo.MaxMsg)
		if err != nil {
			max = 1
		}
	}

	if pullInfo.RetImm == "false" {
		retImm = false
	}

	// Init Received Message List
	recList := messages.RecList{}

	msgs, err := refBrk.Consume(r.Context(), fullTopic, targetSub.Offset, retImm, int64(max))
	if err != nil {
		// If tracked offset is off
		if err == brokers.ErrOffsetOff {
			log.Debug("Will increment now...")
			// Increment tracked offset to current min offset
			targetSub.Offset = refBrk.GetMinOffset(fullTopic)
			refStr.UpdateSubOffset(projectUUID, targetSub.Name, targetSub.Offset)
			// Try again to consume
			msgs, err = refBrk.Consume(r.Context(), fullTopic, targetSub.Offset, retImm, int64(max))
			// If still error respond and return
			if err != nil {
				log.Errorf("Couldn't consume messages for subscription %v, %v", targetSub.FullName, err.Error())
				err := APIErrGenericBackend()
				respondErr(w, err)
				return
			}
		} else {
			log.Errorf("Couldn't consume messages for subscription %v, %v", targetSub.FullName, err.Error())
			err := APIErrGenericBackend()
			respondErr(w, err)
			return
		}
	}
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
			err := APIErrGenericInternal("Message retrieved from broker network has invalid JSON Structure")
			respondErr(w, err)
			return
		}
		// calc the message id = message's kafka offset (read offst + msg position)
		idOff := targetSub.Offset + int64(i)
		curMsg.ID = strconv.FormatInt(idOff, 10)
		curRec := messages.RecMsg{AckID: ackPrefix + curMsg.ID, Msg: curMsg}
		recList.RecMsgs = append(recList.RecMsgs, curRec)
	}

	// amount of messages consumed
	msgCount := int64(len(msgs))

	log.Debug(msgCount)

	// consumption time
	consumeTime := time.Now().UTC()

	// increment subscription number of message metric
	refStr.IncrementSubMsgNum(projectUUID, urlSub, msgCount)
	refStr.IncrementSubBytes(projectUUID, urlSub, recList.TotalSize())
	refStr.UpdateSubLatestConsume(projectUUID, targetSub.Name, consumeTime)

	// count the rate of consumed messages per sec between the last two consume events
	var dt float64 = 1
	// if its the first consume to the subscription
	// skip the subtraction that computes the DT between the last two consume events
	if !targetSub.LatestConsume.IsZero() {
		dt = consumeTime.Sub(targetSub.LatestConsume).Seconds()
	}

	refStr.UpdateSubConsumeRate(projectUUID, targetSub.Name, float64(msgCount)/dt)

	resJSON, err := recList.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Stamp time to UTC Z to seconds
	zSec := "2006-01-02T15:04:05Z"
	t := time.Now().UTC()
	ts := t.Format(zSec)
	refStr.UpdateSubPull(targetSub.ProjectUUID, targetSub.Name, int64(len(recList.RecMsgs))+targetSub.Offset, ts)

	output = []byte(resJSON)
	respondOK(w, output)
}

// HealthCheck returns an ok message to make sure the service is up and running
func HealthCheck(w http.ResponseWriter, r *http.Request) {

	var err error
	var bytes []byte

	apsc := gorillaContext.Get(r, "apsc").(push.Client)

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	healthMsg := HealthStatus{
		Status: "ok",
	}

	pwToken := gorillaContext.Get(r, "push_worker_token").(string)
	pushEnabled := gorillaContext.Get(r, "push_enabled").(bool)
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	if pushEnabled {
		_, err := auth.GetPushWorker(pwToken, refStr)
		if err != nil {
			healthMsg.Status = "warning"
		}

		healthMsg.PushServers = []PushServerInfo{
			{
				Endpoint: apsc.Target(),
				Status:   apsc.HealthCheck(context.TODO()).Result(),
			},
		}

	} else {
		healthMsg.PushFunctionality = "disabled"
	}

	if bytes, err = json.MarshalIndent(healthMsg, "", " "); err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	respondOK(w, bytes)
}

// SchemaCreate(POST) handles the creation of a new schema
func SchemaCreate(w http.ResponseWriter, r *http.Request) {

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
		respondErr(w, err)
		return
	}

	schema, err = schemas.Create(projectUUID, schemaUUID, schemaName, schema.Type, schema.RawSchema, refStr)
	if err != nil {
		if err.Error() == "exists" {
			err := APIErrorConflict("Schema")
			respondErr(w, err)
			return

		}

		if err.Error() == "unsupported" {
			err := APIErrorInvalidData(schemas.UnsupportedSchemaError)
			respondErr(w, err)
			return

		}

		err := APIErrorInvalidData(err.Error())
		respondErr(w, err)
		return
	}

	output, _ := json.MarshalIndent(schema, "", " ")
	respondOK(w, output)
}

// SchemaListOne(GET) retrieves information about the requested schema
func SchemaListOne(w http.ResponseWriter, r *http.Request) {

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
	schemasList, err := schemas.Find(projectUUID, "", schemaName, refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	if schemasList.Empty() {
		err := APIErrorNotFound("Schema")
		respondErr(w, err)
		return
	}

	output, _ := json.MarshalIndent(schemasList.Schemas[0], "", " ")
	respondOK(w, output)
}

// SchemaLisAll(GET) retrieves all the schemas under the given project
func SchemaListAll(w http.ResponseWriter, r *http.Request) {

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)
	schemasList, err := schemas.Find(projectUUID, "", "", refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	output, _ := json.MarshalIndent(schemasList, "", " ")
	respondOK(w, output)
}

// SchemaUpdate(PUT) updates the given schema
func SchemaUpdate(w http.ResponseWriter, r *http.Request) {

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
	schemasList, err := schemas.Find(projectUUID, "", schemaName, refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	if schemasList.Empty() {
		err := APIErrorNotFound("Schema")
		respondErr(w, err)
		return
	}

	updatedSchema := schemas.Schema{}

	err = json.NewDecoder(r.Body).Decode(&updatedSchema)
	if err != nil {
		err := APIErrorInvalidArgument("Schema")
		respondErr(w, err)
		return
	}

	if updatedSchema.FullName != "" {
		_, schemaName, err := schemas.ExtractSchema(updatedSchema.FullName)
		if err != nil {
			err := APIErrorInvalidData(err.Error())
			respondErr(w, err)
			return
		}
		updatedSchema.Name = schemaName
	}

	schema, err := schemas.Update(schemasList.Schemas[0], updatedSchema.Name, updatedSchema.Type, updatedSchema.RawSchema, refStr)
	if err != nil {
		if err.Error() == "exists" {
			err := APIErrorConflict("Schema")
			respondErr(w, err)
			return

		}

		if err.Error() == "unsupported" {
			err := APIErrorInvalidData(schemas.UnsupportedSchemaError)
			respondErr(w, err)
			return

		}

		err := APIErrorInvalidData(err.Error())
		respondErr(w, err)
		return
	}

	output, _ := json.MarshalIndent(schema, "", " ")
	respondOK(w, output)
}

func SchemaDelete(w http.ResponseWriter, r *http.Request) {

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

	schemasList, err := schemas.Find(projectUUID, "", schemaName, refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	if schemasList.Empty() {
		err := APIErrorNotFound("Schema")
		respondErr(w, err)
		return
	}

	err = schemas.Delete(schemasList.Schemas[0].UUID, refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	respondOK(w, nil)
}

// SchemaValidateMessage(POST) validates the given message against the schema
func SchemaValidateMessage(w http.ResponseWriter, r *http.Request) {

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
	schemasList, err := schemas.Find(projectUUID, "", schemaName, refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	if schemasList.Empty() {
		err := APIErrorNotFound("Schema")
		respondErr(w, err)
		return
	}

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		err := APIErrorInvalidData(err.Error())
		respondErr(w, err)
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
			respondErr(w, err)
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
			respondErr(w, err)
			return
		}
	}

	err = schemas.ValidateMessages(schemasList.Schemas[0], msgList)
	if err != nil {
		if err.Error() == "500" {
			err := APIErrGenericInternal(schemas.GenericError)
			respondErr(w, err)
			return
		} else {
			err := APIErrorInvalidData(err.Error())
			respondErr(w, err)
			return
		}
	}

	res, _ := json.MarshalIndent(map[string]string{"message": "Message validated successfully"}, "", " ")

	respondOK(w, res)
}

// ListVersion displays version information about the service
func ListVersion(w http.ResponseWriter, r *http.Request) {

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	v := version.Model{
		Release:   version.Release,
		Commit:    version.Commit,
		BuildTime: version.BuildTime,
		GO:        version.GO,
		Compiler:  version.Compiler,
		OS:        version.OS,
		Arch:      version.Arch,
	}

	output, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}
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
func respondErr(w http.ResponseWriter, apiErr APIErrorRoot) {
	log.Error(apiErr.Body.Code, "\t", apiErr.Body.Message)
	// set the response code
	w.WriteHeader(apiErr.Body.Code)
	// Output API Erorr object to JSON
	output, _ := json.MarshalIndent(apiErr, "", "   ")
	w.Write(output)
}

type HealthStatus struct {
	Status            string           `json:"status,omitempty"`
	PushServers       []PushServerInfo `json:"push_servers,omitempty"`
	PushFunctionality string           `json:"push_functionality,omitempty"`
}

type PushServerInfo struct {
	Endpoint string `json:"endpoint"`
	Status   string `json:"status"`
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

// IsValidHTTPS checks if a url string is valid https url
func isValidHTTPS(urlStr string) bool {
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return false
	}
	// If a valid url is in form without slashes after scheme consider it invalid.
	// If a valid url doesn't have https as a scheme consider it invalid
	if u.Host == "" || u.Scheme != "https" {
		return false
	}

	return true
}

// api err to be used when dealing with an invalid request body
var APIErrorInvalidRequestBody = func() APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusBadRequest, Message: "Invalid Request Body", Status: "BAD_REQUEST"}
	return APIErrorRoot{Body: apiErrBody}
}

// api err to be used when a name provided through the url parameters is not valid
var APIErrorInvalidName = func(key string) APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusBadRequest, Message: fmt.Sprintf("Invalid %v name", key), Status: "INVALID_ARGUMENT"}
	return APIErrorRoot{Body: apiErrBody}
}

// api err to be used when data provided is invalid
var APIErrorInvalidData = func(msg string) APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusBadRequest, Message: msg, Status: "INVALID_ARGUMENT"}
	return APIErrorRoot{Body: apiErrBody}
}

// api err to be used when argument's provided are invalid according to the resource
var APIErrorInvalidArgument = func(resource string) APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusBadRequest, Message: fmt.Sprintf("Invalid %v Arguments", resource), Status: "INVALID_ARGUMENT"}
	return APIErrorRoot{Body: apiErrBody}
}

// api err to be used when a user is unauthorized
var APIErrorUnauthorized = func() APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusUnauthorized, Message: "Unauthorized", Status: "UNAUTHORIZED"}
	return APIErrorRoot{Body: apiErrBody}
}

// api err to be used when access to a resource is forbidden for the request user
var APIErrorForbidden = func() APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusForbidden, Message: "Access to this resource is forbidden", Status: "FORBIDDEN"}
	return APIErrorRoot{Body: apiErrBody}
}

// api err to be used when access to a resource is forbidden for the request user
var APIErrorForbiddenWithMsg = func(msg string) APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusForbidden, Message: fmt.Sprintf("Access to this resource is forbidden. %v", msg), Status: "FORBIDDEN"}
	return APIErrorRoot{Body: apiErrBody}
}

// api err for dealing with absent resources
var APIErrorNotFound = func(resource string) APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusNotFound, Message: fmt.Sprintf("%v doesn't exist", resource), Status: "NOT_FOUND"}
	return APIErrorRoot{Body: apiErrBody}
}

// api err for dealing with  timeouts
var APIErrorTimeout = func(msg string) APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusRequestTimeout, Message: msg, Status: "TIMEOUT"}
	return APIErrorRoot{Body: apiErrBody}
}

// api err for dealing with already existing resources
var APIErrorConflict = func(resource string) APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusConflict, Message: fmt.Sprintf("%v already exists", resource), Status: "ALREADY_EXISTS"}
	return APIErrorRoot{Body: apiErrBody}
}

// api error to be used when push enabled false
var APIErrorPushConflict = func() APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusConflict,
		Message: "Push functionality is currently disabled",
		Status:  "CONFLICT",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api error to be used to format generic conflict errors
var APIErrorGenericConflict = func(msg string) APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusConflict,
		Message: msg,
		Status:  "CONFLICT",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api error to be used when push enabled false
var APIErrorPullNoTopic = func() APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusConflict,
		Message: "Subscription's topic doesn't exist",
		Status:  "CONFLICT",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api err for dealing with too large messages
var APIErrTooLargeMessage = func(resource string) APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusRequestEntityTooLarge, Message: "Message size is too large", Status: "INVALID_ARGUMENT"}
	return APIErrorRoot{Body: apiErrBody}
}

// api err for dealing with generic internal errors
var APIErrGenericInternal = func(msg string) APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusInternalServerError, Message: msg, Status: "INTERNAL_SERVER_ERROR"}
	return APIErrorRoot{Body: apiErrBody}
}

// api err for dealing with generic internal errors
var APIErrPushVerification = func(msg string) APIErrorRoot {
	apiErrBody := APIErrorBody{
		Code:    http.StatusUnauthorized,
		Message: fmt.Sprintf("Endpoint verification failed.%v", msg),
		Status:  "UNAUTHORIZED",
	}
	return APIErrorRoot{Body: apiErrBody}
}

// api err for dealing with internal errors when marshaling json to struct
var APIErrExportJSON = func() APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusInternalServerError, Message: "Error exporting data to JSON", Status: "INTERNAL_SERVER_ERROR"}
	return APIErrorRoot{Body: apiErrBody}
}

// api err for dealing with internal errors when querying the datastore
var APIErrQueryDatastore = func() APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusInternalServerError, Message: "Internal error while querying datastore", Status: "INTERNAL_SERVER_ERROR"}
	return APIErrorRoot{Body: apiErrBody}
}

// api err for dealing with internal errors related to acknowledgement
var APIErrHandlingAcknowledgement = func() APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusInternalServerError, Message: "Error handling acknowledgement", Status: "INTERNAL_SERVER_ERROR"}
	return APIErrorRoot{Body: apiErrBody}
}

// api err for dealing with generic backend errors
var APIErrGenericBackend = func() APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusInternalServerError, Message: "Backend Error", Status: "INTERNAL_SERVER_ERROR"}
	return APIErrorRoot{Body: apiErrBody}
}

// api error to be used when push enabled true but push worker was not able to be retrieved
var APIErrInternalPush = func() APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusInternalServerError,
		Message: "Push functionality is currently unavailable",
		Status:  "INTERNAL_SERVER_ERROR",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}
