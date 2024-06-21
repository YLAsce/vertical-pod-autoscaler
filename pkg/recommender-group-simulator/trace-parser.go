package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

type TraceDatapoint struct {
	startTime   time.Time
	duration    time.Duration
	cpuUsage    model.ResourceAmount
	memoryUsage model.ResourceAmount
}

type TraceInfo struct {
	traceFilePath string
	traceData     []*TraceDatapoint
}

func (t *TraceInfo) Max() model.Resources {
	result := make(model.Resources)
	result[model.ResourceCPU] = model.ResourceAmount(0)
	result[model.ResourceMemory] = model.ResourceAmount(0)
	for _, d := range t.traceData {
		result[model.ResourceCPU] = model.ResourceAmountMax(result[model.ResourceCPU], d.cpuUsage)
		result[model.ResourceMemory] = model.ResourceAmountMax(result[model.ResourceMemory], d.memoryUsage)
	}
	return result
}

func (t *TraceInfo) GetTraceData(initialTime time.Time) {
	traceFile, err := os.Open(t.traceFilePath)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(traceFile)
	for scanner.Scan() {
		line := scanner.Text()
		nums := strings.Fields(line)
		if len(nums) != 4 {
			panic("Error in trace file, not 4 numbers in one line: " + line)
		}

		var startTime int64 = 0
		var period int = 0
		var cpu float64 = 0.0
		var memory int64 = 0
		var err error

		if startTime, err = strconv.ParseInt(nums[0], 10, 64); err != nil {
			panic(err)
		}

		if period, err = strconv.Atoi(nums[1]); err != nil {
			panic(err)
		}

		if cpu, err = strconv.ParseFloat(nums[2], 64); err != nil {
			panic(err)
		}

		if memory, err = strconv.ParseInt(nums[3], 10, 64); err != nil {
			panic(err)
		}
		t.traceData = append(t.traceData, &TraceDatapoint{
			startTime:   initialTime.Add(time.Duration(startTime) * time.Second),
			duration:    time.Duration(period) * time.Second,
			cpuUsage:    model.CPUAmountFromCores(cpu),
			memoryUsage: model.MemoryAmountFromBytes(float64(memory)),
		})
	}
	traceFile.Close()
}

func GetTraceInfoMap(initialTime time.Time) map[string]*TraceInfo {
	folderPath := "trace"

	ret := make(map[string]*TraceInfo)
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}
		if !strings.HasPrefix(info.Name(), "0") {
			return nil
		}
		parts := strings.Split(info.Name(), ".")
		if len(parts) != 3 {
			panic("Error file name " + info.Name())
		}
		curStr := ""
		if parts[1] == "0" {
			curStr = "serverless_" + parts[2]
		} else {
			curStr = "daemon_" + parts[2]
		}
		curTraceInfo := &TraceInfo{
			traceFilePath: path,
			traceData:     make([]*TraceDatapoint, 0),
		}
		curTraceInfo.GetTraceData(initialTime)
		ret[curStr] = curTraceInfo
		return nil
	})

	if err != nil {
		panic(err.Error())
	}

	return ret
}

func GetMaxTraceTime(traceInfoMap map[string]*TraceInfo, initialTime time.Time) time.Time {
	ret := initialTime
	for _, traceInfo := range traceInfoMap {
		lastTraceData := traceInfo.traceData[len(traceInfo.traceData)-1]
		if lastTraceData.startTime.Add(lastTraceData.duration).After(ret) {
			ret = lastTraceData.startTime.Add(lastTraceData.duration)
		}
	}
	return ret
}

func DumpTraceInfoSummary(traceInfoMap map[string]*TraceInfo, initialTime time.Time) {
	maxlen := 0
	maxlenkey := ""
	for key, traceInfo := range traceInfoMap {
		if maxlen < len(traceInfo.traceData) {
			maxlen = len(traceInfo.traceData)
			maxlenkey = key
		}
	}

	fullPath := "metrics/" + *nodeFolder + "/metricsinfo.csv"
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
	for _, td := range traceInfoMap[maxlenkey].traceData {
		_, err := writer.WriteString(fmt.Sprintf("%v,0,0,0,0", td.startTime.Unix()) + "\n")
		if err != nil {
			panic(err)
		}
	}

	err = writer.Flush()
	if err != nil {
		panic(err)
	}

	fullPathFirst := "metrics/" + *nodeFolder + "/metricsinfo.info"
	fileFirst, err := os.OpenFile(fullPathFirst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = fmt.Fprintf(fileFirst, "%v", initialTime.Unix())
	if err != nil {
		panic(err)
	}
}
