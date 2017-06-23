package metrics

import (
	"io/ioutil"
	"testing"

	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/stores"
	log "github.com/Sirupsen/logrus"

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

func (suite *MetricsTestSuite) TestGetTopics() {

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	n, _ := GetProjectTopics("argo_uuid", store)
	suite.Equal(int64(3), n)

}
func (suite *MetricsTestSuite) TestGetTopicsACL() {

	APIcfg := config.NewAPICfg()
	APIcfg.LoadStrJSON(suite.cfgStr)
	store := stores.NewMockStore(APIcfg.StoreHost, APIcfg.StoreDB)
	n, _ := GetProjectTopicsACL("argo_uuid", "uuid1", store)
	suite.Equal(int64(2), n)

}

func TestMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(MetricsTestSuite))
}
