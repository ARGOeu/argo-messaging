package handlers

import (
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/metrics"
	oldPush "github.com/ARGOeu/argo-messaging/push"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

type MetricsHandlersTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *MetricsHandlersTestSuite) SetupTest() {
	suite.cfgStr = `{
	"bind_ip":"",
	"port":8080,
	"zookeeper_hosts":["localhost"],
	"kafka_znode":"",
	"store_host":"localhost",
	"store_db":"argo_msg",
	"certificate":"/etc/pki/tls/certs/localhost.crt",
	"certificate_key":"/etc/pki/tls/private/localhost.key",
	"per_resource_auth":"true",
	"push_enabled": "true",
	"push_worker_token": "push_token"
	}`
}

func (suite *MetricsHandlersTestSuite) TestProjectMessageCount() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/metrics/va_metrics?start_date=2018-10-01&end_date=2018-10-04", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
 "projects_metrics": {
  "projects": [
   {
    "project": "ARGO",
    "message_count": 30,
    "average_daily_messages": 7,
    "topics_count": 0,
    "subscriptions_count": 0,
    "users_count": 0
   }
  ],
  "total_message_count": 30,
  "average_daily_messages": 7
 },
 "total_users_count": 0,
 "total_topics_count": 0,
 "total_subscriptions_count": 0
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/metrics/va_metrics", WrapMockAuthConfig(VaMetrics, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *MetricsHandlersTestSuite) TestVaReportFull() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/metrics/va_metrics?start_date=2007-10-01&end_date=2020-11-24", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
 "projects_metrics": {
  "projects": [
   {
    "project": "ARGO",
    "message_count": 140,
    "average_daily_messages": 0,
    "topics_count": 4,
    "subscriptions_count": 4,
    "users_count": 7
   }
  ],
  "total_message_count": 140,
  "average_daily_messages": 0
 },
 "total_users_count": 7,
 "total_topics_count": 4,
 "total_subscriptions_count": 4
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/metrics/va_metrics", WrapMockAuthConfig(VaMetrics, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *MetricsHandlersTestSuite) TestProjectMessageCountErrors() {

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects-message-count", WrapMockAuthConfig(VaMetrics, cfgKafka, &brk, str, &mgr, nil))

	// wrong start date
	expResp1 := `{
   "error": {
      "code": 400,
      "message": "Start date is not in valid format",
      "status": "INVALID_ARGUMENT"
   }
}`
	req1, err := http.NewRequest("GET", "http://localhost:8080/v1/projects-message-count?start_date=ffff", nil)
	if err != nil {
		log.Fatal(err)
	}
	router.ServeHTTP(w, req1)
	suite.Equal(400, w.Code)
	suite.Equal(expResp1, w.Body.String())
	w.Body.Reset()

	// wrong end date
	expResp2 := `{
   "error": {
      "code": 400,
      "message": "End date is not in valid format",
      "status": "INVALID_ARGUMENT"
   }
}`
	req2, err := http.NewRequest("GET", "http://localhost:8080/v1/projects-message-count?end_date=ffff", nil)
	if err != nil {
		log.Fatal(err)
	}
	router.ServeHTTP(w, req2)
	suite.Equal(400, w.Code)
	suite.Equal(expResp2, w.Body.String())
	w.Body.Reset()

	// one of the projects doesn't exist end date
	expResp3 := `{
   "error": {
      "code": 404,
      "message": "Project ffff doesn't exist",
      "status": "NOT_FOUND"
   }
}`
	req3, err := http.NewRequest("GET", "http://localhost:8080/v1/projects-message-count?projects=ARGO,ffff", nil)
	if err != nil {
		log.Fatal(err)
	}
	router.ServeHTTP(w, req3)
	suite.Equal(400, w.Code)
	suite.Equal(expResp3, w.Body.String())
	w.Body.Reset()

	// start date is off
	expResp4 := `{
   "error": {
      "code": 400,
      "message": "Start date cannot be after the end date",
      "status": "INVALID_ARGUMENT"
   }
}`
	req4, err := http.NewRequest("GET", "http://localhost:8080/v1/projects-message-count?start_date=2019-04-04&end_date=2018-01-01", nil)
	if err != nil {
		log.Fatal(err)
	}
	router.ServeHTTP(w, req4)
	suite.Equal(400, w.Code)
	suite.Equal(expResp4, w.Body.String())
}

func (suite *MetricsHandlersTestSuite) TestSubMetrics() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions/sub1:metrics", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "metrics": [
      {
         "metric": "subscription.number_of_messages",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "subscription",
         "resource_name": "sub1",
         "timeseries": [
            {
               "timestamp": "{{TS1}}",
               "value": 0
            }
         ],
         "description": "Counter that displays the number of messages consumed from the specific subscription"
      },
      {
         "metric": "subscription.number_of_bytes",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "subscription",
         "resource_name": "sub1",
         "timeseries": [
            {
               "timestamp": "{{TS2}}",
               "value": 0
            }
         ],
         "description": "Counter that displays the total size of data (in bytes) consumed from the specific subscription"
      },
      {
         "metric": "subscription.consumption_rate",
         "metric_type": "rate",
         "value_type": "float64",
         "resource_type": "subscription",
         "resource_name": "sub1",
         "timeseries": [
            {
               "timestamp": "2019-05-06T00:00:00Z",
               "value": 10
            }
         ],
         "description": "A rate that displays how many messages were consumed per second between the last two consume events"
      }
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:metrics", WrapMockAuthConfig(SubMetrics, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)

	metricOut, _ := metrics.GetMetricsFromJSON([]byte(w.Body.String()))
	ts1 := metricOut.Metrics[0].Timeseries[0].Timestamp
	ts2 := metricOut.Metrics[1].Timeseries[0].Timestamp
	expResp = strings.Replace(expResp, "{{TS1}}", ts1, -1)
	expResp = strings.Replace(expResp, "{{TS2}}", ts2, -1)
	suite.Equal(expResp, w.Body.String())

}

func (suite *MetricsHandlersTestSuite) TestSubMetricsNotFound() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/subscriptions/unknown_sub:metrics", nil)
	if err != nil {
		log.Fatal(err)
	}

	expRes := `{
   "error": {
      "code": 404,
      "message": "Subscription doesn't exist",
      "status": "NOT_FOUND"
   }
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	// temporarily disable auth for this test case
	cfgKafka.ResAuth = false
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}:metrics", WrapMockAuthConfig(SubMetrics, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expRes, w.Body.String())

}

func (suite *MetricsHandlersTestSuite) TestProjectMetrics() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO:metrics", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "metrics": [
      {
         "metric": "project.number_of_topics",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project",
         "resource_name": "ARGO",
         "timeseries": [
            {
               "timestamp": "{{TS1}}",
               "value": 4
            }
         ],
         "description": "Counter that displays the number of topics belonging to the specific project"
      },
      {
         "metric": "project.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project",
         "resource_name": "ARGO",
         "timeseries": [
            {
               "timestamp": "{{TS2}}",
               "value": 4
            }
         ],
         "description": "Counter that displays the number of subscriptions belonging to the specific project"
      },
      {
         "metric": "project.user.number_of_topics",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserA",
         "timeseries": [
            {
               "timestamp": "{{TS3}}",
               "value": 2
            }
         ],
         "description": "Counter that displays the number of topics that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_topics",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserB",
         "timeseries": [
            {
               "timestamp": "{{TS4}}",
               "value": 2
            }
         ],
         "description": "Counter that displays the number of topics that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_topics",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserX",
         "timeseries": [
            {
               "timestamp": "{{TS5}}",
               "value": 1
            }
         ],
         "description": "Counter that displays the number of topics that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_topics",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserZ",
         "timeseries": [
            {
               "timestamp": "{{TS6}}",
               "value": 1
            }
         ],
         "description": "Counter that displays the number of topics that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserA",
         "timeseries": [
            {
               "timestamp": "{{TS7}}",
               "value": 3
            }
         ],
         "description": "Counter that displays the number of subscriptions that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserB",
         "timeseries": [
            {
               "timestamp": "{{TS8}}",
               "value": 3
            }
         ],
         "description": "Counter that displays the number of subscriptions that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserX",
         "timeseries": [
            {
               "timestamp": "{{TS9}}",
               "value": 1
            }
         ],
         "description": "Counter that displays the number of subscriptions that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserZ",
         "timeseries": [
            {
               "timestamp": "{{TS10}}",
               "value": 2
            }
         ],
         "description": "Counter that displays the number of subscriptions that a user has access to the specific project"
      },
      {
         "metric": "project.number_of_daily_messages",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project",
         "resource_name": "ARGO",
         "timeseries": [
            {
               "timestamp": "{{TS11}}",
               "value": 30
            },
            {
               "timestamp": "{{TS12}}",
               "value": 110
            }
         ],
         "description": "A collection of counters that represents the total number of messages published each day to all of the project's topics"
      }
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}:metrics", WrapMockAuthConfig(ProjectMetrics, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	metricOut, _ := metrics.GetMetricsFromJSON([]byte(w.Body.String()))
	ts1 := metricOut.Metrics[0].Timeseries[0].Timestamp
	ts2 := metricOut.Metrics[1].Timeseries[0].Timestamp
	ts3 := metricOut.Metrics[2].Timeseries[0].Timestamp
	ts4 := metricOut.Metrics[3].Timeseries[0].Timestamp
	ts5 := metricOut.Metrics[4].Timeseries[0].Timestamp
	ts6 := metricOut.Metrics[5].Timeseries[0].Timestamp
	ts7 := metricOut.Metrics[6].Timeseries[0].Timestamp
	ts8 := metricOut.Metrics[7].Timeseries[0].Timestamp
	ts9 := metricOut.Metrics[8].Timeseries[0].Timestamp
	ts10 := metricOut.Metrics[9].Timeseries[0].Timestamp
	ts11 := metricOut.Metrics[10].Timeseries[0].Timestamp
	ts12 := metricOut.Metrics[10].Timeseries[1].Timestamp
	expResp = strings.Replace(expResp, "{{TS1}}", ts1, -1)
	expResp = strings.Replace(expResp, "{{TS2}}", ts2, -1)
	expResp = strings.Replace(expResp, "{{TS3}}", ts3, -1)
	expResp = strings.Replace(expResp, "{{TS4}}", ts4, -1)
	expResp = strings.Replace(expResp, "{{TS5}}", ts5, -1)
	expResp = strings.Replace(expResp, "{{TS6}}", ts6, -1)
	expResp = strings.Replace(expResp, "{{TS7}}", ts7, -1)
	expResp = strings.Replace(expResp, "{{TS8}}", ts8, -1)
	expResp = strings.Replace(expResp, "{{TS9}}", ts9, -1)
	expResp = strings.Replace(expResp, "{{TS10}}", ts10, -1)
	expResp = strings.Replace(expResp, "{{TS11}}", ts11, -1)
	expResp = strings.Replace(expResp, "{{TS12}}", ts12, -1)
	suite.Equal(expResp, w.Body.String())

}

func (suite *MetricsHandlersTestSuite) TestOpMetrics() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/metrics", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "metrics": [
      {
         "metric": "ams_node.cpu_usage",
         "metric_type": "percentage",
         "value_type": "float64",
         "resource_type": "ams_node",
         "resource_name": "{{HOST}}",
         "timeseries": [
            {
               "timestamp": "{{TS1}}",
               "value": {{VAL1}}
            }
         ],
         "description": "Percentage value that displays the CPU usage of ams service in the specific node"
      },
      {
         "metric": "ams_node.memory_usage",
         "metric_type": "percentage",
         "value_type": "float64",
         "resource_type": "ams_node",
         "resource_name": "{{HOST}}",
         "timeseries": [
            {
               "timestamp": "{{TS1}}",
               "value": {{VAL2}}
            }
         ],
         "description": "Percentage value that displays the Memory usage of ams service in the specific node"
      }
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/metrics", WrapMockAuthConfig(OpMetrics, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	metricOut, _ := metrics.GetMetricsFromJSON([]byte(w.Body.String()))
	ts1 := metricOut.Metrics[0].Timeseries[0].Timestamp
	val1 := metricOut.Metrics[0].Timeseries[0].Value.(float64)
	ts2 := metricOut.Metrics[1].Timeseries[0].Timestamp
	val2 := metricOut.Metrics[1].Timeseries[0].Value.(float64)
	host := metricOut.Metrics[0].Resource
	expResp = strings.Replace(expResp, "{{TS1}}", ts1, -1)
	expResp = strings.Replace(expResp, "{{TS2}}", ts2, -1)
	expResp = strings.Replace(expResp, "{{VAL1}}", strconv.FormatFloat(val1, 'g', 1, 64), -1)
	expResp = strings.Replace(expResp, "{{VAL2}}", strconv.FormatFloat(val2, 'g', 1, 64), -1)
	expResp = strings.Replace(expResp, "{{HOST}}", host, -1)
	suite.Equal(expResp, w.Body.String())

}

func (suite *MetricsHandlersTestSuite) TestTopicMetrics() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics/topic1:metrics", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "metrics": [
      {
         "metric": "topic.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "topic",
         "resource_name": "topic1",
         "timeseries": [
            {
               "timestamp": "{{TIMESTAMP1}}",
               "value": 1
            }
         ],
         "description": "Counter that displays the number of subscriptions belonging to a specific topic"
      },
      {
         "metric": "topic.number_of_messages",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "topic",
         "resource_name": "topic1",
         "timeseries": [
            {
               "timestamp": "{{TIMESTAMP2}}",
               "value": 0
            }
         ],
         "description": "Counter that displays the number of messages published to the specific topic"
      },
      {
         "metric": "topic.number_of_bytes",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "topic",
         "resource_name": "topic1",
         "timeseries": [
            {
               "timestamp": "{{TIMESTAMP3}}",
               "value": 0
            }
         ],
         "description": "Counter that displays the total size of data (in bytes) published to the specific topic"
      },
      {
         "metric": "topic.number_of_daily_messages",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "topic",
         "resource_name": "topic1",
         "timeseries": [
            {
               "timestamp": "{{TIMESTAMP4}}",
               "value": 30
            },
            {
               "timestamp": "{{TIMESTAMP5}}",
               "value": 40
            }
         ],
         "description": "A collection of counters that represents the total number of messages published each day to a specific topic"
      },
      {
         "metric": "topic.publishing_rate",
         "metric_type": "rate",
         "value_type": "float64",
         "resource_type": "topic",
         "resource_name": "topic1",
         "timeseries": [
            {
               "timestamp": "2019-05-06T00:00:00Z",
               "value": 10
            }
         ],
         "description": "A rate that displays how many messages were published per second between the last two publish events"
      }
   ]
}`

	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:metrics", WrapMockAuthConfig(TopicMetrics, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(200, w.Code)
	metricOut, _ := metrics.GetMetricsFromJSON([]byte(w.Body.String()))
	ts1 := metricOut.Metrics[0].Timeseries[0].Timestamp
	ts2 := metricOut.Metrics[1].Timeseries[0].Timestamp
	ts3 := metricOut.Metrics[2].Timeseries[0].Timestamp
	ts4 := metricOut.Metrics[3].Timeseries[0].Timestamp
	ts5 := metricOut.Metrics[3].Timeseries[1].Timestamp
	expResp = strings.Replace(expResp, "{{TIMESTAMP1}}", ts1, -1)
	expResp = strings.Replace(expResp, "{{TIMESTAMP2}}", ts2, -1)
	expResp = strings.Replace(expResp, "{{TIMESTAMP3}}", ts3, -1)
	expResp = strings.Replace(expResp, "{{TIMESTAMP4}}", ts4, -1)
	expResp = strings.Replace(expResp, "{{TIMESTAMP5}}", ts5, -1)

	suite.Equal(expResp, w.Body.String())

}

func (suite *MetricsHandlersTestSuite) TestTopicMetricsNotFound() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics/topic_not_found:metrics", nil)
	if err != nil {
		log.Fatal(err)
	}

	expRes := `{
   "error": {
      "code": 404,
      "message": "Topic doesn't exist",
      "status": "NOT_FOUND"
   }
}`
	cfgKafka := config.NewAPICfg()
	cfgKafka.LoadStrJSON(suite.cfgStr)
	// deactivate auth for this specific test case
	cfgKafka.ResAuth = false
	brk := brokers.MockBroker{}
	str := stores.NewMockStore("whatever", "argo_mgs")
	router := mux.NewRouter().StrictSlash(true)
	w := httptest.NewRecorder()
	mgr := oldPush.Manager{}
	router.HandleFunc("/v1/projects/{project}/topics/{topic}:metrics", WrapMockAuthConfig(TopicMetrics, cfgKafka, &brk, str, &mgr, nil))
	router.ServeHTTP(w, req)
	suite.Equal(404, w.Code)
	suite.Equal(expRes, w.Body.String())

}

func TestMetricsHandlersTestSuite(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	suite.Run(t, new(MetricsHandlersTestSuite))
}
