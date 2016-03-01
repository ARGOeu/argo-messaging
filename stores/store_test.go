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

	eTopList := []QTopics{QTopics{"ARGO", "topic1"},
		QTopics{"ARGO", "topic2"},
		QTopics{"ARGO", "topic3"}}

	eSubList := []QSubs{QSubs{"ARGO", "sub1", "topic1", 0},
		QSubs{"ARGO", "sub2", "topic2", 0},
		QSubs{"ARGO", "sub3", "topic3", 0}}

	suite.Equal(eTopList, store.QueryTopics())
	suite.Equal(eSubList, store.QuerySubs())

}

func TestStoresTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}
