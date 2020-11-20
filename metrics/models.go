package metrics

import (
	"encoding/json"
	"time"
)

// Metric names and descriptions
const (
	DescProjectTopics     = "Counter that displays the number of topics belonging to the specific project"
	NameProjectTopics     = "project.number_of_topics"
	DescProjectSubs       = "Counter that displays the number of subscriptions belonging to the specific project"
	NameProjectSubs       = "project.number_of_subscriptions"
	NameDailyProjectMsgs  = "project.number_of_daily_messages"
	DescDailyProjectMsgs  = "A collection of counters that represents the total number of messages published each day to all of the project's topics"
	DescTopicSubs         = "Counter that displays the number of subscriptions belonging to a specific topic"
	NameTopicSubs         = "topic.number_of_subscriptions"
	DescTopicRate         = "A rate that displays how many messages were published per second between the last two publish events"
	NameTopicRate         = "topic.publishing_rate"
	DescSubRate           = "A rate that displays how many messages were consumed per second between the last two consume events"
	NameSubRate           = "subscription.consumption_rate"
	DescTopicMsgs         = "Counter that displays the number of messages published to the specific topic"
	NameTopicMsgs         = "topic.number_of_messages"
	DescDailyTopicMsgs    = "A collection of counters that represents the total number of messages published each day to a specific topic"
	NameDailyTopicMsgs    = "topic.number_of_daily_messages"
	DescTopicBytes        = "Counter that displays the total size of data (in bytes) published to the specific topic"
	NameTopicBytes        = "topic.number_of_bytes"
	DescProjectUserSubs   = "Counter that displays the number of subscriptions that a user has access to the specific project"
	NameProjectUserSubs   = "project.user.number_of_subscriptions"
	DescProjectUserTopics = "Counter that displays the number of topics that a user has access to the specific project"
	NameProjectUserTopics = "project.user.number_of_topics"
	DescSubMsgs           = "Counter that displays the number of messages consumed from the specific subscription"
	NameSubMsgs           = "subscription.number_of_messages"
	DescSubBytes          = "Counter that displays the total size of data (in bytes) consumed from the specific subscription"
	NameSubBytes          = "subscription.number_of_bytes"
	DescOpNodeCPU         = "Percentage value that displays the CPU usage of ams service in the specific node"
	NameOpNodeCPU         = "ams_node.cpu_usage"
	DescOpNodeMEM         = "Percentage value that displays the Memory usage of ams service in the specific node"
	NameOpNodeMEM         = "ams_node.memory_usage"
)

type MetricList struct {
	Metrics []Metric `json:"metrics"`
}

type Metric struct {
	Metric       string      `json:"metric"`
	MetricType   string      `json:"metric_type"`
	ValueType    string      `json:"value_type"`
	ResourceType string      `json:"resource_type"`
	Resource     string      `json:"resource_name"`
	Timeseries   []Timepoint `json:"timeseries"`
	Description  string      `json:"description"`
}

type ProjectMessageCount struct {
	Project              string  `json:"project"`
	MessageCount         int64   `json:"message_count"`
	AverageDailyMessages float64 `json:"average_daily_messages"`
}

type TotalProjectsMessageCount struct {
	Projects             []ProjectMessageCount `json:"projects"`
	TotalCount           int64                 `json:"total_message_count"`
	AverageDailyMessages float64               `json:"average_daily_messages"`
}

type VAReport struct {
	ProjectsMetrics    TotalProjectsMessageCount `json:"projects_metrics"`
	UsersCount         int                       `json:"users_count"`
	TopicsCount        int                       `json:"topics_count"`
	SubscriptionsCount int                       `json:"subscriptions_count"`
}

type Timepoint struct {
	Timestamp string      `json:"timestamp"`
	Value     interface{} `json:"value"`
}

// ExportJSON exports whole ProjectTopic structure
func (m *Metric) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(m, "", "   ")
	return string(output[:]), err
}

// ExportJSON exports whole ProjectTopic structure
func (ml *MetricList) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(ml, "", "   ")
	return string(output[:]), err
}

func NewMetricList(m Metric) MetricList {
	ml := MetricList{Metrics: []Metric{m}}
	return ml
}

func NewProjectTopics(project string, value int64, tstamp string) Metric {
	// Initialize single point timeseries with the latest timestamp and value
	ts := []Timepoint{Timepoint{Timestamp: tstamp, Value: value}}
	m := Metric{Metric: NameProjectTopics, MetricType: "counter", ValueType: "int64", ResourceType: "project", Resource: project, Timeseries: ts, Description: DescProjectTopics}
	return m
}

func NewProjectSubs(project string, value int64, tstamp string) Metric {
	// Initialize single point timeseries with the latest timestamp and value
	ts := []Timepoint{Timepoint{Timestamp: tstamp, Value: value}}
	m := Metric{Metric: NameProjectSubs, MetricType: "counter", ValueType: "int64", ResourceType: "project", Resource: project, Timeseries: ts, Description: DescProjectSubs}
	return m
}

func NewDailyProjectMsgCount(project string, timePoints []Timepoint) Metric {
	m := Metric{Metric: NameDailyProjectMsgs, MetricType: "counter", ValueType: "int64", ResourceType: "project", Resource: project, Timeseries: timePoints, Description: DescDailyProjectMsgs}
	return m
}

func NewTopicSubs(topic string, value int64, tstamp string) Metric {
	// Initialize single point timeseries with the latest timestamp and value
	ts := []Timepoint{Timepoint{Timestamp: tstamp, Value: value}}
	m := Metric{Metric: NameTopicSubs, MetricType: "counter", ValueType: "int64", ResourceType: "topic", Resource: topic, Timeseries: ts, Description: DescTopicSubs}
	return m
}

func NewTopicRate(topic string, value float64, tstamp string) Metric {
	// Initialize single point timeseries with the latest timestamp and value
	ts := []Timepoint{Timepoint{Timestamp: tstamp, Value: value}}
	m := Metric{Metric: NameTopicRate, MetricType: "rate", ValueType: "float64", ResourceType: "topic", Resource: topic, Timeseries: ts, Description: DescTopicRate}
	return m
}

func NewSubRate(sub string, value float64, tstamp string) Metric {
	// Initialize single point timeseries with the latest timestamp and value
	ts := []Timepoint{Timepoint{Timestamp: tstamp, Value: value}}
	m := Metric{Metric: NameSubRate, MetricType: "rate", ValueType: "float64", ResourceType: "subscription", Resource: sub, Timeseries: ts, Description: DescSubRate}
	return m
}

func NewSubMsgs(topic string, value int64, tstamp string) Metric {
	// Initialize single point timeseries with the latest timestamp and value
	ts := []Timepoint{Timepoint{Timestamp: tstamp, Value: value}}

	m := Metric{Metric: NameSubMsgs, MetricType: "counter", ValueType: "int64", ResourceType: "subscription", Resource: topic, Timeseries: ts, Description: DescSubMsgs}

	return m
}

func NewSubBytes(topic string, value int64, tstamp string) Metric {
	// Initialize single point timeseries with the latest timestamp and value
	ts := []Timepoint{Timepoint{Timestamp: tstamp, Value: value}}
	m := Metric{Metric: NameSubBytes, MetricType: "counter", ValueType: "int64", ResourceType: "subscription", Resource: topic, Timeseries: ts, Description: DescSubBytes}

	return m
}

func NewTopicMsgs(topic string, value int64, tstamp string) Metric {
	// Initialize single point timeseries with the latest timestamp and value
	ts := []Timepoint{Timepoint{Timestamp: tstamp, Value: value}}
	m := Metric{Metric: NameTopicMsgs, MetricType: "counter", ValueType: "int64", ResourceType: "topic", Resource: topic, Timeseries: ts, Description: DescTopicMsgs}
	return m
}

func NewDailyTopicMsgCount(topic string, timePoints []Timepoint) Metric {

	m := Metric{Metric: NameDailyTopicMsgs, MetricType: "counter", ValueType: "int64", ResourceType: "topic", Resource: topic, Timeseries: timePoints, Description: DescDailyTopicMsgs}
	return m
}

func NewTopicBytes(topic string, value int64, tstamp string) Metric {
	// Initialize single point timeseries with the latest timestamp and value
	ts := []Timepoint{Timepoint{Timestamp: tstamp, Value: value}}
	m := Metric{Metric: NameTopicBytes, MetricType: "counter", ValueType: "int64", ResourceType: "topic", Resource: topic, Timeseries: ts, Description: DescTopicBytes}
	return m
}

func NewProjectUserSubs(project string, user string, value int64, tstamp string) Metric {
	// Initialize single point timeseries with the latest timestamp and value
	ts := []Timepoint{Timepoint{Timestamp: tstamp, Value: value}}
	m := Metric{Metric: NameProjectUserSubs, MetricType: "counter", ValueType: "int64", ResourceType: "project.user", Resource: project + "." + user, Timeseries: ts, Description: DescProjectUserSubs}

	return m
}

func NewProjectUserTopics(project string, user string, value int64, tstamp string) Metric {
	// Initialize single point timeseries with the latest timestamp and value
	ts := []Timepoint{Timepoint{Timestamp: tstamp, Value: value}}
	m := Metric{Metric: NameProjectUserTopics, MetricType: "counter", ValueType: "int64", ResourceType: "project.user", Resource: project + "." + user, Timeseries: ts, Description: DescProjectUserTopics}

	return m
}

// Initialize single point timeseries with the latest timestamp and value
func NewOpNodeCPU(hostname string, value float64, tstamp string) Metric {
	ts := []Timepoint{Timepoint{Timestamp: tstamp, Value: value}}
	m := Metric{Metric: NameOpNodeCPU, MetricType: "percentage", ValueType: "float64", ResourceType: "ams_node", Resource: hostname, Timeseries: ts, Description: DescOpNodeCPU}

	return m
}

func NewOpNodeMEM(hostname string, value float64, tstamp string) Metric {
	ts := []Timepoint{Timepoint{Timestamp: tstamp, Value: value}}
	m := Metric{Metric: NameOpNodeMEM, MetricType: "percentage", ValueType: "float64", ResourceType: "ams_node", Resource: hostname, Timeseries: ts, Description: DescOpNodeMEM}

	return m
}

// GetUserFromJSON retrieves User info From JSON string
func GetMetricsFromJSON(input []byte) (MetricList, error) {
	ml := MetricList{}
	err := json.Unmarshal([]byte(input), &ml)
	return ml, err
}

func GetTimeNowZulu() string {
	zSec := "2006-01-02T15:04:05Z"
	t := time.Now().UTC()
	ts := t.Format(zSec)
	return ts
}
