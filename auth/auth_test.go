package auth

import (
	"testing"

	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/stretchr/testify/suite"
	"github.com/ARGOeu/argo-messaging/stores"
)

type AuthTestSuite struct {
	suite.Suite
}

func (suite *AuthTestSuite) TestMockStore() {

	store := stores.NewMockStore("mockhost", "mockbase")
	suite.Equal([]string{"admin", "member"}, Authenticate("ARGO", "S3CR3T", store))
	suite.Equal([]string{}, Authenticate("ARGO", "falseSecret", store))

	suite.Equal(true, Authorize("topics:list_all", []string{"admin"}, store))
	suite.Equal(true, Authorize("topics:list_all", []string{"admin", "reader"}, store))
	suite.Equal(true, Authorize("topics:list_all", []string{"admin", "foo"}, store))
	suite.Equal(false, Authorize("topics:list_all", []string{"foo"}, store))
	suite.Equal(false, Authorize("topics:publish", []string{"reader"}, store))
	suite.Equal(true, Authorize("topics:list_all", []string{"publisher"}, store))
	suite.Equal(true, Authorize("topics:publish", []string{"publisher"}, store))

}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
