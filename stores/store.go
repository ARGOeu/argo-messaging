package stores

// Store encapsulates the generic store interface
type Store interface {
	Initialize(server string, database string)
	QuerySubs() []QSub
	QueryTopics() []QTopic
	RemoveTopic(project string, name string) error
	RemoveSub(project string, name string) error
	InsertTopic(project string, name string) error
	InsertSub(project string, name string, topic string, offest int64) error
	HasProject(project string) bool
	HasResourceRoles(resource string, roles []string) bool
	GetUserRoles(project string, token string) []string
	UpdateSubOffset(name string, offset int64)
	Close()
}
