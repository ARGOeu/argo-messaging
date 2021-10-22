package stores

import (
	"time"
)

// QSub are the results of the Qsub query
type QSub struct {
	ID                  interface{} `bson:"_id,omitempty"`
	ProjectUUID         string      `bson:"project_uuid"`
	Name                string      `bson:"name"`
	Topic               string      `bson:"topic"`
	Offset              int64       `bson:"offset"`
	NextOffset          int64       `bson:"next_offset"`
	PendingAck          string      `bson:"pending_ack"`
	PushEndpoint        string      `bson:"push_endpoint"`
	MaxMessages         int64       `bson:"max_messages"`
	AuthorizationType   string      `bson:"authorization_type"`
	AuthorizationHeader string      `bson:"authorization_header"`
	Ack                 int         `bson:"ack"`
	RetPolicy           string      `bson:"retry_policy"`
	RetPeriod           int         `bson:"retry_period"`
	MsgNum              int64       `bson:"msg_num"`
	TotalBytes          int64       `bson:"total_bytes"`
	VerificationHash    string      `bson:"verification_hash"`
	Verified            bool        `bson:"verified"`
	LatestConsume       time.Time   `bson:"latest_consume"`
	ConsumeRate         float64     `bson:"consume_rate"`
	CreatedOn           time.Time   `bson:"created_on"`
	ACL                 []string    `bson:"acl"`
}

// QAcl holds a list of authorized users queried from topic or subscription collections
type QAcl struct {
	ACL []string `bson:"acl"`
}

// QopMetric are the results of the QopMetric query
type QopMetric struct {
	Hostname string  `bson:"hostname"`
	CPU      float64 `bson:"cpu"`
	MEM      float64 `bson:"mem"`
}

// QProject are the results of the QProject query
type QProject struct {
	UUID        string    `bson:"uuid"`
	Name        string    `bson:"name"`
	CreatedOn   time.Time `bson:"created_on"`
	ModifiedOn  time.Time `bson:"modified_on"`
	CreatedBy   string    `bson:"created_by"`
	Description string    `bson:"description"`
}

// QUserRegistration holds information about a UserRegister query
type QUserRegistration struct {
	UUID            string `bson:"uuid"`
	Name            string `bson:"name"`
	FirstName       string `bson:"first_name"`
	LastName        string `bson:"last_name"`
	Organization    string `bson:"organization"`
	Description     string `bson:"description"`
	Email           string `bson:"email"`
	ActivationToken string `bson:"activation_token"`
	Status          string `bson:"status"`
	RegisteredAt    string `bson:"registered_at"`
	ModifiedBy      string `bson:"modified_by"`
	ModifiedAt      string `bson:"modified_at"`
}

// QUser are the results of the QUser query
type QUser struct {
	ID           interface{}     `bson:"_id,omitempty"`
	UUID         string          `bson:"uuid"`
	Projects     []QProjectRoles `bson:"projects"`
	Name         string          `bson:"name"`
	FirstName    string          `bson:"first_name,omitempty"`
	LastName     string          `bson:"last_name,omitempty"`
	Organization string          `bson:"organization,omitempty"`
	Description  string          `bson:"description,omitempty"`
	Token        string          `bson:"token"`
	Email        string          `bson:"email"`
	ServiceRoles []string        `bson:"service_roles"`
	CreatedOn    time.Time       `bson:"created_on"`
	ModifiedOn   time.Time       `bson:"modified_on"`
	CreatedBy    string          `bson:"created_by"`
}

//QProjectRoles include information about projects and roles that user has
type QProjectRoles struct {
	ProjectUUID string   `bson:"project_uuid"`
	Roles       []string `bson:"roles"`
}

// QRole holds roles resources relationships
type QRole struct {
	Name  string   `bson:"resource"`
	Roles []string `bson:"roles"`
}

// QTopic are the results of the QTopic query
type QTopic struct {
	ID            interface{} `bson:"_id,omitempty"`
	ProjectUUID   string      `bson:"project_uuid"`
	Name          string      `bson:"name"`
	MsgNum        int64       `bson:"msg_num"`
	TotalBytes    int64       `bson:"total_bytes"`
	LatestPublish time.Time   `bson:"latest_publish"`
	PublishRate   float64     `bson:"publish_rate"`
	SchemaUUID    string      `bson:"schema_uuid"`
	CreatedOn     time.Time   `bson:"created_on"`
	ACL           []string    `bson:"acl"`
}

// QDailyTopicMsgCount holds information about the daily number of messages published to a topic
type QDailyTopicMsgCount struct {
	Date             time.Time `bson:"date"`
	ProjectUUID      string    `bson:"project_uuid"`
	TopicName        string    `bson:"topic_name"`
	NumberOfMessages int64     `bson:"msg_count"`
}

// QDailyProjectMsgCount holds information about the total amount of messages published to all of a project's topics daily
type QDailyProjectMsgCount struct {
	Date             time.Time `bson:"date"`
	NumberOfMessages int64     `bson:"msg_count"`
}

// QProjectMessageCount holds information about the total messages and average daily messages for a specific project
type QProjectMessageCount struct {
	ProjectUUID          string  `bson:"project_uuid"`
	NumberOfMessages     int64   `bson:"msg_count"`
	AverageDailyMessages float64 `bson:"avg_daily_msg"`
}

type QProjectResourceCount struct {
	ProjectUUID string `bson:"project_uuid"`
	Count       int64  `bson:"resource_count"`
}

// QSchema is the query model representing a schema
type QSchema struct {
	ProjectUUID string `bson:"project_uuid"`
	UUID        string `bson:"uuid"`
	Name        string `bson:"name"`
	Type        string `bson:"type"`
	RawSchema   string `bson:"raw_schema"`
}

func (qUsr *QUser) isInProject(projectUUID string) bool {
	for _, item := range qUsr.Projects {
		if item.ProjectUUID == projectUUID {
			return true
		}
	}

	return false
}

func (qUsr *QUser) getProjectRoles(projectUUID string) []string {

	result := []string{}
	for _, item := range qUsr.Projects {
		if item.ProjectUUID == projectUUID {
			result = item.Roles
		}
	}

	// if Service admin add this also to the roles regardless of the project

	if len(qUsr.ServiceRoles) > 0 {
		result = append(result, qUsr.ServiceRoles...)
	}

	return result
}
