package metric

import (
	"code/platform/v5/lock"
)

type ProcessMetric struct {
	pid    string
	cpu    float64
	memory float64
}

type Metric struct {
	agentMetric ProcessMetric
	plugins     map[string]map[string]ProcessMetric //pluginID:timestamp:ProcessMetric
}

type Manager struct {
	lock    lock.Table
	metrics map[string]*Metric // key: agentID
}
