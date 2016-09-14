package stores

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type StoreTestSuite struct {
	suite.Suite
}

func (suite *StoreTestSuite) TestMockStore() {

	store := NewMockStore("mockhost", "mockbase")
	suite.Equal("mockhost", store.Server)
	suite.Equal("mockbase", store.Database)

	eTopList := []QTopic{QTopic{"argo_uuid", "topic1"},
		QTopic{"argo_uuid", "topic2"},
		QTopic{"argo_uuid", "topic3"}}

	eSubList := []QSub{QSub{"argo_uuid", "sub1", "topic1", 0, 0, "", "", 10, "linear", 300},
		QSub{"argo_uuid", "sub2", "topic2", 0, 0, "", "", 10, "linear", 300},
		QSub{"argo_uuid", "sub3", "topic3", 0, 0, "", "", 10, "linear", 300},
		QSub{"argo_uuid", "sub4", "topic4", 0, 0, "", "endpoint.foo", 10, "linear", 300}}

	tpList, _ := store.QueryTopics("argo_uuid", "")
	suite.Equal(eTopList, tpList)
	subList, _ := store.QuerySubs("argo_uuid", "")
	suite.Equal(eSubList, subList)

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

	store.InsertTopic("argo_uuid", "topicFresh")
	store.InsertSub("argo_uuid", "subFresh", "topicFresh", 0, 10, "", "linear", 300)

	eTopList2 := []QTopic{QTopic{"argo_uuid", "topic1"},
		QTopic{"argo_uuid", "topic2"},
		QTopic{"argo_uuid", "topic3"},
		QTopic{"argo_uuid", "topicFresh"}}

	eSubList2 := []QSub{QSub{"argo_uuid", "sub1", "topic1", 0, 0, "", "", 10, "linear", 300},
		QSub{"argo_uuid", "sub2", "topic2", 0, 0, "", "", 10, "linear", 300},
		QSub{"argo_uuid", "sub3", "topic3", 0, 0, "", "", 10, "linear", 300},
		QSub{"argo_uuid", "sub4", "topic4", 0, 0, "", "endpoint.foo", 10, "linear", 300},
		QSub{"argo_uuid", "subFresh", "topicFresh", 0, 0, "", "", 10, "linear", 300}}

	tpList, _ = store.QueryTopics("argo_uuid", "")
	suite.Equal(eTopList2, tpList)
	subList, _ = store.QuerySubs("argo_uuid", "")
	suite.Equal(eSubList2, subList)

	// Test delete on topic
	err := store.RemoveTopic("argo_uuid", "topicFresh")
	suite.Equal(nil, err)
	tpList, _ = store.QueryTopics("argo_uuid", "")
	suite.Equal(eTopList, tpList)
	err = store.RemoveTopic("argo_uuid", "topicFresh")
	suite.Equal("not found", err.Error())

	// Test delete on subscription
	err = store.RemoveSub("argo_uuid", "subFresh")
	suite.Equal(nil, err)
	subList, _ = store.QuerySubs("argo_uuid", "")
	suite.Equal(eSubList, subList)
	err = store.RemoveSub("argo_uuid", "subFresh")
	suite.Equal("not found", err.Error())

	sb, err := store.QueryOneSub("argo_uuid", "sub1")
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

	//Check has users
	allFound, notFound := store.HasUsers("ARGO", []string{"UserA", "UserB", "FooUser"})
	suite.Equal(false, allFound)
	suite.Equal([]string{"FooUser"}, notFound)

	allFound, notFound = store.HasUsers("ARGO", []string{"UserA", "UserB"})
	suite.Equal(true, allFound)
	suite.Equal([]string(nil), notFound)

	created := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	modified := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	// Test Projects
	qPr := QProject{UUID: "argo_uuid", Name: "ARGO", CreatedOn: created, ModifiedOn: modified, CreatedBy: "userA", Description: "simple project"}
	qPr2 := QProject{UUID: "argo_uuid2", Name: "ARGO2", CreatedOn: created, ModifiedOn: modified, CreatedBy: "userA", Description: "simple project"}
	expProj1 := []QProject{qPr}
	expProj2 := []QProject{qPr2}
	expProj3 := []QProject{qPr, qPr2}
	expProj4 := []QProject{}
	projectOut1, err := store.QueryProjects("ARGO", "")
	suite.Equal(expProj1, projectOut1)
	suite.Equal(nil, err)
	projectOut2, err := store.QueryProjects("ARGO2", "")
	suite.Equal(expProj2, projectOut2)
	suite.Equal(nil, err)
	projectOut3, err := store.QueryProjects("", "")
	suite.Equal(expProj3, projectOut3)
	suite.Equal(nil, err)
	projectOut4, err := store.QueryProjects("FOO", "")
	suite.Equal(expProj4, projectOut4)
	suite.Equal(errors.New("not found"), err)
	// Test insert project
	qPr3 := QProject{UUID: "argo_uuid3", Name: "ARGO3", CreatedOn: created, ModifiedOn: modified, CreatedBy: "userA", Description: "simple project"}
	expProj5 := []QProject{qPr, qPr2, qPr3}
	expProj6 := []QProject{qPr3}
	store.InsertProject("argo_uuid3", "ARGO3", created, modified, "userA", "simple project")
	projectOut5, err := store.QueryProjects("", "")
	suite.Equal(expProj5, projectOut5)
	suite.Equal(nil, err)
	projectOut6, err := store.QueryProjects("ARGO3", "argo_uuid2")
	suite.Equal(expProj6, projectOut6)
	suite.Equal(nil, err)
	// Test queries by uuid
	projectOut7, err := store.QueryProjects("", "argo_uuid2")
	suite.Equal(expProj2, projectOut7)
	suite.Equal(nil, err)
	projectOut8, err := store.QueryProjects("", "foo_uuidNone")
	suite.Equal(expProj4, projectOut8)
	suite.Equal(errors.New("not found"), err)

}

func TestStoresTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}
