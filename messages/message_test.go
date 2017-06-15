package messages

import (
	b64 "encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MsgTestSuite struct {
	suite.Suite
}

func (suite *MsgTestSuite) TestNewMessage() {

	testMsg := New("this is a test")
	suite.Equal("this is a test", testMsg.Data)

}

func (suite *MsgTestSuite) TestAttributes() {

	testMsg := New("this is a test")
	suite.Equal("this is a test", testMsg.Data)

	testMsg.InsertAttribute("bruce", "wayne")
	testMsg.InsertAttribute("clark", "kent")
	expAttr := Attributes{"bruce": "wayne", "clark": "kent"}
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

func (suite *MsgTestSuite) TestMsgListBytes() {

	testMsg1 := New("this is a test")
	testMsg2 := New("this is another test")
	testMsg3 := New("this is another test")
	sz1 := testMsg1.Size()
	sz2 := testMsg2.Size()
	sz3 := testMsg3.Size()
	ml := MsgList{Msgs: []Message{}}
	ml.Msgs = append(ml.Msgs, testMsg1)
	ml.Msgs = append(ml.Msgs, testMsg2)
	ml.Msgs = append(ml.Msgs, testMsg3)
	suite.Equal(sz1+sz2+sz3, ml.TotalSize())
}

func (suite *MsgTestSuite) TestLoadMsgJson() {
	txtJSON := `{
   "messageId": "35",
   "attributes": {"tick":"tock","flip":"flop"},
   "data": "aGVsbG8gd29ybGQh"
 }`

	testMsg, err := LoadMsgJSON([]byte(txtJSON))
	suite.Equal(nil, err)
	suite.Equal("35", testMsg.ID)
	expAttr := Attributes{"tick": "tock", "flip": "flop"}
	suite.Equal(expAttr, testMsg.Attr)
	suite.Equal("aGVsbG8gd29ybGQh", testMsg.Data)

}

func (suite *MsgTestSuite) TestExportJson() {
	expJSON := `{
   "messageId": "0",
   "attributes": {
      "color": "blue",
      "foo": "bar"
   },
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

func (suite *MsgTestSuite) TestGetDecodedData() {

	origData := "hello world!"
	b64Data := b64.StdEncoding.EncodeToString([]byte(origData))
	testMsg := New(b64Data)
	suite.Equal("aGVsbG8gd29ybGQh", testMsg.Data)     // expected base64 data
	suite.Equal("hello world!", testMsg.GetDecoded()) // expected decoded data

}

func TestMsgTestSuite(t *testing.T) {
	suite.Run(t, new(MsgTestSuite))
}
