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

	eSubList := []QSub{QSub{"ARGO", "sub1", "topic1", 0},
		QSub{"ARGO", "sub2", "topic2", 0},
		QSub{"ARGO", "sub3", "topic3", 0}}

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
	suite.Equal(true, store.HasResourceRoles("topics:list_all", []string{"publisher"}))
	suite.Equal(true, store.HasResourceRoles("topics:publish", []string{"publisher"}))

}

func TestStoresTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}
