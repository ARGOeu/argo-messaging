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
		"broker_host":"localhost:9092",
		"store_host":"localhost",
		"store_db":"argo_msg"
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
