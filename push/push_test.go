package push

import (
	"errors"
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/stores"
)

type PushTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *PushTestSuite) SetupTest() {
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

func (suite *PushTestSuite) TestPusher() {
	sndr := NewMockSender(false)
	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	pushMgr := NewManager(nil, nil, nil)
	suite.Equal(false, pushMgr.isSet())
	pushMgr = NewManager(&brk, str, sndr)
	suite.Equal(true, pushMgr.isSet())
	err := pushMgr.Add("ARGO", "tralala")
	suite.Equal(errors.New("No sub found"), err)
	err = pushMgr.Add("ARGO", "sub4")
	suite.Equal(nil, err)
	p, err := pushMgr.Get("foo/bar")
	suite.Equal(errors.New("not found"), err)
	suite.Equal(true, p == nil)
	p, err = pushMgr.Get("ARGO/sub4")
	suite.Equal(nil, err)
	suite.Equal(0, p.id)
	suite.Equal("ARGO", p.sub.Project)
	suite.Equal("topic4", p.sub.Topic)
	suite.Equal("/projects/ARGO/subscriptions/sub4", p.sub.FullName)
	suite.Equal("/projects/ARGO/topics/topic4", p.sub.FullTopic)
	suite.Equal(10, p.sub.Ack)
	suite.Equal("endpoint.foo", p.sub.PushCfg.Pend)
}

func TestPushTestSuite(t *testing.T) {
	suite.Run(t, new(PushTestSuite))
}
