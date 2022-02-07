package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/ARGOeu/argo-messaging/auth"
	"github.com/ARGOeu/argo-messaging/stores"
	gorillaContext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
	"io/ioutil"
	"net/http"
	"time"
)

// RegisterUser(POST) registers a new user
func RegisterUser(w http.ResponseWriter, r *http.Request) {

	// Init output
	output := []byte("")

	// Add content type header to the response
	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables

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
	requestBody := auth.UserRegistration{}
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		err := APIErrorInvalidArgument("User")
		respondErr(w, err)
		return
	}

	// check if a user with that name already exists
	if auth.ExistsWithName(requestBody.Name, refStr) {
		err := APIErrorConflict("User")
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

	ur, err := auth.RegisterUser(uuid, requestBody.Name, requestBody.FirstName, requestBody.LastName, requestBody.Email,
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
	regUUID := urlVars["uuid"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refUserUUID := gorillaContext.Get(r, "auth_user_uuid").(string)

	ru, err := auth.FindUserRegistration(regUUID, auth.PendingRegistrationStatus, refStr)
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
	err = auth.UpdateUserRegistration(regUUID, auth.AcceptedRegistrationStatus, "", refUserUUID, created, refStr)
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

func DeclineRegisterUser(w http.ResponseWriter, r *http.Request) {

	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	regUUID := urlVars["uuid"]
	refUserUUID := gorillaContext.Get(r, "auth_user_uuid").(string)

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	_, err := auth.FindUserRegistration(regUUID, auth.PendingRegistrationStatus, refStr)
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

	reqBody := make(map[string]string)

	// check the validity of the JSON
	if r.Body != nil {
		err = json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			err := APIErrorInvalidRequestBody()
			respondErr(w, err)
			return
		}
	}

	err = auth.UpdateUserRegistration(regUUID, auth.DeclinedRegistrationStatus, reqBody["comment"], refUserUUID, time.Now().UTC(), refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	respondOK(w, []byte("{}"))

}

// ListOneRegistration(GET) retrieves information for a specific registration based on the provided activation token
func ListOneRegistration(w http.ResponseWriter, r *http.Request) {

	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab url path variables
	urlVars := mux.Vars(r)
	regUUID := urlVars["uuid"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	ur, err := auth.FindUserRegistration(regUUID, "", refStr)
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

	urb, err := json.MarshalIndent(ur, "", "   ")
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	respondOK(w, urb)
}

// ListAllRegistrations(GET) retrieves information about all the registrations in the service
func ListAllRegistrations(w http.ResponseWriter, r *http.Request) {

	contentType := "application/json"
	charset := "utf-8"
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	name := r.URL.Query().Get("name")
	status := r.URL.Query().Get("status")
	email := r.URL.Query().Get("email")
	org := r.URL.Query().Get("organization")
	activationToken := r.URL.Query().Get("activation_token")

	ur, err := auth.FindUserRegistrations(status, activationToken, name, email, org, refStr)
	if err != nil {

		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	urb, err := json.MarshalIndent(ur, "", "   ")
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(w, err)
		return
	}

	respondOK(w, urb)
}
