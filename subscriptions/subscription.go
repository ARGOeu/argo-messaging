package subscriptions

import (
	"encoding/json"

	"github.com/ARGOeu/argo-messaging/stores"
)

// Subscription struct to hold information for a given topic
type Subscription struct {
	Project   string     `json:"-"`
	Name      string     `json:"-"`
	Topic     string     `json:"-"`
	FullName  string     `json:"name"`
	FullTopic string     `json:"topic"`
	PushCfg   PushConfig `json:"pushConfig"`
	Ack       int        `json:"ackDeadlineSeconds"`
	Offset    int64      `json:"-"`
}

// PushConfig holds optional configuration for push operations
type PushConfig struct {
	Pend string `json:"pushEndpoint"`
}

// Subscriptions holds a list of Topic items
type Subscriptions struct {
	List []Subscription `json:"subscriptions"`
}

// SubPullOptions holds info about a pull operation on a subscription
type SubPullOptions struct {
	RetImm string `json:"returnImmediately,omitempty"`
	MaxMsg string `json:"maxMessages,omitempty"`
}

// GetPullOptionsJSON retrieves pull information
func GetPullOptionsJSON(input []byte) (SubPullOptions, error) {
	s := SubPullOptions{}
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
	defer store.Close()
	subs := store.QuerySubs()
	for _, item := range subs {
		curSub := New(item.Project, item.Name, item.Topic)
		curSub.Offset = item.Offset
		sl.List = append(sl.List, curSub)
	}

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
