package handlers

import (
	"context"
	"fmt"
	"github.com/ARGOeu/argo-messaging/auth"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/projects"
	"github.com/ARGOeu/argo-messaging/stores"
	gorillaContext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// UserProfile returns a user's profile based on the provided url parameter(key)
func UserProfile(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

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
		respondErr(rCTX, w, err)
		return
	}

	result, err := auth.GetUserByToken(rCTX, token, refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorUnauthorized()
			respondErr(rCTX, w, err)
			return
		}
		err := APIErrQueryDatastore()
		respondErr(rCTX, w, err)
		return
	}

	// Output result to JSON
	resJSON, err := result.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(rCTX, w, err)
		return
	}

	// Write response
	respondOK(w, []byte(resJSON))

}

// RefreshToken (POST) refreshes user's token
func RefreshToken(w http.ResponseWriter, r *http.Request) {
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
	urlUser := urlVars["user"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get Result Object
	userUUID := auth.GetUUIDByName(rCTX, urlUser, refStr)
	token, err := auth.GenToken() // generate a new user token

	res, err := auth.UpdateUserToken(rCTX, userUUID, token, refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(rCTX, w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
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

// UserUpdate (PUT) updates the user information
func UserUpdate(w http.ResponseWriter, r *http.Request) {
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
	urlUser := urlVars["user"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Read POST JSON body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(rCTX, w, err)
		return
	}

	// Parse pull options
	postBody, err := auth.GetUserFromJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("User")
		respondErr(rCTX, w, err)
		return
	}

	// Get Result Object
	userUUID := auth.GetUUIDByName(rCTX, urlUser, refStr)
	modified := time.Now().UTC()
	res, err := auth.UpdateUser(rCTX, userUUID, postBody.FirstName, postBody.LastName, postBody.Organization, postBody.Description,
		postBody.Name, postBody.Projects, postBody.Email, postBody.ServiceRoles, modified, true, refStr)

	if err != nil {

		// In case of invalid project or role in post body

		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(rCTX, w, err)
			return
		}

		if strings.HasPrefix(err.Error(), "invalid") {
			err := APIErrorInvalidData(err.Error())
			respondErr(rCTX, w, err)
			return
		}

		if strings.HasPrefix(err.Error(), "duplicate") {
			err := APIErrorInvalidData(err.Error())
			respondErr(rCTX, w, err)
			return
		}

		err := APIErrGenericInternal(err.Error())
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

// UserCreate (POST) creates a new user inside a project
func UserCreate(w http.ResponseWriter, r *http.Request) {
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
	urlUser := urlVars["user"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refUserUUID := gorillaContext.Get(r, "auth_user_uuid").(string)

	// Read POST JSON body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(rCTX, w, err)
		return
	}

	// Parse pull options
	postBody, err := auth.GetUserFromJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("User")
		respondErr(rCTX, w, err)
		return
	}

	uuid := uuid.NewV4().String() // generate a new uuid to attach to the new project
	token, err := auth.GenToken() // generate a new user token
	created := time.Now().UTC()
	// Get Result Object
	res, err := auth.CreateUser(rCTX, uuid, urlUser, postBody.FirstName, postBody.LastName, postBody.Organization, postBody.Description,
		postBody.Projects, token, postBody.Email, postBody.ServiceRoles, created, refUserUUID, refStr)

	if err != nil {
		if err.Error() == "exists" {
			err := APIErrorConflict("User")
			respondErr(rCTX, w, err)
			return
		}

		if strings.HasPrefix(err.Error(), "duplicate") {
			err := APIErrorInvalidData(err.Error())
			respondErr(rCTX, w, err)
			return
		}

		if strings.HasPrefix(err.Error(), "invalid") {
			err := APIErrorInvalidData(err.Error())
			respondErr(rCTX, w, err)
			return
		}

		err := APIErrGenericInternal(err.Error())
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

// UserListByToken (GET) one user by his token
func UserListByToken(w http.ResponseWriter, r *http.Request) {
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
	urlToken := urlVars["token"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get Results Object
	result, err := auth.GetUserByToken(rCTX, urlToken, refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(rCTX, w, err)
			return
		}
		err := APIErrQueryDatastore()
		respondErr(rCTX, w, err)
		return
	}

	// Output result to JSON
	resJSON, err := result.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(rCTX, w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)

}

// UserListOne (GET) one user
func UserListOne(w http.ResponseWriter, r *http.Request) {
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
	urlUser := urlVars["user"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get Results Object
	results, err := auth.FindUsers(rCTX, "", "", urlUser, true, refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(rCTX, w, err)
			return
		}

		err := APIErrQueryDatastore()
		respondErr(rCTX, w, err)
		return
	}

	res := results.One()

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

// UserListByUUID (GET) one user by uuid
func UserListByUUID(w http.ResponseWriter, r *http.Request) {
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

	// Get Results Object
	result, err := auth.GetUserByUUID(rCTX, urlVars["uuid"], refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(rCTX, w, err)
			return
		}

		if err.Error() == "multiple uuids" {
			err := APIErrGenericInternal("Multiple users found with the same uuid")
			respondErr(rCTX, w, err)
			return
		}

		err := APIErrQueryDatastore()
		respondErr(rCTX, w, err)
		return
	}

	// Output result to JSON
	resJSON, err := result.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(rCTX, w, err)
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)
}

// UserListAll (GET) all users - or users belonging to a project
func UserListAll(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

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
		projectUUID = projects.GetUUIDByName(rCTX, projectName, refStr)
		if projectUUID == "" {
			err := APIErrorNotFound("ProjectUUID")
			respondErr(rCTX, w, err)
			return
		}
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

	// check that user is indeed a service admin in order to be privileged to see full user info
	privileged := auth.IsServiceAdmin(refRoles)

	// Get Results Object - call is always privileged because this handler is only accessible by service admins
	paginatedUsers, err =
		auth.PaginatedFindUsers(rCTX, pageToken, int64(pageSize), projectUUID, privileged, usersDetailedView, refStr)

	if err != nil {
		err := APIErrorInvalidData("Invalid page token")
		respondErr(rCTX, w, err)
		return
	}

	// Output result to JSON
	resJSON, err := paginatedUsers.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(rCTX, w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)
}

// UserDelete (DEL) deletes an existing user
func UserDelete(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

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

	userUUID := auth.GetUUIDByName(rCTX, urlUser, refStr)

	err := auth.RemoveUser(rCTX, userUUID, refStr)
	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("User")
			respondErr(rCTX, w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	// Write empty response if anything ok
	respondOK(w, output)
}
