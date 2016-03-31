package subscriptions

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/stretchr/testify/suite"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/stores"
)

type SubTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *SubTestSuite) SetupTest() {
	suite.cfgStr = `{
	  "port":8080,
		"broker_hosts":["localhost:9092"],
		"store_host":"localhost",
		"store_db":"argo_msg",
		"use_authorization":true,
		"use_authentication":true,
		"use_ack":true
	}`

	log.SetOutput(ioutil.Discard)
}

func (suite *SubTestSuite) TestCreate() {
	mySub := New("ARGO", "test-sub", "topic1")
	suite.Equal("test-sub", mySub.Name)
	suite.Equal("ARGO", mySub.Project)
	suite.Equal("/projects/ARGO/subscriptions/test-sub", mySub.FullName)
	suite.Equal("/projects/ARGO/topics/topic1", mySub.FullTopic)
}

func (suite *SubTestSuite) TestGetSubByName() {
	cfgAPI := config.NewAPICfg()
	cfgAPI.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(cfgAPI.StoreHost, cfgAPI.StoreDB)
	mySubs := Subscriptions{}
	mySubs.LoadFromStore(store)
	result := mySubs.GetSubByName("ARGO", "sub1")
	expSub := New("ARGO", "sub1", "topic1")
	suite.Equal(expSub, result)
}

func (suite *SubTestSuite) TestHasProjectTopic() {
	cfgAPI := config.NewAPICfg()
	cfgAPI.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(cfgAPI.StoreHost, cfgAPI.StoreDB)
	mySubs := Subscriptions{}
	mySubs.LoadFromStore(store)

	suite.Equal(false, mySubs.HasSub("ARGO", "FOO"))
	suite.Equal(true, mySubs.HasSub("ARGO", "sub1"))
}

func (suite *SubTestSuite) TestGetSubsByProject() {
	cfgAPI := config.NewAPICfg()
	cfgAPI.LoadStrJSON(suite.cfgStr)
	mySubs := Subscriptions{}
	store := stores.NewMockStore(cfgAPI.StoreHost, cfgAPI.StoreDB)
	mySubs.LoadFromStore(store)
	result := mySubs.GetSubsByProject("ARGO")
	expSub1 := New("ARGO", "sub1", "topic1")
	expSub2 := New("ARGO", "sub2", "topic2")
	expSub3 := New("ARGO", "sub3", "topic3")
	expSubs := Subscriptions{}
	expSubs.List = append(expSubs.List, expSub1)
	expSubs.List = append(expSubs.List, expSub2)
	expSubs.List = append(expSubs.List, expSub3)
	suite.Equal(expSubs, result)
}

func (suite *SubTestSuite) TestLoadFromCfg() {
	cfgAPI := config.NewAPICfg()
	cfgAPI.LoadStrJSON(suite.cfgStr)
	mySubs := Subscriptions{}
	store := stores.NewMockStore(cfgAPI.StoreHost, cfgAPI.StoreDB)
	mySubs.LoadFromStore(store)
	expSub1 := New("ARGO", "sub1", "topic1")
	expSub2 := New("ARGO", "sub2", "topic2")
	expSub3 := New("ARGO", "sub3", "topic3")
	expSubs := Subscriptions{}
	expSubs.List = append(expSubs.List, expSub1)
	expSubs.List = append(expSubs.List, expSub2)
	expSubs.List = append(expSubs.List, expSub3)
	suite.Equal(expSubs, mySubs)

}

func (suite *SubTestSuite) TestRemoveSubStore() {

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	mySubs := Subscriptions{}
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	mySubs.LoadFromStore(store)

	suite.Equal(true, mySubs.HasSub("ARGO", "sub1"))

	suite.Equal("not found", mySubs.RemoveSub("ARGO", "subFoo", store).Error())
	suite.Equal(nil, mySubs.RemoveSub("ARGO", "sub1", store))
	mySubs.LoadFromStore(store)
	suite.Equal(false, mySubs.HasSub("ARGO", "sub1"))
}

func (suite *SubTestSuite) TestCreateSubStore() {

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	mySubs := Subscriptions{}
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	mySubs.LoadFromStore(store)

	sub, err := mySubs.CreateSub("ARGO", "sub1", "topic1", 0, 0, store)
	suite.Equal(Subscription{}, sub)
	suite.Equal("exists", err.Error())

	sub2, err2 := mySubs.CreateSub("ARGO", "subNew", "topicNew", 0, 0, store)
	expSub := New("ARGO", "subNew", "topicNew")
	suite.Equal(expSub, sub2)
	suite.Equal(nil, err2)

}

func (suite *SubTestSuite) TestExtractFullTopic() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	mySubs := Subscriptions{}
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	mySubs.LoadFromStore(store)

	project, topic, err := ExtractFullTopicRef("projects/ARGO/topics/topic1")
	suite.Equal("ARGO", project)
	suite.Equal("topic1", topic)
	suite.Equal(nil, err)

	project2, topic2, err2 := ExtractFullTopicRef("proje/ARGO/topic/topic1")
	suite.Equal("", project2)
	suite.Equal("", topic2)
	suite.Equal("wrong topic name declaration", err2.Error())

	project3, topic3, err3 := ExtractFullTopicRef("projects/ARGO/topics/topic1/lalala")
	suite.Equal("", project3)
	suite.Equal("", topic3)
	suite.Equal("wrong topic name declaration", err3.Error())

}

func (suite *SubTestSuite) TestExportJson() {
	cfgAPI := config.NewAPICfg()
	cfgAPI.LoadStrJSON(suite.cfgStr)
	mySubs := Subscriptions{}
	store := stores.NewMockStore(cfgAPI.StoreHost, cfgAPI.StoreDB)
	mySubs.LoadFromStore(store)

	outJSON, _ := mySubs.List[0].ExportJSON()
	expJSON := `{
   "name": "/projects/ARGO/subscriptions/sub1",
   "topic": "/projects/ARGO/topics/topic1",
   "pushConfig": {
      "pushEndpoint": ""
   },
   "ackDeadlineSeconds": 10
}`
	suite.Equal(expJSON, outJSON)

	expJSON2 := `{
   "subscriptions": [
      {
         "name": "/projects/ARGO/subscriptions/sub1",
         "topic": "/projects/ARGO/topics/topic1",
         "pushConfig": {
            "pushEndpoint": ""
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub2",
         "topic": "/projects/ARGO/topics/topic2",
         "pushConfig": {
            "pushEndpoint": ""
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub3",
         "topic": "/projects/ARGO/topics/topic3",
         "pushConfig": {
            "pushEndpoint": ""
         },
         "ackDeadlineSeconds": 10
      }
   ]
}`

	outJSON2, _ := mySubs.ExportJSON()
	suite.Equal(expJSON2, outJSON2)

}

func TestSubTestSuite(t *testing.T) {
	suite.Run(t, new(SubTestSuite))
}
