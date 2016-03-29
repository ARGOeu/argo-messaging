package topics

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/stretchr/testify/suite"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/stores"
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
	myTopic := New("ARGO", "test-topic")
	suite.Equal("test-topic", myTopic.Name)
	suite.Equal("ARGO", myTopic.Project)
	suite.Equal("/projects/ARGO/topics/test-topic", myTopic.FullName)
}

func (suite *TopicTestSuite) TestGetTopicByName() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	myTopics := Topics{}
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	myTopics.LoadFromStore(store)
	result := myTopics.GetTopicByName("ARGO", "topic1")
	expTopic := New("ARGO", "topic1")
	suite.Equal(expTopic, result)
}

func (suite *TopicTestSuite) TestCreateTopicStore() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	myTopics := Topics{}
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	myTopics.LoadFromStore(store)

	tp, err := myTopics.CreateTopic("ARGO", "topic1", store)
	suite.Equal(Topic{}, tp)
	suite.Equal("exists", err.Error())

	tp2, err2 := myTopics.CreateTopic("ARGO", "topicNew", store)
	expTopic := New("ARGO", "topicNew")
	suite.Equal(expTopic, tp2)
	suite.Equal(nil, err2)
}

func (suite *TopicTestSuite) TestRemoveTopicStore() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	myTopics := Topics{}
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	myTopics.LoadFromStore(store)

	suite.Equal(true, myTopics.HasTopic("ARGO", "topic1"))

	suite.Equal("not found", myTopics.RemoveTopic("ARGO", "topicFoo", store).Error())
	suite.Equal(nil, myTopics.RemoveTopic("ARGO", "topic1", store))
	myTopics.LoadFromStore(store)
	suite.Equal(false, myTopics.HasTopic("ARGO", "topic1"))
}

func (suite *TopicTestSuite) TestHasProjectTopic() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	myTopics := Topics{}
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	myTopics.LoadFromStore(store)

	suite.Equal(false, myTopics.HasTopic("ARGO", "FOO"))
	suite.Equal(true, myTopics.HasTopic("ARGO", "topic1"))
}

func (suite *TopicTestSuite) TestGetTopicsByProject() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	myTopics := Topics{}
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	myTopics.LoadFromStore(store)
	expTopics := Topics{}
	expTopics.LoadFromStore(store)
	result := myTopics.GetTopicsByProject("ARGO")
	suite.Equal(expTopics, result)
}

func (suite *TopicTestSuite) TestExportJson() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	myTopics := Topics{}
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	myTopics.LoadFromStore(store)

	outJSON, _ := myTopics.List[0].ExportJSON()
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

	outJSON2, _ := myTopics.ExportJSON()
	suite.Equal(expJSON2, outJSON2)

}

func TestTopicsTestSuite(t *testing.T) {
	suite.Run(t, new(TopicTestSuite))
}
