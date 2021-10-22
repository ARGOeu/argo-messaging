package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/ARGOeu/argo-messaging/auth"
	"github.com/ARGOeu/argo-messaging/metrics"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/ARGOeu/argo-messaging/subscriptions"
	"github.com/ARGOeu/argo-messaging/topics"
	gorillaContext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
	"time"
)

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

// VaMetrics (GET) retrieves metrics regrading projects, users, subscriptions, topics
func VaMetrics(w http.ResponseWriter, r *http.Request) {

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

	// if no end date was provided, set it to today
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

	vr, err := metrics.GetVAReport(projectsList, startDate, endDate, refStr)
	if err != nil {
		err := APIErrorNotFound(err.Error())
		respondErr(w, err)
		return
	}

	output, err := json.MarshalIndent(vr, "", " ")
	if err != nil {
		err := APIErrExportJSON()
		respondErr(w, err)
		return
	}

	// Write response
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
