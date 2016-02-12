package subscriptions

import (
	"encoding/json"
	"sort"

	"github.com/ARGOeu/argo-messaging/config"
)

// Subscription struct to hold information for a given topic
type Subscription struct {
	Project   string     `json:"-"`
	Name      string     `json:"-"`
	FullName  string     `json:"name"`
	FullTopic string     `json:"topic"`
	PushCfg   PushConfig `json:"pushConfig"`
	Ack       int        `json:"ackDeadlineSeconds"`
}

// PushConfig holds optional configuration for push operations
type PushConfig struct {
	Pend string `json:"pushEndpoint"`
}

// Subscriptions holds a list of Topic items
type Subscriptions struct {
	List []Subscription `json:"subscriptions"`
}

// New creates a new subscription based on name
func New(name string, topic string) Subscription {
	pr := "ARGO" // Projects as entities will be handled later.
	fsn := "/projects/" + pr + "/subscriptions/" + name
	ftn := "/projects/" + pr + "/topics/" + topic
	ps := PushConfig{}
	s := Subscription{Project: pr, Name: name, FullName: fsn, FullTopic: ftn, PushCfg: ps, Ack: 10}
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

// LoadFromCfg returns all subscriptions defined in configuration
func (sl *Subscriptions) LoadFromCfg(cfg *config.KafkaCfg) {
	var keys []string
	for key := range cfg.Subs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		curSub := New(key, cfg.Subs[key])
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
