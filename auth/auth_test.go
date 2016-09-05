package auth

import (
	"testing"

	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/stretchr/testify/suite"
)

type AuthTestSuite struct {
	suite.Suite
}

func (suite *AuthTestSuite) TestAuth() {

	store := stores.NewMockStore("mockhost", "mockbase")
	authen01, user01 := Authenticate("ARGO", "S3CR3T", store)
	authen02, user02 := Authenticate("ARGO", "falseSECRET", store)
	suite.Equal("userA", user01)
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
	suite.Equal(true, PerResource("ARGO", "topic", "topic1", "userA", store))
	suite.Equal(true, PerResource("ARGO", "topic", "topic2", "userA", store))
	suite.Equal(false, PerResource("ARGO", "topic", "topic3", "userA", store))

	// Check authorization per topic for userB
	suite.Equal(true, PerResource("ARGO", "topic", "topic1", "userB", store))
	suite.Equal(true, PerResource("ARGO", "topic", "topic2", "userB", store))
	suite.Equal(false, PerResource("ARGO", "topic", "topic3", "userB", store))

	// Check authorization per topic for userC
	suite.Equal(false, PerResource("ARGO", "topic", "topic1", "userC", store))
	suite.Equal(false, PerResource("ARGO", "topic", "topic2", "userC", store))
	suite.Equal(true, PerResource("ARGO", "topic", "topic3", "userC", store))

	// Check authorization per topic for userD
	suite.Equal(false, PerResource("ARGO", "topic", "topic1", "userD", store))
	suite.Equal(true, PerResource("ARGO", "topic", "topic2", "userD", store))
	suite.Equal(false, PerResource("ARGO", "topic", "topic3", "userD", store))

	// Check user authorization per subscription
	//
	// sub1: userA, userB
	// sub2: userA, userC
	// sub3: userA, userB, userD
	// sub4: userB, userD

	// Check authorization per subscription for userA
	suite.Equal(true, PerResource("ARGO", "subscription", "sub1", "userA", store))
	suite.Equal(true, PerResource("ARGO", "subscription", "sub2", "userA", store))
	suite.Equal(true, PerResource("ARGO", "subscription", "sub3", "userA", store))
	suite.Equal(false, PerResource("ARGO", "subscription", "sub4", "userA", store))

	// Check authorization per subscription for userB
	suite.Equal(true, PerResource("ARGO", "subscription", "sub1", "userB", store))
	suite.Equal(false, PerResource("ARGO", "subscription", "sub2", "userB", store))
	suite.Equal(true, PerResource("ARGO", "subscription", "sub3", "userB", store))
	suite.Equal(true, PerResource("ARGO", "subscription", "sub4", "userB", store))
	// Check authorization per subscription for userC
	suite.Equal(false, PerResource("ARGO", "subscription", "sub1", "userC", store))
	suite.Equal(true, PerResource("ARGO", "subscription", "sub2", "userC", store))
	suite.Equal(false, PerResource("ARGO", "subscription", "sub3", "userC", store))
	suite.Equal(false, PerResource("ARGO", "subscription", "sub4", "userC", store))
	// Check authorization per subscription for userD
	suite.Equal(false, PerResource("ARGO", "subscription", "sub1", "userD", store))
	suite.Equal(false, PerResource("ARGO", "subscription", "sub2", "userD", store))
	suite.Equal(true, PerResource("ARGO", "subscription", "sub3", "userD", store))
	suite.Equal(true, PerResource("ARGO", "subscription", "sub4", "userD", store))

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

}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
