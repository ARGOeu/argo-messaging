package subscriptions

import (
	"testing"

	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/stretchr/testify/suite"
	"github.com/ARGOeu/argo-messaging/config"
)

type SubTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *SubTestSuite) SetupTest() {
	suite.cfgStr = `
	{
	  "server":"localhost:9092",
	  "topics":["topic1","topic2"],
    "subscriptions":{"sub1":"topic1","sub2":"topic2"}
	}
	`
}

func (suite *SubTestSuite) TestCreate() {
	mySub := New("test-sub", "topic1")
	suite.Equal("test-sub", mySub.Name)
	suite.Equal("ARGO", mySub.Project)
	suite.Equal("/projects/ARGO/subscriptions/test-sub", mySub.FullName)
	suite.Equal("/projects/ARGO/topics/topic1", mySub.FullTopic)
}

func (suite *SubTestSuite) TestGetSubByName() {
	cfgKafka := config.NewKafkaCfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	mySubs := Subscriptions{}
	mySubs.LoadFromCfg(cfgKafka)
	result := mySubs.GetSubByName("ARGO", "sub1")
	expSub := New("sub1", "topic1")
	suite.Equal(expSub, result)
}

func (suite *SubTestSuite) TestGetSubsByProject() {
	cfgKafka := config.NewKafkaCfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	mySubs := Subscriptions{}
	mySubs.LoadFromCfg(cfgKafka)
	result := mySubs.GetSubsByProject("ARGO")
	expSub1 := New("sub1", "topic1")
	expSub2 := New("sub2", "topic2")
	expSubs := Subscriptions{}
	expSubs.List = append(expSubs.List, expSub1)
	expSubs.List = append(expSubs.List, expSub2)
	suite.Equal(expSubs, result)
}

func (suite *SubTestSuite) TestLoadFromCfg() {
	cfgKafka := config.NewKafkaCfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	mySubs := Subscriptions{}
	mySubs.LoadFromCfg(cfgKafka)
	expSub1 := New("sub1", "topic1")
	expSub2 := New("sub2", "topic2")
	expSubs := Subscriptions{}
	expSubs.List = append(expSubs.List, expSub1)
	expSubs.List = append(expSubs.List, expSub2)
	suite.Equal(expSubs, mySubs)

}

func (suite *SubTestSuite) TestExportJson() {
	cfgKafka := config.NewKafkaCfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	mySubs := Subscriptions{}
	mySubs.LoadFromCfg(cfgKafka)

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
      }
   ]
}`

	outJSON2, _ := mySubs.ExportJSON()
	suite.Equal(expJSON2, outJSON2)

}

func TestSubTestSuite(t *testing.T) {
	suite.Run(t, new(SubTestSuite))
}
