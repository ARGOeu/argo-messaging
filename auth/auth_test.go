package auth

import (
	"errors"
	"io/ioutil"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

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
	suite.Equal(true, PerResource("argo_uuid", "topics", "topic1", "uuid1", store))
	suite.Equal(true, PerResource("argo_uuid", "topics", "topic2", "uuid1", store))
	suite.Equal(false, PerResource("argo_uuid", "topics", "topic3", "uuid1", store))

	// Check authorization per topic for userB
	suite.Equal(true, PerResource("argo_uuid", "topics", "topic1", "uuid2", store))
	suite.Equal(true, PerResource("argo_uuid", "topics", "topic2", "uuid2", store))
	suite.Equal(false, PerResource("argo_uuid", "topics", "topic3", "uuid2", store))

	// Check authorization per topic for userC
	suite.Equal(false, PerResource("argo_uuid", "topics", "topic1", "uuid3", store))
	suite.Equal(false, PerResource("argo_uuid", "topics", "topic2", "uuid3", store))
	suite.Equal(true, PerResource("argo_uuid", "topics", "topic3", "uuid3", store))

	// Check authorization per topic for userD
	suite.Equal(false, PerResource("argo_uuid", "topics", "topic1", "uuid4", store))
	suite.Equal(true, PerResource("argo_uuid", "topics", "topic2", "uuid4", store))
	suite.Equal(false, PerResource("argo_uuid", "topics", "topic3", "uuid4", store))

	// Check user authorization per subscription
	//
	// sub1: userA, userB
	// sub2: userA, userC
	// sub3: userA, userB, userD
	// sub4: userB, userD

	// Check authorization per subscription for userA
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub1", "uuid1", store))
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub2", "uuid1", store))
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub3", "uuid1", store))
	suite.Equal(false, PerResource("argo_uuid", "subscriptions", "sub4", "uuid1", store))

	// Check authorization per subscription for userB
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub1", "uuid2", store))
	suite.Equal(false, PerResource("argo_uuid", "subscriptions", "sub2", "uuid2", store))
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub3", "uuid2", store))
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub4", "uuid2", store))
	// Check authorization per subscription for userC
	suite.Equal(false, PerResource("argo_uuid", "subscriptions", "sub1", "uuid3", store))
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub2", "uuid3", store))
	suite.Equal(false, PerResource("argo_uuid", "subscriptions", "sub3", "uuid3", store))
	suite.Equal(false, PerResource("argo_uuid", "subscriptions", "sub4", "uuid3", store))
	// Check authorization per subscription for userD
	suite.Equal(false, PerResource("argo_uuid", "subscriptions", "sub1", "uuid4", store))
	suite.Equal(false, PerResource("argo_uuid", "subscriptions", "sub2", "uuid4", store))
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub3", "uuid4", store))
	suite.Equal(true, PerResource("argo_uuid", "subscriptions", "sub4", "uuid4", store))

	suite.Equal(true, IsConsumer([]string{"consumer"}))
	suite.Equal(true, IsConsumer([]string{"consumer", "publisher"}))
	suite.Equal(false, IsConsumer([]string{"publisher"}))

	suite.Equal(false, IsPublisher([]string{"consumer"}))
	suite.Equal(true, IsPublisher([]string{"consumer", "publisher"}))
	suite.Equal(true, IsPublisher([]string{"publisher"}))

	suite.Equal(true, IsProjectAdmin([]string{"project_admin"}))
	suite.Equal(true, IsProjectAdmin([]string{"project_admin", "publisher"}))
	suite.Equal(false, IsProjectAdmin([]string{"publisher"}))

	suite.Equal(true, IsServiceAdmin([]string{"service_admin"}))
	suite.Equal(true, IsServiceAdmin([]string{"service_admin", "publisher"}))
	suite.Equal(false, IsServiceAdmin([]string{"publisher"}))

	suite.Equal(true, IsPushWorker([]string{"push_worker"}))
	suite.Equal(true, IsPushWorker([]string{"push_worker", "publisher"}))
	suite.Equal(false, IsPushWorker([]string{"publisher"}))

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
         "first_name": "FirstA",
         "last_name": "LastA",
         "organization": "OrgA",
         "description": "DescA",
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

	users, _ := FindUsers("argo_uuid", "", "", true, store)
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
   "first_name": "firstdoe",
   "last_name": "lastdoe",
   "organization": "orgdoe",
   "description": "descdoe",
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
	CreateUser("uuid12", "johndoe", "firstdoe", "lastdoe", "orgdoe", "descdoe", []ProjectRoles{ProjectRoles{Project: "ARGO", Roles: []string{"consumer"}}}, "johndoe@fake.email.foo", "TOK3N", []string{"service_admin"}, tm, "", store)
	usrs, _ := FindUsers("", "uuid12", "", true, store)
	usrJSON, _ := usrs.List[0].ExportJSON()
	suite.Equal(expUsrJSON, usrJSON)

	// Test Create with empty project list
	CreateUser("uuid13", "empty-proj", "", "", "", "", []ProjectRoles{{Project: "", Roles: []string{"consumer"}}}, "TOK3N", "johndoe@fake.email.foo", []string{"service_admin"}, tm, "", store)
	usrs2, _ := FindUsers("", "uuid13", "", true, store)
	expusrs2 := Users{List: []User{{UUID: "uuid13", Projects: []ProjectRoles{}, Name: "empty-proj", Token: "TOK3N", Email: "johndoe@fake.email.foo", ServiceRoles: []string{"service_admin"}, CreatedOn: "2009-11-10T23:00:00Z", ModifiedOn: "2009-11-10T23:00:00Z", CreatedBy: ""}}}
	suite.Equal(expusrs2, usrs2)

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
   "first_name": "firstdoe",
   "last_name": "lastdoe",
   "organization": "orgdoe",
   "description": "descdoe",
   "token": "johndoe@fake.email.foo",
   "email": "TOK3N",
   "service_roles": [
      "consumer",
      "producer"
   ],
   "created_on": "2009-11-10T23:00:00Z",
   "modified_on": "2009-11-10T23:00:00Z"
}`
	UpdateUser("uuid12", "johnny_doe", nil, "", []string{"consumer", "producer"}, tm, false, store)
	usrUpd, _ := FindUsers("", "uuid12", "", true, store)
	usrUpdJSON, _ := usrUpd.List[0].ExportJSON()
	suite.Equal(expUpdate, usrUpdJSON)

	// reflect obj true
	usrUpd2, _ := UpdateUser("uuid12", "johnny_doe", nil, "", []string{"consumer", "producer"}, tm, true, store)
	usrUpdJSON2, _ := usrUpd2.ExportJSON()
	suite.Equal(expUpdate, usrUpdJSON2)

	// Test update with empty project
	UpdateUser("uuid13", "empty-proj", []ProjectRoles{{Project: "", Roles: []string{"consumer"}}}, "johndoe@fake.email.foo", []string{"service_admin"}, tm, false, store)
	usrs2, _ = FindUsers("", "uuid13", "", true, store)
	expusrs2 = Users{List: []User{{UUID: "uuid13", Projects: []ProjectRoles{}, Name: "empty-proj", Token: "TOK3N", Email: "johndoe@fake.email.foo", ServiceRoles: []string{"service_admin"}, CreatedOn: "2009-11-10T23:00:00Z", ModifiedOn: "2009-11-10T23:00:00Z", CreatedBy: ""}}}
	suite.Equal(expusrs2, usrs2)

	RemoveUser("uuid12", store)
	_, err = FindUsers("", "uuid12", "", true, store)
	suite.Equal(errors.New("not found"), err)

	store2 := stores.NewMockStore("", "")

	created := "2009-11-10T23:00:00Z"
	modified := "2009-11-10T23:00:00Z"

	var qUsers1 []User
	qUsers1 = append(qUsers1, User{"uuid8", []ProjectRoles{{"ARGO2", []string{"consumer", "publisher"}, []string{}, []string{}}}, "UserZ", "", "", "", "", "S3CR3T1", "foo-email", []string{}, created, modified, ""})
	qUsers1 = append(qUsers1, User{"uuid7", []ProjectRoles{}, "push_worker_0", "", "", "", "", "push_token", "foo-email", []string{"push_worker"}, created, modified, ""})
	qUsers1 = append(qUsers1, User{"same_uuid", []ProjectRoles{{"ARGO", []string{"publisher", "consumer"}, []string{}, []string{}}}, "UserSame2", "", "", "", "", "S3CR3T42", "foo-email", []string{}, created, modified, "UserA"})
	qUsers1 = append(qUsers1, User{"same_uuid", []ProjectRoles{{"ARGO", []string{"publisher", "consumer"}, []string{}, []string{}}}, "UserSame1", "", "", "", "", "S3CR3T41", "foo-email", []string{}, created, modified, "UserA"})
	qUsers1 = append(qUsers1, User{"uuid4", []ProjectRoles{{"ARGO", []string{"publisher", "consumer"}, []string{"topic2"}, []string{"sub3", "sub4"}}}, "UserZ", "", "", "", "", "S3CR3T4", "foo-email", []string{}, created, modified, "UserA"})
	qUsers1 = append(qUsers1, User{"uuid3", []ProjectRoles{{"ARGO", []string{"publisher", "consumer"}, []string{"topic3"}, []string{"sub2"}}}, "UserX", "", "", "", "", "S3CR3T3", "foo-email", []string{}, created, modified, "UserA"})
	qUsers1 = append(qUsers1, User{"uuid2", []ProjectRoles{{"ARGO", []string{"consumer", "publisher"}, []string{"topic1", "topic2"}, []string{"sub1", "sub3", "sub4"}}}, "UserB", "", "", "", "", "S3CR3T2", "foo-email", []string{}, created, modified, "UserA"})
	qUsers1 = append(qUsers1, User{"uuid1", []ProjectRoles{{"ARGO", []string{"consumer", "publisher"}, []string{"topic1", "topic2"}, []string{"sub1", "sub2", "sub3"}}}, "UserA", "FirstA", "LastA", "OrgA", "DescA", "S3CR3T1", "foo-email", []string{}, created, modified, ""})
	qUsers1 = append(qUsers1, User{"uuid0", []ProjectRoles{{"ARGO", []string{"consumer", "publisher"}, []string{}, []string{}}}, "Test", "", "", "", "", "S3CR3T", "Test@test.com", []string{}, created, modified, ""})
	// return all users
	pu1, e1 := PaginatedFindUsers("", 0, "", true, store2)

	var qUsers2 []User
	qUsers2 = append(qUsers2, User{"uuid8", []ProjectRoles{{"ARGO2", []string{"consumer", "publisher"}, []string{}, []string{}}}, "UserZ", "", "", "", "", "S3CR3T1", "foo-email", []string{}, created, modified, ""})
	qUsers2 = append(qUsers2, User{"uuid7", []ProjectRoles{}, "push_worker_0", "", "", "", "", "push_token", "foo-email", []string{"push_worker"}, created, modified, ""})
	qUsers2 = append(qUsers2, User{"same_uuid", []ProjectRoles{{"ARGO", []string{"publisher", "consumer"}, []string{}, []string{}}}, "UserSame2", "", "", "", "", "S3CR3T42", "foo-email", []string{}, created, modified, "UserA"})

	// return the first page with 2 users
	pu2, e2 := PaginatedFindUsers("", 3, "", true, store2)

	var qUsers3 []User
	qUsers3 = append(qUsers3, User{"uuid4", []ProjectRoles{{"ARGO", []string{"publisher", "consumer"}, []string{"topic2"}, []string{"sub3", "sub4"}}}, "UserZ", "", "", "", "", "S3CR3T4", "foo-email", []string{}, created, modified, "UserA"})
	qUsers3 = append(qUsers3, User{"uuid3", []ProjectRoles{{"ARGO", []string{"publisher", "consumer"}, []string{"topic3"}, []string{"sub2"}}}, "UserX", "", "", "", "", "S3CR3T3", "foo-email", []string{}, created, modified, "UserA"})
	// return the next 2 users
	pu3, e3 := PaginatedFindUsers("NA==", 2, "", true, store2)

	// empty collection
	store3 := stores.NewMockStore("", "")
	store3.UserList = []stores.QUser{}
	pu4, e4 := PaginatedFindUsers("", 0, "", true, store3)

	// invalid id
	_, e5 := PaginatedFindUsers("invalid", 0, "", true, store2)

	// check user list by project
	var qUsersB []User
	qUsersB = append(qUsersB, User{"uuid8", []ProjectRoles{{"ARGO2", []string{"consumer", "publisher"}, []string{}, []string{}}}, "UserZ", "", "", "", "", "S3CR3T1", "foo-email", []string{}, created, modified, ""})

	// check user list by project and with unprivileged mode (token redacted)
	var qUsersC []User
	qUsersC = append(qUsersC, User{"uuid8", []ProjectRoles{{"ARGO2", []string{"consumer", "publisher"}, []string{}, []string{}}}, "UserZ", "", "", "", "", "", "foo-email", []string{}, created, modified, ""})

	puC, e1 := PaginatedFindUsers("", 1, "argo_uuid2", false, store2)
	suite.Equal(qUsersC, puC.Users)
	suite.Equal(int32(1), puC.TotalSize)
	suite.Equal("", puC.NextPageToken)

	suite.Equal(qUsers1, pu1.Users)
	suite.Equal(int32(9), pu1.TotalSize)
	suite.Equal("", pu1.NextPageToken)
	suite.Nil(e1)

	suite.Equal(qUsers2, pu2.Users)
	suite.Equal(int32(9), pu2.TotalSize)
	suite.Equal("NQ==", pu2.NextPageToken)
	suite.Nil(e2)

	suite.Equal(qUsers3, pu3.Users)
	suite.Equal(int32(9), pu3.TotalSize)
	suite.Equal("Mg==", pu3.NextPageToken)
	suite.Nil(e3)

	suite.Equal(0, len(pu4.Users))
	suite.Equal(int32(0), pu4.TotalSize)
	suite.Equal("", pu4.NextPageToken)
	suite.Nil(e4)

	suite.Equal("illegal base64 data at input byte 4", e5.Error())
}

func (suite *AuthTestSuite) TestAppendToUserProjects() {

	store := stores.NewMockStore("", "")
	store.ProjectList = append(store.ProjectList, stores.QProject{UUID: "append_uuid", Name: "append_project"})
	store.UserList = append(store.UserList, stores.QUser{UUID: "append_uuid"})

	err1 := AppendToUserProjects("append_uuid", "append_uuid", store, "publisher")
	u, _ := store.QueryUsers("append_uuid", "append_uuid", "")
	suite.Equal([]stores.QProjectRoles{
		{
			ProjectUUID: "append_uuid",
			Roles:       []string{"publisher"},
		},
	}, u[0].Projects)
	suite.Nil(err1)

	// invalid project
	err2 := AppendToUserProjects("", "unknown", store)
	suite.Equal("invalid project unknown", err2.Error())

	// invalid role
	err3 := AppendToUserProjects("append_uuid", "append_uuid", store, "r1")
	suite.Equal("invalid role r1", err3.Error())

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
      "UserZ",
      "push_worker_0"
   ]
}`

	expJSON05 := `{
   "authorized_users": []
}`

	expJSON01deleted := `{
   "authorized_users": [
      "UserB"
   ]
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

	// make sure that the acl doesn't contain empty "" in the spot of the deleted user
	store.RemoveUser("uuid1")
	dACL, _ := GetACL("argo_uuid", "subscriptions", "sub1", store)
	outJSONd, _ := dACL.ExportJSON()
	suite.Equal(expJSON01deleted, outJSONd)

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

	expJSON01deleted := `{
   "authorized_users": [
      "UserB"
   ]
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

	// make sure that the acl doesn't contain empty "" in the spot of the deleted user
	store.RemoveUser("uuid1")
	dACL, _ := GetACL("argo_uuid", "topics", "topic1", store)
	outJSONd, _ := dACL.ExportJSON()
	suite.Equal(expJSON01deleted, outJSONd)
}

func (suite *AuthTestSuite) TestModACL() {

	store := stores.NewMockStore("", "")

	e1 := ModACL("argo_uuid", "topics", "topic1", []string{"UserX", "UserZ"}, store)
	suite.Nil(e1)

	tACL1, _ := store.TopicsACL["topic1"]
	suite.Equal([]string{"uuid3", "uuid4"}, tACL1.ACL)

	e2 := ModACL("argo_uuid", "subscriptions", "sub1", []string{"UserX", "UserZ"}, store)
	suite.Nil(e2)

	sACL1, _ := store.SubsACL["sub1"]
	suite.Equal([]string{"uuid3", "uuid4"}, sACL1.ACL)

	e3 := ModACL("argo_uuid", "mistype", "sub1", []string{"UserX", "UserZ"}, store)
	suite.Equal("wrong resource type", e3.Error())
}

func (suite *AuthTestSuite) TestAppendToACL() {

	store := stores.NewMockStore("", "")

	e1 := AppendToACL("argo_uuid", "topics", "topic1", []string{"UserX", "UserZ", "UserZ"}, store)
	suite.Nil(e1)

	tACL1, _ := store.TopicsACL["topic1"]
	suite.Equal([]string{"uuid1", "uuid2", "uuid3", "uuid4"}, tACL1.ACL)

	e2 := AppendToACL("argo_uuid", "subscriptions", "sub1", []string{"UserX", "UserZ", "UserZ"}, store)
	suite.Nil(e2)

	sACL1, _ := store.SubsACL["sub1"]
	suite.Equal([]string{"uuid1", "uuid2", "uuid3", "uuid4"}, sACL1.ACL)

	e3 := AppendToACL("argo_uuid", "mistype", "sub1", []string{"UserX", "UserZ"}, store)
	suite.Equal("wrong resource type", e3.Error())
}

func (suite *AuthTestSuite) TestRemoveFromACL() {

	store := stores.NewMockStore("", "")

	e1 := RemoveFromACL("argo_uuid", "topics", "topic1", []string{"UserA", "UserK"}, store)
	suite.Nil(e1)

	tACL1, _ := store.TopicsACL["topic1"]
	suite.Equal([]string{"uuid2"}, tACL1.ACL)

	e2 := RemoveFromACL("argo_uuid", "subscriptions", "sub1", []string{"UserA", "UserK"}, store)
	suite.Nil(e2)

	sACL1, _ := store.SubsACL["sub1"]
	suite.Equal([]string{"uuid2"}, sACL1.ACL)

	e3 := RemoveFromACL("argo_uuid", "mistype", "sub1", []string{"UserX", "UserZ"}, store)
	suite.Equal("wrong resource type", e3.Error())
}

func (suite *AuthTestSuite) TestGetPushWorkerToken() {

	store := stores.NewMockStore("", "")

	// normal case of push enabled true and correct push worker token
	u1, err1 := GetPushWorker("push_token", store)
	suite.Equal(User{"uuid7", []ProjectRoles{}, "push_worker_0", "", "", "", "", "push_token", "foo-email", []string{"push_worker"}, "2009-11-10T23:00:00Z", "2009-11-10T23:00:00Z", ""}, u1)
	suite.Nil(err1)

	//  incorrect push worker token
	u4, err4 := GetPushWorker("missing", store)
	suite.Equal(User{}, u4)
	suite.Equal("push_500", err4.Error())
}

func (suite *AuthTestSuite) TestRegisterUser() {

	store := stores.NewMockStore("", "")

	ur, err := RegisterUser("ruuid1", "n1", "f1", "l1", "e1", "o1", "d1", "time", "atkn", PendingRegistrationStatus, store)
	suite.Nil(err)
	suite.Equal(UserRegistration{
		UUID:            "ruuid1",
		Name:            "n1",
		FirstName:       "f1",
		LastName:        "l1",
		Email:           "e1",
		Organization:    "o1",
		Description:     "d1",
		RegisteredAt:    "time",
		ActivationToken: "atkn",
		Status:          PendingRegistrationStatus,
	}, ur)

}

func (suite *AuthTestSuite) TestFindUserRegistration() {

	store := stores.NewMockStore("", "")

	ur1, e1 := FindUserRegistration("ur-uuid1", "pending", store)
	expur1 := UserRegistration{
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
		ModifiedBy:      "UserA",
		ModifiedAt:      "2020-05-15T22:26:58Z",
	}
	suite.Nil(e1)
	suite.Equal(expur1, ur1)

	// not found
	_, e2 := FindUserRegistration("unknown", "pending", store)
	suite.Equal(errors.New("not found"), e2)
}

func (suite *AuthTestSuite) TestUpdateUserRegistration() {

	store := stores.NewMockStore("", "")
	m := time.Date(2020, 8, 5, 11, 33, 45, 0, time.UTC)
	e1 := UpdateUserRegistration("ur-uuid1", AcceptedRegistrationStatus, "uuid1", m, store)
	ur1, _ := FindUserRegistration("ur-uuid1", "accepted", store)
	expur1 := UserRegistration{
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
		ModifiedBy:      "UserA",
		ModifiedAt:      "2020-08-05T11:33:45Z",
	}
	suite.Nil(e1)
	suite.Equal(expur1, ur1)
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
