package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ARGOeu/argo-messaging/auth"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/messages"
	"github.com/ARGOeu/argo-messaging/projects"
	push "github.com/ARGOeu/argo-messaging/push/grpc/client"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/ARGOeu/argo-messaging/subscriptions"
	"github.com/ARGOeu/argo-messaging/topics"
	"github.com/ARGOeu/argo-messaging/validation"
	gorillaContext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// SubAck (POST) acknowledge the consumption of specific messages
func SubAck(w http.ResponseWriter, r *http.Request) {
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
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)
	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(rCTX, w, err)
		return
	}

	// Parse pull options
	postBody, err := subscriptions.GetAckFromJSON(body)
	if err != nil {
		err := APIErrorInvalidData("Invalid ack parameter")
		respondErr(rCTX, w, err)
		return
	}

	// Get urlParams
	projectName := urlVars["project"]
	subName := urlVars["subscription"]

	// Check if sub exists

	cur_sub, err := subscriptions.Find(rCTX, projectUUID, "", subName, "", 0, refStr)
	if err != nil {
		err := APIErrHandlingAcknowledgement()
		respondErr(rCTX, w, err)
		return
	}
	if len(cur_sub.Subscriptions) == 0 {
		err := APIErrorNotFound("Subscription")
		respondErr(rCTX, w, err)
		return
	}

	// Get list of AckIDs
	if postBody.IDs == nil {
		err := APIErrorInvalidData("Invalid ack id")
		respondErr(rCTX, w, err)
		return
	}

	// Check if each AckID is valid
	for _, ackID := range postBody.IDs {
		if validation.ValidAckID(projectName, subName, ackID) == false {
			err := APIErrorInvalidData("Invalid ack id")
			respondErr(rCTX, w, err)
			return
		}
	}

	// Get Max ackID
	maxAckID, err := subscriptions.GetMaxAckID(postBody.IDs)
	if err != nil {
		err := APIErrHandlingAcknowledgement()
		respondErr(rCTX, w, err)
		return
	}
	// Extract offset from max ackID
	off, err := subscriptions.GetOffsetFromAckID(maxAckID)

	if err != nil {
		err := APIErrorInvalidData("Invalid ack id")
		respondErr(rCTX, w, err)
		return
	}

	zSec := "2006-01-02T15:04:05Z"
	t := time.Now().UTC()
	ts := t.Format(zSec)

	err = refStr.UpdateSubOffsetAck(rCTX, projectUUID, urlVars["subscription"], int64(off+1), ts)
	if err != nil {

		if err.Error() == "ack timeout" {
			err := APIErrorTimeout(err.Error())
			respondErr(rCTX, w, err)
			return
		}

		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
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
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	results, err := subscriptions.Find(rCTX, projectUUID, "", urlVars["subscription"], "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(rCTX, w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Subscription")
		respondErr(rCTX, w, err)
		return
	}

	// if its a push enabled sub and it has a verified endpoint
	// call the push server to find its real time push status
	if results.Subscriptions[0].PushCfg != (subscriptions.PushConfig{}) {
		if results.Subscriptions[0].PushCfg.Verified {
			apsc := gorillaContext.Get(r, "apsc").(push.Client)
			results.Subscriptions[0].PushStatus = apsc.SubscriptionStatus(context.TODO(), results.Subscriptions[0].FullName).Result(false)
		}
	}

	// Output result to JSON
	resJSON, err := results.Subscriptions[0].ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(rCTX, w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)
}

// SubSetOffset (PUT) sets subscriptions current offset
func SubSetOffset(w http.ResponseWriter, r *http.Request) {
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
	// Get Result Object
	urlSub := urlVars["subscription"]

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(rCTX, w, err)
		return
	}

	// Parse pull options
	postBody, err := subscriptions.GetSetOffsetJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("Offset")
		respondErr(rCTX, w, err)
		return
	}

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	refBrk := gorillaContext.Get(r, "brk").(brokers.Broker)
	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Find Subscription
	results, err := subscriptions.Find(rCTX, projectUUID, "", urlVars["subscription"], "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(rCTX, w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Subscription")
		respondErr(rCTX, w, err)
		return
	}
	brk_topic := projectUUID + "." + results.Subscriptions[0].Topic
	min_offset := refBrk.GetMinOffset(rCTX, brk_topic)
	max_offset := refBrk.GetMaxOffset(rCTX, brk_topic)

	//Check if given offset is between min max
	if postBody.Offset < min_offset || postBody.Offset > max_offset {
		err := APIErrorInvalidData("Offset out of bounds")
		respondErr(rCTX, w, err)
	}

	// Get subscription offsets
	refStr.UpdateSubOffset(rCTX, projectUUID, urlSub, postBody.Offset)

	respondOK(w, output)
}

// SubGetOffsets (GET) gets offset indices from a subscription
func SubGetOffsets(w http.ResponseWriter, r *http.Request) {
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
	refBrk := gorillaContext.Get(r, "brk").(brokers.Broker)

	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	results, err := subscriptions.Find(rCTX, projectUUID, "", urlVars["subscription"], "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(rCTX, w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Subscription")
		respondErr(rCTX, w, err)
		return
	}

	// Output result to JSON
	brkTopic := projectUUID + "." + results.Subscriptions[0].Topic
	curOffset := results.Subscriptions[0].Offset
	minOffset := refBrk.GetMinOffset(rCTX, brkTopic)
	maxOffset := refBrk.GetMaxOffset(rCTX, brkTopic)

	// if the current subscription offset is behind the min available offset for the topic
	// update it
	if curOffset < minOffset {
		refStr.UpdateSubOffset(rCTX, projectUUID, urlVars["subscription"], minOffset)
		curOffset = minOffset
	}

	// Create offset struct
	offResult := subscriptions.Offsets{
		Current: curOffset,
		Min:     minOffset,
		Max:     maxOffset,
	}

	resJSON, err := offResult.ExportJSON()
	if err != nil {
		err := APIErrExportJSON()
		respondErr(rCTX, w, err)
		return
	}

	// Write response
	output = []byte(resJSON)
	respondOK(w, output)
}

// SubTimeToOffset (GET) gets offset indices closest to a timestamp
func SubTimeToOffset(w http.ResponseWriter, r *http.Request) {
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
	refBrk := gorillaContext.Get(r, "brk").(brokers.Broker)

	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	results, err := subscriptions.Find(rCTX, projectUUID, "", urlVars["subscription"], "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(rCTX, w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Subscription")
		respondErr(rCTX, w, err)
		return
	}

	t, err := time.Parse("2006-01-02T15:04:05.000Z", r.URL.Query().Get("time"))
	if err != nil {
		err := APIErrorInvalidData("Time is not in valid Zulu format.")
		respondErr(rCTX, w, err)
		return
	}

	// Output result to JSON
	brkTopic := projectUUID + "." + results.Subscriptions[0].Topic
	off, err := refBrk.TimeToOffset(rCTX, brkTopic, t.Local())

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(rCTX, w, err)
		return
	}

	if off < 0 {
		err := APIErrorGenericConflict("Timestamp is out of bounds for the subscription's topic/partition")
		respondErr(rCTX, w, err)
		return
	}

	topicOffset := brokers.TopicOffset{Offset: off}
	output, err = json.Marshal(topicOffset)
	if err != nil {
		err := APIErrExportJSON()
		respondErr(rCTX, w, err)
		return
	}

	respondOK(w, output)
}

// SubDelete (DEL) deletes an existing subscription
func SubDelete(w http.ResponseWriter, r *http.Request) {
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
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Get Result Object
	results, err := subscriptions.Find(rCTX, projectUUID, "", urlVars["subscription"], "", 0, refStr)
	if err != nil {
		err := APIErrGenericBackend()
		respondErr(rCTX, w, err)
		return
	}

	// If not found
	if results.Empty() {
		err := APIErrorNotFound("Subscription")
		respondErr(rCTX, w, err)
		return
	}

	err = subscriptions.RemoveSub(rCTX, projectUUID, urlVars["subscription"], refStr)
	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("Subscription")
			respondErr(rCTX, w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	// if it is a push sub and it is also has a verified push endpoint, deactivate it
	if results.Subscriptions[0].PushCfg != (subscriptions.PushConfig{}) {
		if results.Subscriptions[0].PushCfg.Verified {
			pr := make(map[string]string)
			apsc := gorillaContext.Get(r, "apsc").(push.Client)
			pr["message"] = apsc.DeactivateSubscription(context.TODO(), results.Subscriptions[0].FullName).Result(false)
			b, _ := json.Marshal(pr)
			output = b
		}
	}
	respondOK(w, output)
}

// SubModACL (POST) modifies the ACL
func SubModACL(w http.ResponseWriter, r *http.Request) {
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
	// Get Result Object
	urlSub := urlVars["subscription"]

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(rCTX, w, err)
		return
	}

	// Parse pull options
	postBody, err := auth.GetACLFromJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("Subscription ACL")
		respondErr(rCTX, w, err)
		return
	}

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// check if user list contain valid users for the given project
	_, err = auth.AreValidUsers(rCTX, projectUUID, postBody.AuthUsers, refStr)
	if err != nil {
		err := APIErrorRoot{Body: APIErrorBody{Code: http.StatusNotFound, Message: err.Error(), Status: "NOT_FOUND"}}
		respondErr(rCTX, w, err)
		return
	}

	err = auth.ModACL(rCTX, projectUUID, "subscriptions", urlSub, postBody.AuthUsers, refStr)

	if err != nil {

		if err.Error() == "not found" {
			err := APIErrorNotFound("Subscription")
			respondErr(rCTX, w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	respondOK(w, output)
}

// SubModPush (POST) modifies the push configuration
func SubModPush(w http.ResponseWriter, r *http.Request) {
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
	subName := urlVars["subscription"]

	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(rCTX, w, err)
		return
	}

	// Parse pull options
	postBody, err := subscriptions.GetFromJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("Subscription")
		respondErr(rCTX, w, err)
		return
	}

	// Get Result Object
	res, err := subscriptions.Find(rCTX, projectUUID, "", subName, "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(rCTX, w, err)
		return
	}

	if res.Empty() {
		err := APIErrorNotFound("Subscription")
		respondErr(rCTX, w, err)
		return
	}

	existingSub := res.Subscriptions[0]

	pushEnd := ""
	rPolicy := ""
	rPeriod := 0
	vhash := ""
	verified := false
	authzType := subscriptions.AutoGenerationAuthorizationHeader
	authzHeaderValue := ""
	maxMessages := int64(1)
	pushWorker := auth.User{}
	pwToken := gorillaContext.Get(r, "push_worker_token").(string)
	mattermostUrl := ""
	mattermostUsername := ""
	mattermostChannel := ""
	pushType := ""
	base64Decode := false

	if postBody.PushCfg != (subscriptions.PushConfig{}) {

		pushEnabled := gorillaContext.Get(r, "push_enabled").(bool)

		// check the state of the push functionality
		if !pushEnabled {
			err := APIErrorPushConflict()
			respondErr(rCTX, w, err)
			return
		}

		pushWorker, err = auth.GetPushWorker(rCTX, pwToken, refStr)
		if err != nil {
			err := APIErrInternalPush()
			respondErr(rCTX, w, err)
			return
		}

		base64Decode = postBody.PushCfg.Base64Decode

		rPolicy = postBody.PushCfg.RetPol.PolicyType
		rPeriod = postBody.PushCfg.RetPol.Period

		if rPolicy == "" {
			rPolicy = subscriptions.LinearRetryPolicyType
		}
		if rPeriod <= 0 {
			rPeriod = 3000
		}

		if !subscriptions.IsRetryPolicySupported(rPolicy) {
			err := APIErrorInvalidData(subscriptions.UnSupportedRetryPolicyError)
			respondErr(rCTX, w, err)
			return
		}

		pushType = postBody.PushCfg.Type

		if pushType == subscriptions.HttpEndpointPushConfig {

			pushEnd = postBody.PushCfg.Pend
			// Check if push endpoint is not a valid https:// endpoint
			if !(validation.IsValidHTTPS(pushEnd)) {
				err := APIErrorInvalidData("Push endpoint should be addressed by a valid https url")
				respondErr(rCTX, w, err)
				return
			}

			// if the request wants to transform a pull subscription to a push one
			// we need to begin the verification process

			// if the endpoint in not the same with the old one, we need to verify it again
			if postBody.PushCfg.Pend != existingSub.PushCfg.Pend {
				vhash, err = auth.GenToken()
				if err != nil {
					log.WithFields(
						log.Fields{
							"trace_id":     rCTX.Value("trace_id"),
							"type":         "service_log",
							"subscription": urlVars["subscription"],
							"error":        err.Error(),
						},
					).Error("Could not generate verification hash for subscription")
					err := APIErrGenericInternal("Could not generate verification hash")
					respondErr(rCTX, w, err)
					return
				}
				// else keep the already existing data
			} else {
				vhash = existingSub.PushCfg.VerificationHash
				verified = existingSub.PushCfg.Verified
			}

			authzType = postBody.PushCfg.AuthorizationHeader.Type
			// if there is a given authorization type check if its supported by the service
			if authzType != "" {
				if !subscriptions.IsAuthorizationHeaderTypeSupported(authzType) {
					err := APIErrorInvalidData(subscriptions.UnSupportedAuthorizationHeader)
					respondErr(rCTX, w, err)
					return
				}
			}

			// if the subscription was not push enabled before
			// and no authorization_header has been specified
			// use autogen
			if authzType == "" && (existingSub.PushCfg == subscriptions.PushConfig{}) {
				authzType = subscriptions.AutoGenerationAuthorizationHeader
			}

			// if the provided authorization_header is of autogen type
			// generate a new header
			if authzType == subscriptions.AutoGenerationAuthorizationHeader {
				authzHeaderValue, err = auth.GenToken()
				if err != nil {
					log.WithFields(
						log.Fields{
							"trace_id":     rCTX.Value("trace_id"),
							"type":         "service_log",
							"subscription": urlVars["subscription"],
							"error":        err.Error(),
						},
					).Error("Could not generate auth header for subscription")
					err := APIErrGenericInternal("Could not generate authorization header")
					respondErr(rCTX, w, err)
					return
				}
			}

			// if the provided authorization_header is of disabled type
			if authzType == subscriptions.DisabledAuthorizationHeader {
				authzHeaderValue = ""
			}

			// if there is no authorization_type provided and the push cfg has an empty value because the sub
			// was push activated before the implementation of the feature
			// declare it disabled
			if authzType == "" && existingSub.PushCfg.AuthorizationHeader.Type == "" {
				authzType = subscriptions.DisabledAuthorizationHeader
			}

			// if there is no authorization_header provided but the existing one is of disabled type
			// preserve it
			if authzType == "" && existingSub.PushCfg.AuthorizationHeader.Type == subscriptions.DisabledAuthorizationHeader {
				authzType = subscriptions.DisabledAuthorizationHeader
			}

			// if there is no authorization_header provided but the existing one is of autogen type
			// preserve the value and type
			if authzType == "" && existingSub.PushCfg.AuthorizationHeader.Type == subscriptions.AutoGenerationAuthorizationHeader {
				authzType = subscriptions.AutoGenerationAuthorizationHeader
				authzHeaderValue = existingSub.PushCfg.AuthorizationHeader.Value
			}

			maxMessages = postBody.PushCfg.MaxMessages
			if maxMessages == 0 {
				if existingSub.PushCfg.MaxMessages == 0 {
					maxMessages = int64(1)
				} else {
					maxMessages = existingSub.PushCfg.MaxMessages
				}
			}
		} else if pushType == subscriptions.MattermostPushConfig {
			mattermostUrl = postBody.PushCfg.MattermostUrl
			mattermostChannel = postBody.PushCfg.MattermostChannel
			mattermostUsername = postBody.PushCfg.MattermostUsername
			verified = true

			if postBody.PushCfg.MattermostUrl == "" {
				err := APIErrorInvalidData("Field mattermostUrl cannot be empty")
				respondErr(rCTX, w, err)
				return
			}

		} else {
			err := APIErrorInvalidData(subscriptions.UnsupportedPushConfig)
			respondErr(rCTX, w, err)
			return
		}

	}

	cfg := subscriptions.PushConfig{
		Type:        pushType,
		Pend:        pushEnd,
		MaxMessages: maxMessages,
		AuthorizationHeader: subscriptions.AuthorizationHeader{
			Type:  authzType,
			Value: authzHeaderValue,
		},
		RetPol: subscriptions.RetryPolicy{
			PolicyType: rPolicy,
			Period:     rPeriod,
		},
		VerificationHash:   vhash,
		Verified:           verified,
		MattermostUrl:      mattermostUrl,
		MattermostUsername: mattermostUsername,
		MattermostChannel:  mattermostChannel,
		Base64Decode:       base64Decode,
	}
	err = subscriptions.ModSubPush(rCTX, projectUUID, subName, cfg, refStr)

	if err != nil {
		if err.Error() == "not found" {
			err := APIErrorNotFound("Subscription")
			respondErr(rCTX, w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	// if this is an deactivate request, try to retrieve the push worker in order to remove him from the sub's acl
	if existingSub.PushCfg != (subscriptions.PushConfig{}) && postBody.PushCfg == (subscriptions.PushConfig{}) {
		pushWorker, _ = auth.GetPushWorker(rCTX, pwToken, refStr)
	}

	// if the sub, was push enabled before the update and the endpoint was verified
	// (also works the same for mattermost push subs since they are always verified)
	// we need to deactivate it on the push server
	if existingSub.PushCfg != (subscriptions.PushConfig{}) {
		if existingSub.PushCfg.Verified {
			// deactivate the subscription on the push backend
			apsc := gorillaContext.Get(r, "apsc").(push.Client)
			apsc.DeactivateSubscription(context.TODO(), existingSub.FullName).Result(false)

			// remove the push worker user from the sub's acl
			err = auth.RemoveFromACL(rCTX, projectUUID, "subscriptions", existingSub.Name, []string{pushWorker.Name}, refStr)
			if err != nil {
				err := APIErrGenericInternal(err.Error())
				respondErr(rCTX, w, err)
				return
			}
		}
	}
	// if the update on push configuration is not intended to stop the push functionality
	// activate the subscription with the new values
	if postBody.PushCfg != (subscriptions.PushConfig{}) {

		// reactivate only if the push endpoint hasn't changed and it was already verified
		// otherwise we need to verify the ownership again before wee activate it
		if (postBody.PushCfg.Type == subscriptions.HttpEndpointPushConfig &&
			postBody.PushCfg.Pend == existingSub.PushCfg.Pend && existingSub.PushCfg.Verified) ||
			(postBody.PushCfg.Type == subscriptions.MattermostPushConfig) {

			//activate the subscription on the push backend
			apsc := gorillaContext.Get(r, "apsc").(push.Client)
			s := subscriptions.Subscription{
				FullName:  existingSub.FullName,
				FullTopic: existingSub.FullTopic,
				PushCfg: subscriptions.PushConfig{
					Type:        pushType,
					Pend:        pushEnd,
					MaxMessages: maxMessages,
					AuthorizationHeader: subscriptions.AuthorizationHeader{
						Value: authzHeaderValue,
					},
					RetPol: subscriptions.RetryPolicy{
						PolicyType: rPolicy,
						Period:     rPeriod,
					},
					MattermostUrl:      mattermostUrl,
					MattermostUsername: mattermostUsername,
					MattermostChannel:  mattermostChannel,
				},
			}
			apsc.ActivateSubscription(context.TODO(), s).Result(false)

			// modify the sub's acl with the push worker's uuid
			err = auth.AppendToACL(rCTX, projectUUID, "subscriptions", existingSub.Name, []string{pushWorker.Name}, refStr)
			if err != nil {
				err := APIErrGenericInternal(err.Error())
				respondErr(rCTX, w, err)
				return
			}

			// link the sub's project with the push worker
			err = auth.AppendToUserProjects(rCTX, pushWorker.UUID, projectUUID, refStr)
			if err != nil {
				err := APIErrGenericInternal(err.Error())
				respondErr(rCTX, w, err)
				return
			}
		}
	}

	// Write empty response if everything's ok
	respondOK(w, output)
}

// SubVerifyPushEndpoint (POST) verifies the ownership of a push endpoint registered in a push enabled subscription
func SubVerifyPushEndpoint(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

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
		respondErr(rCTX, w, err)
		return
	}

	pushW, err := auth.GetPushWorker(rCTX, pwToken, refStr)
	if err != nil {
		err := APIErrInternalPush()
		respondErr(rCTX, w, err)
		return
	}

	// Get Result Object
	res, err := subscriptions.Find(rCTX, projectUUID, "", subName, "", 0, refStr)

	if err != nil {
		err := APIErrGenericBackend()
		respondErr(rCTX, w, err)
		return
	}

	if res.Empty() {
		err := APIErrorNotFound("Subscription")
		respondErr(rCTX, w, err)
		return
	}

	sub := res.Subscriptions[0]

	// check that the subscription is push enabled
	if sub.PushCfg.Type != subscriptions.HttpEndpointPushConfig {
		err := APIErrorGenericConflict("Subscription is not in http push mode")
		respondErr(rCTX, w, err)
		return
	}

	// check that the endpoint isn't already verified
	if sub.PushCfg.Verified {
		err := APIErrorGenericConflict("Push endpoint is already verified")
		respondErr(rCTX, w, err)
		return
	}

	// verify the push endpoint
	c := new(http.Client)
	err = subscriptions.VerifyPushEndpoint(rCTX, sub, c, refStr)
	if err != nil {
		err := APIErrPushVerification(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	// activate the subscription on the push backend
	apsc := gorillaContext.Get(r, "apsc").(push.Client)
	err = activatePushSubscription(rCTX, sub, pushW, apsc, refStr)
	if err != nil {
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	respondOK(w, []byte{})
}

// SubModAck (POST) modifies the Ack deadline of the subscription
func SubModAck(w http.ResponseWriter, r *http.Request) {
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
	// Get Result Object
	urlSub := urlVars["subscription"]

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(rCTX, w, err)
		return
	}

	// Parse pull options
	postBody, err := subscriptions.GetAckDeadlineFromJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("ackDeadlineSeconds(needs value between 0 and 600)")
		respondErr(rCTX, w, err)
		return
	}

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)
	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	err = subscriptions.ModAck(rCTX, projectUUID, urlSub, postBody.AckDeadline, refStr)

	if err != nil {
		if err.Error() == "wrong value" {
			respondErr(rCTX, w, APIErrorInvalidArgument("ackDeadlineSeconds(needs value between 0 and 600)"))
			return
		}
		if err.Error() == "not found" {
			err := APIErrorNotFound("Subscription")
			respondErr(rCTX, w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	respondOK(w, output)
}

// SubCreate (PUT) creates a new subscription
func SubCreate(w http.ResponseWriter, r *http.Request) {
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
	refBrk := gorillaContext.Get(r, "brk").(brokers.Broker)
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(rCTX, w, err)
		return
	}

	// Parse pull options
	postBody, err := subscriptions.GetFromJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("Subscription")
		respondErr(rCTX, w, err)
		return
	}

	tProject, tName, err := subscriptions.ExtractFullTopicRef(postBody.FullTopic)

	if err != nil {
		err := APIErrorInvalidName("Topic")
		respondErr(rCTX, w, err)
		return
	}

	if topics.HasTopic(rCTX, projectUUID, tName, refStr) == false {
		err := APIErrorNotFound("Topic")
		respondErr(rCTX, w, err)
		return
	}

	// Get current topic offset
	tProjectUUID := projects.GetUUIDByName(rCTX, tProject, refStr)
	fullTopic := tProjectUUID + "." + tName
	curOff := refBrk.GetMaxOffset(rCTX, fullTopic)

	pushConfig := subscriptions.PushConfig{}

	if postBody.PushCfg != (subscriptions.PushConfig{}) {

		pushConfig.Type = postBody.PushCfg.Type

		// check the state of the push functionality
		pwToken := gorillaContext.Get(r, "push_worker_token").(string)
		pushEnabled := gorillaContext.Get(r, "push_enabled").(bool)

		if !pushEnabled {
			err := APIErrorPushConflict()
			respondErr(rCTX, w, err)
			return
		}

		_, err = auth.GetPushWorker(rCTX, pwToken, refStr)
		if err != nil {
			err := APIErrInternalPush()
			respondErr(rCTX, w, err)
			return
		}

		// checks for http endpoint push subscriptions
		if pushConfig.Type == subscriptions.HttpEndpointPushConfig {

			pushConfig.Pend = postBody.PushCfg.Pend

			// Check if push endpoint is not a valid https:// endpoint
			if !(validation.IsValidHTTPS(pushConfig.Pend)) {
				err := APIErrorInvalidData("Push endpoint should be addressed by a valid https url")
				respondErr(rCTX, w, err)
				return
			}

			pushConfig.MaxMessages = postBody.PushCfg.MaxMessages
			if pushConfig.MaxMessages == 0 {
				pushConfig.MaxMessages = int64(1)
			}

			pushConfig.AuthorizationHeader.Type = postBody.PushCfg.AuthorizationHeader.Type

			if pushConfig.AuthorizationHeader.Type == "" {
				pushConfig.AuthorizationHeader.Type = subscriptions.AutoGenerationAuthorizationHeader
			}

			if !subscriptions.IsAuthorizationHeaderTypeSupported(pushConfig.AuthorizationHeader.Type) {
				err := APIErrorInvalidData(subscriptions.UnSupportedAuthorizationHeader)
				respondErr(rCTX, w, err)
				return
			}

			switch pushConfig.AuthorizationHeader.Type {
			case subscriptions.AutoGenerationAuthorizationHeader:
				pushConfig.AuthorizationHeader.Value, err = auth.GenToken()
				if err != nil {
					log.WithFields(
						log.Fields{
							"trace_id":     rCTX.Value("trace_id"),
							"type":         "service_log",
							"subscription": urlVars["subscription"],
							"error":        err.Error(),
						},
					).Error("Could not generate auth header for subscription")
					err := APIErrGenericInternal("Could not generate authorization header")
					respondErr(rCTX, w, err)
					return
				}
			case subscriptions.DisabledAuthorizationHeader:
				pushConfig.AuthorizationHeader.Value = ""
			}

			pushConfig.VerificationHash, err = auth.GenToken()
			if err != nil {
				log.WithFields(
					log.Fields{
						"trace_id":     rCTX.Value("trace_id"),
						"type":         "service_log",
						"subscription": urlVars["subscription"],
						"error":        err.Error(),
					},
				).Error("Could not generate verification hash for subscription")
				err := APIErrGenericInternal("Could not generate verification hash")
				respondErr(rCTX, w, err)
				return
			}
			pushConfig.Verified = false
		} else if pushConfig.Type == subscriptions.MattermostPushConfig {
			if postBody.PushCfg.MattermostUrl == "" {
				err := APIErrorInvalidData("Field mattermostUrl cannot be empty")
				respondErr(rCTX, w, err)
				return
			}
			pushConfig.MattermostUrl = postBody.PushCfg.MattermostUrl
			pushConfig.MattermostUsername = postBody.PushCfg.MattermostUsername
			pushConfig.MattermostChannel = postBody.PushCfg.MattermostChannel
			pushConfig.Verified = true
		} else {
			err := APIErrorInvalidData(subscriptions.UnsupportedPushConfig)
			respondErr(rCTX, w, err)
			return
		}

		pushConfig.Base64Decode = postBody.PushCfg.Base64Decode

		pushConfig.RetPol.PolicyType = postBody.PushCfg.RetPol.PolicyType
		pushConfig.RetPol.Period = postBody.PushCfg.RetPol.Period

		if pushConfig.RetPol.PolicyType == "" {
			pushConfig.RetPol.PolicyType = subscriptions.LinearRetryPolicyType
		}

		if pushConfig.RetPol.Period <= 0 {
			pushConfig.RetPol.Period = 3000
		}

		if !subscriptions.IsRetryPolicySupported(pushConfig.RetPol.PolicyType) {
			err := APIErrorInvalidData(subscriptions.UnSupportedRetryPolicyError)
			respondErr(rCTX, w, err)
			return
		}

	}

	created := time.Now().UTC()

	// Get Result Object
	res, err := subscriptions.Create(rCTX, projectUUID, urlVars["subscription"], tName, curOff,
		postBody.Ack, pushConfig, created, refStr)

	if err != nil {
		if err.Error() == "exists" {
			err := APIErrorConflict("Subscription")
			respondErr(rCTX, w, err)
			return
		}
		err := APIErrGenericInternal(err.Error())
		respondErr(rCTX, w, err)
		return
	}

	// if the subscription is push enabled mattermost, activate it
	// on the push backend
	if res.PushCfg.Type == subscriptions.MattermostPushConfig {

		pwToken := gorillaContext.Get(r, "push_worker_token").(string)

		pushWorker, err := auth.GetPushWorker(rCTX, pwToken, refStr)
		if err != nil {
			err := APIErrInternalPush()
			respondErr(rCTX, w, err)
			return
		}

		apsc := gorillaContext.Get(r, "apsc").(push.Client)
		activatePushSubscription(rCTX, res, pushWorker, apsc, refStr)
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

// SubACL (GET) one sub's authorized users
func SubACL(w http.ResponseWriter, r *http.Request) {
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
	urlSub := urlVars["subscription"]

	// Grab context references
	refStr := gorillaContext.Get(r, "str").(stores.Store)

	// Get project UUID First to use as reference
	projectUUID := gorillaContext.Get(r, "auth_project_uuid").(string)
	res, err := auth.GetACL(rCTX, projectUUID, "subscriptions", urlSub, refStr)

	// If not found
	if err != nil {
		err := APIErrorNotFound("Subscription")
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

// SubListAll (GET) all subscriptions
func SubListAll(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)

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

	res, err = subscriptions.Find(rCTX, projectUUID, userUUID, "", pageToken, int64(pageSize), refStr)
	if err != nil {
		err := APIErrorInvalidData("Invalid page token")
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

	// Write Response
	output = []byte(resJSON)
	respondOK(w, output)
}

// SubPull (POST) consumes messages from the underlying topic
func SubPull(w http.ResponseWriter, r *http.Request) {
	traceId := gorillaContext.Get(r, "trace_id").(string)
	rCTX := context.WithValue(context.Background(), "trace_id", traceId)
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
	results, err := subscriptions.Find(rCTX, projectUUID, "", urlSub, "", 0, refStr)
	if err != nil {
		err := APIErrGenericBackend()
		respondErr(rCTX, w, err)
		return
	}

	if results.Empty() {
		err := APIErrorNotFound("Subscription")
		respondErr(rCTX, w, err)
		return
	}

	targetSub := results.Subscriptions[0]
	fullTopic := targetSub.ProjectUUID + "." + targetSub.Topic
	retImm := true
	max := 1

	// if the subscription is push enabled but push enabled is false, don't allow push worker user to consume
	if targetSub.PushCfg != (subscriptions.PushConfig{}) && !pushEnabled && auth.IsPushWorker(refRoles) {
		err := APIErrorPushConflict()
		respondErr(rCTX, w, err)
		return
	}

	// if the subscription is push enabled, allow only push worker and service_admin users to pull from it
	if targetSub.PushCfg != (subscriptions.PushConfig{}) && !auth.IsPushWorker(refRoles) && !auth.IsServiceAdmin(refRoles) {
		err := APIErrorForbidden()
		respondErr(rCTX, w, err)
		return
	}

	// Check Authorization per subscription
	// - if enabled in config
	// - if user has only consumer role
	if refAuthResource && auth.IsConsumer(refRoles) {
		if auth.PerResource(rCTX, projectUUID, "subscriptions", targetSub.Name, refUserUUID, refStr) == false {
			err := APIErrorForbidden()
			respondErr(rCTX, w, err)
			return
		}
	}

	// check if the subscription's topic exists
	if !topics.HasTopic(rCTX, projectUUID, targetSub.Topic, refStr) {
		err := APIErrorPullNoTopic()
		respondErr(rCTX, w, err)
		return
	}

	// Read POST JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := APIErrorInvalidRequestBody()
		respondErr(rCTX, w, err)
		return
	}

	// Parse pull options
	pullInfo, err := subscriptions.GetPullOptionsJSON(body)
	if err != nil {
		err := APIErrorInvalidArgument("Pull Parameters")
		respondErr(rCTX, w, err)
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

	msgs, err := refBrk.Consume(rCTX, fullTopic, targetSub.Offset, retImm, int64(max))
	if err != nil {
		// If tracked offset is off
		if err == brokers.ErrOffsetOff {
			log.WithFields(
				log.Fields{
					"trace_id":     rCTX.Value("trace_id"),
					"type":         "sservice_log",
					"subscription": targetSub.FullName,
				},
			).Debug("Will increment now . . .")
			// Increment tracked offset to current min offset
			targetSub.Offset = refBrk.GetMinOffset(rCTX, fullTopic)
			refStr.UpdateSubOffset(rCTX, projectUUID, targetSub.Name, targetSub.Offset)
			// Try again to consume
			msgs, err = refBrk.Consume(rCTX, fullTopic, targetSub.Offset, retImm, int64(max))
			// If still error respond and return
			if err != nil {
				log.WithFields(
					log.Fields{
						"trace_id":     rCTX.Value("trace_id"),
						"type":         "service_log",
						"error":        err.Error(),
						"subscription": targetSub.FullName,
					},
				).Error("Couldn't consume messages for subscription")
				err := APIErrGenericBackend()
				respondErr(rCTX, w, err)
				return
			}
		} else {
			log.WithFields(
				log.Fields{
					"trace_id":     rCTX.Value("trace_id"),
					"type":         "service_log",
					"error":        err.Error(),
					"subscription": targetSub.FullName,
				},
			).Error("Couldn't consume messages for subscription")
			err := APIErrGenericBackend()
			respondErr(rCTX, w, err)
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
			respondErr(rCTX, w, err)
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

	// consumption time
	consumeTime := time.Now().UTC()

	// increment subscription number of message metric
	refStr.IncrementSubMsgNum(rCTX, projectUUID, urlSub, msgCount)
	refStr.IncrementSubBytes(rCTX, projectUUID, urlSub, recList.TotalSize())
	refStr.UpdateSubLatestConsume(rCTX, projectUUID, targetSub.Name, consumeTime)

	// count the rate of consumed messages per sec between the last two consume events
	var dt float64 = 1
	// if its the first consume to the subscription
	// skip the subtraction that computes the DT between the last two consume events
	if !targetSub.LatestConsume.IsZero() {
		dt = consumeTime.Sub(targetSub.LatestConsume).Seconds()
	}

	refStr.UpdateSubConsumeRate(rCTX, projectUUID, targetSub.Name, float64(msgCount)/dt)

	resJSON, err := recList.ExportJSON()

	if err != nil {
		err := APIErrExportJSON()
		respondErr(rCTX, w, err)
		return
	}

	// Stamp time to UTC Z to seconds
	zSec := "2006-01-02T15:04:05Z"
	t := time.Now().UTC()
	ts := t.Format(zSec)
	refStr.UpdateSubPull(rCTX, targetSub.ProjectUUID, targetSub.Name, int64(len(recList.RecMsgs))+targetSub.Offset, ts)

	output = []byte(resJSON)
	respondOK(w, output)
}

func activatePushSubscription(rCTX context.Context, sub subscriptions.Subscription, pushW auth.User,
	apsc push.Client, refStr stores.Store) error {

	// activate the subscription on the push server
	apsc.ActivateSubscription(context.TODO(), sub).Result(false)

	// modify the sub's acl with the push worker's uuid
	err := auth.AppendToACL(rCTX, sub.ProjectUUID, "subscriptions", sub.Name, []string{pushW.Name}, refStr)
	if err != nil {
		return err
	}

	// link the sub's project with the push worker
	err = auth.AppendToUserProjects(rCTX, pushW.UUID, sub.ProjectUUID, refStr)
	if err != nil {
		return err
	}

	return nil
}
