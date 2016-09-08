package stores

// QSub are the results of the Qsub query
type QSub struct {
	Project      string `bson:"project"`
	Name         string `bson:"name"`
	Topic        string `bson:"topic"`
	Offset       int64  `bson:"offset"`
	NextOffset   int64  `bson:"next_offset"`
	PendingAck   string `bson:"pending_ack"`
	PushEndpoint string `bson:"push_endpoint"`
	PushMaxMsg   int    `bson:"push_max_messages"`
	Ack          int    `bson:"ack"`
	RetPolicy    string `bson:"retry_policy"`
	RetPeriod    int    `bson:"retry_period"`
}

// QAcl holds a list of authorized users queried from topic or subscription collections
type QAcl struct {
	ACL []string `bson:"acl"`
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

// QProject are the results of the QProject query
type QProject struct {
	Name string `bson:"name"`
}

// QTopic are the results of the QTopic query
type QTopic struct {
	Project string `bson:"project"`
	Name    string `bson:"name"`
}
