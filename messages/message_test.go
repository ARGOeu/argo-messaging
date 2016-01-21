package messages

import (
	b64 "encoding/base64"
	"errors"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ConfigTestSuite struct {
	suite.Suite
}

func (suite *ConfigTestSuite) TestNewMessage() {

	testMsg := New("this is a test")
	suite.Equal("this is a test", testMsg.Data)

}

func (suite *ConfigTestSuite) TestAttributes() {

	testMsg := New("this is a test")
	suite.Equal("this is a test", testMsg.Data)

	testMsg.InsertAttribute("bruce", "wayne")
	testMsg.InsertAttribute("clark", "kent")
	expAttr := []Attribute{{Key: "bruce", Value: "wayne"}, {Key: "clark", Value: "kent"}}
	suite.Equal(expAttr, testMsg.Attr)
	// Test GetAttribute
	val1, err1 := testMsg.GetAttribute("clark")
	val2, err2 := testMsg.GetAttribute("bruce")
	val3, err3 := testMsg.GetAttribute("peter")
	suite.Equal("kent", val1)
	suite.Equal(nil, err1)
	suite.Equal("wayne", val2)
	suite.Equal(nil, err2)
	suite.Equal("", val3)
	suite.Equal(errors.New("Attribute doesn't exist"), err3)
	// Test update attribute
	testMsg.UpdateAttribute("bruce", "doe")
	val1, err1 = testMsg.GetAttribute("bruce")
	suite.Equal("doe", val1)
	suite.Equal(nil, err1)
	// Test delete attribute
	err1 = testMsg.RemoveAttribute("bruce")
	suite.Equal(nil, err1)
	val1, err1 = testMsg.GetAttribute("bruce")
	suite.Equal("", val1)
	suite.Equal(errors.New("Attribute doesn't exist"), err1)
}

func (suite *ConfigTestSuite) TestLoadJson() {
	txtJSON := `{
   "messageId": 35,
   "attributes": [
     {
       "key": "tick",
       "value": "tock"
     },
     {
       "key": "flip",
       "value": "flop"
     }
   ],
   "data": "aGVsbG8gd29ybGQh"
 }`

	testMsg, err := LoadJSON(txtJSON)
	suite.Equal(nil, err)
	suite.Equal(int64(35), testMsg.Id)
	expAttr := []Attribute{{Key: "tick", Value: "tock"}, {Key: "flip", Value: "flop"}}
	suite.Equal(expAttr, testMsg.Attr)
	suite.Equal("aGVsbG8gd29ybGQh", testMsg.Data)

}

func (suite *ConfigTestSuite) TestExportJson() {
	expJSON := `{
   "messageId": 0,
   "attributes": [
     {
       "key": "foo",
       "value": "bar"
     },
     {
       "key": "color",
       "value": "blue"
     }
   ],
   "data": "aGVsbG8gd29ybGQh"
 }`

	origData := "hello world!"
	b64Data := b64.StdEncoding.EncodeToString([]byte(origData))
	testMsg := New(b64Data)
	testMsg.InsertAttribute("foo", "bar")
	testMsg.InsertAttribute("color", "blue")
	outJSON, err := testMsg.ExportJSON()
	suite.Equal(nil, err)
	suite.Equal(expJSON, outJSON)
}

func (suite *ConfigTestSuite) TestGetDecodedData() {

	origData := "hello world!"
	b64Data := b64.StdEncoding.EncodeToString([]byte(origData))
	testMsg := New(b64Data)
	suite.Equal("aGVsbG8gd29ybGQh", testMsg.Data)     // expected base64 data
	suite.Equal("hello world!", testMsg.GetDecoded()) // expected decoded data

}

func TestFactorsTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
