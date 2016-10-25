package subscriptions

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/ARGOeu/argo-messaging/projects"
	"github.com/ARGOeu/argo-messaging/stores"
)

// Subscription struct to hold information for a given topic
type Subscription struct {
	ProjectUUID string     `json:"-"`
	Name        string     `json:"-"`
	Topic       string     `json:"-"`
	FullName    string     `json:"name"`
	FullTopic   string     `json:"topic"`
	PushCfg     PushConfig `json:"pushConfig"`
	Ack         int        `json:"ackDeadlineSeconds,omitempty"`
	Offset      int64      `json:"-"`
	NextOffset  int64      `json:"-"`
	PendingAck  string     `json:"-"`
}

// PushConfig holds optional configuration for push operations
type PushConfig struct {
	Pend   string      `json:"pushEndpoint"`
	RetPol RetryPolicy `json:"retryPolicy"`
}

// RetryPolicy holds information on retry policies
type RetryPolicy struct {
	PolicyType string `json:"type,omitempty"`
	Period     int    `json:"period,omitempty"`
}

// Subscriptions holds a list of Topic items
type Subscriptions struct {
	List []Subscription `json:"subscriptions,omitempty"`
}

// SubPullOptions holds info about a pull operation on a subscription
type SubPullOptions struct {
	RetImm string `json:"returnImmediately,omitempty"`
	MaxMsg string `json:"maxMessages,omitempty"`
}

// AckIDs utility struct
type AckIDs struct {
	IDs []string `json:"AckIds"`
}

// Empty returns true if Subscriptions list has no items
func (sl *Subscriptions) Empty() bool {
	return len(sl.List) <= 0
}

// GetAckFromJSON retrieves ack ids from json
func GetAckFromJSON(input []byte) (AckIDs, error) {
	s := AckIDs{}
	err := json.Unmarshal([]byte(input), &s)
	return s, err
}

// GetPullOptionsJSON retrieves pull information
func GetPullOptionsJSON(input []byte) (SubPullOptions, error) {
	s := SubPullOptions{}
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

// ExportJSON exports whole sub Structure as a json string
func (sub *Subscription) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(sub, "", "   ")
	return string(output[:]), err
}

// ExportJSON exports whole sub List Structure as a json string
func (sl *Subscriptions) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(sl, "", "   ")
	return string(output[:]), err
}

// Find searches the store for all subscriptions of a given project or a specific one
func Find(projectUUID string, name string, store stores.Store) (Subscriptions, error) {
	result := Subscriptions{}
	subs, err := store.QuerySubs(projectUUID, name)
	for _, item := range subs {
		projectName := projects.GetNameByUUID(item.ProjectUUID, store)
		if projectName == "" {
			return result, errors.New("invalid project")
		}
		curSub := New(item.ProjectUUID, projectName, item.Name, item.Topic)
		curSub.Offset = item.Offset
		curSub.NextOffset = item.NextOffset
		curSub.Ack = item.Ack
		if item.PushEndpoint != "" {
			rp := RetryPolicy{item.RetPolicy, item.RetPeriod}
			curSub.PushCfg = PushConfig{item.PushEndpoint, rp}
		}
		result.List = append(result.List, curSub)
	}
	return result, err
}

// LoadPushSubs returns all subscriptions defined in store that have a push configuration
func LoadPushSubs(store stores.Store) Subscriptions {
	result := Subscriptions{}
	result.List = []Subscription{}
	subs := store.QueryPushSubs()
	for _, item := range subs {
		projectName := projects.GetNameByUUID(item.ProjectUUID, store)
		curSub := New(item.ProjectUUID, projectName, item.Name, item.Topic)
		curSub.Offset = item.Offset
		curSub.NextOffset = item.NextOffset
		curSub.Ack = item.Ack
		rp := RetryPolicy{item.RetPolicy, item.RetPeriod}
		curSub.PushCfg = PushConfig{item.PushEndpoint, rp}

		result.List = append(result.List, curSub)
	}
	return result
}

// CreateSub creates a new subscription
func CreateSub(projectUUID string, name string, topic string, push string, offset int64, ack int, retPolicy string, retPeriod int, store stores.Store) (Subscription, error) {

	if HasSub(projectUUID, name, store) {
		return Subscription{}, errors.New("exists")
	}

	if ack == 0 {
		ack = 10
	}
	err := store.InsertSub(projectUUID, name, topic, offset, ack, push, retPolicy, retPeriod)
	if err != nil {
		return Subscription{}, errors.New("backend error")
	}

	results, err := Find(projectUUID, name, store)
	if len(results.List) != 1 {
		return Subscription{}, errors.New("backend error")
	}

	return results.List[0], err
}

// ModSubPush updates the subscription push config
func ModSubPush(projectUUID string, name string, push string, retPolicy string, retPeriod int, store stores.Store) error {

	if HasSub(projectUUID, name, store) == false {
		return errors.New("not found")
	}

	return store.ModSubPush(projectUUID, name, push, retPolicy, retPeriod)
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
	res, err := Find(projectUUID, name, store)
	if len(res.List) > 0 && err == nil {
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
