package handlers

import (
	"fmt"
	"github.com/ARGOeu/argo-messaging/auth"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/projects"
	"github.com/ARGOeu/argo-messaging/stores"
	gorillaContext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// UserProfile returns a user's profile based on the provided url parameter(key)
func UserProfile(w http.ResponseWriter, r *http.Request) {

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	authOption := gorillaContext.Get(r, "authOption").(config.AuthOption)

	tokenExtractStrategy := GetRequestTokenExtractStrategy(authOption)
	token := tokenExtractStrategy(r)

	if token == "" {
		err := APIErrorUnauthorized()
		respondErr(w, err)
		return
	}

	result, err := auth.GetUserByToken(token, refStr)

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
	res, err := auth.UpdateUser(userUUID, postBody.FirstName, postBody.LastName, postBody.Organization, postBody.Description,
		postBody.Name, postBody.Projects, postBody.Email, postBody.ServiceRoles, modified, true, refStr)

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
	res, err := auth.CreateUser(uuid, urlUser, postBody.FirstName, postBody.LastName, postBody.Organization, postBody.Description,
		postBody.Projects, token, postBody.Email, postBody.ServiceRoles, created, refUserUUID, refStr)

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

// UserListByUUID (GET) one user by uuid
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
	usersDetailedView := false

	// Grab url path variables
	urlValues := r.URL.Query()
	pageToken := urlValues.Get("pageToken")
	strPageSize := urlValues.Get("pageSize")
	projectName := urlValues.Get("project")
	details := urlValues.Get("details")
	projectUUID := ""

	if details == "true" {
		usersDetailedView = true
	}

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

	// check that user is indeed a service admin in order to be privileged to see full user info
	privileged := auth.IsServiceAdmin(refRoles)

	// Get Results Object - call is always privileged because this handler is only accessible by service admins
	paginatedUsers, err =
		auth.PaginatedFindUsers(pageToken, int64(pageSize), projectUUID, privileged, usersDetailedView, refStr)

	if err != nil {
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
