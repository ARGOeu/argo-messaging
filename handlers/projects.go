package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/ARGOeu/argo-messaging/auth"
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
	userFN := u.One().FirstName
	userLN := u.One().LastName
	userOrg := u.One().Organization
	userDesc := u.One().Description

	_, err = auth.UpdateUser(userUUID, userFN, userLN, userOrg, userDesc, userName, userProjects, userEmail, userSRoles, modified, false, refStr)

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
	userFN := u.One().FirstName
	userLN := u.One().LastName
	userOrg := u.One().Organization
	userDesc := u.One().Description

	_, err = auth.UpdateUser(userUUID, userFN, userLN, userOrg, userDesc, userName, userProjects, userEmail, userSRoles, modified, false, refStr)

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
	userFN := u.One().FirstName
	userLN := u.One().LastName
	userOrg := u.One().Organization
	userDesc := u.One().Description

	userProjects = append(userProjects, auth.ProjectRoles{
		Project: projName,
		Roles:   data.Roles,
	})

	_, err = auth.UpdateUser(userUUID, userFN, userLN, userOrg, userDesc, userName, userProjects, userEmail, userSRoles, modified, false, refStr)

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
	details := urlValues.Get("details")
	usersDetailedView := false

	if details == "true" {
		usersDetailedView = true
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
	if paginatedUsers, err = auth.PaginatedFindUsers(pageToken, int32(pageSize), projectUUID, priviledged, usersDetailedView, refStr); err != nil {
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
