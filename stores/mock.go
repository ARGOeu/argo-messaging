package stores

// MockStore holds configuration
type MockStore struct {
	Server    string
	Database  string
	SubList   []QSubs
	TopicList []QTopics
	Session   bool
}

// NewMockStore creates new mock store
func NewMockStore(server string, database string) *MockStore {
	mk := MockStore{}
	mk.Initialize(server, database)
	mk.Session = true
	return &mk
}

// Close is used to close session
func (mk *MockStore) Close() {
	mk.Session = false
}

// UpdateSubOffset updates the offset of the current subscription
func (mk *MockStore) UpdateSubOffset(name string, offset int64) {

}

// Initialize is used to initalize the mock
func (mk *MockStore) Initialize(server string, database string) {
	mk.Server = server
	mk.Database = database
	// populate topics
	qtop1 := QTopics{"ARGO", "topic1"}
	qtop2 := QTopics{"ARGO", "topic2"}
	qtop3 := QTopics{"ARGO", "topic3"}
	mk.TopicList = append(mk.TopicList, qtop1)
	mk.TopicList = append(mk.TopicList, qtop2)
	mk.TopicList = append(mk.TopicList, qtop3)

	// populate Subscriptions
	qsub1 := QSubs{"ARGO", "sub1", "topic1", 0}
	qsub2 := QSubs{"ARGO", "sub2", "topic2", 0}
	qsub3 := QSubs{"ARGO", "sub3", "topic3", 0}
	mk.SubList = append(mk.SubList, qsub1)
	mk.SubList = append(mk.SubList, qsub2)
	mk.SubList = append(mk.SubList, qsub3)
}

// QuerySubs Query Subscription info from store
func (mk *MockStore) QuerySubs() []QSubs {
	return mk.SubList
}

// QueryTopics Query Subscription info from store
func (mk *MockStore) QueryTopics() []QTopics {
	return mk.TopicList
}
