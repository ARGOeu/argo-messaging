package subscriptions

import (
	"errors"
	"io/ioutil"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/stretchr/testify/suite"
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
	mySub := New("argo_uuid", "ARGO", "test-sub", "topic1")
	suite.Equal("test-sub", mySub.Name)
	suite.Equal("argo_uuid", mySub.ProjectUUID)
	suite.Equal("/projects/ARGO/subscriptions/test-sub", mySub.FullName)
	suite.Equal("/projects/ARGO/topics/topic1", mySub.FullTopic)
}

func (suite *SubTestSuite) TestGetPushConfigFromJSON() {
	cfgAPI := config.NewAPICfg()
	cfgAPI.LoadStrJSON(suite.cfgStr)
	pJSON := `
	 {
     "pushConfig": {
	     "pushEndpoint": "exemplar.foo",
	     "retryPolicy":{"type":"linear","period":6000}
    }
   }
	`
	s, err := GetFromJSON([]byte(pJSON))
	suite.Equal(nil, err)
	suite.Equal("exemplar.foo", s.PushCfg.Pend)
	suite.Equal("linear", s.PushCfg.RetPol.PolicyType)
	suite.Equal(6000, s.PushCfg.RetPol.Period)

}

func (suite *SubTestSuite) TestGetSubByName() {
	cfgAPI := config.NewAPICfg()
	cfgAPI.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(cfgAPI.StoreHost, cfgAPI.StoreDB)

	result, _ := Find("argo_uuid", "sub1", store)
	expSub := New("argo_uuid", "ARGO", "sub1", "topic1")
	expSub.PushCfg.RetPol.PolicyType = ""
	expSub.PushCfg.RetPol.Period = 0
	suite.Equal(expSub, result.List[0])

}

func (suite *SubTestSuite) TestGetSubMetric() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	mySubM, _ := FindMetric("argo_uuid", "sub1", store)
	expTopic := SubMetrics{MsgNum: 0}
	suite.Equal(expTopic, mySubM)
}

func (suite *SubTestSuite) TestHasProjectTopic() {
	cfgAPI := config.NewAPICfg()
	cfgAPI.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(cfgAPI.StoreHost, cfgAPI.StoreDB)

	suite.Equal(false, HasSub("argo_uuid", "FOO", store))
	suite.Equal(true, HasSub("argo_uuid", "sub1", store))
}

func (suite *SubTestSuite) TestGetSubsByProject() {
	cfgAPI := config.NewAPICfg()
	cfgAPI.LoadStrJSON(suite.cfgStr)

	store := stores.NewMockStore(cfgAPI.StoreHost, cfgAPI.StoreDB)

	result, _ := Find("argo_uuid", "", store)
	expSub1 := New("argo_uuid", "ARGO", "sub1", "topic1")
	expSub1.PushCfg.RetPol.PolicyType = ""
	expSub1.PushCfg.RetPol.Period = 0
	expSub2 := New("argo_uuid", "ARGO", "sub2", "topic2")
	expSub2.PushCfg.RetPol.PolicyType = ""
	expSub2.PushCfg.RetPol.Period = 0
	expSub3 := New("argo_uuid", "ARGO", "sub3", "topic3")
	expSub3.PushCfg.RetPol.PolicyType = ""
	expSub3.PushCfg.RetPol.Period = 0
	expSub4 := New("argo_uuid", "ARGO", "sub4", "topic4")
	expSub4.PushCfg.RetPol.PolicyType = "linear"
	expSub4.PushCfg.RetPol.Period = 300
	rp := RetryPolicy{"linear", 300}
	expSub4.PushCfg = PushConfig{"endpoint.foo", rp}
	expSubs := Subscriptions{}
	expSubs.List = append(expSubs.List, expSub1)
	expSubs.List = append(expSubs.List, expSub2)
	expSubs.List = append(expSubs.List, expSub3)
	expSubs.List = append(expSubs.List, expSub4)
	suite.Equal(expSubs, result)
}

func (suite *SubTestSuite) TestLoadFromCfg() {
	cfgAPI := config.NewAPICfg()
	cfgAPI.LoadStrJSON(suite.cfgStr)

	store := stores.NewMockStore(cfgAPI.StoreHost, cfgAPI.StoreDB)
	results, _ := Find("argo_uuid", "", store)
	expSub1 := New("argo_uuid", "ARGO", "sub1", "topic1")
	expSub1.PushCfg.RetPol.PolicyType = ""
	expSub1.PushCfg.RetPol.Period = 0
	expSub2 := New("argo_uuid", "ARGO", "sub2", "topic2")
	expSub2.PushCfg.RetPol.PolicyType = ""
	expSub2.PushCfg.RetPol.Period = 0
	expSub3 := New("argo_uuid", "ARGO", "sub3", "topic3")
	expSub3.PushCfg.RetPol.PolicyType = ""
	expSub3.PushCfg.RetPol.Period = 0
	expSub4 := New("argo_uuid", "ARGO", "sub4", "topic4")
	expSub4.PushCfg.RetPol.PolicyType = "linear"
	expSub4.PushCfg.RetPol.Period = 300
	rp := RetryPolicy{"linear", 300}
	expSub4.PushCfg = PushConfig{"endpoint.foo", rp}
	expSubs := Subscriptions{}
	expSubs.List = append(expSubs.List, expSub1)
	expSubs.List = append(expSubs.List, expSub2)
	expSubs.List = append(expSubs.List, expSub3)
	expSubs.List = append(expSubs.List, expSub4)
	suite.Equal(expSubs, results)

}

func (suite *SubTestSuite) TestRemoveSubStore() {

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)

	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)

	suite.Equal(true, HasSub("argo_uuid", "sub1", store))

	suite.Equal("not found", RemoveSub("argo_uuid", "subFoo", store).Error())
	suite.Equal(nil, RemoveSub("argo_uuid", "sub1", store))

	suite.Equal(false, HasSub("ARGO", "sub1", store))
}

func (suite *SubTestSuite) TestCreateSubStore() {

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)

	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)

	sub, err := CreateSub("argo_uuid", "sub1", "topic1", "", 0, 0, "linear", 300, store)
	suite.Equal(Subscription{}, sub)
	suite.Equal("exists", err.Error())

	sub2, err2 := CreateSub("argo_uuid", "subNew", "topicNew", "", 0, 0, "linear", 300, store)
	expSub := New("argo_uuid", "ARGO", "subNew", "topicNew")
	suite.Equal(expSub, sub2)
	suite.Equal(nil, err2)

}

func (suite *SubTestSuite) TestModAck() {

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)

	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)

	err := ModAck("argo_uuid", "sub1", 300, store)
	suite.Equal(nil, err)

	err = ModAck("argo_uuid", "sub1", 0, store)
	suite.Equal(nil, err)

	err = ModAck("argo_uuid", "sub1", -300, store)
	suite.Equal(errors.New("wrong value"), err)

	err = ModAck("argo_uuid", "sub1", 601, store)
	suite.Equal(errors.New("wrong value"), err)
}

func (suite *SubTestSuite) TestExtractFullTopic() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)

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

	store := stores.NewMockStore(cfgAPI.StoreHost, cfgAPI.StoreDB)

	res, _ := Find("argo_uuid", "sub1", store)

	outJSON, _ := res.List[0].ExportJSON()
	expJSON := `{
   "name": "/projects/ARGO/subscriptions/sub1",
   "topic": "/projects/ARGO/topics/topic1",
   "pushConfig": {
      "pushEndpoint": "",
      "retryPolicy": {}
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
            "pushEndpoint": "",
            "retryPolicy": {}
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub2",
         "topic": "/projects/ARGO/topics/topic2",
         "pushConfig": {
            "pushEndpoint": "",
            "retryPolicy": {}
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub3",
         "topic": "/projects/ARGO/topics/topic3",
         "pushConfig": {
            "pushEndpoint": "",
            "retryPolicy": {}
         },
         "ackDeadlineSeconds": 10
      },
      {
         "name": "/projects/ARGO/subscriptions/sub4",
         "topic": "/projects/ARGO/topics/topic4",
         "pushConfig": {
            "pushEndpoint": "endpoint.foo",
            "retryPolicy": {
               "type": "linear",
               "period": 300
            }
         },
         "ackDeadlineSeconds": 10
      }
   ]
}`
	results, _ := Find("argo_uuid", "", store)
	outJSON2, _ := results.ExportJSON()
	suite.Equal(expJSON2, outJSON2)

}

func (suite *SubTestSuite) TestGetMaxAckID() {
	ackIDs := []string{"projects/ARGO/subscriptions/sub1:2",
		"projects/ARGO/subscriptions/sub1:4",
		"projects/ARGO/subscriptions/sub1:1555",
		"projects/ARGO/subscriptions/sub1:5",
		"projects/ARGO/subscriptions/sub1:3"}

	max, err := GetMaxAckID(ackIDs)
	suite.Equal(nil, err)
	suite.Equal("projects/ARGO/subscriptions/sub1:1555", max)

}

func (suite *SubTestSuite) TestGetOffsetFromAckID() {
	ackIDs := []string{"projects/ARGO/subscriptions/sub1:2",
		"projects/ARGO/subscriptions/sub1:4",
		"projects/ARGO/subscriptions/sub1:1555",
		"projects/ARGO/subscriptions/sub1:5",
		"projects/ARGO/subscriptions/sub1:3"}

	expOffsets := []int64{2, 4, 1555, 5, 3}

	for i := range ackIDs {
		off, err := GetOffsetFromAckID(ackIDs[i])
		suite.Equal(nil, err)
		suite.Equal(expOffsets[i], off)
	}

}

func TestSubTestSuite(t *testing.T) {
	suite.Run(t, new(SubTestSuite))
}
