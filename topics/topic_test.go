package topics

import (
	"testing"

	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/stretchr/testify/suite"
	"github.com/ARGOeu/argo-messaging/config"
)

type TopicTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *TopicTestSuite) SetupTest() {
	suite.cfgStr = `
	{
	  "server":"localhost:9092",
	  "topics":["topic1","topic2"]
	}
	`
}

func (suite *TopicTestSuite) TestCreate() {
	myTopic := New("test-topic")
	suite.Equal("test-topic", myTopic.Name)
	suite.Equal("ARGO", myTopic.Project)
	suite.Equal("/projects/ARGO/topics/test-topic", myTopic.FullName)
}

func (suite *TopicTestSuite) TestGetTopicByName() {
	cfgKafka := config.NewKafkaCfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	myTopics := Topics{}
	myTopics.LoadFromCfg(cfgKafka)
	result := myTopics.GetTopicByName("ARGO", "topic1")
	expTopic := New("topic1")
	suite.Equal(expTopic, result)
}

func (suite *TopicTestSuite) TestGetTopicsByProject() {
	cfgKafka := config.NewKafkaCfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	myTopics := Topics{}
	myTopics.LoadFromCfg(cfgKafka)
	expTopics := Topics{}
	expTopics.LoadFromCfg(cfgKafka)
	result := myTopics.GetTopicsByProject("ARGO")
	suite.Equal(expTopics, result)
}

func (suite *TopicTestSuite) TestLoadFromCfg() {
	cfgKafka := config.NewKafkaCfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	myTopics := Topics{}
	myTopics.LoadFromCfg(cfgKafka)
	expTopics := Topics{}
	expTopic1 := New("topic1")
	expTopic2 := New("topic2")
	expTopics.List = append(expTopics.List, expTopic1)
	expTopics.List = append(expTopics.List, expTopic2)
	suite.Equal(expTopics, myTopics)

}

func (suite *TopicTestSuite) TestExportJson() {
	cfgKafka := config.NewKafkaCfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	myTopics := Topics{}
	myTopics.LoadFromCfg(cfgKafka)

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
      }
   ]
}`

	outJSON2, _ := myTopics.ExportJSON()
	suite.Equal(expJSON2, outJSON2)

}

func TestTopicsTestSuite(t *testing.T) {
	suite.Run(t, new(TopicTestSuite))
}
