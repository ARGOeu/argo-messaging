package stores

// Store encapsulates the generic store interface
type Store interface {
	Initialize(server string, database string)
	QuerySubs() []QSubs
	QueryTopics() []QTopics
	UpdateSubOffset(name string, offset int64)
	Close()
}
