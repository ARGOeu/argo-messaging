package topics

import (
	"io/ioutil"
	"testing"

	log "github.com/Sirupsen/logrus"

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

func TestTopicsTestSuite(t *testing.T) {
	suite.Run(t, new(TopicTestSuite))
}
