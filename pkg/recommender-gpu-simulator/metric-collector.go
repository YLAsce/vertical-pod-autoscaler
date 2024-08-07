package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

var (
	metricsSummaryIgnoreHead = flag.Int("metrics-summary-ignore-head", 1800, "ignore the first metrics when calculating summary")
	metricsFolder            = flag.String("metrics-folder", "group", "metrics parent folder")
)

// 单个pod的指标

type metricPoint struct {
	SMUsage       model.ResourceAmount
	MemoryUsage   model.ResourceAmount
	MemoryOverrun bool
	SMOverrun     bool
	Ts            time.Time
	NodeId        int
	RequestId     float64
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (m *metricPoint) String() string {
	return fmt.Sprintf("%v,%d,%d,%d,%f", m.Ts.Unix(), bool2int(m.SMOverrun), bool2int(m.MemoryOverrun), m.NodeId, m.RequestId)
}

type MetricsCollector struct {
	metricArr []metricPoint
	name      string
}

func NewMetricsCollector(name string) *MetricsCollector {
	return &MetricsCollector{
		metricArr: make([]metricPoint, 0),
		name:      name,
	}
}

func (m *MetricsCollector) Record(metric metricPoint) {
	m.metricArr = append(m.metricArr, metric)
}

func (m *MetricsCollector) Dump() {

	fullPath := "metrics/" + *metricsFolder + "/" + m.name + ".csv"
	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s", "ts", "sm_overrun", "mem_overrun", "nodeid", "requestid") + "\n")
	if err != nil {
		panic(err)
	}
	for _, metric := range m.metricArr {
		_, err := writer.WriteString(metric.String() + "\n")
		if err != nil {
			panic(err)
		}
	}

	err = writer.Flush()
	if err != nil {
		panic(err)
	}
}
