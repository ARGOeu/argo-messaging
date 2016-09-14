package stores

import "time"

// QSub are the results of the Qsub query
type QSub struct {
	ProjectUUID  string `bson:"project_uuid"`
	Name         string `bson:"name"`
	Topic        string `bson:"topic"`
	Offset       int64  `bson:"offset"`
	NextOffset   int64  `bson:"next_offset"`
	PendingAck   string `bson:"pending_ack"`
	PushEndpoint string `bson:"push_endpoint"`
	Ack          int    `bson:"ack"`
	RetPolicy    string `bson:"retry_policy"`
	RetPeriod    int    `bson:"retry_period"`
}

// QAcl holds a list of authorized users queried from topic or subscription collections
type QAcl struct {
	ACL []string `bson:"acl"`
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

// QUser are the results of the QUser query
type QUser struct {
	Name    string   `bson:"name"`
	Email   string   `bson:"email"`
	Project string   `bson:"project"`
	Token   string   `bson:"token"`
	Roles   []string `bson:"roles"`
}

// QRole holds roles resources relationships
type QRole struct {
	Name  string   `bson:"resource"`
	Roles []string `bson:"roles"`
}

// QTopic are the results of the QTopic query
type QTopic struct {
	ProjectUUID string `bson:"project_uuid"`
	Name        string `bson:"name"`
}
