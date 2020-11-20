package metrics

import (
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/stores"
	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/suite"
)

type MetricsTestSuite struct {
	suite.Suite
	cfgStr string
}

func (suite *MetricsTestSuite) SetupTest() {
	suite.cfgStr = `{
		"broker_host":"localhost:9092",
		"store_host":"localhost",
		"store_db":"argo_msg"
	}`
	log.SetOutput(ioutil.Discard)
}

func (suite *MetricsTestSuite) TestCreateMetric() {
	expJson := `{
   "metric": "project.number_of_topics",
   "metric_type": "counter",
   "value_type": "int64",
   "resource_type": "project",
   "resource_name": "test_project",
   "timeseries": [
      {
         "timestamp": "2017-06-23T03:42:44Z",
         "value": 32
      }
   ],
   "description": "Counter that displays the number of topics belonging to the specific project"
}`

	ts := "2017-06-23T03:42:44Z"
	myMetric := NewProjectTopics("test_project", 32, ts)
	outputJSON, _ := myMetric.ExportJSON()
	suite.Equal(expJson, outputJSON)
}

func (suite *MetricsTestSuite) TestCreateMetricList() {
	expJson := `{
   "metrics": [
      {
         "metric": "project.number_of_topics",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project",
         "resource_name": "test_project",
         "timeseries": [
            {
               "timestamp": "2017-06-23T03:42:44Z",
               "value": 32
            }
         ],
         "description": "Counter that displays the number of topics belonging to the specific project"
      }
   ]
}`
	ts := "2017-06-23T03:42:44Z"
	myMetric := NewProjectTopics("test_project", 32, ts)
	myList := NewMetricList(myMetric)
	outputJSON, _ := myList.ExportJSON()

	suite.Equal(expJson, outputJSON)
}

func (suite *MetricsTestSuite) TestOperational() {

	expJSON := `{
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
               "timestamp": "{{TS2}}",
               "value": {{VAL2}}
            }
         ],
         "description": "Percentage value that displays the Memory usage of ams service in the specific node"
      }
   ]
}`

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	ml, _ := GetUsageCpuMem(store)
	outJSON, _ := ml.ExportJSON()

	ts1 := ml.Metrics[0].Timeseries[0].Timestamp
	ts2 := ml.Metrics[1].Timeseries[0].Timestamp
	val1 := ml.Metrics[0].Timeseries[0].Value.(float64)
	val2 := ml.Metrics[1].Timeseries[0].Value.(float64)
	host := ml.Metrics[0].Resource
	expJSON = strings.Replace(expJSON, "{{TS1}}", ts1, -1)
	expJSON = strings.Replace(expJSON, "{{TS2}}", ts2, -1)
	expJSON = strings.Replace(expJSON, "{{VAL1}}", strconv.FormatFloat(val1, 'g', 1, 64), -1)
	expJSON = strings.Replace(expJSON, "{{VAL2}}", strconv.FormatFloat(val2, 'g', 1, 64), -1)
	expJSON = strings.Replace(expJSON, "{{HOST}}", host, -1)

	suite.Equal(expJSON, outJSON)

}

func (suite *MetricsTestSuite) TestGetTopics() {

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	n, _ := GetProjectTopics("argo_uuid", store)
	suite.Equal(int64(4), n)

}
func (suite *MetricsTestSuite) TestGetTopicsACL() {

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	n, _ := GetProjectTopicsACL("argo_uuid", "uuid1", store)
	suite.Equal(int64(2), n)

}

func (suite *MetricsTestSuite) TestGetSubs() {

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	n, _ := GetProjectSubs("argo_uuid", store)
	suite.Equal(int64(4), n)

}
func (suite *MetricsTestSuite) TestGetSubsACL() {

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	n, _ := GetProjectSubsACL("argo_uuid", "uuid1", store)
	suite.Equal(int64(3), n)

}

func (suite *MetricsTestSuite) TestGetSubsByTopic() {

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	n, _ := GetProjectSubsByTopic("argo_uuid", "topic1", store)
	suite.Equal(int64(1), n)

}

func (suite *MetricsTestSuite) TestAggrProjectUserSubTest() {

	expJSON := `{
   "metrics": [
      {
         "metric": "project.user.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserA",
         "timeseries": [
            {
               "timestamp": "{{TS1}}",
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
               "timestamp": "{{TS2}}",
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
               "timestamp": "{{TS3}}",
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
               "timestamp": "{{TS4}}",
               "value": 2
            }
         ],
         "description": "Counter that displays the number of subscriptions that a user has access to the specific project"
      }
   ]
}`

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	ml, _ := AggrProjectUserSubs("argo_uuid", store)

	ts1 := ml.Metrics[0].Timeseries[0].Timestamp
	ts2 := ml.Metrics[1].Timeseries[0].Timestamp
	ts3 := ml.Metrics[2].Timeseries[0].Timestamp
	ts4 := ml.Metrics[3].Timeseries[0].Timestamp

	expJSON = strings.Replace(expJSON, "{{TS1}}", ts1, -1)
	expJSON = strings.Replace(expJSON, "{{TS2}}", ts2, -1)
	expJSON = strings.Replace(expJSON, "{{TS3}}", ts3, -1)
	expJSON = strings.Replace(expJSON, "{{TS4}}", ts4, -1)

	outJSON, _ := ml.ExportJSON()

	suite.Equal(expJSON, outJSON)

}

func (suite *MetricsTestSuite) TestAggrProjectUserTopicsTest() {

	expJSON := `{
   "metrics": [
      {
         "metric": "project.user.number_of_topics",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserA",
         "timeseries": [
            {
               "timestamp": "{{TS1}}",
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
               "timestamp": "{{TS2}}",
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
               "timestamp": "{{TS3}}",
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
               "timestamp": "{{TS4}}",
               "value": 1
            }
         ],
         "description": "Counter that displays the number of topics that a user has access to the specific project"
      }
   ]
}`

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	ml, _ := AggrProjectUserTopics("argo_uuid", store)

	ts1 := ml.Metrics[0].Timeseries[0].Timestamp
	ts2 := ml.Metrics[0].Timeseries[0].Timestamp
	ts3 := ml.Metrics[0].Timeseries[0].Timestamp
	ts4 := ml.Metrics[0].Timeseries[0].Timestamp

	expJSON = strings.Replace(expJSON, "{{TS1}}", ts1, -1)
	expJSON = strings.Replace(expJSON, "{{TS2}}", ts2, -1)
	expJSON = strings.Replace(expJSON, "{{TS3}}", ts3, -1)
	expJSON = strings.Replace(expJSON, "{{TS4}}", ts4, -1)

	outJSON, _ := ml.ExportJSON()

	suite.Equal(expJSON, outJSON)

}

func (suite *MetricsTestSuite) TestGetProjectsMessageCount() {

	store := stores.NewMockStore("", "")
	store.Initialize()

	// test total message count per project
	expectedTmpc := TotalProjectsMessageCount{
		Projects: []ProjectMessageCount{
			{
				Project:              "ARGO",
				MessageCount:         60,
				AverageDailyMessages: 15,
			},
		},
		TotalCount:           60,
		AverageDailyMessages: 15,
	}

	tmpc, tmpcerr := GetProjectsMessageCount(
		[]string{"ARGO"},
		time.Date(2018, 10, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2018, 10, 4, 0, 0, 0, 0, time.UTC),
		store,
	)

	suite.Equal(expectedTmpc, tmpc)
	suite.Nil(tmpcerr)
}

func (suite *MetricsTestSuite) TestGetVAReport() {

	store := stores.NewMockStore("", "")
	store.Initialize()

	// test total message count per project
	expectedTmpc := TotalProjectsMessageCount{
		Projects: []ProjectMessageCount{
			{
				Project:              "ARGO",
				MessageCount:         280,
				AverageDailyMessages: 0,
			},
		},
		TotalCount:           280,
		AverageDailyMessages: 0,
	}

	va, tmpcerr := GetVAReport(
		[]string{"ARGO"},
		time.Date(2007, 10, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, 20, 4, 0, 0, 0, 0, time.UTC),
		store,
	)

	expectedVA := VAReport{
		ProjectsMetrics:    expectedTmpc,
		UsersCount:         18,
		TopicsCount:        8,
		SubscriptionsCount: 8,
	}

	suite.Equal(expectedVA, va)
	suite.Nil(tmpcerr)
}

func TestMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(MetricsTestSuite))
}
