package topics

import (
	"encoding/json"
	"errors"

	"encoding/base64"
	"github.com/ARGOeu/argo-messaging/projects"
	"github.com/ARGOeu/argo-messaging/schemas"
	"github.com/ARGOeu/argo-messaging/stores"
	log "github.com/sirupsen/logrus"
	"time"
)

// Topic struct to hold information for a given topic
type Topic struct {
	ProjectUUID   string    `json:"-"`
	Name          string    `json:"-"`
	FullName      string    `json:"name"`
	LatestPublish time.Time `json:"-"`
	PublishRate   float64   `json:"-"`
	Schema        string    `json:"schema,omitempty"`
	CreatedOn     string    `json:"created_on"`
}

type TopicMetrics struct {
	MsgNum        int64     `json:"number_of_messages"`
	TotalBytes    int64     `json:"total_bytes"`
	LatestPublish time.Time `json:"-"`
	PublishRate   float64   `json:"-"`
}

// PaginatedTopics holds information about a topics' page and how to access the next page
type PaginatedTopics struct {
	Topics        []Topic `json:"topics"`
	NextPageToken string  `json:"nextPageToken"`
	TotalSize     int32   `json:"totalSize"`
}

// Empty returns true if Topics has no items
func (tl *PaginatedTopics) Empty() bool {
	return len(tl.Topics) <= 0
}

// New creates a new topic based on name
func New(projectUUID string, projectName string, name string) Topic {
	ftn := "/projects/" + projectName + "/topics/" + name
	t := Topic{
		ProjectUUID:   projectUUID,
		Name:          name,
		FullName:      ftn,
		LatestPublish: time.Time{},
		PublishRate:   0,
	}
	return t
}

// Find searches and returns a specific topic or all topics of a given project
func FindMetric(projectUUID string, name string, store stores.Store) (TopicMetrics, error) {
	result := TopicMetrics{MsgNum: 0}
	topics, _, _, err := store.QueryTopics(projectUUID, "", name, "", 0)

	// check if the topic exists
	if len(topics) == 0 {
		return result, errors.New("not found")
	}

	for _, item := range topics {
		projectName := projects.GetNameByUUID(item.ProjectUUID, store)
		if projectName == "" {
			return result, errors.New("invalid project")
		}

		result.MsgNum = item.MsgNum
		result.TotalBytes = item.TotalBytes
		result.PublishRate = item.PublishRate
		result.LatestPublish = item.LatestPublish
	}
	return result, err
}

// Find searches and returns a specific topic or all topics of a given project
func Find(projectUUID, userUUID, name, pageToken string, pageSize int32, store stores.Store) (PaginatedTopics, error) {

	var err error
	var qTopics []stores.QTopic
	var totalSize int32
	var nextPageToken string
	var pageTokenBytes []byte

	result := PaginatedTopics{Topics: []Topic{}}

	// decode the base64 pageToken
	if pageTokenBytes, err = base64.StdEncoding.DecodeString(pageToken); err != nil {
		log.WithFields(
			log.Fields{
				"type": "service_log",
			},
		).Errorf("Page token %v produced an error while being decoded to base64: %v", pageToken, err.Error())
		return result, err
	}

	if qTopics, totalSize, nextPageToken, err = store.QueryTopics(projectUUID, userUUID, name, string(pageTokenBytes), pageSize); err != nil {
		return result, err
	}

	projectName := projects.GetNameByUUID(projectUUID, store)

	if projectName == "" {
		return result, errors.New("invalid project")

	}

	for _, item := range qTopics {
		curTop := New(item.ProjectUUID, projectName, item.Name)
		curTop.LatestPublish = item.LatestPublish
		curTop.PublishRate = item.PublishRate
		curTop.CreatedOn = item.CreatedOn.Format("2006-01-02T15:04:05Z")

		if item.SchemaUUID != "" {
			sl, err := schemas.Find(projectUUID, item.SchemaUUID, "", store)
			if err == nil {
				if !sl.Empty() {
					curTop.Schema = schemas.FormatSchemaRef(projectName, sl.Schemas[0].Name)
				}
			} else {
				log.WithFields(
					log.Fields{
						"type":         "service_log",
						"topic_name":   item.Name,
						"project_uuid": projectUUID,
						"error":        err.Error(),
					},
				).Error("Could not retrieve schema")
			}
		}

		result.Topics = append(result.Topics, curTop)
	}

	result.NextPageToken = base64.StdEncoding.EncodeToString([]byte(nextPageToken))
	result.TotalSize = totalSize

	return result, err
}

// ExportJSON exports whole TopicMetrics Structure as a json string
func (tp *TopicMetrics) ExportJSON() (string, error) {

	output, err := json.MarshalIndent(tp, "", "   ")
	return string(output[:]), err
}

// ExportJSON exports whole Topic Structure as a json string
func (tp *Topic) ExportJSON() (string, error) {

	output, err := json.MarshalIndent(tp, "", "   ")
	return string(output[:]), err
}

// ExportJSON exports whole Topics List Structure as a json string
func (tl *PaginatedTopics) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(tl, "", "   ")
	return string(output[:]), err
}

// CreateTopic creates a new topic
func CreateTopic(projectUUID string, name string, schemaUUID string, createdOn time.Time, store stores.Store) (Topic, error) {

	if HasTopic(projectUUID, name, store) {
		return Topic{}, errors.New("exists")
	}

	err := store.InsertTopic(projectUUID, name, schemaUUID, createdOn)
	if err != nil {
		return Topic{}, errors.New("backend error")
	}

	results, err := Find(projectUUID, "", name, "", 0, store)

	if len(results.Topics) != 1 {
		return Topic{}, errors.New("backend error")
	}

	return results.Topics[0], err
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
	res, err := Find(projectUUID, "", name, "", 0, store)
	if len(res.Topics) > 0 && err == nil {
		return true
	}
	return false
}
