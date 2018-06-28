package auth

import (
	"errors"
	"io/ioutil"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/stretchr/testify/suite"
)

type AuthTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *AuthTestSuite) SetupTest() {
	suite.cfgStr = `{
		"broker_host":"localhost:9092",
		"store_host":"localhost",
		"store_db":"argo_msg"
	}`
	log.SetOutput(ioutil.Discard)
}

func (suite *AuthTestSuite) TestAuth() {

	store := stores.NewMockStore("mockhost", "mockbase")
	authen01, user01 := Authenticate("argo_uuid", "S3CR3T1", store)
	authen02, user02 := Authenticate("argo_uuid", "falseSECRET", store)
	suite.Equal("UserA", user01)
	suite.Equal("", user02)
	suite.Equal([]string{"consumer", "publisher"}, authen01)
	suite.Equal([]string{}, authen02)

	suite.Equal(true, Authorize("topics:list_all", []string{"admin"}, store))
	suite.Equal(true, Authorize("topics:list_all", []string{"admin", "reader"}, store))
	suite.Equal(true, Authorize("topics:list_all", []string{"admin", "foo"}, store))
	suite.Equal(false, Authorize("topics:list_all", []string{"foo"}, store))
	suite.Equal(false, Authorize("topics:publish", []string{"reader"}, store))
	suite.Equal(true, Authorize("topics:list_all", []string{"admin"}, store))
	suite.Equal(true, Authorize("topics:list_all", []string{"publisher"}, store))
	suite.Equal(true, Authorize("topics:publish", []string{"publisher"}, store))

	// Check user authorization per topic
	//
	// topic1: userA, userB
	// topic2: userA, userB, userD
	// topic3: userC

	// Check authorization per topic for userA
	suite.Equal(true, PerResource("argo_uuid", "topics", "topic1", "UserA", store))
	suite.Equal(true, PerResource("argo_uuid", "topics", "topic2", "UserA", store))
	suite.Equal(false, PerResource("argo_uuid", "topics", "topic3", "UserA", store))

	// Check authorization per topic for userB
	suite.Equal(true, PerResource("argo_uuid", "topics", "topic1", "UserB", store))
	suite.Equal(true, PerResource("argo_uuid", "topics", "topic2", "UserB", store))
	suite.Equal(false, PerResource("argo_uuid", "topics", "topic3", "UserB", store))

	// Check authorization per topic for userC
	suite.Equal(false, PerResource("argo_uuid", "topics", "topic1", "UserX", store))
	suite.Equal(false, PerResource("argo_uuid", "topics", "topic2", "UserX", store))
	suite.Equal(true, PerResource("argo_uuid", "topics", "topic3", "UserX", store))

	// Check authorization per topic for userD
	suite.Equal(false, PerResource("argo_uuid", "topics", "topic1", "UserZ", store))
	suite.Equal(true, PerResource("argo_uuid", "topics", "topic2", "UserZ", store))
	suite.Equal(false, PerResource("argo_uuid", "topics", "topic3", "UserZ", store))

	// Check user authorization per subscription
	//
	// sub1: userA, userB
	// sub2: userA, userC
	// sub3: userA, userB, userD
	// sub4: userB, userD

	// Check authorization per subscription for userA
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub1", "UserA", store))
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub2", "UserA", store))
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub3", "UserA", store))
	suite.Equal(false, PerResource("argo_uuid", "subscriptions", "sub4", "UserA", store))

	// Check authorization per subscription for userB
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub1", "UserB", store))
	suite.Equal(false, PerResource("argo_uuid", "subscriptions", "sub2", "UserB", store))
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub3", "UserB", store))
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub4", "UserB", store))
	// Check authorization per subscription for userC
	suite.Equal(false, PerResource("argo_uuid", "subscriptions", "sub1", "UserX", store))
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub2", "UserX", store))
	suite.Equal(false, PerResource("argo_uuid", "subscriptions", "sub3", "UserX", store))
	suite.Equal(false, PerResource("argo_uuid", "subscriptions", "sub4", "UserX", store))
	// Check authorization per subscription for userD
	suite.Equal(false, PerResource("argo_uuid", "subscriptions", "sub1", "UserZ", store))
	suite.Equal(false, PerResource("argo_uuid", "subscriptions", "sub2", "UserZ", store))
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub3", "UserZ", store))
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub4", "UserZ", store))

	suite.Equal(true, IsConsumer([]string{"consumer"}))
	suite.Equal(true, IsConsumer([]string{"consumer", "publisher"}))
	suite.Equal(false, IsConsumer([]string{"publisher"}))

	suite.Equal(false, IsPublisher([]string{"consumer"}))
	suite.Equal(true, IsPublisher([]string{"consumer", "publisher"}))
	suite.Equal(true, IsPublisher([]string{"publisher"}))

	// Check ValidUsers mechanism
	v, err := AreValidUsers("ARGO", []string{"UserA", "foo", "bar"}, store)
	suite.Equal(false, v)
	suite.Equal("User(s): foo, bar do not exist", err.Error())

	// Check ValidUsers mechanism
	v, err = AreValidUsers("ARGO", []string{"UserA", "UserB"}, store)
	suite.Equal(true, v)
	suite.Equal(nil, err)

	// Test Find Method
	expUserList := `{
   "users": [
      {
         "uuid": "uuid0",
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "consumer",
                  "publisher"
               ],
               "topics": [],
               "subscriptions": []
            }
         ],
         "name": "Test",
         "token": "S3CR3T",
         "email": "Test@test.com",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "uuid": "uuid1",
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "consumer",
                  "publisher"
               ],
               "topics": [
                  "topic1",
                  "topic2"
               ],
               "subscriptions": [
                  "sub1",
                  "sub2",
                  "sub3"
               ]
            }
         ],
         "name": "UserA",
         "token": "S3CR3T1",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z"
      },
      {
         "uuid": "uuid2",
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "consumer",
                  "publisher"
               ],
               "topics": [
                  "topic1",
                  "topic2"
               ],
               "subscriptions": [
                  "sub1",
                  "sub3",
                  "sub4"
               ]
            }
         ],
         "name": "UserB",
         "token": "S3CR3T2",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA"
      },
      {
         "uuid": "uuid3",
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "publisher",
                  "consumer"
               ],
               "topics": [
                  "topic3"
               ],
               "subscriptions": [
                  "sub2"
               ]
            }
         ],
         "name": "UserX",
         "token": "S3CR3T3",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA"
      },
      {
         "uuid": "uuid4",
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "publisher",
                  "consumer"
               ],
               "topics": [
                  "topic2"
               ],
               "subscriptions": [
                  "sub3",
                  "sub4"
               ]
            }
         ],
         "name": "UserZ",
         "token": "S3CR3T4",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA"
      },
      {
         "uuid": "same_uuid",
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "publisher",
                  "consumer"
               ],
               "topics": [],
               "subscriptions": []
            }
         ],
         "name": "UserSame1",
         "token": "S3CR3T41",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA"
      },
      {
         "uuid": "same_uuid",
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "publisher",
                  "consumer"
               ],
               "topics": [],
               "subscriptions": []
            }
         ],
         "name": "UserSame2",
         "token": "S3CR3T42",
         "email": "foo-email",
         "service_roles": [],
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA"
      }
   ]
}`

	users, _ := FindUsers("argo_uuid", "", "", store)
	outUserList, _ := users.ExportJSON()
	suite.Equal(expUserList, outUserList)

	expUsrTkJSON := `{
   "uuid": "uuid4",
   "projects": [
      {
         "project": "ARGO",
         "roles": [
            "publisher",
            "consumer"
         ],
         "topics": [
            "topic2"
         ],
         "subscriptions": [
            "sub3",
            "sub4"
         ]
      }
   ],
   "name": "UserZ",
   "token": "S3CR3T4",
   "email": "foo-email",
   "service_roles": [],
   "created_on": "2009-11-10T23:00:00Z",
   "modified_on": "2009-11-10T23:00:00Z",
   "created_by": "UserA"
}`

	// Test GetUserByToken
	userTk, _ := GetUserByToken("S3CR3T4", store)
	usrTkJSON, _ := userTk.ExportJSON()
	suite.Equal(expUsrTkJSON, usrTkJSON)

	suite.Equal(true, ExistsWithName("UserA", store))
	suite.Equal(false, ExistsWithName("userA", store))
	suite.Equal(true, ExistsWithName("UserB", store))
	suite.Equal(true, ExistsWithUUID("uuid1", store))
	suite.Equal(false, ExistsWithUUID("foouuuid", store))
	suite.Equal(true, ExistsWithUUID("uuid2", store))

	suite.Equal("UserA", GetNameByUUID("uuid1", store))
	suite.Equal("UserB", GetNameByUUID("uuid2", store))
	suite.Equal("UserX", GetNameByUUID("uuid3", store))
	suite.Equal("UserZ", GetNameByUUID("uuid4", store))

	suite.Equal("uuid1", GetUUIDByName("UserA", store))
	suite.Equal("uuid2", GetUUIDByName("UserB", store))
	suite.Equal("uuid3", GetUUIDByName("UserX", store))
	suite.Equal("uuid4", GetUUIDByName("UserZ", store))

	// Test GetUserByUUID
	expUsrUUIDJSON := `{
   "uuid": "uuid4",
   "projects": [
      {
         "project": "ARGO",
         "roles": [
            "publisher",
            "consumer"
         ],
         "topics": [
            "topic2"
         ],
         "subscriptions": [
            "sub3",
            "sub4"
         ]
      }
   ],
   "name": "UserZ",
   "token": "S3CR3T4",
   "email": "foo-email",
   "service_roles": [],
   "created_on": "2009-11-10T23:00:00Z",
   "modified_on": "2009-11-10T23:00:00Z",
   "created_by": "UserA"
}`
	// normal use case
	expUsrUUID, expNilErr := GetUserByUUID("uuid4", store)
	usrUUIDJson, _ := expUsrUUID.ExportJSON()

	suite.Equal(usrUUIDJson, expUsrUUIDJSON)
	suite.Nil(expNilErr)

	// different users have the same uuid
	expUsrMultipleUUID, expErrMultipleUUIDS := GetUserByUUID("same_uuid", store)

	suite.Equal("multiple uuids", expErrMultipleUUIDS.Error())
	suite.Equal(User{}, expUsrMultipleUUID)

	// user with given uuid doesn't exist
	expUsrNotFoundUUID, expErrNotFoundUUID := GetUserByUUID("uuid10", store)

	suite.Equal("not found", expErrNotFoundUUID.Error())
	suite.Equal(User{}, expUsrNotFoundUUID)

	// Test TokenGeneration
	tk1, _ := GenToken()
	tk2, _ := GenToken()
	tk3, _ := GenToken()

	suite.Equal(false, tk1 == tk2)
	suite.Equal(false, tk1 == tk3)
	suite.Equal(false, tk2 == tk3)

	expUsrJSON := `{
   "uuid": "uuid12",
   "projects": [
      {
         "project": "ARGO",
         "roles": [
            "consumer"
         ],
         "topics": [],
         "subscriptions": []
      }
   ],
   "name": "johndoe",
   "token": "johndoe@fake.email.foo",
   "email": "TOK3N",
   "service_roles": [
      "service_admin"
   ],
   "created_on": "2009-11-10T23:00:00Z",
   "modified_on": "2009-11-10T23:00:00Z"
}`

	tm := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	// Test Create
	CreateUser("uuid12", "johndoe", []ProjectRoles{ProjectRoles{Project: "ARGO", Roles: []string{"consumer"}}}, "johndoe@fake.email.foo", "TOK3N", []string{"service_admin"}, tm, "", store)
	usrs, _ := FindUsers("", "uuid12", "", store)
	usrJSON, _ := usrs.List[0].ExportJSON()
	suite.Equal(expUsrJSON, usrJSON)

	// Test Update
	expUpdate := `{
   "uuid": "uuid12",
   "projects": [
      {
         "project": "ARGO",
         "roles": [
            "consumer"
         ],
         "topics": [],
         "subscriptions": []
      }
   ],
   "name": "johnny_doe",
   "token": "johndoe@fake.email.foo",
   "email": "TOK3N",
   "service_roles": [
      "consumer",
      "producer"
   ],
   "created_on": "2009-11-10T23:00:00Z",
   "modified_on": "2009-11-10T23:00:00Z"
}`
	UpdateUser("uuid12", "johnny_doe", nil, "", []string{"consumer", "producer"}, tm, store)
	usrUpd, _ := FindUsers("", "uuid12", "", store)
	usrUpdJSON, _ := usrUpd.List[0].ExportJSON()
	suite.Equal(expUpdate, usrUpdJSON)

	RemoveUser("uuid12", store)
	_, err = FindUsers("", "uuid12", "", store)
	suite.Equal(errors.New("not found"), err)

}

func (suite *AuthTestSuite) TestSubACL() {
	expJSON01 := `{
   "authorized_users": [
      "UserA",
      "UserB"
   ]
}`

	expJSON02 := `{
   "authorized_users": [
      "UserA",
      "UserX"
   ]
}`

	expJSON03 := `{
   "authorized_users": [
      "UserZ",
      "UserB",
      "UserA"
   ]
}`

	expJSON04 := `{
   "authorized_users": [
      "UserB",
      "UserZ"
   ]
}`

	expJSON05 := `{
   "authorized_users": []
}`

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)

	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)

	sACL, _ := GetACL("argo_uuid", "subscriptions", "sub1", store)
	outJSON, _ := sACL.ExportJSON()
	suite.Equal(expJSON01, outJSON)

	sACL2, _ := GetACL("argo_uuid", "subscriptions", "sub2", store)
	outJSON2, _ := sACL2.ExportJSON()
	suite.Equal(expJSON02, outJSON2)

	sACL3, _ := GetACL("argo_uuid", "subscriptions", "sub3", store)
	outJSON3, _ := sACL3.ExportJSON()
	suite.Equal(expJSON03, outJSON3)

	sACL4, _ := GetACL("argo_uuid", "subscriptions", "sub4", store)
	outJSON4, _ := sACL4.ExportJSON()
	suite.Equal(expJSON04, outJSON4)

	sACL5 := ACL{}
	outJSON5, _ := sACL5.ExportJSON()
	suite.Equal(expJSON05, outJSON5)

}

func (suite *AuthTestSuite) TestTopicACL() {
	expJSON01 := `{
   "authorized_users": [
      "UserA",
      "UserB"
   ]
}`

	expJSON02 := `{
   "authorized_users": [
      "UserA",
      "UserB",
      "UserZ"
   ]
}`

	expJSON03 := `{
   "authorized_users": [
      "UserX"
   ]
}`

	expJSON04 := `{
   "authorized_users": []
}`

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)

	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)

	tACL, _ := GetACL("argo_uuid", "topics", "topic1", store)
	outJSON, _ := tACL.ExportJSON()
	suite.Equal(expJSON01, outJSON)

	tACL2, _ := GetACL("argo_uuid", "topics", "topic2", store)
	outJSON2, _ := tACL2.ExportJSON()
	suite.Equal(expJSON02, outJSON2)

	tACL3, _ := GetACL("argo_uuid", "topics", "topic3", store)
	outJSON3, _ := tACL3.ExportJSON()
	suite.Equal(expJSON03, outJSON3)

	tACL4 := ACL{}
	outJSON4, _ := tACL4.ExportJSON()
	suite.Equal(expJSON04, outJSON4)

}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
