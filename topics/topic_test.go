package topics

import (
	"io/ioutil"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/stretchr/testify/suite"
	"time"
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
	myTopics, _ := Find("argo_uuid", "", "topic1", "", 0, store)
	expTopic := New("argo_uuid", "ARGO", "topic1")
	expTopic.PublishRate = 10
	expTopic.LatestPublish = time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local)
	suite.Equal(expTopic, myTopics.Topics[0])
}

func (suite *TopicTestSuite) TestGetPaginatedTopics() {

	store := stores.NewMockStore("", "")

	// retrieve all topics
	expPt1 := PaginatedTopics{Topics: []Topic{
		{"argo_uuid", "topic4", "/projects/ARGO/topics/topic4", time.Date(0, 0, 0, 0, 0, 0, 0, time.Local), 0, ""},
		{"argo_uuid", "topic3", "/projects/ARGO/topics/topic3", time.Date(2019, 5, 7, 0, 0, 0, 0, time.Local), 8.99, ""},
		{"argo_uuid", "topic2", "/projects/ARGO/topics/topic2", time.Date(2019, 5, 8, 0, 0, 0, 0, time.Local), 5.45, "schema-1"},
		{"argo_uuid", "topic1", "/projects/ARGO/topics/topic1", time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local), 10, ""}},
		NextPageToken: "", TotalSize: 4}
	pgTopics1, err1 := Find("argo_uuid", "", "", "", 0, store)

	// retrieve first 2 topics
	expPt2 := PaginatedTopics{Topics: []Topic{
		{"argo_uuid", "topic4", "/projects/ARGO/topics/topic4", time.Date(0, 0, 0, 0, 0, 0, 0, time.Local), 0, ""},
		{"argo_uuid", "topic3", "/projects/ARGO/topics/topic3", time.Date(2019, 5, 7, 0, 0, 0, 0, time.Local), 8.99, ""}},
		NextPageToken: "MQ==", TotalSize: 4}
	pgTopics2, err2 := Find("argo_uuid", "", "", "", 2, store)

	// retrieve the next topic
	expPt3 := PaginatedTopics{Topics: []Topic{
		{"argo_uuid", "topic1", "/projects/ARGO/topics/topic1", time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local), 10, ""}},
		NextPageToken: "", TotalSize: 4}
	pgTopics3, err3 := Find("argo_uuid", "", "", "MA==", 1, store)

	// invalid page token
	_, err4 := Find("", "", "", "invalid", 0, store)

	// retrieve topics for a specific user
	expPt5 := PaginatedTopics{Topics: []Topic{
		{"argo_uuid", "topic2", "/projects/ARGO/topics/topic2", time.Date(2019, 5, 8, 0, 0, 0, 0, time.Local), 5.45, "schema-1"},
		{"argo_uuid", "topic1", "/projects/ARGO/topics/topic1", time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local), 10, ""}},
		NextPageToken: "", TotalSize: 2}
	pgTopics5, err5 := Find("argo_uuid", "uuid1", "", "", 2, store)

	// retrieve topics for a specific user with pagination
	expPt6 := PaginatedTopics{Topics: []Topic{
		{"argo_uuid", "topic2", "/projects/ARGO/topics/topic2", time.Date(2019, 5, 8, 0, 0, 0, 0, time.Local), 5.45, "schema-1"}},
		NextPageToken: "MA==", TotalSize: 2}
	pgTopics6, err6 := Find("argo_uuid", "uuid1", "", "", 1, store)

	suite.Equal(expPt1, pgTopics1)
	suite.Equal(expPt2, pgTopics2)
	suite.Equal(expPt3, pgTopics3)
	suite.Equal(expPt5, pgTopics5)
	suite.Equal(expPt6, pgTopics6)

	suite.Nil(err1)
	suite.Nil(err2)
	suite.Nil(err3)
	suite.Equal("illegal base64 data at input byte 4", err4.Error())
	suite.Nil(err5)
	suite.Nil(err6)
}

func (suite *TopicTestSuite) TestGetTopicMetric() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	myTopics, _ := FindMetric("argo_uuid", "topic1", store)
	expTopic := TopicMetrics{MsgNum: 0}
	expTopic.LatestPublish = time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local)
	expTopic.PublishRate = 10
	suite.Equal(expTopic, myTopics)
}

// Find searches and returns a specific topic metric
func (suite *TopicTestSuite) TestGetTopicMetrics() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	myTopics, _ := FindMetric("argo_uuid", "topic1", store)
	expTopic := TopicMetrics{MsgNum: 0}
	expTopic.PublishRate = 10
	expTopic.LatestPublish = time.Date(2019, 5, 6, 0, 0, 0, 0, time.Local)
	suite.Equal(expTopic, myTopics)
}

func (suite *TopicTestSuite) TestCreateTopicStore() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)

	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)

	tp, err := CreateTopic("argo_uuid", "topic1", "", store)
	suite.Equal(Topic{}, tp)
	suite.Equal("exists", err.Error())

	tp2, err2 := CreateTopic("argo_uuid", "topicNew", "schema_uuid_1", store)
	expTopic := New("argo_uuid", "ARGO", "topicNew")
	expTopic.Schema = "schema-1"
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

	topics, _ := Find("argo_uuid", "", "topic1", "", 0, store)
	outJSON, _ := topics.Topics[0].ExportJSON()
	expJSON := `{
   "name": "/projects/ARGO/topics/topic1"
}`
	suite.Equal(expJSON, outJSON)

	expJSON2 := `{
   "topics": [
      {
         "name": "/projects/ARGO/topics/topic4"
      },
      {
         "name": "/projects/ARGO/topics/topic3"
      },
      {
         "name": "/projects/ARGO/topics/topic2",
         "schema": "schema-1"
      },
      {
         "name": "/projects/ARGO/topics/topic1"
      }
   ],
   "nextPageToken": "",
   "totalSize": 4
}`
	topics2, _ := Find("argo_uuid", "", "", "", 0, store)
	outJSON2, _ := topics2.ExportJSON()
	suite.Equal(expJSON2, outJSON2)

}

func TestTopicsTestSuite(t *testing.T) {
	suite.Run(t, new(TopicTestSuite))
}
