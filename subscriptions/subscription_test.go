package subscriptions

import (
	"errors"
	"io/ioutil"
	"testing"

	log "github.com/sirupsen/logrus"

	"net/http"
	"strings"
	"time"

	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/stretchr/testify/suite"
)

type SubTestSuite struct {
	suite.Suite
	cfgStr string
}

type MockPushRoundTripper struct{}

func (m *MockPushRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {

	var resp *http.Response

	header := make(http.Header)
	header.Set("Content-type", "text/plain")

	switch r.URL.Host {

	case "example.com":

		resp = &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader("vhash-1")),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "example_error.com":

		resp = &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader("Internal error")),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "example_mismatch.com":

		resp = &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader("wrong_vhash")),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	}

	return resp, nil
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

func (suite *SubTestSuite) TestNewNamesLIst() {
	nl := NewNamesList()
	// make sure that the subscriptions slice has been initialised
	suite.NotNil(nl.Subscriptions)
}

func (suite *SubTestSuite) TestFindByTopic() {

	store := stores.NewMockStore("", "")

	nl1, err1 := FindByTopic("argo_uuid", "topic1", store)
	suite.Equal([]string{"/projects/ARGO/subscriptions/sub1"}, nl1.Subscriptions)
	suite.Nil(err1)

	// check empty case
	store.SubList = nil
	nl2, err2 := FindByTopic("argo_uuid", "topic1", store)
	suite.Equal([]string{}, nl2.Subscriptions)
	suite.Nil(err2)
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

	result, _ := Find("argo_uuid", "", "sub1", "", 0, store)
	expSub := New("argo_uuid", "ARGO", "sub1", "topic1")
	expSub.PushCfg.RetPol.PolicyType = ""
	expSub.PushCfg.RetPol.Period = 0
	expSub.CreatedOn = "2020-11-19T00:00:00Z"
	expSub.LatestConsume = time.Date(2019, 5, 6, 0, 0, 0, 0, time.UTC)
	expSub.ConsumeRate = 10
	suite.Equal(expSub, result.Subscriptions[0])

}

func (suite *SubTestSuite) TestGetSubMetric() {
	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	mySubM, _ := FindMetric("argo_uuid", "sub1", store)
	expTopic := SubMetrics{
		MsgNum:        0,
		LatestConsume: time.Date(2019, 5, 6, 0, 0, 0, 0, time.UTC),
		ConsumeRate:   10,
	}
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

	expSub1 := New("argo_uuid", "ARGO", "sub1", "topic1")
	expSub1.PushCfg.RetPol.PolicyType = ""
	expSub1.PushCfg.RetPol.Period = 0
	expSub1.PushCfg.MaxMessages = 0
	expSub1.LatestConsume = time.Date(2019, 5, 6, 0, 0, 0, 0, time.UTC)
	expSub1.ConsumeRate = 10
	expSub1.CreatedOn = "2020-11-19T00:00:00Z"
	expSub2 := New("argo_uuid", "ARGO", "sub2", "topic2")
	expSub2.PushCfg.RetPol.PolicyType = ""
	expSub2.PushCfg.RetPol.Period = 0
	expSub2.PushCfg.MaxMessages = 0
	expSub2.LatestConsume = time.Date(2019, 5, 7, 0, 0, 0, 0, time.UTC)
	expSub2.ConsumeRate = 8.99
	expSub2.CreatedOn = "2020-11-20T00:00:00Z"
	expSub3 := New("argo_uuid", "ARGO", "sub3", "topic3")
	expSub3.PushCfg.RetPol.PolicyType = ""
	expSub3.PushCfg.RetPol.Period = 0
	expSub3.PushCfg.MaxMessages = 0
	expSub3.LatestConsume = time.Date(2019, 5, 8, 0, 0, 0, 0, time.UTC)
	expSub3.ConsumeRate = 5.45
	expSub3.CreatedOn = "2020-11-21T00:00:00Z"
	expSub4 := New("argo_uuid", "ARGO", "sub4", "topic4")
	expSub4.PushCfg.RetPol.PolicyType = "linear"
	expSub4.PushCfg.RetPol.Period = 300
	expSub4.CreatedOn = "2020-11-22T00:00:00Z"
	rp := RetryPolicy{"linear", 300}
	authCFG := AuthorizationHeader{"autogen", "auth-header-1"}
	expSub4.PushCfg = PushConfig{
		Type:                "http_endpoint",
		Pend:                "endpoint.foo",
		AuthorizationHeader: authCFG,
		RetPol:              rp,
		VerificationHash:    "push-id-1",
		Verified:            true,
		MaxMessages:         1,
	}
	expSub4.LatestConsume = time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
	expSub4.ConsumeRate = 0

	// retrieve all subs
	expSubs1 := []Subscription{}
	expSubs1 = append(expSubs1, expSub4)
	expSubs1 = append(expSubs1, expSub3)
	expSubs1 = append(expSubs1, expSub2)
	expSubs1 = append(expSubs1, expSub1)
	result1, _ := Find("argo_uuid", "", "", "", 0, store)

	// retrieve first two subs
	expSubs2 := []Subscription{}
	expSubs2 = append(expSubs2, expSub4)
	expSubs2 = append(expSubs2, expSub3)
	result2, _ := Find("argo_uuid", "", "", "", 2, store)

	//retrieve the next two subs
	expSubs3 := []Subscription{}
	expSubs3 = append(expSubs3, expSub2)
	expSubs3 = append(expSubs3, expSub1)
	result3, _ := Find("argo_uuid", "", "", "MQ==", 2, store)

	// provide an invalid page token
	_, err := Find("", "", "", "invalid", 0, store)

	// retrieve user's subs
	expSubs4 := []Subscription{}
	expSubs4 = append(expSubs4, expSub4)
	expSubs4 = append(expSubs4, expSub3)
	expSubs4 = append(expSubs4, expSub2)
	result4, _ := Find("argo_uuid", "uuid1", "", "", 0, store)

	// retrieve user's subs with pagination
	expSubs5 := []Subscription{}
	expSubs5 = append(expSubs5, expSub4)
	expSubs5 = append(expSubs5, expSub3)
	result5, _ := Find("argo_uuid", "uuid1", "", "", 2, store)

	suite.Equal(expSubs1, result1.Subscriptions)
	suite.Equal("", result1.NextPageToken)
	suite.Equal(int32(4), result1.TotalSize)

	suite.Equal(expSubs2, result2.Subscriptions)
	suite.Equal("MQ==", result2.NextPageToken)
	suite.Equal(int32(4), result2.TotalSize)

	suite.Equal(expSubs3, result3.Subscriptions)
	suite.Equal("", result3.NextPageToken)
	suite.Equal(int32(4), result3.TotalSize)

	suite.Equal("illegal base64 data at input byte 4", err.Error())

	suite.Equal(expSubs4, result4.Subscriptions)
	suite.Equal("", result4.NextPageToken)
	suite.Equal(int32(3), result4.TotalSize)

	suite.Equal(expSubs5, result5.Subscriptions)
	suite.Equal("MQ==", result5.NextPageToken)
	suite.Equal(int32(3), result5.TotalSize)
}

func (suite *SubTestSuite) TestLoadFromCfg() {
	cfgAPI := config.NewAPICfg()
	cfgAPI.LoadStrJSON(suite.cfgStr)

	store := stores.NewMockStore(cfgAPI.StoreHost, cfgAPI.StoreDB)
	results, _ := Find("argo_uuid", "", "", "", 0, store)
	expSub1 := New("argo_uuid", "ARGO", "sub1", "topic1")
	expSub1.PushCfg.RetPol.PolicyType = ""
	expSub1.PushCfg.RetPol.Period = 0
	expSub1.PushCfg.MaxMessages = 0
	expSub1.LatestConsume = time.Date(2019, 5, 6, 0, 0, 0, 0, time.UTC)
	expSub1.ConsumeRate = 10
	expSub1.CreatedOn = "2020-11-19T00:00:00Z"
	expSub2 := New("argo_uuid", "ARGO", "sub2", "topic2")
	expSub2.PushCfg.RetPol.PolicyType = ""
	expSub2.PushCfg.RetPol.Period = 0
	expSub2.PushCfg.MaxMessages = 0
	expSub2.LatestConsume = time.Date(2019, 5, 7, 0, 0, 0, 0, time.UTC)
	expSub2.ConsumeRate = 8.99
	expSub2.CreatedOn = "2020-11-20T00:00:00Z"
	expSub3 := New("argo_uuid", "ARGO", "sub3", "topic3")
	expSub3.PushCfg.RetPol.PolicyType = ""
	expSub3.PushCfg.RetPol.Period = 0
	expSub3.PushCfg.MaxMessages = 0
	expSub3.LatestConsume = time.Date(2019, 5, 8, 0, 0, 0, 0, time.UTC)
	expSub3.ConsumeRate = 5.45
	expSub3.CreatedOn = "2020-11-21T00:00:00Z"
	expSub4 := New("argo_uuid", "ARGO", "sub4", "topic4")
	expSub4.PushCfg.RetPol.PolicyType = "linear"
	expSub4.PushCfg.RetPol.Period = 300
	authCFG := AuthorizationHeader{"autogen", "auth-header-1"}
	expSub4.CreatedOn = "2020-11-22T00:00:00Z"
	rp := RetryPolicy{"linear", 300}
	expSub4.PushCfg = PushConfig{
		Type:                "http_endpoint",
		Pend:                "endpoint.foo",
		AuthorizationHeader: authCFG,
		RetPol:              rp,
		VerificationHash:    "push-id-1",
		Verified:            true,
		MaxMessages:         1,
	}
	expSub4.LatestConsume = time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
	expSub4.ConsumeRate = 0
	expSubs := []Subscription{}
	expSubs = append(expSubs, expSub4)
	expSubs = append(expSubs, expSub3)
	expSubs = append(expSubs, expSub2)
	expSubs = append(expSubs, expSub1)
	suite.Equal(expSubs, results.Subscriptions)

}

func (suite *SubTestSuite) TestIsAuthzTypeSupported() {
	suite.True(IsAuthorizationHeaderTypeSupported("autogen"))
	suite.True(IsAuthorizationHeaderTypeSupported("disabled"))
	suite.False(IsRetryPolicySupported("unknown"))
}

func (suite *SubTestSuite) TestIsRetPolSupported() {
	suite.True(IsRetryPolicySupported("linear"))
	suite.False(IsRetryPolicySupported("unknown"))
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

	sub, err := Create("argo_uuid", "sub1", "topic1", 0, 300,
		PushConfig{}, time.Date(2019, 7, 7, 0, 0, 0, 0, time.UTC), store)
	suite.Equal(Subscription{}, sub)
	suite.Equal("exists", err.Error())

	sub2, err2 := Create("argo_uuid", "subNew", "topicNew", 0, 0,
		PushConfig{}, time.Date(2019, 7, 7, 0, 0, 0, 0, time.UTC), store)
	expSub := New("argo_uuid", "ARGO", "subNew", "topicNew")
	expSub.CreatedOn = "2019-07-07T00:00:00Z"
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

func (suite *SubTestSuite) TestModSubPush() {

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)

	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)

	// modify push config
	err1 := ModSubPush("argo_uuid", "sub1", "example.com", "autogen", "auth-h", 2, "linear", 400, "hash-1", true, store)

	suite.Nil(err1)

	sub1, _ := store.QueryOneSub("argo_uuid", "sub1")
	suite.Equal("example.com", sub1.PushEndpoint)
	suite.Equal(int64(2), sub1.MaxMessages)
	suite.Equal("linear", sub1.RetPolicy)
	suite.Equal(400, sub1.RetPeriod)
	suite.Equal("hash-1", sub1.VerificationHash)
	suite.True(sub1.Verified)

	// test error case
	err2 := ModSubPush("argo_uuid", "unknown", "", "", "", 0, "", 0, "", false, store)
	suite.Equal("not found", err2.Error())
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

func (suite *SubTestSuite) TestPushEndpointHost() {

	sub := Subscription{
		PushCfg: PushConfig{
			Pend: "https://example.com:8084/receive_here",
		},
	}

	u1 := sub.PushEndpointHost()
	suite.Equal("example.com:8084", u1)
}

func (suite *SubTestSuite) TestVerifyPushEndpoint() {

	str := stores.NewMockStore("", "")

	// normal case
	s1 := Subscription{
		Name:        "push-sub-v1",
		ProjectUUID: "argo_uuid",
		PushCfg: PushConfig{
			Pend:             "https://example.com/receive_here",
			VerificationHash: "vhash-1",
		},
	}

	// add a temporary subscription
	q1 := stores.QSub{
		Name:             "push-sub-v1",
		ProjectUUID:      "argo_uuid",
		PushEndpoint:     "https://example.com/receive_here",
		VerificationHash: "vhash-1",
		Verified:         false,
	}

	str.SubList = append(str.SubList, q1)

	c1 := &http.Client{
		Transport: new(MockPushRoundTripper),
	}

	e1 := VerifyPushEndpoint(s1, c1, str)

	qs1, _ := str.QueryOneSub("argo_uuid", "push-sub-v1")

	suite.Nil(e1)
	suite.True(qs1.Verified)

	// wrong response from remote endpoint
	s2 := Subscription{
		PushCfg: PushConfig{
			Pend:             "https://example_error.com/receive_here",
			VerificationHash: "vhash-1",
		},
	}

	c2 := &http.Client{
		Transport: new(MockPushRoundTripper),
	}

	e2 := VerifyPushEndpoint(s2, c2, nil)

	suite.Equal("Wrong response status code", e2.Error())

	// mismatch
	s3 := Subscription{
		PushCfg: PushConfig{
			Pend:             "https://example_mismatch.com/receive_here",
			VerificationHash: "vhash-1",
		},
	}

	c3 := &http.Client{
		Transport: new(MockPushRoundTripper),
	}

	e3 := VerifyPushEndpoint(s3, c3, nil)

	suite.Equal("Wrong verification hash", e3.Error())
}

func (suite *SubTestSuite) TestExportJson() {
	cfgAPI := config.NewAPICfg()
	cfgAPI.LoadStrJSON(suite.cfgStr)

	store := stores.NewMockStore(cfgAPI.StoreHost, cfgAPI.StoreDB)

	res, _ := Find("argo_uuid", "", "sub1", "", 0, store)

	outJSON, _ := res.Subscriptions[0].ExportJSON()
	expJSON := `{
   "name": "/projects/ARGO/subscriptions/sub1",
   "topic": "/projects/ARGO/topics/topic1",
   "pushConfig": {
      "type": "",
      "pushEndpoint": "",
      "maxMessages": 0,
      "authorizationHeader": {},
      "retryPolicy": {},
      "verificationHash": "",
      "verified": false,
      "mattermostUrl": "",
      "mattermostUsername": "",
      "mattermostChannel": ""
   },
   "ackDeadlineSeconds": 10,
   "createdOn": "2020-11-19T00:00:00Z"
}`
	suite.Equal(expJSON, outJSON)

	expJSON2 := `{
   "subscriptions": [
      {
         "name": "/projects/ARGO/subscriptions/sub4",
         "topic": "/projects/ARGO/topics/topic4",
         "pushConfig": {
            "type": "http_endpoint",
            "pushEndpoint": "endpoint.foo",
            "maxMessages": 1,
            "authorizationHeader": {
               "type": "autogen",
               "value": "auth-header-1"
            },
            "retryPolicy": {
               "type": "linear",
               "period": 300
            },
            "verificationHash": "push-id-1",
            "verified": true,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-22T00:00:00Z"
      },
      {
         "name": "/projects/ARGO/subscriptions/sub3",
         "topic": "/projects/ARGO/topics/topic3",
         "pushConfig": {
            "type": "",
            "pushEndpoint": "",
            "maxMessages": 0,
            "authorizationHeader": {},
            "retryPolicy": {},
            "verificationHash": "",
            "verified": false,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-21T00:00:00Z"
      },
      {
         "name": "/projects/ARGO/subscriptions/sub2",
         "topic": "/projects/ARGO/topics/topic2",
         "pushConfig": {
            "type": "",
            "pushEndpoint": "",
            "maxMessages": 0,
            "authorizationHeader": {},
            "retryPolicy": {},
            "verificationHash": "",
            "verified": false,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-20T00:00:00Z"
      },
      {
         "name": "/projects/ARGO/subscriptions/sub1",
         "topic": "/projects/ARGO/topics/topic1",
         "pushConfig": {
            "type": "",
            "pushEndpoint": "",
            "maxMessages": 0,
            "authorizationHeader": {},
            "retryPolicy": {},
            "verificationHash": "",
            "verified": false,
            "mattermostUrl": "",
            "mattermostUsername": "",
            "mattermostChannel": ""
         },
         "ackDeadlineSeconds": 10,
         "createdOn": "2020-11-19T00:00:00Z"
      }
   ],
   "nextPageToken": "",
   "totalSize": 4
}`
	results, _ := Find("argo_uuid", "", "", "", 0, store)
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
