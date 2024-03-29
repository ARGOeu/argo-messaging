package subscriptions

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ARGOeu/argo-messaging/projects"
	"github.com/ARGOeu/argo-messaging/stores"
	log "github.com/sirupsen/logrus"
)

const (
	LinearRetryPolicyType             = "linear"
	SlowStartRetryPolicyType          = "slowstart"
	AutoGenerationAuthorizationHeader = "autogen"
	DisabledAuthorizationHeader       = "disabled"
	HttpEndpointPushConfig            = "http_endpoint"
	MattermostPushConfig              = "mattermost"
	UnSupportedRetryPolicyError       = `Retry policy can only be of 'linear' or 'slowstart' type`
	UnSupportedAuthorizationHeader    = `Authorization header type can only be of 'autogen' or 'disabled' type`
	UnsupportedPushConfig             = `Push configuration type can only be of 'http_endpoint' or 'mattermost'`
)

var supportedRetryPolicyTypes = []string{
	LinearRetryPolicyType,
	SlowStartRetryPolicyType,
}

var supportedAuthorizationHeaderTypes = []string{
	AutoGenerationAuthorizationHeader,
	DisabledAuthorizationHeader,
}

var supportedPushConfigTypes = []string{
	HttpEndpointPushConfig,
	MattermostPushConfig,
}

// Subscription struct to hold information for a given topic
type Subscription struct {
	ProjectUUID   string     `json:"-"`
	Name          string     `json:"-"`
	Topic         string     `json:"-"`
	FullName      string     `json:"name"`
	FullTopic     string     `json:"topic"`
	PushCfg       PushConfig `json:"pushConfig"`
	Ack           int        `json:"ackDeadlineSeconds"`
	Offset        int64      `json:"-"`
	NextOffset    int64      `json:"-"`
	PendingAck    string     `json:"-"`
	PushStatus    string     `json:"pushStatus,omitempty"`
	CreatedOn     string     `json:"createdOn"`
	LatestConsume time.Time  `json:"-"`
	ConsumeRate   float64    `json:"-"`
}

// PushConfig holds optional configuration for push operations
type PushConfig struct {
	Type                string              `json:"type"`
	Pend                string              `json:"pushEndpoint"`
	MaxMessages         int64               `json:"maxMessages"`
	AuthorizationHeader AuthorizationHeader `json:"authorizationHeader"`
	RetPol              RetryPolicy         `json:"retryPolicy"`
	VerificationHash    string              `json:"verificationHash"`
	Verified            bool                `json:"verified"`
	MattermostUrl       string              `json:"mattermostUrl"`
	MattermostUsername  string              `json:"mattermostUsername"`
	MattermostChannel   string              `json:"mattermostChannel"`
	Base64Decode        bool                `json:"base64Decode"`
}

// SubMetrics holds the subscription's metric details
type SubMetrics struct {
	MsgNum        int64     `json:"number_of_messages"`
	TotalBytes    int64     `json:"total_bytes"`
	LatestConsume time.Time `json:"-"`
	ConsumeRate   float64   `json:"-"`
}

// RetryPolicy holds information on retry policies
type RetryPolicy struct {
	PolicyType string `json:"type,omitempty"`
	Period     int    `json:"period,omitempty"`
}

type AuthorizationHeader struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// PaginatedSubscriptions holds information about a subscriptions' page and how to access the next page
type PaginatedSubscriptions struct {
	Subscriptions []Subscription `json:"subscriptions"`
	NextPageToken string         `json:"nextPageToken"`
	TotalSize     int64          `json:"totalSize"`
}

// SubPullOptions holds info about a pull operation on a subscription
type SubPullOptions struct {
	RetImm string `json:"returnImmediately,omitempty"`
	MaxMsg string `json:"maxMessages,omitempty"`
}

// SetOffset structure is used for input in set Offset Request
type SetOffset struct {
	Offset int64 `json:"offset"`
}

// Offsets is used as a json structure for show offsets Response
type Offsets struct {
	Max     int64 `json:"max"`
	Min     int64 `json:"min"`
	Current int64 `json:"current"`
}

// AckIDs utility struct
type AckIDs struct {
	IDs []string `json:"AckIds"`
}

// Ack utility struct
type AckDeadline struct {
	AckDeadline int `json:"ackDeadlineSeconds"`
}

type NamesList struct {
	Subscriptions []string `json:"subscriptions"`
}

func NewNamesList() NamesList {
	return NamesList{
		Subscriptions: make([]string, 0),
	}
}

// IsRetryPolicySupported checks if the provided retry policy is supported by the service
func IsRetryPolicySupported(retPol string) bool {

	for _, rp := range supportedRetryPolicyTypes {
		if rp == retPol {
			return true
		}
	}
	return false
}

// IsAuthorizationHeaderTypeSupported checks if the provided authorization header type is supported by the service
func IsAuthorizationHeaderTypeSupported(authzType string) bool {

	for _, aht := range supportedAuthorizationHeaderTypes {
		if authzType == aht {
			return true
		}
	}
	return false
}

// FindMetric returns the metric of a specific subscription
func FindMetric(ctx context.Context, projectUUID string, name string, store stores.Store) (SubMetrics, error) {
	result := SubMetrics{MsgNum: 0}
	subs, _, _, err := store.QuerySubs(ctx, projectUUID, "", name, "", 0)

	// check if sub exists
	if len(subs) == 0 {
		return result, errors.New("not found")
	}

	for _, item := range subs {
		projectName := projects.GetNameByUUID(ctx, item.ProjectUUID, store)
		if projectName == "" {
			return result, errors.New("invalid project")
		}

		result.MsgNum = item.MsgNum
		result.TotalBytes = item.TotalBytes
		result.LatestConsume = item.LatestConsume
		result.ConsumeRate = item.ConsumeRate

	}
	return result, err
}

// Empty returns true if Subscriptions list has no items
func (sl *PaginatedSubscriptions) Empty() bool {
	return len(sl.Subscriptions) <= 0
}

// GetAckFromJSON retrieves ack ids from json
func GetAckFromJSON(input []byte) (AckIDs, error) {
	s := AckIDs{}
	err := json.Unmarshal([]byte(input), &s)
	return s, err
}

// GetSetOffsetJSON retrieves set offset information
func GetSetOffsetJSON(input []byte) (SetOffset, error) {
	s := SetOffset{}
	err := json.Unmarshal([]byte(input), &s)
	return s, err
}

// GetPullOptionsJSON retrieves pull information
func GetPullOptionsJSON(input []byte) (SubPullOptions, error) {
	s := SubPullOptions{}
	err := json.Unmarshal([]byte(input), &s)
	return s, err
}

// GetAckDeadlineFromJSON retrieves ack deadline from json input
func GetAckDeadlineFromJSON(input []byte) (AckDeadline, error) {
	s := AckDeadline{}
	err := json.Unmarshal([]byte(input), &s)
	return s, err
}

// GetFromJSON retrieves Sub Info From Json
func GetFromJSON(input []byte) (Subscription, error) {
	s := Subscription{}
	err := json.Unmarshal([]byte(input), &s)
	return s, err
}

// New creates a new subscription based on name
func New(projectUUID string, projectName string, name string, topic string) Subscription {
	fsn := "/projects/" + projectName + "/subscriptions/" + name
	ftn := "/projects/" + projectName + "/topics/" + topic
	ps := PushConfig{}
	s := Subscription{ProjectUUID: projectUUID, Name: name, Topic: topic, FullName: fsn, FullTopic: ftn, PushCfg: ps, Ack: 10}
	return s
}

// ExportJSON exports metrics as a json string
func (offs *SubMetrics) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(offs, "", "   ")
	return string(output[:]), err
}

// ExportJSON exports offsets structure as a json string
func (offs *Offsets) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(offs, "", "   ")
	return string(output[:]), err
}

// ExportJSON exports whole sub Structure as a json string
func (sub *Subscription) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(sub, "", "   ")
	return string(output[:]), err
}

// PushEndpointHost extracts the host:port of a push endpoint
func (sub *Subscription) PushEndpointHost() string {

	if sub.PushCfg.Pend == "" {
		return ""
	}

	u, err := url.Parse(sub.PushCfg.Pend)
	if err != nil {
		return ""
	}

	return u.Host
}

// ExportJSON exports whole sub List Structure as a json string
func (sl *PaginatedSubscriptions) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(sl, "", "   ")
	return string(output[:]), err
}

// VerifyPushEndpoint verifies the ownership of a push endpoint
func VerifyPushEndpoint(ctx context.Context, sub Subscription, c *http.Client, store stores.Store) error {

	// extract the push endpoint host
	if sub.PushCfg.Pend == "" {
		return errors.New("Could not retrieve push endpoint host")
	}

	u1 := &url.URL{}
	u1, err := url.Parse(sub.PushCfg.Pend)
	if err != nil {
		return err
	}

	// create a new url that will be used to retrieve the verification hash
	u := url.URL{
		Scheme: u1.Scheme,
		Host:   u1.Host,
		Path:   "ams_verification_hash",
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		log.WithFields(
			log.Fields{
				"trace_id":        ctx.Value("trace_id"),
				"type":            "backend_log",
				"backend_service": "remote_endpoint",
				"backend_hosts":   u.String(),
				"subscription":    sub.FullName,
				"status":          resp.StatusCode,
			},
		).Error("failed to verify push endpoint for subscription")
		return errors.New("Wrong response status code")
	} else {
		// read the response
		buf := bytes.Buffer{}
		buf.ReadFrom(resp.Body)

		defer resp.Body.Close()

		if sub.PushCfg.VerificationHash != buf.String() {
			log.WithFields(
				log.Fields{
					"trace_id":        ctx.Value("trace_id"),
					"type":            "backend_log",
					"backend_service": "remote_endpoint",
					"backend_hosts":   u.String(),
					"subscription":    sub.FullName,
					"status":          resp.StatusCode,
					"expected_hash":   sub.PushCfg.VerificationHash,
					"actual_hash":     buf.String(),
				},
			).Error("failed to verify hash for push endpoint of subscription")
			return errors.New("Wrong verification hash")
		}
	}

	// update the push config with verified true
	cfg := PushConfig{
		Type:                sub.PushCfg.Type,
		Pend:                sub.PushCfg.Pend,
		MaxMessages:         sub.PushCfg.MaxMessages,
		AuthorizationHeader: sub.PushCfg.AuthorizationHeader,
		RetPol:              sub.PushCfg.RetPol,
		VerificationHash:    sub.PushCfg.VerificationHash,
		Verified:            true,
		MattermostUrl:       sub.PushCfg.MattermostUrl,
		MattermostUsername:  sub.PushCfg.MattermostUsername,
		MattermostChannel:   sub.PushCfg.MattermostChannel,
	}
	err = ModSubPush(ctx, sub.ProjectUUID, sub.Name, cfg, store)
	if err != nil {
		return err
	}

	return nil
}

// Find searches the store for all subscriptions of a given project or a specific one
func Find(ctx context.Context, projectUUID, userUUID, name, pageToken string, pageSize int64, store stores.Store) (PaginatedSubscriptions, error) {

	var err error
	var qSubs []stores.QSub
	var totalSize int64
	var nextPageToken string
	var pageTokenBytes []byte

	result := PaginatedSubscriptions{Subscriptions: []Subscription{}}

	// decode the base64 pageToken
	if pageTokenBytes, err = base64.StdEncoding.DecodeString(pageToken); err != nil {
		log.WithFields(
			log.Fields{
				"trace_id":   ctx.Value("trace_id"),
				"type":       "request_log",
				"page_token": pageToken,
				"error":      err.Error(),
			},
		).Error("error while decoding to base64")
		return result, err
	}

	if qSubs, totalSize, nextPageToken, err = store.QuerySubs(ctx, projectUUID, userUUID, name, string(pageTokenBytes), pageSize); err != nil {
		return result, err
	}

	projectName := projects.GetNameByUUID(ctx, projectUUID, store)

	if projectName == "" {
		return result, errors.New("invalid project")
	}

	for _, item := range qSubs {
		curSub := New(item.ProjectUUID, projectName, item.Name, item.Topic)
		curSub.Offset = item.Offset
		curSub.NextOffset = item.NextOffset
		curSub.Ack = item.Ack
		curSub.CreatedOn = item.CreatedOn.UTC().Format("2006-01-02T15:04:05Z")
		if item.PushType != "" {
			rp := RetryPolicy{
				PolicyType: item.RetPolicy,
			}

			if item.RetPolicy != SlowStartRetryPolicyType {
				rp.Period = item.RetPeriod
			}

			maxM := int64(1)
			if item.MaxMessages != 0 {
				maxM = item.MaxMessages
			}

			authzCFG := AuthorizationHeader{
				Type:  item.AuthorizationType,
				Value: item.AuthorizationHeader,
			}

			curSub.PushCfg = PushConfig{
				Pend:                item.PushEndpoint,
				MaxMessages:         maxM,
				AuthorizationHeader: authzCFG,
				RetPol:              rp,
				VerificationHash:    item.VerificationHash,
				Verified:            item.Verified,
				MattermostUrl:       item.MattermostUrl,
				MattermostChannel:   item.MattermostChannel,
				MattermostUsername:  item.MattermostUsername,
				Type:                item.PushType,
				Base64Decode:        item.Base64Decode,
			}
		}
		curSub.LatestConsume = item.LatestConsume
		curSub.ConsumeRate = item.ConsumeRate
		result.Subscriptions = append(result.Subscriptions, curSub)
	}

	result.NextPageToken = base64.StdEncoding.EncodeToString([]byte(nextPageToken))
	result.TotalSize = totalSize

	return result, err
}

// FindByTopic retrieves all subscriptions associated with the given topic
func FindByTopic(ctx context.Context, projectUUID string, topicName string, store stores.Store) (NamesList, error) {

	subs, err := store.QuerySubsByTopic(ctx, projectUUID, topicName)
	if err != nil {
		return NewNamesList(), err
	}

	subNames := NewNamesList()
	projectName := projects.GetNameByUUID(ctx, projectUUID, store)

	for _, sub := range subs {
		subNames.Subscriptions = append(subNames.Subscriptions,
			fmt.Sprintf("/projects/%v/subscriptions/%v", projectName, sub.Name))
	}

	return subNames, nil
}

// LoadPushSubs returns all subscriptions defined in store that have a push configuration
func LoadPushSubs(store stores.Store) PaginatedSubscriptions {
	result := PaginatedSubscriptions{Subscriptions: []Subscription{}}
	subs := store.QueryPushSubs(context.Background())
	for _, item := range subs {
		projectName := projects.GetNameByUUID(context.Background(), item.ProjectUUID, store)
		curSub := New(item.ProjectUUID, projectName, item.Name, item.Topic)
		curSub.Offset = item.Offset
		curSub.NextOffset = item.NextOffset
		curSub.Ack = item.Ack
		rp := RetryPolicy{item.RetPolicy, item.RetPeriod}
		curSub.PushCfg = PushConfig{Pend: item.PushEndpoint, RetPol: rp}

		result.Subscriptions = append(result.Subscriptions, curSub)
	}
	return result
}

// Create creates a new subscription
func Create(ctx context.Context, projectUUID string, name string, topic string, offset int64, ack int,
	pushCfg PushConfig, createdOn time.Time, store stores.Store) (Subscription, error) {

	if HasSub(ctx, projectUUID, name, store) {
		return Subscription{}, errors.New("exists")
	}

	if ack == 0 {
		ack = 10
	}

	if pushCfg.RetPol.PolicyType == SlowStartRetryPolicyType {
		pushCfg.RetPol.Period = 0
	}

	qPushCfg := stores.QPushConfig{
		Type:                pushCfg.Type,
		PushEndpoint:        pushCfg.Pend,
		MaxMessages:         pushCfg.MaxMessages,
		AuthorizationType:   pushCfg.AuthorizationHeader.Type,
		AuthorizationHeader: pushCfg.AuthorizationHeader.Value,
		RetPolicy:           pushCfg.RetPol.PolicyType,
		RetPeriod:           pushCfg.RetPol.Period,
		VerificationHash:    pushCfg.VerificationHash,
		Verified:            pushCfg.Verified,
		MattermostChannel:   pushCfg.MattermostChannel,
		MattermostUrl:       pushCfg.MattermostUrl,
		MattermostUsername:  pushCfg.MattermostUsername,
		Base64Decode:        pushCfg.Base64Decode,
	}

	err := store.InsertSub(ctx, projectUUID, name, topic, offset, ack, qPushCfg, createdOn)
	if err != nil {
		return Subscription{}, errors.New("backend error")
	}

	results, err := Find(ctx, projectUUID, "", name, "", 0, store)
	if len(results.Subscriptions) != 1 {
		return Subscription{}, errors.New("backend error")
	}

	return results.Subscriptions[0], err
}

// ModAck updates the subscription's acknowledgment timeout
func ModAck(ctx context.Context, projectUUID string, name string, ack int, store stores.Store) error {
	// minimum deadline allowed 0 seconds, maximum: 600 sec (10 minutes)
	if ack < 0 || ack > 600 {
		return errors.New("wrong value")
	}

	if HasSub(ctx, projectUUID, name, store) == false {
		return errors.New("not found")
	}

	log.WithFields(
		log.Fields{
			"trace_id": ctx.Value("trace_id"),
			"type":     "service_log",
			"deadline": ack,
		},
	).Info("modifying ack deadline")

	return store.ModAck(ctx, projectUUID, name, ack)
}

// ModSubPush updates the subscription push config
func ModSubPush(ctx context.Context, projectUUID string, name string, pushCfg PushConfig, store stores.Store) error {

	if HasSub(ctx, projectUUID, name, store) == false {
		return errors.New("not found")
	}

	if pushCfg.RetPol.PolicyType == SlowStartRetryPolicyType {
		pushCfg.RetPol.Period = 0
	}

	qPushCfg := stores.QPushConfig{
		Type:                pushCfg.Type,
		PushEndpoint:        pushCfg.Pend,
		MaxMessages:         pushCfg.MaxMessages,
		AuthorizationType:   pushCfg.AuthorizationHeader.Type,
		AuthorizationHeader: pushCfg.AuthorizationHeader.Value,
		RetPolicy:           pushCfg.RetPol.PolicyType,
		RetPeriod:           pushCfg.RetPol.Period,
		VerificationHash:    pushCfg.VerificationHash,
		Verified:            pushCfg.Verified,
		MattermostChannel:   pushCfg.MattermostChannel,
		MattermostUrl:       pushCfg.MattermostUrl,
		MattermostUsername:  pushCfg.MattermostUsername,
	}

	return store.ModSubPush(ctx, projectUUID, name, qPushCfg)
}

// RemoveSub removes an existing subscription
func RemoveSub(ctx context.Context, projectUUID string, name string, store stores.Store) error {

	if HasSub(ctx, projectUUID, name, store) == false {
		return errors.New("not found")
	}

	return store.RemoveSub(ctx, projectUUID, name)
}

// HasSub returns true if project & subscription combination exist
func HasSub(ctx context.Context, projectUUID string, name string, store stores.Store) bool {
	res, err := Find(ctx, projectUUID, "", name, "", 0, store)
	if len(res.Subscriptions) > 0 && err == nil {
		return true
	}

	return false
}

// ExtractFullTopicRef gets a full topic ref and extracts project and topic refs
func ExtractFullTopicRef(fTopicRef string) (string, string, error) {
	items := strings.Split(fTopicRef, "/")
	if len(items) != 4 {
		return "", "", errors.New("wrong topic name declaration")
	}

	if items[0] != "projects" && items[2] != "topics" {
		return "", "", errors.New("wrong topic name declaration")
	}

	return items[1], items[3], nil

}

// GetMaxAckID gets a list of ack ids and selects the maximum one
func GetMaxAckID(ackIDs []string) (string, error) {
	var max int64
	var maxAckID string

	for _, ackID := range ackIDs {

		offNum, err := GetOffsetFromAckID(ackID)

		if err != nil {
			return "", errors.New("invalid argument")
		}

		if offNum >= max {
			max = offNum
			maxAckID = ackID
		}

	}

	return maxAckID, nil

}

// GetOffsetFromAckID extracts an offset from an ackID
func GetOffsetFromAckID(ackID string) (int64, error) {

	var num int64
	tokens := strings.Split(ackID, "/")
	if len(tokens) != 4 {
		return num, errors.New("invalid argument")
	}
	subTokens := strings.Split(tokens[3], ":")
	num, err := strconv.ParseInt(subTokens[1], 10, 64)

	return num, err
}
