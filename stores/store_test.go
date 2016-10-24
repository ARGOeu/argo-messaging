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
	roles01, _ := store.GetUserRoles("argo_uuid", "S3CR3T")
	roles02, _ := store.GetUserRoles("argo_uuid", "SecretKey")
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
	ExpectedACL01 := QAcl{[]string{"uuid1", "uuid2"}}
	QAcl01, _ := store.QueryACL("argo_uuid", "topics", "topic1")
	suite.Equal(ExpectedACL01, QAcl01)

	ExpectedACL02 := QAcl{[]string{"uuid1", "uuid2", "uuid4"}}
	QAcl02, _ := store.QueryACL("argo_uuid", "topics", "topic2")
	suite.Equal(ExpectedACL02, QAcl02)

	ExpectedACL03 := QAcl{[]string{"uuid3"}}
	QAcl03, _ := store.QueryACL("argo_uuid", "topics", "topic3")
	suite.Equal(ExpectedACL03, QAcl03)

	ExpectedACL04 := QAcl{[]string{"uuid1", "uuid2"}}
	QAcl04, _ := store.QueryACL("argo_uuid", "subscriptions", "sub1")
	suite.Equal(ExpectedACL04, QAcl04)

	ExpectedACL05 := QAcl{[]string{"uuid1", "uuid3"}}
	QAcl05, _ := store.QueryACL("argo_uuid", "subscriptions", "sub2")
	suite.Equal(ExpectedACL05, QAcl05)

	ExpectedACL06 := QAcl{[]string{"uuid4", "uuid2", "uuid1"}}
	QAcl06, _ := store.QueryACL("argo_uuid", "subscriptions", "sub3")
	suite.Equal(ExpectedACL06, QAcl06)

	ExpectedACL07 := QAcl{[]string{"uuid2", "uuid4"}}
	QAcl07, _ := store.QueryACL("argo_uuid", "subscriptions", "sub4")
	suite.Equal(ExpectedACL07, QAcl07)

	QAcl08, err08 := store.QueryACL("argo_uuid", "subscr", "sub4ss")
	suite.Equal(QAcl{}, QAcl08)
	suite.Equal(errors.New("not found"), err08)

	//Check has users
	allFound, notFound := store.HasUsers("argo_uuid", []string{"UserA", "UserB", "FooUser"})
	suite.Equal(false, allFound)
	suite.Equal([]string{"FooUser"}, notFound)

	allFound, notFound = store.HasUsers("argo_uuid", []string{"UserA", "UserB"})
	suite.Equal(true, allFound)
	suite.Equal([]string(nil), notFound)

	created := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	modified := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	// Test Projects
	qPr := QProject{UUID: "argo_uuid", Name: "ARGO", CreatedOn: created, ModifiedOn: modified, CreatedBy: "uuid1", Description: "simple project"}
	qPr2 := QProject{UUID: "argo_uuid2", Name: "ARGO2", CreatedOn: created, ModifiedOn: modified, CreatedBy: "uuid1", Description: "simple project"}
	expProj1 := []QProject{qPr}
	expProj2 := []QProject{qPr2}
	expProj3 := []QProject{qPr, qPr2}
	expProj4 := []QProject{}

	projectOut1, err := store.QueryProjects("", "ARGO")
	suite.Equal(expProj1, projectOut1)
	suite.Equal(nil, err)
	projectOut2, err := store.QueryProjects("", "ARGO2")
	suite.Equal(expProj2, projectOut2)
	suite.Equal(nil, err)
	projectOut3, err := store.QueryProjects("", "")
	suite.Equal(expProj3, projectOut3)
	suite.Equal(nil, err)

	projectOut4, err := store.QueryProjects("", "FOO")

	suite.Equal(expProj4, projectOut4)
	suite.Equal(errors.New("not found"), err)
	// Test insert project
	qPr3 := QProject{UUID: "argo_uuid3", Name: "ARGO3", CreatedOn: created, ModifiedOn: modified, CreatedBy: "uuid1", Description: "simple project"}
	expProj5 := []QProject{qPr, qPr2, qPr3}
	expProj6 := []QProject{qPr3}
	store.InsertProject("argo_uuid3", "ARGO3", created, modified, "uuid1", "simple project")
	projectOut5, err := store.QueryProjects("", "")
	suite.Equal(expProj5, projectOut5)
	suite.Equal(nil, err)

	projectOut6, err := store.QueryProjects("argo_uuid2", "ARGO3")
	suite.Equal(expProj6, projectOut6)
	suite.Equal(nil, err)
	// Test queries by uuid
	projectOut7, err := store.QueryProjects("argo_uuid2", "")
	suite.Equal(expProj2, projectOut7)
	suite.Equal(nil, err)
	projectOut8, err := store.QueryProjects("foo_uuidNone", "")
	suite.Equal(expProj4, projectOut8)
	suite.Equal(errors.New("not found"), err)

	// Test update project
	modified = time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)
	expPr1 := QProject{UUID: "argo_uuid3", Name: "ARGO3", CreatedOn: created, ModifiedOn: modified, CreatedBy: "uuid1", Description: "a modified description"}
	store.UpdateProject("argo_uuid3", "", "a modified description", modified)
	prUp1, _ := store.QueryProjects("argo_uuid3", "")
	suite.Equal(expPr1, prUp1[0])
	expPr2 := QProject{UUID: "argo_uuid3", Name: "ARGO_updated3", CreatedOn: created, ModifiedOn: modified, CreatedBy: "uuid1", Description: "a modified description"}
	store.UpdateProject("argo_uuid3", "ARGO_updated3", "", modified)
	prUp2, _ := store.QueryProjects("argo_uuid3", "")
	suite.Equal(expPr2, prUp2[0])
	expPr3 := QProject{UUID: "argo_uuid3", Name: "ARGO_3", CreatedOn: created, ModifiedOn: modified, CreatedBy: "uuid1", Description: "a newly modified description"}
	store.UpdateProject("argo_uuid3", "ARGO_3", "a newly modified description", modified)
	prUp3, _ := store.QueryProjects("argo_uuid3", "")
	suite.Equal(expPr3, prUp3[0])

	// Test RemoveProjectTopics
	store.RemoveProjectTopics("argo_uuid")
	resTop, _ := store.QueryTopics("argo_uuid", "")
	suite.Equal([]QTopic{}, resTop)
	store.RemoveProjectSubs("argo_uuid")
	resSub, _ := store.QuerySubs("argo_uuid", "")
	suite.Equal([]QSub{}, resSub)

	// Test RemoveProject
	store.RemoveProject("argo_uuid")
	resProj, err := store.QueryProjects("argo_uuid", "")
	suite.Equal([]QProject{}, resProj)
	suite.Equal(errors.New("not found"), err)

	// Test Insert User
	qRoleAdmin1 := []QProjectRoles{QProjectRoles{"argo_uuid", []string{"admin"}}}
	qRoles := []QProjectRoles{QProjectRoles{"argo_uuid", []string{"admin"}}, QProjectRoles{"argo_uuid2", []string{"admin", "viewer"}}}
	expUsr10 := QUser{"user_uuid10", qRoleAdmin1, "newUser1", "A3B94A94V3A", "fake@email.com", []string{}}
	expUsr11 := QUser{"user_uuid11", qRoles, "newUser2", "BX312Z34NLQ", "fake@email.com", []string{}}
	store.InsertUser("user_uuid10", qRoleAdmin1, "newUser1", "A3B94A94V3A", "fake@email.com", []string{})
	store.InsertUser("user_uuid11", qRoles, "newUser2", "BX312Z34NLQ", "fake@email.com", []string{})
	usr10, _ := store.QueryUsers("argo_uuid", "user_uuid10", "")
	usr11, _ := store.QueryUsers("argo_uuid", "", "newUser2")

	suite.Equal(expUsr10, usr10[0])
	suite.Equal(expUsr11, usr11[0])

	rolesA, usernameA := store.GetUserRoles("argo_uuid", "BX312Z34NLQ")
	rolesB, usernameB := store.GetUserRoles("argo_uuid2", "BX312Z34NLQ")
	suite.Equal("newUser2", usernameA)
	suite.Equal("newUser2", usernameB)

	suite.Equal([]string{"admin"}, rolesA)
	suite.Equal([]string{"admin", "viewer"}, rolesB)

	// Test Update User
	usrUpdated := QUser{"user_uuid11", qRoles, "updated_name", "BX312Z34NLQ", "fake@email.com", []string{"service_admin"}}
	store.UpdateUser("user_uuid11", nil, "updated_name", "", []string{"service_admin"})
	usr11, _ = store.QueryUsers("", "user_uuid11", "")
	suite.Equal(usrUpdated, usr11[0])

	// Test Remove User
	store.RemoveUser("user_uuid11")
	usr11, err = store.QueryUsers("", "user_uuid11", "")
	suite.Equal(errors.New("not found"), err)

}

func TestStoresTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}
