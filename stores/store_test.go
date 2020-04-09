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

	eTopList := []QTopic{
		{3, "argo_uuid", "topic4", 0, 0, time.Date(0, 0, 0, 0, 0, 0, 0, time.Local), 0, ""},
		{2, "argo_uuid", "topic3", 0, 0, time.Date(2019, 5, 7, 0, 0, 0, 0, time.Local), 8.99, "schema_uuid_3"},
		{1, "argo_uuid", "topic2", 0, 0, time.Date(2019, 5, 8, 0, 0, 0, 0, time.Local), 5.45, "schema_uuid_1"},
		{0, "argo_uuid", "topic1", 0, 0, time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local), 10, ""},
	}

	eSubList := []QSub{
		{3, "argo_uuid", "sub4", "topic4", 0, 0, "", "endpoint.foo", 1, 10, "linear", 300, 0, 0, "push-id-1", true, time.Date(0, 0, 0, 0, 0, 0, 0, time.Local), 0},
		{2, "argo_uuid", "sub3", "topic3", 0, 0, "", "", 0, 10, "", 0, 0, 0, "", false, time.Date(2019, 5, 8, 0, 0, 0, 0, time.Local), 5.45},
		{1, "argo_uuid", "sub2", "topic2", 0, 0, "", "", 0, 10, "", 0, 0, 0, "", false, time.Date(2019, 5, 7, 0, 0, 0, 0, time.Local), 8.99},
		{0, "argo_uuid", "sub1", "topic1", 0, 0, "", "", 0, 10, "", 0, 0, 0, "", false, time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local), 10},
	}
	// retrieve all topics
	tpList, ts1, pg1, _ := store.QueryTopics("argo_uuid", "", "", "", 0)
	suite.Equal(eTopList, tpList)
	suite.Equal(int32(4), ts1)
	suite.Equal("", pg1)

	// retrieve first 2
	eTopList1st2 := []QTopic{
		{3, "argo_uuid", "topic4", 0, 0, time.Date(0, 0, 0, 0, 0, 0, 0, time.Local), 0, ""},
		{2, "argo_uuid", "topic3", 0, 0, time.Date(2019, 5, 7, 0, 0, 0, 0, time.Local), 8.99, "schema_uuid_3"},
	}
	tpList2, ts2, pg2, _ := store.QueryTopics("argo_uuid", "", "", "", 2)
	suite.Equal(eTopList1st2, tpList2)
	suite.Equal(int32(4), ts2)
	suite.Equal("1", pg2)

	// retrieve the last one
	eTopList3 := []QTopic{
		{0, "argo_uuid", "topic1", 0, 0, time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local), 10, ""},
	}
	tpList3, ts3, pg3, _ := store.QueryTopics("argo_uuid", "", "", "0", 1)
	suite.Equal(eTopList3, tpList3)
	suite.Equal(int32(4), ts3)
	suite.Equal("", pg3)

	// retrieve a single topic
	eTopList4 := []QTopic{
		{0, "argo_uuid", "topic1", 0, 0, time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local), 10, ""},
	}
	tpList4, ts4, pg4, _ := store.QueryTopics("argo_uuid", "", "topic1", "", 0)
	suite.Equal(eTopList4, tpList4)
	suite.Equal(int32(0), ts4)
	suite.Equal("", pg4)

	// retrieve user's topics
	eTopList5 := []QTopic{
		{1, "argo_uuid", "topic2", 0, 0, time.Date(2019, 5, 8, 0, 0, 0, 0, time.Local), 5.45, "schema_uuid_1"},
		{0, "argo_uuid", "topic1", 0, 0, time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local), 10, ""},
	}
	tpList5, ts5, pg5, _ := store.QueryTopics("argo_uuid", "uuid1", "", "", 0)
	suite.Equal(eTopList5, tpList5)
	suite.Equal(int32(2), ts5)
	suite.Equal("", pg5)

	// retrieve use's topic with pagination
	eTopList6 := []QTopic{
		{1, "argo_uuid", "topic2", 0, 0, time.Date(2019, 5, 8, 0, 0, 0, 0, time.Local), 5.45, "schema_uuid_1"},
	}

	tpList6, ts6, pg6, _ := store.QueryTopics("argo_uuid", "uuid1", "", "", 1)
	suite.Equal(eTopList6, tpList6)
	suite.Equal(int32(2), ts6)
	suite.Equal("0", pg6)

	// retrieve all subs
	subList, ts1, pg1, err1 := store.QuerySubs("argo_uuid", "", "", "", 0)
	suite.Equal(eSubList, subList)
	suite.Equal(int32(4), ts1)
	suite.Equal("", pg3)

	// retrieve first 2 subs
	eSubListFirstPage := []QSub{
		{3, "argo_uuid", "sub4", "topic4", 0, 0, "", "endpoint.foo", 1, 10, "linear", 300, 0, 0, "push-id-1", true, time.Date(0, 0, 0, 0, 0, 0, 0, time.Local), 0},
		{2, "argo_uuid", "sub3", "topic3", 0, 0, "", "", 0, 10, "", 0, 0, 0, "", false, time.Date(2019, 5, 8, 0, 0, 0, 0, time.Local), 5.45}}

	subList2, ts2, pg2, err2 := store.QuerySubs("argo_uuid", "", "", "", 2)
	suite.Equal(eSubListFirstPage, subList2)
	suite.Equal(int32(4), ts2)
	suite.Equal("1", pg2)

	// retrieve next 2 subs
	eSubListNextPage := []QSub{
		{1, "argo_uuid", "sub2", "topic2", 0, 0, "", "", 0, 10, "", 0, 0, 0, "", false, time.Date(2019, 5, 7, 0, 0, 0, 0, time.Local), 8.99},
		{0, "argo_uuid", "sub1", "topic1", 0, 0, "", "", 0, 10, "", 0, 0, 0, "", false, time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local), 10},
	}

	subList3, ts3, pg3, err3 := store.QuerySubs("argo_uuid", "", "", "1", 2)
	suite.Equal(eSubListNextPage, subList3)
	suite.Equal(int32(4), ts3)
	suite.Equal("", pg3)

	// retrieve user's subs
	eSubList4 := []QSub{
		{ID: 3, ProjectUUID: "argo_uuid", Name: "sub4", Topic: "topic4", Offset: 0, NextOffset: 0, PendingAck: "", PushEndpoint: "endpoint.foo", MaxMessages: 1, Ack: 10, RetPolicy: "linear", RetPeriod: 300, MsgNum: 0, TotalBytes: 0, VerificationHash: "push-id-1", Verified: true, LatestConsume: time.Date(0, 0, 0, 0, 0, 0, 0, time.Local), ConsumeRate: 0},
		{ID: 2, ProjectUUID: "argo_uuid", Name: "sub3", Topic: "topic3", Offset: 0, NextOffset: 0, PendingAck: "", PushEndpoint: "", MaxMessages: 0, Ack: 10, RetPolicy: "", RetPeriod: 0, MsgNum: 0, TotalBytes: 0, LatestConsume: time.Date(2019, 5, 8, 0, 0, 0, 0, time.Local), ConsumeRate: 5.45},
		{ID: 1, ProjectUUID: "argo_uuid", Name: "sub2", Topic: "topic2", Offset: 0, NextOffset: 0, PendingAck: "", PushEndpoint: "", MaxMessages: 0, Ack: 10, RetPolicy: "", RetPeriod: 0, MsgNum: 0, TotalBytes: 0, LatestConsume: time.Date(2019, 5, 7, 0, 0, 0, 0, time.Local), ConsumeRate: 8.99},
	}

	subList4, ts4, pg4, err4 := store.QuerySubs("argo_uuid", "uuid1", "", "", 0)

	suite.Equal(int32(3), ts4)
	suite.Equal("", pg4)
	suite.Equal(eSubList4, subList4)

	// retrieve user's subs
	eSubList5 := []QSub{
		{ID: 3, ProjectUUID: "argo_uuid", Name: "sub4", Topic: "topic4", Offset: 0, NextOffset: 0, PendingAck: "", PushEndpoint: "endpoint.foo", MaxMessages: 1, Ack: 10, RetPolicy: "linear", RetPeriod: 300, MsgNum: 0, TotalBytes: 0, VerificationHash: "push-id-1", Verified: true, LatestConsume: time.Date(0, 0, 0, 0, 0, 0, 0, time.Local), ConsumeRate: 0},
		{ID: 2, ProjectUUID: "argo_uuid", Name: "sub3", Topic: "topic3", Offset: 0, NextOffset: 0, PendingAck: "", PushEndpoint: "", MaxMessages: 0, Ack: 10, RetPolicy: "", RetPeriod: 0, MsgNum: 0, TotalBytes: 0, LatestConsume: time.Date(2019, 5, 8, 0, 0, 0, 0, time.Local), ConsumeRate: 5.45},
	}
	subList5, ts5, pg5, err5 := store.QuerySubs("argo_uuid", "uuid1", "", "", 2)

	suite.Equal(int32(3), ts5)
	suite.Equal("1", pg5)
	suite.Equal(eSubList5, subList5)

	suite.Nil(err1)
	suite.Nil(err2)
	suite.Nil(err3)
	suite.Nil(err4)
	suite.Nil(err5)

	// test retrieve subs by topic
	subListByTopic, errSublistByTopic := store.QuerySubsByTopic("argo_uuid", "topic1")
	suite.Equal([]QSub{
		{
			ID:            0,
			ProjectUUID:   "argo_uuid",
			Name:          "sub1",
			Topic:         "topic1",
			Offset:        0,
			NextOffset:    0,
			PendingAck:    "",
			PushEndpoint:  "",
			Ack:           10,
			RetPolicy:     "",
			RetPeriod:     0,
			MsgNum:        0,
			TotalBytes:    0,
			LatestConsume: time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local),
			ConsumeRate:   10,
		},
	}, subListByTopic)
	suite.Nil(errSublistByTopic)

	// Test ProjectUUID
	suite.Equal(true, store.HasProject("ARGO"))
	suite.Equal(false, store.HasProject("FOO"))

	// check query all
	qdsAll, _ := store.QueryDailyTopicMsgCount("", "", time.Time{})
	suite.Equal(store.DailyTopicMsgCount, qdsAll)

	// test daily count
	store.IncrementDailyTopicMsgCount("argo_uuid", "topic1", 40, time.Date(2018, 10, 1, 0, 0, 0, 0, time.UTC))
	qds, _ := store.QueryDailyTopicMsgCount("argo_uuid", "topic1", time.Date(2018, 10, 1, 0, 0, 0, 0, time.UTC))
	suite.Equal(int64(80), qds[0].NumberOfMessages)

	// check if the it was inserted since it wasn't present
	store.IncrementDailyTopicMsgCount("argo_uuid", "some_other_topic", 70, time.Date(2018, 10, 1, 0, 0, 0, 0, time.UTC))
	qds2, _ := store.QueryDailyTopicMsgCount("argo_uuid", "some_other_topic", time.Date(2018, 10, 1, 0, 0, 0, 0, time.UTC))
	suite.Equal(int64(70), qds2[0].NumberOfMessages)

	// Test user
	roles01, _ := store.GetUserRoles("argo_uuid", "S3CR3T")
	roles02, _ := store.GetUserRoles("argo_uuid", "SecretKey")
	suite.Equal([]string{"consumer", "publisher"}, roles01)
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

	store.InsertTopic("argo_uuid", "topicFresh", "")
	store.InsertSub("argo_uuid", "subFresh", "topicFresh", 0, 0, 10, "", "", 0, "", false)

	eTopList2 := []QTopic{
		{4, "argo_uuid", "topicFresh", 0, 0, time.Time{}, 0, ""},
		{3, "argo_uuid", "topic4", 0, 0, time.Date(0, 0, 0, 0, 0, 0, 0, time.Local), 0, ""},
		{2, "argo_uuid", "topic3", 0, 0, time.Date(2019, 5, 7, 0, 0, 0, 0, time.Local), 8.99, "schema_uuid_3"},
		{1, "argo_uuid", "topic2", 0, 0, time.Date(2019, 5, 8, 0, 0, 0, 0, time.Local), 5.45, "schema_uuid_1"},
		{0, "argo_uuid", "topic1", 0, 0, time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local), 10, ""},
	}

	eSubList2 := []QSub{
		{4, "argo_uuid", "subFresh", "topicFresh", 0, 0, "", "", 0, 10, "", 0, 0, 0, "", false, time.Time{}, 0},
		{3, "argo_uuid", "sub4", "topic4", 0, 0, "", "endpoint.foo", 1, 10, "linear", 300, 0, 0, "push-id-1", true, time.Date(0, 0, 0, 0, 0, 0, 0, time.Local), 0},
		{2, "argo_uuid", "sub3", "topic3", 0, 0, "", "", 0, 10, "", 0, 0, 0, "", false, time.Date(2019, 5, 8, 0, 0, 0, 0, time.Local), 5.45},
		{1, "argo_uuid", "sub2", "topic2", 0, 0, "", "", 0, 10, "", 0, 0, 0, "", false, time.Date(2019, 5, 7, 0, 0, 0, 0, time.Local), 8.99},
		{0, "argo_uuid", "sub1", "topic1", 0, 0, "", "", 0, 10, "", 0, 0, 0, "", false, time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local), 10}}

	tpList, _, _, _ = store.QueryTopics("argo_uuid", "", "", "", 0)
	suite.Equal(eTopList2, tpList)
	subList, _, _, _ = store.QuerySubs("argo_uuid", "", "", "", 0)
	suite.Equal(eSubList2, subList)

	// Test delete on topic
	err := store.RemoveTopic("argo_uuid", "topicFresh")
	suite.Equal(nil, err)
	tpList, _, _, _ = store.QueryTopics("argo_uuid", "", "", "", 0)
	suite.Equal(eTopList, tpList)
	err = store.RemoveTopic("argo_uuid", "topicFresh")
	suite.Equal("not found", err.Error())

	// Test delete on subscription
	err = store.RemoveSub("argo_uuid", "subFresh")
	suite.Equal(nil, err)
	subList, _, _, _ = store.QuerySubs("argo_uuid", "", "", "", 0)
	suite.Equal(eSubList, subList)
	err = store.RemoveSub("argo_uuid", "subFresh")
	suite.Equal("not found", err.Error())

	sb, err := store.QueryOneSub("argo_uuid", "sub1")
	esb := QSub{0, "argo_uuid", "sub1", "topic1", 0, 0, "", "", 0, 10, "", 0, 0, 0, "", false, time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local), 10}
	suite.Equal(esb, sb)

	// Test modify ack deadline in store
	store.ModAck("argo_uuid", "sub1", 66)
	subAck, _ := store.QueryOneSub("argo_uuid", "sub1")
	suite.Equal(66, subAck.Ack)

	// Test mod push sub
	e1 := store.ModSubPush("argo_uuid", "sub1", "example.com", 3, "linear", 400, "hash-1", true)
	sub1, _ := store.QueryOneSub("argo_uuid", "sub1")
	suite.Nil(e1)
	suite.Equal("example.com", sub1.PushEndpoint)
	suite.Equal(int64(3), sub1.MaxMessages)
	suite.Equal("linear", sub1.RetPolicy)
	suite.Equal(400, sub1.RetPeriod)
	suite.Equal("hash-1", sub1.VerificationHash)
	suite.True(sub1.Verified)

	e2 := store.ModSubPush("argo_uuid", "unknown", "", 0, "", 0, "", false)
	suite.Equal("not found", e2.Error())

	// exists in acl
	existsE1 := store.ExistsInACL("argo_uuid", "topics", "topic1", "uuid1")
	suite.Nil(existsE1)

	existsE2 := store.ExistsInACL("argo_uuid", "topics", "topic1", "unknown")
	suite.Equal("not found", existsE2.Error())

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

	ExpectedACL07 := QAcl{[]string{"uuid2", "uuid4", "uuid7"}}
	QAcl07, _ := store.QueryACL("argo_uuid", "subscriptions", "sub4")
	suite.Equal(ExpectedACL07, QAcl07)

	QAcl08, err08 := store.QueryACL("argo_uuid", "subscr", "sub4ss")
	suite.Equal(QAcl{}, QAcl08)
	suite.Equal(errors.New("not found"), err08)

	// test mod acl
	eModACL1 := store.ModACL("argo_uuid", "topics", "topic1", []string{"u1", "u2"})
	suite.Nil(eModACL1)
	tACL := store.TopicsACL["topic1"].ACL
	suite.Equal([]string{"u1", "u2"}, tACL)

	eModACL2 := store.ModACL("argo_uuid", "subscriptions", "sub1", []string{"u1", "u2"})
	suite.Nil(eModACL2)
	sACL := store.SubsACL["sub1"].ACL
	suite.Equal([]string{"u1", "u2"}, sACL)

	eModACL3 := store.ModACL("argo_uuid", "mistype", "sub1", []string{"u1", "u2"})
	suite.Equal("wrong resource type", eModACL3.Error())

	// test append acl
	eAppACL1 := store.AppendToACL("argo_uuid", "topics", "topic1", []string{"u3", "u4", "u4"})
	suite.Nil(eAppACL1)
	tACLapp := store.TopicsACL["topic1"].ACL
	suite.Equal([]string{"u1", "u2", "u3", "u4"}, tACLapp)

	eAppACL2 := store.AppendToACL("argo_uuid", "subscriptions", "sub1", []string{"u3", "u4", "u4"})
	suite.Nil(eAppACL2)
	sACLapp := store.SubsACL["sub1"].ACL
	suite.Equal([]string{"u1", "u2", "u3", "u4"}, sACLapp)

	eAppACL3 := store.AppendToACL("argo_uuid", "mistype", "sub1", []string{"u3", "u4", "u4"})
	suite.Equal("wrong resource type", eAppACL3.Error())

	// test remove acl
	eRemACL1 := store.RemoveFromACL("argo_uuid", "topics", "topic1", []string{"u1", "u4", "u5"})
	suite.Nil(eRemACL1)
	tACLRem := store.TopicsACL["topic1"].ACL
	suite.Equal([]string{"u2", "u3"}, tACLRem)

	eRemACL2 := store.RemoveFromACL("argo_uuid", "subscriptions", "sub1", []string{"u1", "u4", "u5"})
	suite.Nil(eRemACL2)
	sACLRem := store.SubsACL["sub1"].ACL
	suite.Equal([]string{"u2", "u3"}, sACLRem)

	eRemACL3 := store.RemoveFromACL("argo_uuid", "mistype", "sub1", []string{"u3", "u4", "u4"})
	suite.Equal("wrong resource type", eRemACL3.Error())

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

	// Test Sub Update Pull
	err = store.UpdateSubPull("argo_uuid", "sub4", 4, "2016-10-11T12:00:35:15Z")
	qSubUpd, _, _, err := store.QuerySubs("argo_uuid", "", "sub4", "", 0)
	var nxtOff int64 = 4
	suite.Equal(qSubUpd[0].NextOffset, nxtOff)
	suite.Equal("2016-10-11T12:00:35:15Z", qSubUpd[0].PendingAck)
	// Test RemoveProjectTopics
	store.RemoveProjectTopics("argo_uuid")
	resTop, _, _, _ := store.QueryTopics("argo_uuid", "", "", "", 0)
	suite.Equal(0, len(resTop))
	store.RemoveProjectSubs("argo_uuid")
	resSub, _, _, _ := store.QuerySubs("argo_uuid", "", "", "", 0)
	suite.Equal(0, len(resSub))

	// Test RemoveProject
	store.RemoveProject("argo_uuid")
	resProj, err := store.QueryProjects("argo_uuid", "")
	suite.Equal([]QProject{}, resProj)
	suite.Equal(errors.New("not found"), err)

	// Test Insert User
	qRoleAdmin1 := []QProjectRoles{QProjectRoles{"argo_uuid", []string{"admin"}}}
	qRoles := []QProjectRoles{QProjectRoles{"argo_uuid", []string{"admin"}}, QProjectRoles{"argo_uuid2", []string{"admin", "viewer"}}}
	expUsr10 := QUser{UUID: "user_uuid10", Projects: qRoleAdmin1, Name: "newUser1", FirstName: "fname", LastName: "lname", Organization: "org1", Description: "desc1", Token: "A3B94A94V3A", Email: "fake@email.com", ServiceRoles: []string{}, CreatedOn: created, ModifiedOn: modified, CreatedBy: "uuid1"}
	expUsr11 := QUser{UUID: "user_uuid11", Projects: qRoles, Name: "newUser2", Token: "BX312Z34NLQ", Email: "fake@email.com", ServiceRoles: []string{}, CreatedOn: created, ModifiedOn: modified, CreatedBy: "uuid1"}
	store.InsertUser("user_uuid10", qRoleAdmin1, "newUser1", "fname", "lname", "org1", "desc1", "A3B94A94V3A", "fake@email.com", []string{}, created, modified, "uuid1")
	store.InsertUser("user_uuid11", qRoles, "newUser2", "", "", "", "", "BX312Z34NLQ", "fake@email.com", []string{}, created, modified, "uuid1")
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
	usrUpdated := QUser{UUID: "user_uuid11", Projects: qRoles, Name: "updated_name", Token: "BX312Z34NLQ", Email: "fake@email.com", ServiceRoles: []string{"service_admin"}, CreatedOn: created, ModifiedOn: modified, CreatedBy: "uuid1"}
	store.UpdateUser("user_uuid11", nil, "updated_name", "", []string{"service_admin"}, modified)
	usr11, _ = store.QueryUsers("", "user_uuid11", "")
	suite.Equal(usrUpdated, usr11[0])

	// test append project to user
	errUserPrj := store.AppendToUserProjects("uuid1", "p1_uuid", "r1", "r2")
	usr, _ := store.QueryUsers("", "uuid1", "")
	suite.Equal([]QProjectRoles{
		{
			ProjectUUID: "argo_uuid",
			Roles:       []string{"consumer", "publisher"},
		},
		{
			ProjectUUID: "p1_uuid",
			Roles:       []string{"r1", "r2"},
		},
	}, usr[0].Projects)
	suite.Nil(errUserPrj)

	// Test Remove User
	store.RemoveUser("user_uuid11")
	usr11, err = store.QueryUsers("", "user_uuid11", "")
	suite.Equal(errors.New("not found"), err)

	usrGet, _ := store.GetUserFromToken("A3B94A94V3A")
	suite.Equal(usr10[0], usrGet)

	// test paginated query users
	store2 := NewMockStore("", "")

	// return all users in one page
	qUsers1, ts1, pg1, _ := store2.PaginatedQueryUsers("", 0, "")

	// return a page with the first 2
	qUsers2, ts2, pg2, _ := store2.PaginatedQueryUsers("", 2, "")

	// empty store
	store3 := NewMockStore("", "")
	store3.UserList = []QUser{}
	qUsers3, ts3, pg3, _ := store3.PaginatedQueryUsers("", 0, "")

	// use page token "5" to grab another 2 results
	qUsers4, ts4, pg4, _ := store2.PaginatedQueryUsers("4", 2, "")

	suite.Equal(store2.UserList, qUsers1)
	suite.Equal("", pg1)
	suite.Equal(int32(9), ts1)

	suite.Equal(8, qUsers2[0].ID)
	suite.Equal(7, qUsers2[1].ID)
	suite.Equal("6", pg2)
	suite.Equal(int32(9), ts2)

	suite.Equal(0, len(qUsers3))
	suite.Equal("", pg3)
	suite.Equal(int32(0), ts3)

	suite.Equal(4, qUsers4[0].ID)
	suite.Equal(3, qUsers4[1].ID)
	suite.Equal("2", pg4)
	suite.Equal(int32(9), ts4)

	// test update topic latest publish time
	e1ulp := store2.UpdateTopicLatestPublish("argo_uuid", "topic1", time.Date(2019, 8, 8, 0, 0, 0, 0, time.Local))
	suite.Nil(e1ulp)
	tpc, _, _, _ := store2.QueryTopics("argo_uuid", "", "topic1", "", 0)
	suite.Equal(time.Date(2019, 8, 8, 0, 0, 0, 0, time.Local), tpc[0].LatestPublish)

	// test update topic publishing rate
	e1upr := store2.UpdateTopicPublishRate("argo_uuid", "topic1", 8.44)
	suite.Nil(e1upr)
	tpc2, _, _, _ := store2.QueryTopics("argo_uuid", "", "topic1", "", 0)
	suite.Equal(8.44, tpc2[0].PublishRate)

	// test update topic latest publish time
	scre1 := store2.UpdateSubLatestConsume("argo_uuid", "sub1", time.Date(2019, 8, 8, 0, 0, 0, 0, time.Local))
	suite.Nil(scre1)
	spc, _, _, _ := store2.QuerySubs("argo_uuid", "", "sub1", "", 0)
	suite.Equal(time.Date(2019, 8, 8, 0, 0, 0, 0, time.Local), spc[0].LatestConsume)

	// test update topic publishing rate
	scre2 := store2.UpdateSubConsumeRate("argo_uuid", "sub1", 8.44)
	suite.Nil(scre2)
	spc2, _, _, _ := store2.QuerySubs("argo_uuid", "", "sub1", "", 0)
	suite.Equal(8.44, spc2[0].ConsumeRate)

	// test QueryTotalMessagesPerProject
	expectedQpmc := []QProjectMessageCount{
		{ProjectUUID: "argo_uuid", NumberOfMessages: 30, AverageDailyMessages: 10},
	}
	qpmc, qpmcerr1 := store2.QueryTotalMessagesPerProject([]string{"argo_uuid"}, time.Date(2018, 10, 1, 0, 0, 0, 0, time.UTC), time.Date(2018, 10, 4, 0, 0, 0, 0, time.UTC))
	suite.Equal(expectedQpmc, qpmc)
	suite.Nil(qpmcerr1)

	// test InsertSchema
	eis := store.InsertSchema("argo_uuid", "uuid1", "s1-insert", "json", "raw")
	qs1, _ := store.QuerySchemas("argo_uuid", "uuid1", "s1-insert")
	suite.Equal(QSchema{
		ProjectUUID: "argo_uuid",
		UUID:        "uuid1",
		Name:        "s1-insert",
		Type:        "json",
		RawSchema:   "raw",
	}, qs1[0])
	suite.Nil(eis)

	// test QuerySchemas
	s := "eyJwcm9wZXJ0aWVzIjp7ImFkZHJlc3MiOnsidHlwZSI6InN0cmluZyJ9LCJlbWFpbCI6eyJ0eXBlIjoic3RyaW5nIn0sIm5hbWUiOnsidHlwZSI6InN0cmluZyJ9LCJ0ZWxlcGhvbmUiOnsidHlwZSI6InN0cmluZyJ9fSwicmVxdWlyZWQiOlsibmFtZSIsImVtYWlsIl0sInR5cGUiOiJvYmplY3QifQ=="
	avros := "eyJmaWVsZHMiOlt7Im5hbWUiOiJ1c2VybmFtZSIsInR5cGUiOiJzdHJpbmcifSx7Im5hbWUiOiJwaG9uZSIsInR5cGUiOiJpbnQifV0sIm5hbWUiOiJVc2VyIiwibmFtZXNwYWNlIjoidXNlci5hdnJvIiwidHlwZSI6InJlY29yZCJ9"
	expectedSchemas := []QSchema{
		{UUID: "schema_uuid_1", ProjectUUID: "argo_uuid", Type: "json", Name: "schema-1", RawSchema: s},
		{UUID: "schema_uuid_2", ProjectUUID: "argo_uuid", Type: "json", Name: "schema-2", RawSchema: s},
		{UUID: "schema_uuid_3", ProjectUUID: "argo_uuid", Type: "avro", Name: "schema-3", RawSchema: avros},
	}
	qqs1, _ := store2.QuerySchemas("argo_uuid", "", "")
	qqs2, _ := store2.QuerySchemas("argo_uuid", "schema_uuid_1", "")
	qqs3, _ := store2.QuerySchemas("argo_uuid", "schema_uuid_1", "schema-1")
	suite.Equal(expectedSchemas, qqs1)
	suite.Equal(expectedSchemas[0], qqs2[0])
	suite.Equal(expectedSchemas[0], qqs3[0])

	// test update schema
	store2.UpdateSchema("schema_uuid_1", "new-name", "new-type", "new-raw-schema")
	eus := QSchema{UUID: "schema_uuid_1", ProjectUUID: "argo_uuid", Type: "new-type", Name: "new-name", RawSchema: "new-raw-schema"}
	qus, _ := store2.QuerySchemas("argo_uuid", "schema_uuid_1", "")
	suite.Equal(eus, qus[0])

	//test delete schema
	store4 := NewMockStore("", "")

	ed := store4.DeleteSchema("schema_uuid_1")
	expd, _ := store4.QuerySchemas("argo_uuid", "schema_uuid_1", "")
	// check that topic-1 no longer has any schema_uuid associated with it
	qtd, _, _, _ := store4.QueryTopics("argo_uuid", "", "topic2", "", 1)
	suite.Equal("", qtd[0].SchemaUUID)
	suite.Equal([]QSchema{}, expd)
	suite.Nil(ed)

	// test user registration
	store.RegisterUser("ruuid1", "n1", "f1", "l1", "e1", "o1", "d1", "time", "atkn", "pending")
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

	ur1, _ := store.QueryRegistrations("atkn", "pending")
	suite.Equal(expur1, ur1)

	expur2 := []QUserRegistration{{
		UUID:            "ur-uuid1",
		Name:            "urname",
		FirstName:       "urfname",
		LastName:        "urlname",
		Organization:    "urorg",
		Description:     "urdesc",
		Email:           "uremail",
		ActivationToken: "uratkn-1",
		Status:          "accepted",
		RegisteredAt:    "2019-05-12T22:26:58Z",
		ModifiedBy:      "uuid1",
		ModifiedAt:      "2020-05-17T22:26:58Z",
	}}
	store.UpdateRegistration("uratkn-1", "accepted", "uuid1", "2020-05-17T22:26:58Z")
	ur2, _ := store.QueryRegistrations("uratkn-1", "accepted")
	suite.Equal(expur2, ur2)

}

func TestStoresTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}
