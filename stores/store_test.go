package stores

import (
	"testing"

	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/stretchr/testify/suite"
)

type StoreTestSuite struct {
	suite.Suite
}

func (suite *StoreTestSuite) TestMockStore() {

	store := NewMockStore("mockhost", "mockbase")
	suite.Equal("mockhost", store.Server)
	suite.Equal("mockbase", store.Database)

	eTopList := []QTopic{QTopic{"ARGO", "topic1"},
		QTopic{"ARGO", "topic2"},
		QTopic{"ARGO", "topic3"}}

	eSubList := []QSub{QSub{"ARGO", "sub1", "topic1", 0, 0, "", "", 10},
		QSub{"ARGO", "sub2", "topic2", 0, 0, "", "", 10},
		QSub{"ARGO", "sub3", "topic3", 0, 0, "", "", 10},
		QSub{"ARGO", "sub4", "topic4", 0, 0, "", "endpoint.foo", 10}}

	suite.Equal(eTopList, store.QueryTopics())
	suite.Equal(eSubList, store.QuerySubs())

	// Test Project
	suite.Equal(true, store.HasProject("ARGO"))
	suite.Equal(false, store.HasProject("FOO"))

	// Test user
	suite.Equal([]string{"admin", "member"}, store.GetUserRoles("ARGO", "S3CR3T"))
	suite.Equal([]string{}, store.GetUserRoles("ARGO", "SecretKey"))

	// Test roles
	suite.Equal(true, store.HasResourceRoles("topics:list_all", []string{"admin"}))
	suite.Equal(true, store.HasResourceRoles("topics:list_all", []string{"admin", "reader"}))
	suite.Equal(true, store.HasResourceRoles("topics:list_all", []string{"admin", "foo"}))
	suite.Equal(false, store.HasResourceRoles("topics:list_all", []string{"foo"}))
	suite.Equal(false, store.HasResourceRoles("topics:publish", []string{"reader"}))
	suite.Equal(true, store.HasResourceRoles("topics:list_all", []string{"admin"}))
	suite.Equal(true, store.HasResourceRoles("topics:list_all", []string{"publisher"}))
	suite.Equal(true, store.HasResourceRoles("topics:publish", []string{"publisher"}))

	store.InsertTopic("ARGO", "topicFresh")
	store.InsertSub("ARGO", "subFresh", "topicFresh", 0, 10)

	eTopList2 := []QTopic{QTopic{"ARGO", "topic1"},
		QTopic{"ARGO", "topic2"},
		QTopic{"ARGO", "topic3"},
		QTopic{"ARGO", "topicFresh"}}

	eSubList2 := []QSub{QSub{"ARGO", "sub1", "topic1", 0, 0, "", "", 10},
		QSub{"ARGO", "sub2", "topic2", 0, 0, "", "", 10},
		QSub{"ARGO", "sub3", "topic3", 0, 0, "", "", 10},
		QSub{"ARGO", "sub4", "topic4", 0, 0, "", "endpoint.foo", 10},
		QSub{"ARGO", "subFresh", "topicFresh", 0, 0, "", "", 10}}

	suite.Equal(eTopList2, store.QueryTopics())
	suite.Equal(eSubList2, store.QuerySubs())

	// Test delete on topic
	err := store.RemoveTopic("ARGO", "topicFresh")
	suite.Equal(nil, err)
	suite.Equal(eTopList, store.QueryTopics())
	err = store.RemoveTopic("ARGO", "topicFresh")
	suite.Equal("not found", err.Error())

	// Test delete on subscription
	err = store.RemoveSub("ARGO", "subFresh")
	suite.Equal(nil, err)
	suite.Equal(eSubList, store.QuerySubs())
	err = store.RemoveSub("ARGO", "subFresh")
	suite.Equal("not found", err.Error())

	sb, err := store.QueryOneSub("ARGO", "sub1")
	suite.Equal(sb, eSubList[0])
}

func TestStoresTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}
