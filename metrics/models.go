package metrics

import (
	"encoding/json"
	"time"
)

// Metric names and descriptions
const (
	DescProjectTopics     string = "Counter that displays the number of topics belonging to the specific project"
	NameProjectTopics     string = "project.number_of_topics"
	DescProjectSubs       string = "Counter that displays the number of subscriptions belonging to the specific project"
	NameProjectSubs       string = "project.number_of_subscriptions"
	DescTopicSubs         string = "Counter that displays the number of subscriptions belonging to a specific topic"
	NameTopicSubs         string = "topic.number_of_subscriptions"
	DescTopicMsgs         string = "Counter that displays the number number of messages published to the specific topic"
	NameTopicMsgs         string = "topic.number_of_messages"
	DescTopicBytes        string = "Counter that displays the total size of data (in bytes) published to the specific topic"
	NameTopicBytes        string = "topic.number_of_bytes"
	DescProjectUserSubs   string = "Counter that displays the number of subscriptions that a user has access to the specific project"
	NameProjectUserSubs   string = "project.user.number_of_subscriptions"
	DescProjectUserTopics string = "Counter that displays the number of topics that a user has access to the specific project"
	NameProjectUserTopics string = "project.user.number_of_topics"
	DescSubMsgs           string = "Counter that displays the number number of messages consumed from the specific subscription"
	NameSubMsgs           string = "subscription.number_of_messages"
	DescSubBytes          string = "Counter that displays the total size of data (in bytes) consumed from the specific subscription"
	NameSubBytes          string = "subscription.number_of_bytes"
	DescOpNodeCPU         string = "Percentage value that displays the CPU usage of ams service in the specific node"
	NameOpNodeCPU         string = "ams_node.cpu_usage"
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

func NewTopicSubs(topic string, value int64, tstamp string) Metric {
	// Initialize single point timeseries with the latest timestamp and value
	ts := []Timepoint{Timepoint{Timestamp: tstamp, Value: value}}
	m := Metric{Metric: NameTopicSubs, MetricType: "counter", ValueType: "int64", ResourceType: "topic", Resource: topic, Timeseries: ts, Description: DescTopicSubs}
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

// GetUserFromJSON retrieves User info From JSON string
func GetMetricsFromJSON(input []byte) (MetricList, error) {
	ml := MetricList{}
	err := json.Unmarshal([]byte(input), &ml)
	return ml, err
}

func GetTimeNowZulu() string {
	zSec := "2006-01-02T15:04:05Z"
	t := time.Now()
	ts := t.Format(zSec)
	return ts
}
