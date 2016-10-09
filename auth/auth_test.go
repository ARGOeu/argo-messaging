package auth

import (
	"io/ioutil"
	"log"
	"testing"

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
	suite.Equal([]string{"admin", "member"}, authen01)
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
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "admin",
                  "member"
               ]
            }
         ],
         "name": "Test",
         "token": "S3CR3T",
         "email": "Test@test.com"
      },
      {
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "admin",
                  "member"
               ]
            }
         ],
         "name": "UserA",
         "token": "S3CR3T1",
         "email": "foo-email"
      },
      {
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "admin",
                  "member"
               ]
            }
         ],
         "name": "UserB",
         "token": "S3CR3T2",
         "email": "foo-email"
      },
      {
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "consumer"
               ]
            }
         ],
         "name": "UserX",
         "token": "S3CR3T3",
         "email": "foo-email"
      },
      {
         "projects": [
            {
               "project": "ARGO",
               "roles": [
                  "producer"
               ]
            }
         ],
         "name": "UserZ",
         "token": "S3CR3T4",
         "email": "foo-email"
      }
   ]
}`
	users, _ := FindUsers("argo_uuid", "", "", store)
	outUserList, _ := users.ExportJSON()
	suite.Equal(expUserList, outUserList)

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

	// Test TokenGeneration
	tk1, _ := GenToken()
	tk2, _ := GenToken()
	tk3, _ := GenToken()

	suite.Equal(false, tk1 == tk2)
	suite.Equal(false, tk1 == tk3)
	suite.Equal(false, tk2 == tk3)

	expUsrJSON := `{
   "projects": [
      {
         "project": "ARGO",
         "roles": [
            "consumer"
         ]
      }
   ],
   "name": "johndoe",
   "token": "johndoe@fake.email.foo",
   "email": "TOK3N",
   "service_roles": [
      "service_admin"
   ]
}`

	// Test Create
	CreateUser("uuid12", "johndoe", []ProjectRoles{ProjectRoles{Project: "ARGO", Roles: []string{"consumer"}}}, "johndoe@fake.email.foo", "TOK3N", []string{"service_admin"}, store)
	usrs, _ := FindUsers("", "uuid12", "", store)
	usrJSON, _ := usrs.List[0].ExportJSON()
	suite.Equal(expUsrJSON, usrJSON)

	// Test Update
	expUpdate := `{
   "projects": [
      {
         "project": "ARGO",
         "roles": [
            "consumer"
         ]
      }
   ],
   "name": "johnny_doe",
   "token": "johndoe@fake.email.foo",
   "email": "TOK3N",
   "service_roles": [
      "consumer",
      "producer"
   ]
}`
	UpdateUser("uuid12", "johnny_doe", nil, "", []string{"consumer", "producer"}, store)
	usrUpd, _ := FindUsers("", "uuid12", "", store)
	usrUpdJSON, _ := usrUpd.List[0].ExportJSON()
	suite.Equal(expUpdate, usrUpdJSON)
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
