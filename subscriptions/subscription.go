package subscriptions

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"encoding/base64"
	"github.com/ARGOeu/argo-messaging/projects"
	"github.com/ARGOeu/argo-messaging/stores"
	log "github.com/sirupsen/logrus"
)

// Subscription struct to hold information for a given topic
type Subscription struct {
	ProjectUUID string     `json:"-"`
	Name        string     `json:"-"`
	Topic       string     `json:"-"`
	FullName    string     `json:"name"`
	FullTopic   string     `json:"topic"`
	PushCfg     PushConfig `json:"pushConfig"`
	Ack         int        `json:"ackDeadlineSeconds"`
	Offset      int64      `json:"-"`
	NextOffset  int64      `json:"-"`
	PendingAck  string     `json:"-"`
	PushStatus  string     `json:"push_status,omitempty"`
}

// PushConfig holds optional configuration for push operations
type PushConfig struct {
	Pend   string      `json:"pushEndpoint"`
	RetPol RetryPolicy `json:"retryPolicy"`
}

// SubMetrics holds the subscription's metric details
type SubMetrics struct {
	MsgNum     int64 `json:"number_of_messages"`
	TotalBytes int64 `json:"total_bytes"`
}

// RetryPolicy holds information on retry policies
type RetryPolicy struct {
	PolicyType string `json:"type,omitempty"`
	Period     int    `json:"period,omitempty"`
}

// PaginatedSubscriptions holds information about a subscriptions' page and how to access the next page
type PaginatedSubscriptions struct {
	Subscriptions []Subscription `json:"subscriptions"`
	NextPageToken string         `json:"nextPageToken"`
	TotalSize     int32          `json:"totalSize"`
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

// FindMetric returns the metric of a specific subscription
func FindMetric(projectUUID string, name string, store stores.Store) (SubMetrics, error) {
	result := SubMetrics{MsgNum: 0}
	subs, _, _, err := store.QuerySubs(projectUUID, name, "", 0)

	// check if sub exists
	if len(subs) == 0 {
		return result, errors.New("not found")
	}

	for _, item := range subs {
		projectName := projects.GetNameByUUID(item.ProjectUUID, store)
		if projectName == "" {
			return result, errors.New("invalid project")
		}

		result.MsgNum = item.MsgNum
		result.TotalBytes = item.TotalBytes
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

// GetAckDeadlineFromJson retrieves ack deadline from json input
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

// ExportJSON exports whole sub List Structure as a json string
func (sl *PaginatedSubscriptions) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(sl, "", "   ")
	return string(output[:]), err
}

// Find searches the store for all subscriptions of a given project or a specific one
func Find(projectUUID string, name string, pageToken string, pageSize int32, store stores.Store) (PaginatedSubscriptions, error) {

	var err error
	var qSubs []stores.QSub
	var totalSize int32
	var nextPageToken string
	var pageTokenBytes []byte

	result := PaginatedSubscriptions{Subscriptions: []Subscription{}}

	// decode the base64 pageToken
	if pageTokenBytes, err = base64.StdEncoding.DecodeString(pageToken); err != nil {
		log.Errorf("Page token %v produced an error while being decoded to base64: %v", pageToken, err.Error())
		return result, err
	}

	if qSubs, totalSize, nextPageToken, err = store.QuerySubs(projectUUID, name, string(pageTokenBytes), pageSize); err != nil {
		return result, err
	}

	projectName := projects.GetNameByUUID(projectUUID, store)

	if projectName == "" {
		return result, errors.New("invalid project")
	}

	for _, item := range qSubs {
		curSub := New(item.ProjectUUID, projectName, item.Name, item.Topic)
		curSub.Offset = item.Offset
		curSub.NextOffset = item.NextOffset
		curSub.Ack = item.Ack
		if item.PushEndpoint != "" {
			rp := RetryPolicy{item.RetPolicy, item.RetPeriod}
			curSub.PushCfg = PushConfig{item.PushEndpoint, rp}
			curSub.PushStatus = item.PushStatus
		}
		result.Subscriptions = append(result.Subscriptions, curSub)
	}

	result.NextPageToken = base64.StdEncoding.EncodeToString([]byte(nextPageToken))
	result.TotalSize = totalSize

	return result, err
}

// LoadPushSubs returns all subscriptions defined in store that have a push configuration
func LoadPushSubs(store stores.Store) PaginatedSubscriptions {
	result := PaginatedSubscriptions{Subscriptions: []Subscription{}}
	subs := store.QueryPushSubs()
	for _, item := range subs {
		projectName := projects.GetNameByUUID(item.ProjectUUID, store)
		curSub := New(item.ProjectUUID, projectName, item.Name, item.Topic)
		curSub.Offset = item.Offset
		curSub.NextOffset = item.NextOffset
		curSub.Ack = item.Ack
		rp := RetryPolicy{item.RetPolicy, item.RetPeriod}
		curSub.PushCfg = PushConfig{item.PushEndpoint, rp}

		result.Subscriptions = append(result.Subscriptions, curSub)
	}
	return result
}

// CreateSub creates a new subscription
func CreateSub(projectUUID string, name string, topic string, push string, offset int64, ack int, retPolicy string, retPeriod int, status string, store stores.Store) (Subscription, error) {

	if HasSub(projectUUID, name, store) {
		return Subscription{}, errors.New("exists")
	}

	if ack == 0 {
		ack = 10
	}
	err := store.InsertSub(projectUUID, name, topic, offset, ack, push, retPolicy, retPeriod, status)
	if err != nil {
		return Subscription{}, errors.New("backend error")
	}

	results, err := Find(projectUUID, name, "", 0, store)
	if len(results.Subscriptions) != 1 {
		return Subscription{}, errors.New("backend error")
	}

	return results.Subscriptions[0], err
}

// ModAck updates the subscription's acknowledgment timeout
func ModAck(projectUUID string, name string, ack int, store stores.Store) error {
	// minimum deadline allowed 0 seconds, maximum: 600 sec (10 minutes)
	if ack < 0 || ack > 600 {
		return errors.New("wrong value")
	}

	if HasSub(projectUUID, name, store) == false {
		return errors.New("not found")
	}

	return store.ModAck(projectUUID, name, ack)
}

// ModSubPush updates the subscription push config
func ModSubPush(projectUUID string, name string, push string, retPolicy string, retPeriod int, status string, store stores.Store) error {

	if HasSub(projectUUID, name, store) == false {
		return errors.New("not found")
	}

	return store.ModSubPush(projectUUID, name, push, retPolicy, retPeriod, status)
}

// RemoveSub removes an existing subscription
func RemoveSub(projectUUID string, name string, store stores.Store) error {

	if HasSub(projectUUID, name, store) == false {
		return errors.New("not found")
	}

	return store.RemoveSub(projectUUID, name)
}

// HasSub returns true if project & subscription combination exist
func HasSub(projectUUID string, name string, store stores.Store) bool {
	res, err := Find(projectUUID, name, "", 0, store)
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
