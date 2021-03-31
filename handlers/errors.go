package handlers

import (
	"fmt"
	"net/http"
)

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

// api err to be used when dealing with an invalid request body
var APIErrorInvalidRequestBody = func() APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusBadRequest,
		Message: "Invalid Request Body",
		Status:  "BAD_REQUEST",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api err to be used when a name provided through the url parameters is not valid
var APIErrorInvalidName = func(key string) APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("Invalid %v name", key),
		Status:  "INVALID_ARGUMENT",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api err to be used when data provided is invalid
var APIErrorInvalidData = func(msg string) APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusBadRequest,
		Message: msg,
		Status:  "INVALID_ARGUMENT",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api err to be used when argument's provided are invalid according to the resource
var APIErrorInvalidArgument = func(resource string) APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("Invalid %v Arguments", resource),
		Status:  "INVALID_ARGUMENT",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api err to be used when a user is unauthorized
var APIErrorUnauthorized = func() APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusUnauthorized,
		Message: "Unauthorized",
		Status:  "UNAUTHORIZED",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api err to be used when access to a resource is forbidden for the request user
var APIErrorForbidden = func() APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusForbidden,
		Message: "Access to this resource is forbidden",
		Status:  "FORBIDDEN",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api err to be used when access to a resource is forbidden for the request user
var APIErrorForbiddenWithMsg = func(msg string) APIErrorRoot {
	apiErrBody := APIErrorBody{Code: http.StatusForbidden, Message: fmt.Sprintf("Access to this resource is forbidden. %v", msg), Status: "FORBIDDEN"}
	return APIErrorRoot{Body: apiErrBody}
}

// api err for dealing with absent resources
var APIErrorNotFound = func(resource string) APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("%v doesn't exist", resource),
		Status:  "NOT_FOUND",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api err for dealing with  timeouts
var APIErrorTimeout = func(msg string) APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusRequestTimeout,
		Message: msg,
		Status:  "TIMEOUT",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api err for dealing with already existing resources
var APIErrorConflict = func(resource string) APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusConflict,
		Message: fmt.Sprintf("%v already exists", resource),
		Status:  "ALREADY_EXISTS",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
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

	apiErrBody := APIErrorBody{
		Code:    http.StatusRequestEntityTooLarge,
		Message: "Message size is too large",
		Status:  "INVALID_ARGUMENT",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api err for dealing with generic internal errors
var APIErrGenericInternal = func(msg string) APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusInternalServerError,
		Message: msg,
		Status:  "INTERNAL_SERVER_ERROR",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api err for dealing with generic internal errors
var APIErrPushVerification = func(msg string) APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusUnauthorized,
		Message: fmt.Sprintf("Endpoint verification failed.%v", msg),
		Status:  "UNAUTHORIZED",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api err for dealing with internal errors when marshaling json to struct
var APIErrExportJSON = func() APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusInternalServerError,
		Message: "Error exporting data to JSON",
		Status:  "INTERNAL_SERVER_ERROR",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api err for dealing with internal errors when querying the datastore
var APIErrQueryDatastore = func() APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusInternalServerError,
		Message: "Internal error while querying datastore",
		Status:  "INTERNAL_SERVER_ERROR",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api err for dealing with internal errors related to acknowledgement
var APIErrHandlingAcknowledgement = func() APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusInternalServerError,
		Message: "Error handling acknowledgement",
		Status:  "INTERNAL_SERVER_ERROR",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
}

// api err for dealing with generic backend errors
var APIErrGenericBackend = func() APIErrorRoot {

	apiErrBody := APIErrorBody{
		Code:    http.StatusInternalServerError,
		Message: "Backend Error",
		Status:  "INTERNAL_SERVER_ERROR",
	}

	return APIErrorRoot{
		Body: apiErrBody,
	}
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
