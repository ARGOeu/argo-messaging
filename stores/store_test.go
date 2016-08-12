package stores

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
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

	eSubList := []QSub{QSub{"ARGO", "sub1", "topic1", 0, 0, "", "", 10, "linear", 300},
		QSub{"ARGO", "sub2", "topic2", 0, 0, "", "", 10, "linear", 300},
		QSub{"ARGO", "sub3", "topic3", 0, 0, "", "", 10, "linear", 300},
		QSub{"ARGO", "sub4", "topic4", 0, 0, "", "endpoint.foo", 10, "linear", 300}}

	suite.Equal(eTopList, store.QueryTopics())
	suite.Equal(eSubList, store.QuerySubs())

	// Test Project
	suite.Equal(true, store.HasProject("ARGO"))
	suite.Equal(false, store.HasProject("FOO"))

	// Test user
	roles01, _ := store.GetUserRoles("ARGO", "S3CR3T")
	roles02, _ := store.GetUserRoles("ARGO", "SecretKey")
	suite.Equal([]string{"admin", "member"}, roles01)
	suite.Equal([]string{}, roles02)

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
	store.InsertSub("ARGO", "subFresh", "topicFresh", 0, 10, "", "linear", 300)

	eTopList2 := []QTopic{QTopic{"ARGO", "topic1"},
		QTopic{"ARGO", "topic2"},
		QTopic{"ARGO", "topic3"},
		QTopic{"ARGO", "topicFresh"}}

	eSubList2 := []QSub{QSub{"ARGO", "sub1", "topic1", 0, 0, "", "", 10, "linear", 300},
		QSub{"ARGO", "sub2", "topic2", 0, 0, "", "", 10, "linear", 300},
		QSub{"ARGO", "sub3", "topic3", 0, 0, "", "", 10, "linear", 300},
		QSub{"ARGO", "sub4", "topic4", 0, 0, "", "endpoint.foo", 10, "linear", 300},
		QSub{"ARGO", "subFresh", "topicFresh", 0, 0, "", "", 10, "linear", 300}}

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

	// Query ACLS
	ExpectedACL01 := QAcl{[]string{"userA", "userB"}}
	QAcl01, _ := store.QueryACL("ARGO", "topic", "topic1")
	suite.Equal(ExpectedACL01, QAcl01)

	ExpectedACL02 := QAcl{[]string{"userA", "userB", "userD"}}
	QAcl02, _ := store.QueryACL("ARGO", "topic", "topic2")
	suite.Equal(ExpectedACL02, QAcl02)

	ExpectedACL03 := QAcl{[]string{"userC"}}
	QAcl03, _ := store.QueryACL("ARGO", "topic", "topic3")
	suite.Equal(ExpectedACL03, QAcl03)

	ExpectedACL04 := QAcl{[]string{"userA", "userB"}}
	QAcl04, _ := store.QueryACL("ARGO", "subscription", "sub1")
	suite.Equal(ExpectedACL04, QAcl04)

	ExpectedACL05 := QAcl{[]string{"userA", "userC"}}
	QAcl05, _ := store.QueryACL("ARGO", "subscription", "sub2")
	suite.Equal(ExpectedACL05, QAcl05)

	ExpectedACL06 := QAcl{[]string{"userD", "userB", "userA"}}
	QAcl06, _ := store.QueryACL("ARGO", "subscription", "sub3")
	suite.Equal(ExpectedACL06, QAcl06)

	ExpectedACL07 := QAcl{[]string{"userB", "userD"}}
	QAcl07, _ := store.QueryACL("ARGO", "subscription", "sub4")
	suite.Equal(ExpectedACL07, QAcl07)

	QAcl08, err08 := store.QueryACL("ARGO", "subscr", "sub4ss")
	suite.Equal(QAcl{}, QAcl08)
	suite.Equal(errors.New("not found"), err08)
}

func TestStoresTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}
