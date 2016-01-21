package topics

import (
	"encoding/json"

	"github.com/ARGOeu/argo-messaging/config"
)

// Topic struct to hold information for a given topic
type Topic struct {
	Project  string `json:"-"`
	Name     string `json:"-"`
	FullName string `json:"name"`
}

// Topics holds a list of Topic items
type Topics struct {
	List []Topic `json:"topics"`
}

// New creates a new topic based on name
func New(name string) Topic {
	pr := "ARGO" // Projects as entities will be handled later.
	ftn := "/projects/" + pr + "/topics/" + name
	t := Topic{Project: pr, Name: name, FullName: ftn}
	return t
}

// LoadFromCfg returns all topics defined in configuration
func (tl *Topics) LoadFromCfg(cfg *config.KafkaCfg) {
	for _, value := range cfg.Topics {
		nTopic := New(value)
		tl.List = append(tl.List, nTopic)
	}
}

// ExportJSON exports whole Topic Structure as a json string
func (tp *Topic) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(tp, "", "   ")
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
