package metrics

import (
	"encoding/json"
	"time"
)

// Metric names and descriptions
const (
	DescProjectTopics string = "Counter that displays the number of topics belonging to the specific project"
	NameProjectTopics string = "project.number_of_topics"
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
