package topics

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/stretchr/testify/suite"
)

type TopicTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *TopicTestSuite) SetupTest() {
	suite.cfgStr = `{
		"broker_host":"localhost:9092",
		"store_host":"localhost",
		"store_db":"argo_msg"
	}`
	log.SetOutput(ioutil.Discard)
}

func (suite *TopicTestSuite) TestCreate() {
	myTopic := New("argo_uuid", "ARGO", "test-topic")
	suite.Equal("test-topic", myTopic.Name)
	suite.Equal("argo_uuid", myTopic.ProjectUUID)
	suite.Equal("/projects/ARGO/topics/test-topic", myTopic.FullName)
}

func (suite *TopicTestSuite) TestGetTopicByName() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	myTopics, _ := Find("argo_uuid", "topic1", store)
	expTopic := New("argo_uuid", "ARGO", "topic1")
	suite.Equal(expTopic, myTopics.List[0])
}

func (suite *TopicTestSuite) TestCreateTopicStore() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)

	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)

	tp, err := CreateTopic("argo_uuid", "topic1", store)
	suite.Equal(Topic{}, tp)
	suite.Equal("exists", err.Error())

	tp2, err2 := CreateTopic("argo_uuid", "topicNew", store)
	expTopic := New("argo_uuid", "ARGO", "topicNew")
	suite.Equal(expTopic, tp2)
	suite.Equal(nil, err2)
}

func (suite *TopicTestSuite) TestRemoveTopicStore() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)

	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)

	suite.Equal(true, HasTopic("argo_uuid", "topic1", store))

	suite.Equal("not found", RemoveTopic("argo_uuid", "topicFoo", store).Error())
	suite.Equal(nil, RemoveTopic("argo_uuid", "topic1", store))
	suite.Equal(false, HasTopic("argo_uuid", "topic1", store))
}

func (suite *TopicTestSuite) TestHasProjectTopic() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)

	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)

	suite.Equal(false, HasTopic("argo_uuid", "FOO", store))
	suite.Equal(true, HasTopic("argo_uuid", "topic1", store))
}

func (suite *TopicTestSuite) TestExportJson() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)

	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)

	topics, _ := Find("argo_uuid", "topic1", store)
	outJSON, _ := topics.List[0].ExportJSON()
	expJSON := `{
   "name": "/projects/ARGO/topics/topic1"
}`
	suite.Equal(expJSON, outJSON)

	expJSON2 := `{
   "topics": [
      {
         "name": "/projects/ARGO/topics/topic1"
      },
      {
         "name": "/projects/ARGO/topics/topic2"
      },
      {
         "name": "/projects/ARGO/topics/topic3"
      }
   ]
}`
	topics2, _ := Find("argo_uuid", "", store)
	outJSON2, _ := topics2.ExportJSON()
	suite.Equal(expJSON2, outJSON2)

}

func (suite *TopicTestSuite) TestTopicACL() {
	expJSON01 := `{
   "authorized_users": [
      "userA",
      "userB"
   ]
}`

	expJSON02 := `{
   "authorized_users": [
      "userA",
      "userB",
      "userD"
   ]
}`

	expJSON03 := `{
   "authorized_users": [
      "userC"
   ]
}`

	expJSON04 := `{
   "authorized_users": []
}`

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)

	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)

	tACL, _ := GetTopicACL("ARGO", "topic1", store)
	outJSON, _ := tACL.ExportJSON()
	suite.Equal(expJSON01, outJSON)

	tACL2, _ := GetTopicACL("ARGO", "topic2", store)
	outJSON2, _ := tACL2.ExportJSON()
	suite.Equal(expJSON02, outJSON2)

	tACL3, _ := GetTopicACL("ARGO", "topic3", store)
	outJSON3, _ := tACL3.ExportJSON()
	suite.Equal(expJSON03, outJSON3)

	tACL4 := TopicACL{}
	outJSON4, _ := tACL4.ExportJSON()
	suite.Equal(expJSON04, outJSON4)

	// Test topics empty method
	tpc1, _ := Find("argo_uuid", "FooTopic", store)
	suite.Equal(true, tpc1.Empty())

	tpc2, _ := Find("argo_uuid", "", store)
	suite.Equal(false, tpc2.Empty())

}

func TestTopicsTestSuite(t *testing.T) {
	suite.Run(t, new(TopicTestSuite))
}
