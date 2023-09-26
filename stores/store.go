package stores

import (
	"context"
	"time"
)

// Store encapsulates the generic store interface
type Store interface {
	Initialize()
	QuerySubsByTopic(ctx context.Context, projectUUID, topic string) ([]QSub, error)
	QueryTopicsByACL(ctx context.Context, projectUUID, user string) ([]QTopic, error)
	QuerySubsByACL(ctx context.Context, projectUUID, user string) ([]QSub, error)
	QuerySubs(ctx context.Context, projectUUID string, userUUID string, name string, pageToken string, pageSize int64) ([]QSub, int64, string, error)
	QueryTopics(ctx context.Context, projectUUID string, userUUID string, name string, pageToken string, pageSize int64) ([]QTopic, int64, string, error)
	QueryDailyTopicMsgCount(ctx context.Context, projectUUID string, name string, date time.Time) ([]QDailyTopicMsgCount, error)
	UpdateTopicLatestPublish(ctx context.Context, projectUUID string, name string, date time.Time) error
	UpdateTopicPublishRate(ctx context.Context, projectUUID string, name string, rate float64) error
	UpdateSubLatestConsume(ctx context.Context, projectUUID string, name string, date time.Time) error
	UpdateSubConsumeRate(ctx context.Context, projectUUID string, name string, rate float64) error
	RemoveTopic(ctx context.Context, projectUUID string, name string) error
	RemoveSub(ctx context.Context, projectUUID string, name string) error
	PaginatedQueryUsers(ctx context.Context, pageToken string, pageSize int64, projectUUID string) ([]QUser, int64, string, error)
	QueryUsers(ctx context.Context, projectUUID string, uuid string, name string) ([]QUser, error)
	UpdateUser(ctx context.Context, uuid, fname, lname, org, desc string, projects []QProjectRoles, name string, email string, serviceRoles []string, modifiedOn time.Time) error
	AppendToUserProjects(ctx context.Context, userUUID string, projectUUID string, pRoles ...string) error
	UpdateUserToken(ctx context.Context, uuid string, token string) error
	RemoveUser(ctx context.Context, uuid string) error
	QueryProjects(ctx context.Context, uuid string, name string) ([]QProject, error)
	UpdateProject(ctx context.Context, projectUUID string, name string, description string, modifiedOn time.Time) error
	RemoveProject(ctx context.Context, uuid string) error
	RemoveProjectTopics(ctx context.Context, projectUUID string) error
	RemoveProjectSubs(ctx context.Context, projectUUID string) error
	RemoveProjectDailyMessageCounters(ctx context.Context, projectUUID string) error
	QueryDailyProjectMsgCount(ctx context.Context, projectUUID string) ([]QDailyProjectMsgCount, error)
	QueryTotalMessagesPerProject(ctx context.Context, projectUUIDs []string, startDate time.Time, endDate time.Time) ([]QProjectMessageCount, error)
	RegisterUser(ctx context.Context, uuid, name, firstName, lastName, email, org, desc, registeredAt, atkn, status string) error
	DeleteRegistration(ctx context.Context, uuid string) error
	QueryRegistrations(ctx context.Context, regUUID, status, activationToken, name, email, org string) ([]QUserRegistration, error)
	UpdateRegistration(ctx context.Context, regUUID, status, declineComment, modifiedBy, modifiedAt string) error
	InsertUser(ctx context.Context, uuid string, projects []QProjectRoles, name string, firstName string, lastName string, org string, desc string, token string, email string, serviceRoles []string, createdOn time.Time, modifiedOn time.Time, createdBy string) error
	InsertProject(ctx context.Context, uuid string, name string, createdOn time.Time, modifiedOn time.Time, createdBy string, description string) error
	InsertOpMetric(ctx context.Context, hostname string, cpu float64, mem float64) error
	InsertTopic(ctx context.Context, projectUUID string, name string, schemaUUID string, createdOn time.Time) error
	LinkTopicSchema(ctx context.Context, projectUUID, name, schemaUUID string) error
	IncrementTopicMsgNum(ctx context.Context, projectUUID string, name string, num int64) error
	IncrementDailyTopicMsgCount(ctx context.Context, projectUUID string, topicName string, num int64, date time.Time) error
	IncrementTopicBytes(ctx context.Context, projectUUID string, name string, totalBytes int64) error
	IncrementSubBytes(ctx context.Context, projectUUID string, name string, totalBytes int64) error
	IncrementSubMsgNum(ctx context.Context, projectUUID string, name string, num int64) error
	InsertSub(ctx context.Context, projectUUID string, name string, topic string, offest int64, ack int, pushCfg QPushConfig, createdOn time.Time) error
	HasProject(ctx context.Context, name string) bool
	HasUsers(ctx context.Context, projectUUID string, users []string) (bool, []string)
	QueryOneSub(ctx context.Context, projectUUID string, name string) (QSub, error)
	QueryPushSubs(ctx context.Context) []QSub
	HasResourceRoles(ctx context.Context, resource string, roles []string) bool
	GetOpMetrics(ctx context.Context) []QopMetric
	GetUserRoles(ctx context.Context, projectUUID string, token string) ([]string, string)
	GetUserFromToken(ctx context.Context, token string) (QUser, error)
	UpdateSubOffset(ctx context.Context, projectUUID string, name string, offset int64)
	UpdateSubPull(ctx context.Context, projectUUID string, name string, offset int64, ts string) error
	UpdateSubOffsetAck(ctx context.Context, projectUUID string, name string, offset int64, ts string) error
	ModSubPush(ctx context.Context, projectUUID string, name string, pushCfg QPushConfig) error
	QueryACL(ctx context.Context, projectUUID string, resource string, name string) (QAcl, error)
	ExistsInACL(ctx context.Context, projectUUID string, resource string, resourceName string, userUUID string) error
	ModACL(ctx context.Context, projectUUID string, resource string, name string, acl []string) error
	AppendToACL(ctx context.Context, projectUUID string, resource string, name string, acl []string) error
	RemoveFromACL(ctx context.Context, projectUUID string, resource string, name string, acl []string) error
	ModAck(ctx context.Context, projectUUID string, name string, ack int) error
	GetAllRoles(ctx context.Context) []string
	InsertSchema(ctx context.Context, projectUUID, schemaUUID, name, schemaType, rawSchemaString string) error
	QuerySchemas(ctx context.Context, projectUUID, schemaUUID, name string) ([]QSchema, error)
	UpdateSchema(ctx context.Context, schemaUUID, name, schemaType, rawSchemaString string) error
	DeleteSchema(ctx context.Context, schemaUUID string) error
	UsersCount(ctx context.Context, startDate, endDate time.Time, projectUUIDs []string) (map[string]int64, error)
	TopicsCount(ctx context.Context, startDate, endDate time.Time, projectUUIDs []string) (map[string]int64, error)
	SubscriptionsCount(ctx context.Context, startDate, endDate time.Time, projectUUIDs []string) (map[string]int64, error)
	Clone() Store
	Close()
}
