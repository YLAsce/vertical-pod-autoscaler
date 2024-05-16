package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

var (
	metricsFile              = flag.String("metrics-file", "", "metrics file name")
	metricsSummaryFile       = flag.String("metrics-summary-file", "", "metrics summary file name")
	metricsSummaryIgnoreHead = flag.Int("metrics-summary-ignore-head", 1800, "ignore the first metrics when calculating summary")
)

type metricPoint struct {
	CpuUsage      model.ResourceAmount
	CpuRequest    model.ResourceAmount
	MemoryUsage   model.ResourceAmount
	MemoryRequest model.ResourceAmount
	MemoryOverrun bool
	CpuOverrun    bool
	Oom           bool
	ts            time.Time
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (m *metricPoint) String() string {
	return fmt.Sprintf("%v %v %v %v %v %d %d %d", m.ts.Unix(), model.CoresFromCPUAmount(m.CpuRequest), model.CoresFromCPUAmount(m.CpuUsage), int64(model.BytesFromMemoryAmount(m.MemoryRequest)), int64(model.BytesFromMemoryAmount(m.MemoryUsage)), bool2int(m.CpuOverrun), bool2int(m.MemoryOverrun), bool2int(m.Oom))
}

type metricsCollector struct {
	metricArr          []metricPoint
	memoryLimitRequest float64
}

func NewMetricsCollector(memoryLimitRequest float64) *metricsCollector {
	return &metricsCollector{
		metricArr:          make([]metricPoint, 0),
		memoryLimitRequest: memoryLimitRequest,
	}
}

func (m *metricsCollector) Record(metric metricPoint) {
	m.metricArr = append(m.metricArr, metric)
}

func (m *metricsCollector) Dump() {
	if len(*metricsFile) == 0 {
		return
	}

	fullPath := "metrics/" + *metricsFile + "_" + strconv.FormatFloat(m.memoryLimitRequest, 'f', 2, 64) + ".data"
	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

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

func (m *metricsCollector) DumpSummary() {
	fmt.Printf("Start Dump Summary %s %v\n", *metricsSummaryFile, *metricsSummaryIgnoreHead)
	if len(*metricsSummaryFile) == 0 {
		return
	}
	fullPath := "metrics/" + *metricsSummaryFile + "_" + strconv.FormatFloat(m.memoryLimitRequest, 'f', 2, 64) + ".json"
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
