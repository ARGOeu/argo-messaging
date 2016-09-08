package subscriptions

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/ARGOeu/argo-messaging/stores"
)

// Subscription struct to hold information for a given topic
type Subscription struct {
	Project    string     `json:"-"`
	Name       string     `json:"-"`
	Topic      string     `json:"-"`
	FullName   string     `json:"name"`
	FullTopic  string     `json:"topic"`
	PushCfg    PushConfig `json:"pushConfig,omitempty"`
	Ack        int        `json:"ackDeadlineSeconds,omitempty"`
	Offset     int64      `json:"-"`
	NextOffset int64      `json:"-"`
	PendingAck string     `json:"-"`
}

// PushConfig holds optional configuration for push operations
type PushConfig struct {
	Pend      string      `json:"pushEndpoint"`
	MaxMsgNum int         `json:"maxMessages,omitempty"`
	RetPol    RetryPolicy `json:"retryPolicy"`
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

// SubACL holds the authorized users for a topic
type SubACL struct {
	AuthUsers []string `json:"authorized_users"`
}

// GetSubACL returns an authorized list of users for the topic
func GetSubACL(project string, sub string, store stores.Store) (SubACL, error) {
	result := SubACL{}
	subACL, err := store.QueryACL(project, "subscription", sub)
	if err != nil {
		return result, err
	}
	for _, item := range subACL.ACL {
		result.AuthUsers = append(result.AuthUsers, item)
	}
	return result, nil
}

// GetACLFromJSON retrieves SubACL info from JSON
func GetACLFromJSON(input []byte) (SubACL, error) {
	s := SubACL{}
	err := json.Unmarshal([]byte(input), &s)
	return s, err
}

// ModACL is called to modify a sub's acl
func ModACL(project string, name string, acl []string, store stores.Store) error {
	return store.ModACL(project, "subscriptions", name, acl)
}

// ExportJSON export subscription acl body to json for use in http response
func (sAcl *SubACL) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(sAcl, "", "   ")
	return string(output[:]), err
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
func New(project string, name string, topic string) Subscription {
	fsn := "/projects/" + project + "/subscriptions/" + name
	ftn := "/projects/" + project + "/topics/" + topic
	ps := PushConfig{}
	s := Subscription{Project: project, Name: name, Topic: topic, FullName: fsn, FullTopic: ftn, PushCfg: ps, Ack: 10}
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

// LoadFromStore returns all subscriptions defined in store
func (sl *Subscriptions) LoadFromStore(store stores.Store) {

	sl.List = []Subscription{}
	subs := store.QuerySubs()
	for _, item := range subs {
		curSub := New(item.Project, item.Name, item.Topic)
		curSub.Offset = item.Offset
		curSub.NextOffset = item.NextOffset
		curSub.Ack = item.Ack
		if item.PushEndpoint != "" {
			rp := RetryPolicy{item.RetPolicy, item.RetPeriod}
			curSub.PushCfg = PushConfig{item.PushEndpoint, item.PushMaxMsg, rp}
		}

		sl.List = append(sl.List, curSub)
	}

}

// LoadPushSubs returns all subscriptions defined in store that have a push configuration
func (sl *Subscriptions) LoadPushSubs(store stores.Store) {

	sl.List = []Subscription{}
	subs := store.QueryPushSubs()
	for _, item := range subs {
		curSub := New(item.Project, item.Name, item.Topic)
		curSub.Offset = item.Offset
		curSub.NextOffset = item.NextOffset
		curSub.Ack = item.Ack
		rp := RetryPolicy{item.RetPolicy, item.RetPeriod}
		curSub.PushCfg = PushConfig{item.PushEndpoint, item.PushMaxMsg, rp}

		sl.List = append(sl.List, curSub)
	}

}

// LoadOne loads one subscription
func (sl *Subscriptions) LoadOne(project string, subname string, store stores.Store) error {

	sl.List = []Subscription{}
	sub, err := store.QueryOneSub(project, subname)
	if err != nil {
		return errors.New("not found")
	}
	curSub := New(sub.Project, sub.Name, sub.Topic)
	curSub.Offset = sub.Offset
	curSub.NextOffset = sub.NextOffset
	curSub.Ack = sub.Ack
	if sub.PushEndpoint != "" {
		rp := RetryPolicy{sub.RetPolicy, sub.RetPeriod}
		curSub.PushCfg = PushConfig{sub.PushEndpoint, sub.PushMaxMsg, rp}
	}

	sl.List = append(sl.List, curSub)
	return nil

}

// CreateSub creates a new subscription
func (sl *Subscriptions) CreateSub(project string, name string, topic string, push string, pushMaxMsg int, offset int64, ack int, retPolicy string, retPeriod int, store stores.Store) (Subscription, error) {

	if sl.HasSub(project, name) {
		return Subscription{}, errors.New("exists")
	}

	subNew := New(project, name, topic)
	subNew.Offset = offset
	if ack == 0 {
		ack = 10
	}
	err := store.InsertSub(project, name, topic, offset, ack, push, pushMaxMsg, retPolicy, retPeriod)

	return subNew, err
}

// ModSubPush updates the subscription push config
func (sl *Subscriptions) ModSubPush(project string, name string, push string, retPolicy string, retPeriod int, store stores.Store) error {

	if sl.HasSub(project, name) == false {
		return errors.New("not found")
	}

	return store.ModSubPush(project, name, push, retPolicy, retPeriod)
}

// RemoveSub removes an existing subscription
func (sl *Subscriptions) RemoveSub(project string, name string, store stores.Store) error {

	if sl.HasSub(project, name) == false {
		return errors.New("not found")
	}

	return store.RemoveSub(project, name)
}

// GetSubByName returns a specific topic
func (sl *Subscriptions) GetSubByName(project string, name string) Subscription {

	for _, value := range sl.List {
		if (value.Project == project) && (value.Name == name) {
			return value
		}
	}
	return Subscription{}
}

// GetSubsByProject returns a specific topic
func (sl *Subscriptions) GetSubsByProject(project string) Subscriptions {
	result := Subscriptions{}
	for _, value := range sl.List {
		if value.Project == project {
			result.List = append(result.List, value)
		}
	}

	return result
}

// HasSub returns true if project & subscription combination exist
func (sl *Subscriptions) HasSub(project string, name string) bool {
	res := sl.GetSubByName(project, name)
	if res.Name != "" {
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
