package push

import (
	"errors"
	"io"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/stretchr/testify/suite"
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

	log.SetOutput(io.Discard)
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
	err := pushMgr.Add("argo_uuid", "tralala")
	suite.Equal(errors.New("not found"), err)
	err = pushMgr.Add("argo_uuid", "sub4")
	suite.Equal(nil, err)
	p, err := pushMgr.Get("foo_uuid/bar")
	suite.Equal(errors.New("not found"), err)
	suite.Equal(true, p == nil)
	p, err = pushMgr.Get("argo_uuid/sub4")
	suite.Equal(nil, err)
	suite.Equal(0, p.id)
	suite.Equal("argo_uuid", p.sub.ProjectUUID)
	suite.Equal("topic4", p.sub.Topic)
	suite.Equal("/projects/ARGO/subscriptions/sub4", p.sub.FullName)
	suite.Equal("/projects/ARGO/topics/topic4", p.sub.FullTopic)
	suite.Equal(10, p.sub.Ack)
	suite.Equal("endpoint.foo", p.sub.PushCfg.Pend)
}

func TestPushTestSuite(t *testing.T) {
	suite.Run(t, new(PushTestSuite))
}
