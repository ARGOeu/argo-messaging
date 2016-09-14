package stores

import "time"

// Store encapsulates the generic store interface
type Store interface {
	Initialize()
	QuerySubs(projectUUID string, name string) ([]QSub, error)
	QueryTopics(projectUUID string, name string) ([]QTopic, error)
	QueryProjects(name string, uuid string) ([]QProject, error)
	RemoveTopic(project string, name string) error
	RemoveSub(project string, name string) error
	InsertProject(uuid string, name string, createdOn time.Time, modifiedOn time.Time, createdBy string, description string) error
	InsertTopic(projectUUID string, name string) error
	InsertSub(projectUUID string, name string, topic string, offest int64, ack int, push string, rPolicy string, rPeriod int) error
	HasProject(project string) bool
	HasUsers(project string, users []string) (bool, []string)
	QueryOneSub(projectUUID string, name string) (QSub, error)
	QueryPushSubs() []QSub
	HasResourceRoles(resource string, roles []string) bool
	GetUserRoles(project string, token string) ([]string, string)
	UpdateSubOffset(projectUUID string, name string, offset int64)
	UpdateSubPull(name string, offset int64, ts string)
	UpdateSubOffsetAck(projectUUID string, name string, offset int64, ts string) error
	ModSubPush(project string, name string, push string, rPolicy string, rPeriod int) error
	QueryACL(project string, resource string, name string) (QAcl, error)
	ModACL(project string, resource string, name string, acl []string) error
	Clone() Store
	Close()
}
