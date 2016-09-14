package topics

import (
	"encoding/json"
	"errors"

	"github.com/ARGOeu/argo-messaging/projects"
	"github.com/ARGOeu/argo-messaging/stores"
)

// Topic struct to hold information for a given topic
type Topic struct {
	ProjectUUID string `json:"-"`
	Name        string `json:"-"`
	FullName    string `json:"name"`
}

// TopicACL holds the authorized users for a topic
type TopicACL struct {
	AuthUsers []string `json:"authorized_users"`
}

// Topics holds a list of Topic items
type Topics struct {
	List []Topic `json:"topics,omitempty"`
}

// Empty returns true if Topics has no items
func (tl *Topics) Empty() bool {
	return len(tl.List) <= 0
}

// New creates a new topic based on name
func New(projectUUID string, projectName string, name string) Topic {
	ftn := "/projects/" + projectName + "/topics/" + name
	t := Topic{ProjectUUID: projectUUID, Name: name, FullName: ftn}
	return t
}

// Find searches and returns a specific topic or all topics of a given project
func Find(projectUUID string, name string, store stores.Store) (Topics, error) {
	result := Topics{}
	topics, err := store.QueryTopics(projectUUID, name)
	for _, item := range topics {
		projectName := projects.GetNameByUUID(item.ProjectUUID, store)
		if projectName == "" {
			return result, errors.New("invalid project")
		}
		curTop := New(item.ProjectUUID, projectName, item.Name)
		result.List = append(result.List, curTop)
	}
	return result, err
}

// ExportJSON exports whole Topic Structure as a json string
func (tp *Topic) ExportJSON() (string, error) {

	output, err := json.MarshalIndent(tp, "", "   ")
	return string(output[:]), err
}

// GetTopicACL returns an authorized list of users for the topic
func GetTopicACL(projectUUID string, topic string, store stores.Store) (TopicACL, error) {
	result := TopicACL{}
	topicACL, err := store.QueryACL(projectUUID, "topic", topic)
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
func ModACL(projectUUID string, name string, acl []string, store stores.Store) error {

	return store.ModACL(projectUUID, "topics", name, acl)
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

// CreateTopic creates a new topic
func CreateTopic(projectUUID string, name string, store stores.Store) (Topic, error) {

	if HasTopic(projectUUID, name, store) {
		return Topic{}, errors.New("exists")
	}

	err := store.InsertTopic(projectUUID, name)
	if err != nil {
		return Topic{}, errors.New("backend error")
	}

	results, err := Find(projectUUID, name, store)

	if len(results.List) != 1 {
		return Topic{}, errors.New("backend error")
	}

	return results.List[0], err
}

// RemoveTopic removes an existing topic
func RemoveTopic(projectUUID string, name string, store stores.Store) error {
	if HasTopic(projectUUID, name, store) == false {
		return errors.New("not found")
	}

	return store.RemoveTopic(projectUUID, name)
}

// HasTopic returns true if project & topic combination exist
func HasTopic(projectUUID string, name string, store stores.Store) bool {
	res, err := Find(projectUUID, name, store)
	if len(res.List) > 0 && err == nil {
		return true
	}
	return false
}
