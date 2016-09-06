package topics

import (
	"encoding/json"
	"errors"

	"github.com/ARGOeu/argo-messaging/stores"
)

// Topic struct to hold information for a given topic
type Topic struct {
	Project  string `json:"-"`
	Name     string `json:"-"`
	FullName string `json:"name"`
}

// TopicACL holds the authorized users for a topic
type TopicACL struct {
	AuthUsers []string `json:"authorized_users"`
}

// Topics holds a list of Topic items
type Topics struct {
	List []Topic `json:"topics,omitempty"`
}

// New creates a new topic based on name
func New(project string, name string) Topic {
	pr := project // Projects as entities will be handled later.
	ftn := "/projects/" + pr + "/topics/" + name
	t := Topic{Project: pr, Name: name, FullName: ftn}
	return t
}

// // LoadFromCfg returns all topics defined in configuration
// func (tl *Topics) LoadFromCfg(cfg *config.APICfg) {
// 	for _, value := range cfg.Topics {
// 		nTopic := New(value)
// 		tl.List = append(tl.List, nTopic)
// 	}
// }

// LoadFromStore returns all subscriptions defined in store
func (tl *Topics) LoadFromStore(store stores.Store) {

	tl.List = []Topic{}
	topics := store.QueryTopics()
	for _, item := range topics {
		curTop := New(item.Project, item.Name)
		tl.List = append(tl.List, curTop)
	}

}

// ExportJSON exports whole Topic Structure as a json string
func (tp *Topic) ExportJSON() (string, error) {

	output, err := json.MarshalIndent(tp, "", "   ")
	return string(output[:]), err
}

// GetTopicACL returns an authorized list of users for the topic
func GetTopicACL(project string, topic string, store stores.Store) (TopicACL, error) {
	result := TopicACL{}
	topicACL, err := store.QueryACL(project, "topic", topic)
	if err != nil {
		return result, err
	}
	for _, item := range topicACL.ACL {
		result.AuthUsers = append(result.AuthUsers, item)
	}
	return result, err
}

// GetACLFromJSON retrieves TopicACL info from JSON
func GetACLFromJSON(input []byte) (TopicACL, error) {
	s := TopicACL{}
	err := json.Unmarshal([]byte(input), &s)
	if s.AuthUsers == nil {
		return s, errors.New("wrong argument")
	}
	return s, err
}

// ModACL is called to modify a topic's acl
func ModACL(project string, name string, acl []string, store stores.Store) error {

	return store.ModACL(project, "topics", name, acl)
}

// ExportJSON export topic acl body to json for use in http response
func (tAcl *TopicACL) ExportJSON() (string, error) {
	if tAcl.AuthUsers == nil {
		tAcl.AuthUsers = make([]string, 0)
	}
	output, err := json.MarshalIndent(tAcl, "", "   ")
	return string(output[:]), err
}

// ExportJSON exports whole Topics List Structure as a json string
func (tl *Topics) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(tl, "", "   ")
	return string(output[:]), err
}

// GetTopicByName returns a specific topic
func (tl *Topics) GetTopicByName(project string, name string) Topic {
	for _, value := range tl.List {
		if (value.Project == project) && (value.Name == name) {
			return value
		}
	}
	return Topic{}
}

// GetTopicsByProject returns a specific topic
func (tl *Topics) GetTopicsByProject(project string) Topics {
	result := Topics{}
	for _, value := range tl.List {
		if value.Project == project {
			result.List = append(result.List, value)
		}
	}

	return result
}

// CreateTopic creates a new topic
func (tl *Topics) CreateTopic(project string, name string, store stores.Store) (Topic, error) {
	if tl.HasTopic(project, name) {
		return Topic{}, errors.New("exists")
	}

	topicNew := New(project, name)
	err := store.InsertTopic(project, name)
	return topicNew, err
}

// RemoveTopic removes an existing topic
func (tl *Topics) RemoveTopic(project string, name string, store stores.Store) error {
	if tl.HasTopic(project, name) == false {
		return errors.New("not found")
	}

	return store.RemoveTopic(project, name)
}

// HasTopic returns true if project & topic combination exist
func (tl *Topics) HasTopic(project string, name string) bool {
	res := tl.GetTopicByName(project, name)
	if res.Name != "" {
		return true
	}

	return false
}
