package validation

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type ValidationTestSuite struct {
	suite.Suite
}

func (suite *ValidationTestSuite) TestValidHTTPS() {
	suite.Equal(false, IsValidHTTPS("ht"))
	suite.Equal(false, IsValidHTTPS("www.example.com"))
	suite.Equal(false, IsValidHTTPS("https:www.example.com"))
	suite.Equal(false, IsValidHTTPS("http://www.example.com"))
	suite.Equal(true, IsValidHTTPS("https://www.example.com"))

}

func (suite *ValidationTestSuite) TestValidation() {
	// nameValidations
	suite.Equal(true, ValidName("topic101"))
	suite.Equal(true, ValidName("topic_101"))
	suite.Equal(true, ValidName("topic_101_another_thing"))
	suite.Equal(true, ValidName("topic___343_random"))
	suite.Equal(true, ValidName("topic_dc1cc538-1361-4317-a235-0bf383d4a69f"))
	suite.Equal(false, ValidName("topic_dc1cc538.1361-4317-a235-0bf383d4a69f"))
	suite.Equal(false, ValidName("topic.not.valid"))
	suite.Equal(false, ValidName("spaces are not valid"))
	suite.Equal(false, ValidName("topic/A"))
	suite.Equal(false, ValidName("topic/B"))

	// ackID validations
	suite.Equal(true, ValidAckID("ARGO", "sub101", "projects/ARGO/subscriptions/sub101:5"))
	suite.Equal(false, ValidAckID("ARGO", "sub101", "projects/ARGO/subscriptions/sub101:aaa"))
	suite.Equal(false, ValidAckID("ARGO", "sub101", "projects/FARGO/subscriptions/sub101:5"))
	suite.Equal(false, ValidAckID("ARGO", "sub101", "projects/ARGO/subscriptions/subF00:5"))
	suite.Equal(false, ValidAckID("ARGO", "sub101", "falsepath/ARGO/subscriptions/sub101:5"))
	suite.Equal(true, ValidAckID("FOO", "BAR", "projects/FOO/subscriptions/BAR:11155"))
	suite.Equal(false, ValidAckID("FOO", "BAR", "projects/FOO//subscriptions/BAR:11155"))

}

func TestValidationTestSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}
