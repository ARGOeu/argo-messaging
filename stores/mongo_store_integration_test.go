//go:build integration

package stores

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"gopkg.in/mgo.v2/bson"
	"testing"
	"time"
)

// mongodbContainer represents the mongodb container type used in the module
type mongodbContainer struct {
	testcontainers.Container
}

// startContainer creates an instance of the mongodb container type
func startContainer(ctx context.Context) (*mongodbContainer, error) {

	req := testcontainers.ContainerRequest{
		Name:         "mongodb-4.2-ams",
		Image:        "mongo:4.2",
		ExposedPorts: []string{"27017/tcp"},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &mongodbContainer{Container: container}, nil
}

type MongoStoreIntegrationTestSuite struct {
	suite.Suite
	store              Store
	ctx                context.Context
	UserRegistrations  []QUserRegistration
	SubList            []QSub
	TopicList          []QTopic
	DailyTopicMsgCount []QDailyTopicMsgCount
	ProjectList        []QProject
	UserList           []QUser
	RoleList           []QRole
	SchemaList         []QSchema
	Session            bool
	TopicsACL          map[string]QAcl
	SubsACL            map[string]QAcl
	OpMetrics          map[string]QopMetric
}

type QListReverser[T any] struct{}

func (suite *QListReverser[T]) reverse(s []T) []T {
	a := make([]T, len(s))
	copy(a, s)

	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}

	return a
}

func (suite *MongoStoreIntegrationTestSuite) assertSchemasEqual(expected []QSchema, actual []QSchema) {
	suite.True(len(actual) == len(expected))
	for idx, schema := range expected {
		suite.Equal(schema.UUID, actual[idx].UUID, schema.Name)
		suite.Equal(schema.Name, actual[idx].Name, schema.Name)
		suite.Equal(schema.ProjectUUID, actual[idx].ProjectUUID, schema.Name)
		suite.Equal(schema.Type, actual[idx].Type, schema.Name)
		suite.Equal(schema.RawSchema, actual[idx].RawSchema, schema.Name)
	}
}

func (suite *MongoStoreIntegrationTestSuite) assertUsersEqual(expected []QUser, actual []QUser) {
	suite.True(len(actual) == len(expected))
	for idx, user := range expected {
		suite.Equal(user.UUID, actual[idx].UUID, user.Name)
		suite.Equal(user.Name, actual[idx].Name, user.Name)
		suite.Equal(user.Projects, actual[idx].Projects, user.Name)
		suite.Equal(user.LastName, actual[idx].LastName, user.Name)
		suite.Equal(user.FirstName, actual[idx].FirstName, user.Name)
		suite.Equal(user.Organization, actual[idx].Organization, user.Name)
		suite.Equal(user.Description, actual[idx].Description, user.Name)
		suite.Equal(user.Email, actual[idx].Email, user.Name)
		suite.Equal(user.ServiceRoles, actual[idx].ServiceRoles, user.Name)
		suite.Equal(user.CreatedBy, actual[idx].CreatedBy, user.Name)
		suite.Equal(user.CreatedOn, actual[idx].CreatedOn, user.Name)
		suite.Equal(user.ModifiedOn, actual[idx].ModifiedOn, user.Name)
		suite.Equal(user.Token, actual[idx].Token, user.Name)
		suite.True((actual[idx].ID.(bson.ObjectId)).Valid(), user.Name)
	}
}

func (suite *MongoStoreIntegrationTestSuite) assertTopicsEqual(expected []QTopic, actual []QTopic) {
	suite.True(len(actual) == len(expected))
	for idx, topic := range expected {
		suite.Equal(topic.Name, actual[idx].Name, topic.Name)
		suite.Equal(topic.ProjectUUID, actual[idx].ProjectUUID, topic.Name)
		suite.Equal(topic.SchemaUUID, actual[idx].SchemaUUID, topic.Name)
		suite.Equal(topic.PublishRate, actual[idx].PublishRate, topic.Name)
		suite.Equal(topic.LatestPublish, actual[idx].LatestPublish, topic.Name)
		suite.Equal(topic.CreatedOn, actual[idx].CreatedOn, topic.Name)
		suite.Equal(topic.MsgNum, actual[idx].MsgNum, topic.Name)
		suite.Equal(topic.TotalBytes, actual[idx].TotalBytes, topic.Name)
		suite.Equal(topic.ACL, actual[idx].ACL, topic.Name)
		suite.True((actual[idx].ID.(bson.ObjectId)).Valid(), topic.Name)
	}
}

func (suite *MongoStoreIntegrationTestSuite) assertSubsEqual(expected []QSub, actual []QSub) {
	suite.True(len(actual) == len(expected))
	for idx, sub := range expected {
		suite.Equal(sub.Name, actual[idx].Name, sub.Name)
		suite.Equal(sub.ProjectUUID, actual[idx].ProjectUUID, sub.Name)
		suite.Equal(sub.MattermostChannel, actual[idx].MattermostChannel, sub.Name)
		suite.Equal(sub.MattermostUrl, actual[idx].MattermostUrl, sub.Name)
		suite.Equal(sub.MattermostUsername, actual[idx].MattermostUsername, sub.Name)
		suite.Equal(sub.ConsumeRate, actual[idx].ConsumeRate, sub.Name)
		suite.Equal(sub.LatestConsume, actual[idx].LatestConsume, sub.Name)
		suite.Equal(sub.CreatedOn, actual[idx].CreatedOn, sub.Name)
		suite.Equal(sub.MsgNum, actual[idx].MsgNum, sub.Name)
		suite.Equal(sub.TotalBytes, actual[idx].TotalBytes, sub.Name)
		suite.Equal(sub.ACL, actual[idx].ACL, sub.Name)
		suite.True((actual[idx].ID.(bson.ObjectId)).Valid(), sub.Name)
		suite.Equal(sub.PushType, actual[idx].PushType, sub.Name)
		suite.Equal(sub.PushEndpoint, actual[idx].PushEndpoint, sub.Name)
		suite.Equal(sub.AuthorizationType, actual[idx].AuthorizationType, sub.Name)
		suite.Equal(sub.RetPolicy, actual[idx].RetPolicy, sub.Name)
		suite.Equal(sub.RetPeriod, actual[idx].RetPeriod, sub.Name)
		suite.Equal(sub.Base64Decode, actual[idx].Base64Decode, sub.Name)
		suite.Equal(sub.AuthorizationHeader, actual[idx].AuthorizationHeader, sub.Name)
		suite.Equal(sub.MaxMessages, actual[idx].MaxMessages, sub.Name)
		suite.Equal(sub.VerificationHash, actual[idx].VerificationHash, sub.Name)
		suite.Equal(sub.Verified, actual[idx].Verified, sub.Name)
	}
}

func (suite *MongoStoreIntegrationTestSuite) initDB() {

	created := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.Local)
	modified := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.Local)

	// populate Projects
	qPr := QProject{
		UUID:        "argo_uuid",
		Name:        "ARGO",
		CreatedOn:   created,
		ModifiedOn:  modified,
		CreatedBy:   "uuid1",
		Description: "simple project",
	}
	qPr2 := QProject{
		UUID:        "argo_uuid2",
		Name:        "ARGO2",
		CreatedOn:   created,
		ModifiedOn:  modified,
		CreatedBy:   "uuid1",
		Description: "simple project",
	}
	suite.ProjectList = append(suite.ProjectList, qPr, qPr2)
	for _, project := range suite.ProjectList {
		err := suite.store.InsertProject(suite.ctx, project.UUID, project.Name,
			project.CreatedOn, project.ModifiedOn, project.CreatedBy,
			project.Description)
		if err != nil {
			panic("could not insert project")
		}
	}

	qTopicACL01 := QAcl{[]string{"uuid1", "uuid2"}}
	qTopicACL02 := QAcl{[]string{"uuid1", "uuid2", "uuid4"}}
	qTopicACL03 := QAcl{[]string{"uuid3"}}

	qSubACL01 := QAcl{[]string{"uuid1", "uuid2"}}
	qSubACL02 := QAcl{[]string{"uuid1", "uuid3"}}
	qSubACL03 := QAcl{[]string{"uuid4", "uuid2", "uuid1"}}
	qSubACL04 := QAcl{[]string{"uuid2", "uuid4", "uuid7"}}

	suite.TopicsACL = make(map[string]QAcl)
	suite.SubsACL = make(map[string]QAcl)
	suite.TopicsACL["topic1"] = qTopicACL01
	suite.TopicsACL["topic2"] = qTopicACL02
	suite.TopicsACL["topic3"] = qTopicACL03
	suite.SubsACL["sub1"] = qSubACL01
	suite.SubsACL["sub2"] = qSubACL02
	suite.SubsACL["sub3"] = qSubACL03
	suite.SubsACL["sub4"] = qSubACL04

	// populate topics
	qtop4 := QTopic{
		ProjectUUID:   "argo_uuid",
		Name:          "topic4",
		MsgNum:        0,
		TotalBytes:    0,
		LatestPublish: time.Date(1970, time.January, 1, 2, 0, 0, 0, time.Local),
		PublishRate:   0,
		SchemaUUID:    "",
		CreatedOn:     time.Date(2020, 11, 19, 0, 0, 0, 0, time.Local),
		ACL:           []string{},
	}
	qtop3 := QTopic{
		ProjectUUID:   "argo_uuid",
		Name:          "topic3",
		MsgNum:        0,
		TotalBytes:    0,
		LatestPublish: time.Date(2019, 5, 7, 0, 0, 0, 0, time.Local),
		PublishRate:   8.99,
		SchemaUUID:    "schema_uuid_3",
		CreatedOn:     time.Date(2020, 11, 20, 0, 0, 0, 0, time.Local),
		ACL:           qTopicACL03.ACL,
	}
	qtop2 := QTopic{
		ProjectUUID:   "argo_uuid",
		Name:          "topic2",
		MsgNum:        0,
		TotalBytes:    0,
		LatestPublish: time.Date(2019, 5, 8, 0, 0, 0, 0, time.Local),
		PublishRate:   5.45,
		SchemaUUID:    "schema_uuid_1",
		CreatedOn:     time.Date(2020, 11, 21, 0, 0, 0, 0, time.Local),
		ACL:           qTopicACL02.ACL,
	}
	qtop1 := QTopic{
		ProjectUUID:   "argo_uuid",
		Name:          "topic1",
		MsgNum:        0,
		TotalBytes:    0,
		LatestPublish: time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local),
		PublishRate:   10,
		SchemaUUID:    "",
		CreatedOn:     time.Date(2020, 11, 22, 0, 0, 0, 0, time.Local),
		ACL:           qTopicACL01.ACL,
	}
	suite.TopicList = append(suite.TopicList, qtop1)
	suite.TopicList = append(suite.TopicList, qtop2)
	suite.TopicList = append(suite.TopicList, qtop3)
	suite.TopicList = append(suite.TopicList, qtop4)
	for _, qTopic := range suite.TopicList {
		err := suite.store.InsertTopic(suite.ctx, qTopic.ProjectUUID, qTopic.Name, qTopic.SchemaUUID, qTopic.CreatedOn)
		if err != nil {
			panic("could not insert topics")
		}
		err = suite.store.ModACL(suite.ctx, qTopic.ProjectUUID, "topics", qTopic.Name, qTopic.ACL)
		if err != nil {
			panic("could not mod topics acl")
		}
		if qTopic.SchemaUUID != "" {
			err = suite.store.LinkTopicSchema(suite.ctx, qTopic.ProjectUUID, qTopic.Name, qTopic.SchemaUUID)
			if err != nil {
				panic("could not link topic schema")
			}
		}

		if qTopic.PublishRate != 0 {
			err = suite.store.UpdateTopicPublishRate(suite.ctx, qTopic.ProjectUUID, qTopic.Name, qTopic.PublishRate)
			if err != nil {
				panic("could not update topic publishing rate")
			}
		}

		if !qTopic.LatestPublish.IsZero() {
			err = suite.store.UpdateTopicLatestPublish(suite.ctx, qTopic.ProjectUUID, qTopic.Name, qTopic.LatestPublish)
			if err != nil {
				panic("could not update topic latest publish")
			}
		}
	}

	// populate Subscriptions
	qsub1 := QSub{
		ID:            0,
		ProjectUUID:   "argo_uuid",
		Name:          "sub1",
		Topic:         "topic1",
		Ack:           10,
		LatestConsume: time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local),
		ConsumeRate:   10,
		CreatedOn:     time.Date(2020, 11, 19, 0, 0, 0, 0, time.Local),
		ACL:           qSubACL01.ACL,
	}

	qsub2 := QSub{
		ID:            1,
		ProjectUUID:   "argo_uuid",
		Name:          "sub2",
		Topic:         "topic2",
		Ack:           10,
		LatestConsume: time.Date(2019, 5, 7, 0, 0, 0, 0, time.Local),
		ConsumeRate:   8.99,
		CreatedOn:     time.Date(2020, 11, 20, 0, 0, 0, 0, time.Local),
		ACL:           qSubACL02.ACL,
	}

	qsub3 := QSub{
		ID:            2,
		ProjectUUID:   "argo_uuid",
		Name:          "sub3",
		Topic:         "topic3",
		Ack:           10,
		LatestConsume: time.Date(2019, 5, 8, 0, 0, 0, 0, time.Local),
		ConsumeRate:   5.45,
		CreatedOn:     time.Date(2020, 11, 21, 0, 0, 0, 0, time.Local),
		ACL:           qSubACL03.ACL,
	}

	qsub4 := QSub{
		ID:                  3,
		ProjectUUID:         "argo_uuid",
		Name:                "sub4",
		Topic:               "topic4",
		PushType:            "http_endpoint",
		PushEndpoint:        "endpoint.foo",
		MaxMessages:         1,
		AuthorizationType:   "autogen",
		AuthorizationHeader: "auth-header-1",
		Ack:                 10,
		RetPolicy:           "linear",
		RetPeriod:           300,
		VerificationHash:    "push-id-1",
		Verified:            true,
		Base64Decode:        true,
		CreatedOn:           time.Date(2020, 11, 22, 0, 0, 0, 0, time.Local),
		ACL:                 qSubACL04.ACL,
	}

	suite.SubList = append(suite.SubList, qsub1)
	suite.SubList = append(suite.SubList, qsub2)
	suite.SubList = append(suite.SubList, qsub3)
	suite.SubList = append(suite.SubList, qsub4)
	for _, qSub := range suite.SubList {
		pushCfg := QPushConfig{
			Type:                qSub.PushType,
			PushEndpoint:        qSub.PushEndpoint,
			MaxMessages:         qSub.MaxMessages,
			AuthorizationType:   qSub.AuthorizationType,
			AuthorizationHeader: qSub.AuthorizationHeader,
			RetPolicy:           qSub.RetPolicy,
			RetPeriod:           qSub.RetPeriod,
			VerificationHash:    qSub.VerificationHash,
			Verified:            qSub.Verified,
			MattermostUrl:       qSub.MattermostUrl,
			MattermostUsername:  qSub.MattermostUsername,
			MattermostChannel:   qSub.MattermostChannel,
			Base64Decode:        qSub.Base64Decode,
		}
		err := suite.store.InsertSub(suite.ctx, qSub.ProjectUUID, qSub.Name, qSub.Topic,
			qSub.Offset, qSub.Ack, pushCfg, qSub.CreatedOn)
		if err != nil {
			panic("could not insert subs")
		}
		err = suite.store.ModACL(suite.ctx, qSub.ProjectUUID, "subscriptions", qSub.Name, qSub.ACL)
		if err != nil {
			panic("could not mod subs acl")
		}

		if qSub.ConsumeRate != 0 {
			err = suite.store.UpdateSubConsumeRate(suite.ctx, qSub.ProjectUUID, qSub.Name, qSub.ConsumeRate)
			if err != nil {
				panic("could not update sub consume rate")
			}
		}

		if !qSub.LatestConsume.IsZero() {
			err = suite.store.UpdateSubLatestConsume(suite.ctx, qSub.ProjectUUID, qSub.Name, qSub.LatestConsume)
			if err != nil {
				panic("could not update sub latest consume")
			}
		}

	}

	// populate schemas
	//{
	// 			"type": "object",
	//			 "properties": {
	// 			  "name":        { "type": "string" },
	//  			  "email":        { "type": "string" },
	// 			  "address":    { "type": "string" },
	//  			  "telephone": { "type": "string" }
	//	 },
	// 	"required": ["name", "email"]
	//}
	//}
	// the above schema base64 encoded
	s := "eyJwcm9wZXJ0aWVzIjp7ImFkZHJlc3MiOnsidHlwZSI6InN0cmluZyJ9LCJlbWFpbCI6eyJ0eXBlIjoic3RyaW5nIn0sIm5hbWUiOnsidHlwZSI6InN0cmluZyJ9LCJ0ZWxlcGhvbmUiOnsidHlwZSI6InN0cmluZyJ9fSwicmVxdWlyZWQiOlsibmFtZSIsImVtYWlsIl0sInR5cGUiOiJvYmplY3QifQ=="
	qSchema1 := QSchema{
		UUID:        "schema_uuid_1",
		ProjectUUID: "argo_uuid",
		Type:        "json",
		Name:        "schema-1",
		RawSchema:   s,
	}
	qSchema2 := QSchema{
		UUID:        "schema_uuid_2",
		ProjectUUID: "argo_uuid",
		Type:        "json",
		Name:        "schema-2",
		RawSchema:   s,
	}
	// {
	//		"namespace": "user.avro",
	//		"type": "record",
	//		"name": "User",
	//		"fields": [
	//		{"name": "username", "type":"string"},
	//		{"name": "phone", "type": "int"}
	//      ]
	// }
	avros := "eyJmaWVsZHMiOlt7Im5hbWUiOiJ1c2VybmFtZSIsInR5cGUiOiJzdHJpbmcifSx7Im5hbWUiOiJwaG9uZSIsInR5cGUiOiJpbnQifV0sIm5hbWUiOiJVc2VyIiwibmFtZXNwYWNlIjoidXNlci5hdnJvIiwidHlwZSI6InJlY29yZCJ9"
	qSchema3 := QSchema{
		UUID:        "schema_uuid_3",
		ProjectUUID: "argo_uuid",
		Type:        "avro",
		Name:        "schema-3",
		RawSchema:   avros,
	}
	suite.SchemaList = append(suite.SchemaList, qSchema1, qSchema2, qSchema3)
	for _, qSchema := range suite.SchemaList {
		err := suite.store.InsertSchema(suite.ctx, qSchema.ProjectUUID, qSchema.UUID, qSchema.Name, qSchema.Type, qSchema.RawSchema)
		if err != nil {
			panic("could not insert schema")
		}
	}

	// populate daily msg count for topics
	dc1 := QDailyTopicMsgCount{
		Date:             time.Date(2018, 10, 1, 0, 0, 0, 0, time.Local),
		ProjectUUID:      "argo_uuid",
		TopicName:        "topic1",
		NumberOfMessages: 40,
	}
	dc2 := QDailyTopicMsgCount{
		Date:             time.Date(2018, 10, 2, 0, 0, 0, 0, time.Local),
		ProjectUUID:      "argo_uuid",
		TopicName:        "topic1",
		NumberOfMessages: 30,
	}
	dc3 := QDailyTopicMsgCount{
		Date:             time.Date(2018, 10, 1, 0, 0, 0, 0, time.Local),
		ProjectUUID:      "argo_uuid",
		TopicName:        "topic2",
		NumberOfMessages: 70,
	}
	dc4 := QDailyTopicMsgCount{
		Date:        time.Date(2018, 10, 1, 0, 0, 0, 0, time.Local),
		ProjectUUID: "argo_uuid",
		TopicName:   "topic3",
	}

	suite.DailyTopicMsgCount = append(suite.DailyTopicMsgCount, dc2, dc1, dc3, dc4)
	for _, dc := range suite.DailyTopicMsgCount {
		err := suite.store.IncrementDailyTopicMsgCount(suite.ctx, dc.ProjectUUID, dc.TopicName, dc.NumberOfMessages, dc.Date)
		if err != nil {
			panic(" could not increment daily message count")
		}
	}

	// populate Users
	qRole := []QProjectRoles{{
		ProjectUUID: "argo_uuid",
		Roles:       []string{"consumer", "publisher"},
	}}
	qRoleB := []QProjectRoles{{
		ProjectUUID: "argo_uuid2",
		Roles:       []string{"consumer", "publisher"},
	}}
	qUsr := QUser{
		ID:           0,
		UUID:         "uuid0",
		Projects:     qRole,
		Name:         "Test",
		Token:        "S3CR3T",
		Email:        "Test@test.com",
		ServiceRoles: []string{},
		CreatedOn:    created,
		ModifiedOn:   modified,
	}

	suite.UserList = append(suite.UserList, qUsr)

	qRoleConsumerPub := []QProjectRoles{{
		"argo_uuid",
		[]string{"publisher", "consumer"},
	}}

	suite.UserList = append(suite.UserList, QUser{
		ID:           1,
		UUID:         "uuid1",
		Projects:     qRole,
		Name:         "UserA",
		FirstName:    "FirstA",
		LastName:     "LastA",
		Organization: "OrgA",
		Description:  "DescA",
		Token:        "S3CR3T1",
		Email:        "foo-email",
		ServiceRoles: []string{},
		CreatedOn:    created,
		ModifiedOn:   modified,
	})
	suite.UserList = append(suite.UserList, QUser{
		ID:           2,
		UUID:         "uuid2",
		Projects:     qRole,
		Name:         "UserB",
		Token:        "S3CR3T2",
		Email:        "foo-email",
		ServiceRoles: []string{},
		CreatedOn:    created,
		ModifiedOn:   modified,
		CreatedBy:    "uuid1",
	})
	suite.UserList = append(suite.UserList, QUser{
		ID:           3,
		UUID:         "uuid3",
		Projects:     qRoleConsumerPub,
		Name:         "UserX",
		Token:        "S3CR3T3",
		Email:        "foo-email",
		ServiceRoles: []string{},
		CreatedOn:    created,
		ModifiedOn:   modified,
		CreatedBy:    "uuid1",
	})
	suite.UserList = append(suite.UserList, QUser{
		ID:           4,
		UUID:         "uuid4",
		Projects:     qRoleConsumerPub,
		Name:         "UserZ",
		Token:        "S3CR3T4",
		Email:        "foo-email",
		ServiceRoles: []string{},
		CreatedOn:    created,
		ModifiedOn:   modified,
		CreatedBy:    "uuid1",
	})
	suite.UserList = append(suite.UserList, QUser{
		ID:           5,
		UUID:         "same_uuid",
		Projects:     qRoleConsumerPub,
		Name:         "UserSame1",
		Token:        "S3CR3T41",
		Email:        "foo-email",
		ServiceRoles: []string{},
		CreatedOn:    created,
		ModifiedOn:   modified,
		CreatedBy:    "uuid1",
	})
	suite.UserList = append(suite.UserList, QUser{
		ID:           6,
		UUID:         "same_uuid",
		Projects:     qRoleConsumerPub,
		Name:         "UserSame2",
		Token:        "S3CR3T42",
		Email:        "foo-email",
		ServiceRoles: []string{},
		CreatedOn:    created,
		ModifiedOn:   modified,
		CreatedBy:    "uuid1",
	})
	suite.UserList = append(suite.UserList, QUser{
		ID:           7,
		UUID:         "uuid7",
		Projects:     []QProjectRoles{},
		Name:         "push_worker_0",
		Token:        "push_token",
		Email:        "foo-email",
		ServiceRoles: []string{"push_worker"},
		CreatedOn:    created,
		ModifiedOn:   modified,
	})
	suite.UserList = append(suite.UserList, QUser{
		ID:           8,
		UUID:         "uuid8",
		Projects:     qRoleB,
		Name:         "UserZ",
		Token:        "S3CR3T1",
		Email:        "foo-email",
		ServiceRoles: []string{},
		CreatedOn:    created,
		ModifiedOn:   modified,
	})

	for _, qUser := range suite.UserList {
		err := suite.store.InsertUser(suite.ctx, qUser.UUID, qUser.Projects, qUser.Name, qUser.FirstName, qUser.LastName,
			qUser.Organization, qUser.Description, qUser.Token, qUser.Email, qUser.ServiceRoles,
			qUser.CreatedOn, qUser.ModifiedOn, qUser.CreatedBy)
		if err != nil {
			panic("could not insert user")
		}
	}

	qRole1 := QRole{"topics:list_all", []string{"admin", "reader", "publisher"}}
	qRole2 := QRole{"topics:publish", []string{"admin", "publisher"}}
	suite.RoleList = append(suite.RoleList, qRole1, qRole2)
	for _, role := range suite.RoleList {
		err := suite.store.InsertResourceRoles(suite.ctx, role.Name, role.Roles)
		if err != nil {
			panic("could not insert roles")
		}
	}

	// Populate user registrations
	ur1 := QUserRegistration{
		UUID:            "ur-uuid1",
		Name:            "urname",
		FirstName:       "urfname",
		LastName:        "urlname",
		Organization:    "urorg",
		Description:     "urdesc",
		Email:           "uremail",
		ActivationToken: "uratkn-1",
		Status:          "pending",
		RegisteredAt:    "2019-05-12T22:26:58Z",
		ModifiedBy:      "uuid1",
		ModifiedAt:      "2020-05-15T22:26:58Z",
	}

	suite.UserRegistrations = append(suite.UserRegistrations, ur1)
	for _, ur := range suite.UserRegistrations {
		_ = suite.store.RegisterUser(suite.ctx, ur.UUID, ur.Name, ur.FirstName, ur.LastName, ur.Email,
			ur.Organization, ur.Description, ur.RegisteredAt, ur.ActivationToken, ur.Status)
	}

}

func (suite *MongoStoreIntegrationTestSuite) TestQueryTopics() {

	qTopicListReverser := QListReverser[QTopic]{}

	// retrieve all topics
	tpList, ts1, pg1, _ := suite.store.QueryTopics(suite.ctx, "argo_uuid", "", "", "", 0)
	suite.assertTopicsEqual(qTopicListReverser.reverse(suite.TopicList), tpList)
	suite.Equal(int64(4), ts1)
	suite.Equal("", pg1)

	// retrieve first 2
	eTopList1st2 := []QTopic{suite.TopicList[3], suite.TopicList[2]}
	tpList2, ts2, pg2, _ := suite.store.QueryTopics(suite.ctx, "argo_uuid", "", "", "", 2)
	suite.assertTopicsEqual(eTopList1st2, tpList2)
	suite.Equal(int64(4), ts2)
	suite.True(bson.IsObjectIdHex(pg2))

	// retrieve the next one
	eTopList3 := []QTopic{suite.TopicList[1]}
	tpList3, ts3, pg3, _ := suite.store.QueryTopics(suite.ctx, "argo_uuid", "", "", pg2, 1)
	suite.assertTopicsEqual(eTopList3, tpList3)
	suite.Equal(int64(4), ts3)
	suite.True(bson.IsObjectIdHex(pg3))

	// retrieve a single topic
	eTopList4 := []QTopic{suite.TopicList[0]}
	tpList4, ts4, pg4, _ := suite.store.QueryTopics(suite.ctx, "argo_uuid", "", "topic1", "", 0)
	suite.assertTopicsEqual(eTopList4, tpList4)
	suite.Equal(int64(0), ts4)
	suite.Equal("", pg4)

	// retrieve user's topics
	eTopList5 := []QTopic{suite.TopicList[1], suite.TopicList[0]}
	tpList5, ts5, pg5, _ := suite.store.QueryTopics(suite.ctx, "argo_uuid", "uuid1", "", "", 0)
	suite.assertTopicsEqual(eTopList5, tpList5)
	suite.Equal(int64(2), ts5)
	suite.Equal("", pg5)

	// retrieve use's topic with pagination
	eTopList6 := []QTopic{suite.TopicList[1]}
	tpList6, ts6, pg6, _ := suite.store.QueryTopics(suite.ctx, "argo_uuid", "uuid1", "", "", 1)
	suite.assertTopicsEqual(eTopList6, tpList6)
	suite.Equal(int64(2), ts6)
	suite.True(bson.IsObjectIdHex(pg6))
}

func (suite *MongoStoreIntegrationTestSuite) TestQuerySubs() {

	qSubcListReverser := QListReverser[QSub]{}

	subList, ts1, pg1, err1 := suite.store.QuerySubs(suite.ctx, "argo_uuid", "", "", "", 0)
	suite.assertSubsEqual(qSubcListReverser.reverse(suite.SubList), subList)
	suite.Equal(int64(4), ts1)
	suite.Equal("", pg1)

	// retrieve first 2 subs
	eSubListFirstPage := []QSub{
		suite.SubList[3],
		suite.SubList[2],
	}
	subList2, ts2, pg2, err2 := suite.store.QuerySubs(suite.ctx, "argo_uuid", "", "", "", 2)
	suite.assertSubsEqual(eSubListFirstPage, subList2)
	suite.Equal(int64(4), ts2)
	suite.True(bson.IsObjectIdHex(pg2))

	// retrieve next 2 subs
	eSubListNextPage := []QSub{
		suite.SubList[1],
		suite.SubList[0],
	}

	subList3, ts3, pg3, err3 := suite.store.QuerySubs(suite.ctx, "argo_uuid", "", "", pg2, 2)
	suite.assertSubsEqual(eSubListNextPage, subList3)
	suite.Equal(int64(4), ts3)
	suite.Equal("", pg3)

	// retrieve user's subs
	eSubList4 := []QSub{
		suite.SubList[2],
		suite.SubList[1],
		suite.SubList[0],
	}

	subList4, ts4, pg4, err4 := suite.store.QuerySubs(suite.ctx, "argo_uuid", "uuid1", "", "", 0)

	suite.Equal(int64(3), ts4)
	suite.Equal("", pg4)
	suite.assertSubsEqual(eSubList4, subList4)

	// retrieve user's subs
	eSubList5 := []QSub{
		suite.SubList[2],
		suite.SubList[1],
	}
	subList5, ts5, pg5, err5 := suite.store.QuerySubs(suite.ctx, "argo_uuid", "uuid1", "", "", 2)

	suite.Equal(int64(3), ts5)
	suite.True(bson.IsObjectIdHex(pg5))
	suite.assertSubsEqual(eSubList5, subList5)

	suite.Nil(err1)
	suite.Nil(err2)
	suite.Nil(err3)
	suite.Nil(err4)
	suite.Nil(err5)

	// test retrieve subs by topic
	subListByTopic, errSublistByTopic := suite.store.QuerySubsByTopic(suite.ctx, "argo_uuid", "topic1")
	suite.assertSubsEqual([]QSub{suite.SubList[0]}, subListByTopic)
	suite.Nil(errSublistByTopic)

	sb, err := suite.store.QueryOneSub(suite.ctx, "argo_uuid", "sub1")
	suite.assertSubsEqual([]QSub{suite.SubList[0]}, []QSub{sb})
	suite.Nil(err)
}

func (suite *MongoStoreIntegrationTestSuite) TestDailyTopicMsgCount() {

	// check query all
	qdsAll, _ := suite.store.QueryDailyTopicMsgCount(suite.ctx, "", "", time.Time{})
	suite.Equal(suite.DailyTopicMsgCount, qdsAll)

	// test daily count
	_ = suite.store.IncrementDailyTopicMsgCount(suite.ctx, "argo_uuid", "topic1", 40, time.Date(2018, 10, 1, 0, 0, 0, 0, time.Local))
	qds, _ := suite.store.QueryDailyTopicMsgCount(suite.ctx, "argo_uuid", "topic1", time.Date(2018, 10, 1, 0, 0, 0, 0, time.Local))
	suite.Equal(int64(80), qds[0].NumberOfMessages)

	// check if it was inserted since it wasn't present
	_ = suite.store.IncrementDailyTopicMsgCount(suite.ctx, "argo_uuid", "some_other_topic", 70, time.Date(2018, 10, 1, 0, 0, 0, 0, time.Local))
	qds2, _ := suite.store.QueryDailyTopicMsgCount(suite.ctx, "argo_uuid", "some_other_topic", time.Date(2018, 10, 1, 0, 0, 0, 0, time.Local))
	suite.Equal(int64(70), qds2[0].NumberOfMessages)
}

func (suite *MongoStoreIntegrationTestSuite) TestHasResourceRoles() {
	suite.True(suite.store.HasResourceRoles(suite.ctx, "topics:list_all", []string{"admin"}))
	suite.True(suite.store.HasResourceRoles(suite.ctx, "topics:list_all", []string{"admin", "reader"}))
	suite.True(suite.store.HasResourceRoles(suite.ctx, "topics:list_all", []string{"admin", "foo"}))
	suite.False(suite.store.HasResourceRoles(suite.ctx, "topics:list_all", []string{"foo"}))
	suite.False(suite.store.HasResourceRoles(suite.ctx, "topics:publish", []string{"reader"}))
	suite.True(suite.store.HasResourceRoles(suite.ctx, "topics:list_all", []string{"admin"}))
	suite.True(suite.store.HasResourceRoles(suite.ctx, "topics:list_all", []string{"publisher"}))
	suite.True(suite.store.HasResourceRoles(suite.ctx, "topics:publish", []string{"publisher"}))

}

func (suite *MongoStoreIntegrationTestSuite) TestHasProject() {
	suite.True(suite.store.HasProject(suite.ctx, "ARGO"))
	suite.False(suite.store.HasProject(suite.ctx, "FOO"))
}

func (suite *MongoStoreIntegrationTestSuite) TestGetUserRoles() {
	roles01, _ := suite.store.GetUserRoles(suite.ctx, "argo_uuid", "S3CR3T")
	roles02, _ := suite.store.GetUserRoles(suite.ctx, "argo_uuid", "SecretKey")
	suite.Equal([]string{"consumer", "publisher"}, roles01)
	suite.Equal([]string{}, roles02)
}

func (suite *MongoStoreIntegrationTestSuite) TestRemoveSub() {
	_ = suite.store.InsertSub(suite.ctx, "argo_uuid", "subFresh", "topicFresh", 0, 10, QPushConfig{}, time.Date(2020, 12, 19, 0, 0, 0, 0, time.Local))
	err := suite.store.RemoveSub(suite.ctx, "argo_uuid", "subFresh")
	suite.Equal(nil, err)
	subList, _, _, _ := suite.store.QuerySubs(suite.ctx, "argo_uuid", "", "", "", 0)
	suite.Equal(4, len(subList))
	err = suite.store.RemoveSub(suite.ctx, "argo_uuid", "subFresh")
	suite.Equal("not found", err.Error())
}

func (suite *MongoStoreIntegrationTestSuite) TestRemoveTopic() {
	_ = suite.store.InsertTopic(suite.ctx, "argo_uuid", "topicFresh", "", time.Date(2020, 9, 11, 0, 0, 0, 0, time.Local))
	err := suite.store.RemoveTopic(suite.ctx, "argo_uuid", "topicFresh")
	suite.Equal(nil, err)
	qTopicListReverser := QListReverser[QTopic]{}
	tpList, _, _, _ := suite.store.QueryTopics(suite.ctx, "argo_uuid", "", "", "", 0)
	suite.assertTopicsEqual(qTopicListReverser.reverse(suite.TopicList), tpList)
	err = suite.store.RemoveTopic(suite.ctx, "argo_uuid", "topicFresh")
	suite.Equal("not found", err.Error())
}

func (suite *MongoStoreIntegrationTestSuite) TestModAck() {
	_ = suite.store.ModAck(suite.ctx, "argo_uuid", "sub1", 66)
	subAck, _ := suite.store.QueryOneSub(suite.ctx, "argo_uuid", "sub1")
	suite.Equal(66, subAck.Ack)
}

func (suite *MongoStoreIntegrationTestSuite) TestModPushSub() {
	_ = suite.store.InsertSub(suite.ctx, "argo_uuid", "subFresh", "topicFresh", 0, 10, QPushConfig{}, time.Date(2020, 12, 19, 0, 0, 0, 0, time.Local))
	qCfg := QPushConfig{
		Type:                "http_endpoint",
		PushEndpoint:        "example.com",
		MaxMessages:         3,
		AuthorizationType:   "autogen",
		AuthorizationHeader: "auth-h-1",
		RetPolicy:           "linear",
		RetPeriod:           400,
		VerificationHash:    "hash-1",
		Verified:            true,
		MattermostUrl:       "m-url",
		MattermostUsername:  "m-u",
		MattermostChannel:   "m-c",
		Base64Decode:        true,
	}
	e1 := suite.store.ModSubPush(suite.ctx, "argo_uuid", "subFresh", qCfg)
	sub1, _ := suite.store.QueryOneSub(suite.ctx, "argo_uuid", "subFresh")
	suite.Nil(e1)
	suite.Equal("example.com", sub1.PushEndpoint)
	suite.Equal(int64(3), sub1.MaxMessages)
	suite.Equal("linear", sub1.RetPolicy)
	suite.Equal(400, sub1.RetPeriod)
	suite.Equal("hash-1", sub1.VerificationHash)
	suite.Equal("autogen", sub1.AuthorizationType)
	suite.Equal("auth-h-1", sub1.AuthorizationHeader)
	suite.Equal("m-url", sub1.MattermostUrl)
	suite.Equal("m-c", sub1.MattermostChannel)
	suite.Equal("m-u", sub1.MattermostUsername)
	suite.True(sub1.Verified)
	suite.True(sub1.Base64Decode)

	e2 := suite.store.ModSubPush(suite.ctx, "argo_uuid", "unknown", QPushConfig{})
	suite.Equal("not found", e2.Error())

	_ = suite.store.RemoveSub(suite.ctx, "argo_uuid", "subFresh")
}

func (suite *MongoStoreIntegrationTestSuite) TestExistsInACL() {
	existsE1 := suite.store.ExistsInACL(suite.ctx, "argo_uuid", "topics", "topic1", "uuid1")
	suite.Nil(existsE1)

	existsE2 := suite.store.ExistsInACL(suite.ctx, "argo_uuid", "topics", "topic1", "unknown")
	suite.Equal("not found", existsE2.Error())
}

func (suite *MongoStoreIntegrationTestSuite) TestQueryACL() {
	ExpectedACL01 := QAcl{[]string{"uuid1", "uuid2"}}
	QAcl01, _ := suite.store.QueryACL(suite.ctx, "argo_uuid", "topics", "topic1")
	suite.Equal(ExpectedACL01, QAcl01)

	ExpectedACL02 := QAcl{[]string{"uuid1", "uuid2", "uuid4"}}
	QAcl02, _ := suite.store.QueryACL(suite.ctx, "argo_uuid", "topics", "topic2")
	suite.Equal(ExpectedACL02, QAcl02)

	ExpectedACL03 := QAcl{[]string{"uuid3"}}
	QAcl03, _ := suite.store.QueryACL(suite.ctx, "argo_uuid", "topics", "topic3")
	suite.Equal(ExpectedACL03, QAcl03)

	ExpectedACL04 := QAcl{[]string{"uuid1", "uuid2"}}
	QAcl04, _ := suite.store.QueryACL(suite.ctx, "argo_uuid", "subscriptions", "sub1")
	suite.Equal(ExpectedACL04, QAcl04)

	ExpectedACL05 := QAcl{[]string{"uuid1", "uuid3"}}
	QAcl05, _ := suite.store.QueryACL(suite.ctx, "argo_uuid", "subscriptions", "sub2")
	suite.Equal(ExpectedACL05, QAcl05)

	ExpectedACL06 := QAcl{[]string{"uuid4", "uuid2", "uuid1"}}
	QAcl06, _ := suite.store.QueryACL(suite.ctx, "argo_uuid", "subscriptions", "sub3")
	suite.Equal(ExpectedACL06, QAcl06)

	ExpectedACL07 := QAcl{[]string{"uuid2", "uuid4", "uuid7"}}
	QAcl07, _ := suite.store.QueryACL(suite.ctx, "argo_uuid", "subscriptions", "sub4")
	suite.Equal(ExpectedACL07, QAcl07)

	QAcl08, err08 := suite.store.QueryACL(suite.ctx, "argo_uuid", "subscr", "sub4ss")
	suite.Equal(QAcl{}, QAcl08)
	suite.Equal(errors.New("wrong resource type"), err08)

	QAcl09, err09 := suite.store.QueryACL(suite.ctx, "argo_uuid", "subscriptions", "sub4ss")
	suite.Equal(QAcl{}, QAcl09)
	suite.Equal(errors.New("not found"), err09)
}

func (suite *MongoStoreIntegrationTestSuite) TestUpdateUser() {
	created := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.Local)
	modified := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.Local)
	qRoles := []QProjectRoles{QProjectRoles{"argo_uuid", []string{"admin"}}, QProjectRoles{"argo_uuid2", []string{"admin", "viewer"}}}
	_ = suite.store.InsertUser(suite.ctx, "user_uuid11", qRoles, "newUser2", "", "", "", "", "BX312Z34NLQ", "fake@email.com", []string{}, created, modified, "uuid1")
	usrUpdated := QUser{UUID: "user_uuid11", Projects: qRoles, Name: "updated_name", Token: "BX312Z34NLQ", Email: "fake@email.com", ServiceRoles: []string{"service_admin"}, CreatedOn: created, ModifiedOn: modified, CreatedBy: "uuid1"}
	_ = suite.store.UpdateUser(suite.ctx, "user_uuid11", "", "", "", "", nil, "updated_name", "", []string{"service_admin"}, modified)
	usr11, _ := suite.store.QueryUsers(suite.ctx, "", "user_uuid11", "")
	suite.assertUsersEqual([]QUser{usrUpdated}, usr11)
	// test append project to user
	errUserPrj := suite.store.AppendToUserProjects(suite.ctx, "user_uuid11", "p1_uuid", "r1", "r2")
	usr, _ := suite.store.QueryUsers(suite.ctx, "", "user_uuid11", "")
	suite.Equal([]QProjectRoles{
		{
			ProjectUUID: "argo_uuid",
			Roles:       []string{"admin"},
		},
		{
			ProjectUUID: "argo_uuid2",
			Roles:       []string{"admin", "viewer"},
		},
		{
			ProjectUUID: "p1_uuid",
			Roles:       []string{"r1", "r2"},
		},
	}, usr[0].Projects)
	suite.Nil(errUserPrj)

	// Test Remove User
	_ = suite.store.RemoveUser(suite.ctx, "user_uuid11")
	qu, _ := suite.store.QueryUsers(suite.ctx, "", "user_uuid11", "")
	suite.Empty(qu)
}

func (suite *MongoStoreIntegrationTestSuite) TestGetUserFromToken() {
	usrGet, _ := suite.store.GetUserFromToken(suite.ctx, "S3CR3T")
	suite.assertUsersEqual([]QUser{suite.UserList[0]}, []QUser{usrGet})
}

func (suite *MongoStoreIntegrationTestSuite) TestPaginatedQueryUsers() {

	reverser := QListReverser[QUser]{}

	// return all users in one page
	qUsers1, ts1, pg1, _ := suite.store.PaginatedQueryUsers(suite.ctx, "", 0, "")

	// return a page with the first 2
	qUsers2, ts2, pg2, _ := suite.store.PaginatedQueryUsers(suite.ctx, "", 2, "")

	// empty store
	qUsers3, ts3, pg3, _ := suite.store.PaginatedQueryUsers(suite.ctx, "", 0, "unkknown")

	// use page token to grab another 2 results
	qUsers4, ts4, pg4, _ := suite.store.PaginatedQueryUsers(suite.ctx, pg2, 2, "")

	suite.assertUsersEqual(reverser.reverse(suite.UserList), qUsers1)
	suite.Equal("", pg1)
	suite.Equal(int64(9), ts1)

	suite.assertUsersEqual([]QUser{suite.UserList[8], suite.UserList[7]}, qUsers2)
	suite.True(bson.IsObjectIdHex(pg2))
	suite.Equal(int64(9), ts2)

	suite.Equal(0, len(qUsers3))
	suite.Equal("", pg3)
	suite.Equal(int64(0), ts3)

	suite.assertUsersEqual([]QUser{suite.UserList[6], suite.UserList[5]}, qUsers4)
	suite.True(bson.IsObjectIdHex(pg4))
	suite.Equal(int64(9), ts4)
}

func (suite *MongoStoreIntegrationTestSuite) TestACLModificationActions() {

	_ = suite.store.InsertSub(suite.ctx, "argo_uuid", "subFresh", "topicFresh", 0, 10, QPushConfig{}, time.Date(2020, 12, 19, 0, 0, 0, 0, time.Local))
	_ = suite.store.InsertTopic(suite.ctx, "argo_uuid", "topicFresh", "", time.Date(2020, 9, 11, 0, 0, 0, 0, time.Local))

	// test mod acl
	ExpectedACL01 := QAcl{[]string{"uuid1", "uuid2"}}
	eModACL1 := suite.store.ModACL(suite.ctx, "argo_uuid", "topics", "topicFresh", ExpectedACL01.ACL)
	suite.Nil(eModACL1)
	QAcl01, _ := suite.store.QueryACL(suite.ctx, "argo_uuid", "topics", "topicFresh")
	suite.Equal(ExpectedACL01, QAcl01)

	eModACL2 := suite.store.ModACL(suite.ctx, "argo_uuid", "subscriptions", "subFresh", ExpectedACL01.ACL)
	suite.Nil(eModACL2)
	QAcl01sub, _ := suite.store.QueryACL(suite.ctx, "argo_uuid", "topics", "topicFresh")
	suite.Equal(ExpectedACL01, QAcl01sub)

	eModACL3 := suite.store.ModACL(suite.ctx, "argo_uuid", "mistype", "sub1", []string{"u1", "u2"})
	suite.Equal("wrong resource type", eModACL3.Error())

	// test append acl
	ExpectedACL02 := QAcl{ACL: append(ExpectedACL01.ACL, "u3", "u4")}
	eAppACL1 := suite.store.AppendToACL(suite.ctx, "argo_uuid", "topics", "topicFresh", []string{"u3", "u4", "u4"})
	suite.Nil(eAppACL1)
	QAcl02, _ := suite.store.QueryACL(suite.ctx, "argo_uuid", "topics", "topicFresh")
	suite.Equal(ExpectedACL02, QAcl02)

	eAppACL2 := suite.store.AppendToACL(suite.ctx, "argo_uuid", "subscriptions", "subFresh", []string{"u3", "u4", "u4"})
	suite.Nil(eAppACL2)
	QAcl02sub, _ := suite.store.QueryACL(suite.ctx, "argo_uuid", "topics", "topicFresh")
	suite.Equal(ExpectedACL02, QAcl02sub)

	eAppACL3 := suite.store.AppendToACL(suite.ctx, "argo_uuid", "mistype", "sub1", []string{"u3", "u4", "u4"})
	suite.Equal("wrong resource type", eAppACL3.Error())

	// test remove acl
	ExpectedACL03 := QAcl{ACL: append(ExpectedACL01.ACL, "u3")}
	eRemACL1 := suite.store.RemoveFromACL(suite.ctx, "argo_uuid", "topics", "topicFresh", []string{"u1", "u4", "u5"})
	suite.Nil(eRemACL1)
	QAcl03, _ := suite.store.QueryACL(suite.ctx, "argo_uuid", "topics", "topicFresh")
	suite.Equal(ExpectedACL03, QAcl03)

	eRemACL2 := suite.store.RemoveFromACL(suite.ctx, "argo_uuid", "subscriptions", "subFresh", []string{"u1", "u4", "u5"})
	suite.Nil(eRemACL2)
	QAcl03sub, _ := suite.store.QueryACL(suite.ctx, "argo_uuid", "topics", "topicFresh")
	suite.Equal(ExpectedACL03, QAcl03sub)

	eRemACL3 := suite.store.RemoveFromACL(suite.ctx, "argo_uuid", "mistype", "sub1", []string{"u3", "u4", "u4"})
	suite.Equal("wrong resource type", eRemACL3.Error())

	_ = suite.store.RemoveSub(suite.ctx, "argo_uuid", "subFresh")
	_ = suite.store.RemoveTopic(suite.ctx, "argo_uuid", "topicFresh")
}

func (suite *MongoStoreIntegrationTestSuite) TestHasUsers() {
	allFound, notFound := suite.store.HasUsers(suite.ctx, "argo_uuid", []string{"UserA", "UserB", "FooUser"})
	suite.Equal(false, allFound)
	suite.Equal([]string{"FooUser"}, notFound)

	allFound, notFound = suite.store.HasUsers(suite.ctx, "argo_uuid", []string{"UserA", "UserB"})
	suite.Equal(true, allFound)
	suite.Equal([]string(nil), notFound)
}

func (suite *MongoStoreIntegrationTestSuite) TestCRUDProjects() {

	expProj1 := []QProject{suite.ProjectList[0]}
	expProj2 := []QProject{suite.ProjectList[1]}
	expProj3 := []QProject{suite.ProjectList[0], suite.ProjectList[1]}
	var expProj4 []QProject

	projectOut1, err := suite.store.QueryProjects(suite.ctx, "", "ARGO")
	suite.Equal(expProj1, projectOut1)
	suite.Equal(nil, err)
	projectOut2, err := suite.store.QueryProjects(suite.ctx, "", "ARGO2")
	suite.Equal(expProj2, projectOut2)
	suite.Equal(nil, err)
	projectOut3, err := suite.store.QueryProjects(suite.ctx, "", "")
	suite.Equal(expProj3, projectOut3)
	suite.Equal(nil, err)

	projectOut4, err := suite.store.QueryProjects(suite.ctx, "", "FOO")

	suite.Equal(expProj4, projectOut4)
	suite.Equal(errors.New("not found"), err)

	// Test queries by uuid
	projectOut7, err := suite.store.QueryProjects(suite.ctx, "argo_uuid2", "")
	suite.Equal(expProj2, projectOut7)
	suite.Equal(nil, err)
	projectOut8, err := suite.store.QueryProjects(suite.ctx, "foo_uuidNone", "")
	suite.Equal(expProj4, projectOut8)
	suite.Equal(errors.New("not found"), err)

	created := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.Local)
	modified := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.Local)
	_ = suite.store.InsertProject(suite.ctx, "argo_uuid3", "ARGO3", created, modified, "uuid1", "simple project")
	modified = time.Date(2010, time.November, 10, 23, 0, 0, 0, time.Local)
	expPr1 := QProject{UUID: "argo_uuid3", Name: "ARGO3", CreatedOn: created, ModifiedOn: modified, CreatedBy: "uuid1", Description: "a modified description"}
	_ = suite.store.UpdateProject(suite.ctx, "argo_uuid3", "", "a modified description", modified)
	prUp1, _ := suite.store.QueryProjects(suite.ctx, "argo_uuid3", "")
	suite.Equal(expPr1, prUp1[0])
	expPr2 := QProject{UUID: "argo_uuid3", Name: "ARGO_updated3", CreatedOn: created, ModifiedOn: modified, CreatedBy: "uuid1", Description: "a modified description"}
	_ = suite.store.UpdateProject(suite.ctx, "argo_uuid3", "ARGO_updated3", "", modified)
	prUp2, _ := suite.store.QueryProjects(suite.ctx, "argo_uuid3", "")
	suite.Equal(expPr2, prUp2[0])
	expPr3 := QProject{UUID: "argo_uuid3", Name: "ARGO_3", CreatedOn: created, ModifiedOn: modified, CreatedBy: "uuid1", Description: "a newly modified description"}
	_ = suite.store.UpdateProject(suite.ctx, "argo_uuid3", "ARGO_3", "a newly modified description", modified)
	prUp3, _ := suite.store.QueryProjects(suite.ctx, "argo_uuid3", "")
	suite.Equal(expPr3, prUp3[0])
	// Test RemoveProject
	_ = suite.store.RemoveProject(suite.ctx, "argo_uuid3")
	_, err = suite.store.QueryProjects(suite.ctx, "argo_uuid3", "")
	suite.Equal(errors.New("not found"), err)
}

func (suite *MongoStoreIntegrationTestSuite) TestQueryTotalMessagesPerProject() {
	expectedQpmc := []QProjectMessageCount{
		{ProjectUUID: "argo_uuid", NumberOfMessages: 30, AverageDailyMessages: 7.5},
	}
	qpmc, qpmcerr1 := suite.store.QueryTotalMessagesPerProject(suite.ctx, []string{"argo_uuid"}, time.Date(2018, 10, 1, 0, 0, 0, 0, time.UTC), time.Date(2018, 10, 4, 0, 0, 0, 0, time.UTC))
	suite.Equal(expectedQpmc, qpmc)
	suite.Nil(qpmcerr1)
}

func (suite *MongoStoreIntegrationTestSuite) TestCRUDSchemas() {

	// test QuerySchemas
	expectedSchemas := []QSchema{
		suite.SchemaList[0],
		suite.SchemaList[1],
		suite.SchemaList[2],
	}
	qqs1, _ := suite.store.QuerySchemas(suite.ctx, "argo_uuid", "", "")
	qqs2, _ := suite.store.QuerySchemas(suite.ctx, "argo_uuid", "schema_uuid_1", "")
	qqs3, _ := suite.store.QuerySchemas(suite.ctx, "argo_uuid", "schema_uuid_1", "schema-1")
	suite.assertSchemasEqual(expectedSchemas, qqs1)
	suite.assertSchemasEqual([]QSchema{expectedSchemas[0]}, []QSchema{qqs2[0]})
	suite.assertSchemasEqual([]QSchema{expectedSchemas[0]}, []QSchema{qqs3[0]})

	// test InsertSchema
	eis := suite.store.InsertSchema(suite.ctx, "argo_uuid", "uuid1", "s1-insert", "json", "raw")
	qs1, _ := suite.store.QuerySchemas(suite.ctx, "argo_uuid", "uuid1", "s1-insert")
	suite.Equal(QSchema{
		ProjectUUID: "argo_uuid",
		UUID:        "uuid1",
		Name:        "s1-insert",
		Type:        "json",
		RawSchema:   "raw",
	}, qs1[0])
	suite.Nil(eis)

	// test update schema
	_ = suite.store.InsertTopic(suite.ctx, "argo_uuid", "topicFresh", "", time.Date(2020, 9, 11, 0, 0, 0, 0, time.Local))
	_ = suite.store.UpdateSchema(suite.ctx, "uuid1", "new-name", "new-type", "new-raw-schema")
	eus := QSchema{UUID: "uuid1", ProjectUUID: "argo_uuid", Type: "new-type", Name: "new-name", RawSchema: "new-raw-schema"}
	qus, _ := suite.store.QuerySchemas(suite.ctx, "argo_uuid", "uuid1", "")
	suite.Equal(eus, qus[0])
	_ = suite.store.LinkTopicSchema(suite.ctx, "argo_uuid", "topicFresh", "uuid1")
	t, _, _, _ := suite.store.QueryTopics(suite.ctx, "argo_uuid", "", "topicFresh", "", 0)
	suite.Equal("uuid1", t[0].SchemaUUID)
	ed := suite.store.DeleteSchema(suite.ctx, "uuid1")
	expd, _ := suite.store.QuerySchemas(suite.ctx, "argo_uuid", "uuid1", "")
	// check that topic  no longer has any schema_uuid associated with it
	qtd, _, _, _ := suite.store.QueryTopics(suite.ctx, "argo_uuid", "", "topicFresh", "", 1)
	suite.Equal("", qtd[0].SchemaUUID)
	suite.Empty(expd)
	suite.Nil(ed)
	_ = suite.store.RemoveTopic(suite.ctx, "argo_uuid", "topicFresh")
}

func (suite *MongoStoreIntegrationTestSuite) TestCRUDRegistrations() {

	_ = suite.store.RegisterUser(suite.ctx, "ruuid1", "n1", "f1", "l1", "e1", "o1", "d1", "time", "atkn", "pending")
	expur1 := []QUserRegistration{{
		UUID:            "ruuid1",
		Name:            "n1",
		FirstName:       "f1",
		LastName:        "l1",
		Email:           "e1",
		Organization:    "o1",
		Description:     "d1",
		RegisteredAt:    "time",
		ActivationToken: "atkn",
		Status:          "pending",
	}}

	ur1, _ := suite.store.QueryRegistrations(suite.ctx, "ruuid1", "pending", "atkn", "n1", "e1", "o1")
	suite.Equal(expur1, ur1)

	ur12, _ := suite.store.QueryRegistrations(suite.ctx, "ruuid1", "", "", "", "", "")
	suite.Equal(expur1, ur12)

	expur2 := []QUserRegistration{{
		UUID:            "ur-uuid1",
		Name:            "urname",
		FirstName:       "urfname",
		LastName:        "urlname",
		Organization:    "urorg",
		Description:     "urdesc",
		Email:           "uremail",
		ActivationToken: "",
		Status:          "accepted",
		RegisteredAt:    "2019-05-12T22:26:58Z",
		ModifiedBy:      "uuid1",
		ModifiedAt:      "2020-05-17T22:26:58Z",
	}}
	_ = suite.store.UpdateRegistration(suite.ctx, "ur-uuid1", "accepted", "", "uuid1", "2020-05-17T22:26:58Z")
	ur2, _ := suite.store.QueryRegistrations(suite.ctx, "ur-uuid1", "accepted", "", "", "", "")
	suite.Equal(expur2, ur2)

	ur3, _ := suite.store.QueryRegistrations(suite.ctx, "", "", "", "", "", "")

	suite.Equal(2, len(ur3))
	suite.Equal([]QUserRegistration{expur2[0], expur1[0]}, ur3)
	_ = suite.store.DeleteRegistration(suite.ctx, "ruuid1")
	ur4, _ := suite.store.QueryRegistrations(suite.ctx, "", "", "", "", "", "")
	suite.Equal(1, len(ur4))

	dErr := suite.store.DeleteRegistration(suite.ctx, "unknown")
	suite.Equal("not found", dErr.Error())

}

func (suite *MongoStoreIntegrationTestSuite) TestResourcesCounters() {
	sdate := time.Date(2008, 11, 19, 8, 0, 0, 0, time.UTC)
	edate := time.Date(2020, 11, 21, 6, 0, 0, 0, time.UTC)
	tc, _ := suite.store.TopicsCount(suite.ctx, sdate, edate, []string{})
	sc, _ := suite.store.SubscriptionsCount(suite.ctx, sdate, edate, []string{})
	uc, _ := suite.store.UsersCount(suite.ctx, sdate, edate, []string{})

	suite.Equal(map[string]int64{"argo_uuid": 3}, tc)
	suite.Equal(map[string]int64{"argo_uuid": 3}, sc)
	suite.Equal(map[string]int64{"argo_uuid": 7, "argo_uuid2": 1}, uc)
}

func (suite *MongoStoreIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.store.Initialize()
	suite.initDB()
}

func (suite *MongoStoreIntegrationTestSuite) TearDownSuite() {
	suite.store.Close()
}

func TestMongoStoreIntegrationTestSuite(t *testing.T) {

	container, err := startContainer(context.Background())
	if err != nil {
		panic("Could not start container for mongodb integration tests. " + err.Error())
	}

	p, _ := container.MappedPort(context.Background(), "27017/tcp")

	mongoDBUri := fmt.Sprintf("mongodb://localhost:%s", p.Port())

	mongoStore := &MongoStore{
		Server:   mongoDBUri,
		Database: "argo_ams",
	}
	suite.Run(t, &MongoStoreIntegrationTestSuite{
		store: mongoStore,
		ctx:   context.Background(),
	})
}
