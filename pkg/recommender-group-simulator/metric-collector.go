package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math"
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
	CpuUsage      model.ResourceAmount
	CpuRequest    model.ResourceAmount
	MemoryUsage   model.ResourceAmount
	MemoryRequest model.ResourceAmount
	MemoryOverrun bool
	CpuOverrun    bool
	Oom           bool
	Ts            time.Time
	NodeId        int
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (m *metricPoint) String() string {
	return fmt.Sprintf("%v,%v,%v,%d,%d", m.Ts.Unix(), model.CoresFromCPUAmount(m.CpuRequest), int64(model.BytesFromMemoryAmount(m.MemoryRequest)), bool2int(m.Oom), m.NodeId)
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
	_, err = writer.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s", "ts", "cpu_request", "mem_request", "oom", "nodeid") + "\n")
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

type MetricsSummary struct {
	CPUAverageUtilization    float64 `json:"cpu-average-gap"`
	MemoryAverageUtilization float64 `json:"memory-average-gap"`
	CPUOverrunSeconds        int64   `json:"cpu-overrun-seconds"`
	MemoryOverrunSeconds     int64   `json:"memory-overrun-seconds"`
	OomSeconds               int64   `json:"oom-seconds"`
	CPURequestAdjustTimes    int64   `json:"cpu-request-adjust-times"`
	MemoryRequestAdjustTimes int64   `json:"memory-request-adjust-times"`
}

func (m *MetricsCollector) DumpSummary() {
	// fmt.Printf("Start Dump Summary %s %v\n", *metricsSummaryFile, *metricsSummaryIgnoreHead)

	fullPath := "metrics/" + m.name + ".json"
	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	summary := MetricsSummary{
		CPUAverageUtilization:    0.0,
		MemoryAverageUtilization: 0.0,
		CPUOverrunSeconds:        0,
		MemoryOverrunSeconds:     0,
		OomSeconds:               0,
		CPURequestAdjustTimes:    0,
		MemoryRequestAdjustTimes: 0,
	}

	if len(m.metricArr) < *metricsSummaryIgnoreHead {
		panic("too few data to generate summary")
	}

	var totalCPURequest float64 = 0.0
	var totalMemoryRequest float64 = 0.0
	var totalDataPoints int = 0
	for i := *metricsSummaryIgnoreHead; i < len(m.metricArr); i++ {
		if !m.metricArr[i].Oom {
			// Ignore OOM points when calculating average
			totalDataPoints += 1
			summary.CPUAverageUtilization += math.Min(float64(m.metricArr[i].CpuRequest), float64(m.metricArr[i].CpuUsage))
			summary.MemoryAverageUtilization += math.Min(float64(m.metricArr[i].MemoryRequest), float64(m.metricArr[i].MemoryUsage))
			totalCPURequest += float64(m.metricArr[i].CpuRequest)
			totalMemoryRequest += float64(m.metricArr[i].MemoryRequest)
		} else {
			summary.OomSeconds += 1
		}

		if m.metricArr[i].CpuOverrun {
			summary.CPUOverrunSeconds += 1
		}
		if m.metricArr[i].MemoryOverrun {
			summary.MemoryOverrunSeconds += 1
		}

		if m.metricArr[i].CpuRequest != m.metricArr[i-1].CpuRequest {
			summary.CPURequestAdjustTimes += 1
		}

		if m.metricArr[i].MemoryRequest != m.metricArr[i-1].MemoryRequest {
			summary.MemoryRequestAdjustTimes += 1
		}
	}

	summary.CPUAverageUtilization = (totalCPURequest - summary.CPUAverageUtilization) / float64(totalDataPoints)
	summary.MemoryAverageUtilization = (totalMemoryRequest - summary.MemoryAverageUtilization) / float64(totalDataPoints)

	encoder := json.NewEncoder(file)
	err = encoder.Encode(summary)
	if err != nil {
		panic(err)
	}
}
