package stores

// QSub are the results of the Qsub query
type QSub struct {
	Project    string `bson:"project"`
	Name       string `bson:"name"`
	Topic      string `bson:"topic"`
	Offset     int64  `bson:"offset"`
	NextOffset int64  `bson:"next_offset"`
	PendingAck string `bson:"pending_ack"`
	Ack        int    `bson:"ack"`
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
