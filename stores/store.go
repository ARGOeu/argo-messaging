package stores

import "time"

// Store encapsulates the generic store interface
type Store interface {
	Initialize()
	QuerySubs(projectUUID string, name string) ([]QSub, error)
	QueryTopics(projectUUID string, name string) ([]QTopic, error)
	RemoveTopic(projectUUID string, name string) error
	RemoveSub(projectUUID string, name string) error
	QueryUsers(projectUUID string, uuid string, name string) ([]QUser, error)
	UpdateUser(uuid string, projects []QProjectRoles, name string, email string, serviceRoles []string) error
	UpdateUserToken(uuid string, token string) error
	RemoveUser(uuid string) error
	QueryProjects(uuid string, name string) ([]QProject, error)
	UpdateProject(projectUUID string, name string, description string, modifiedOn time.Time) error
	RemoveProject(uuid string) error
	RemoveProjectTopics(projectUUID string) error
	RemoveProjectSubs(projectUUID string) error
	InsertUser(uuid string, projects []QProjectRoles, name string, token string, email string, serviceRoles []string) error
	InsertProject(uuid string, name string, createdOn time.Time, modifiedOn time.Time, createdBy string, description string) error
	InsertTopic(projectUUID string, name string) error
	InsertSub(projectUUID string, name string, topic string, offest int64, ack int, push string, rPolicy string, rPeriod int) error
	HasProject(name string) bool
	HasUsers(projectUUID string, users []string) (bool, []string)
	QueryOneSub(projectUUID string, name string) (QSub, error)
	QueryPushSubs() []QSub
	HasResourceRoles(resource string, roles []string) bool
	GetUserRoles(projectUUID string, token string) ([]string, string)
	UpdateSubOffset(projectUUID string, name string, offset int64)
	UpdateSubPull(name string, offset int64, ts string)
	UpdateSubOffsetAck(projectUUID string, name string, offset int64, ts string) error
	ModSubPush(projectUUID string, name string, push string, rPolicy string, rPeriod int) error
	QueryACL(projectUUID string, resource string, name string) (QAcl, error)
	ModACL(projectUUID string, resource string, name string, acl []string) error
	GetAllRoles() []string
	Clone() Store
	Close()
}
