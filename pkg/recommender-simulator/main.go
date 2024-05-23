package main

import (
	"flag"
	"fmt"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

var (
	initialCPU              = flag.Float64("initial-cpu", 0.8, "initial cpu cores request&limit in kubernetes yaml, (same value)")
	initialMemory           = flag.Int64("initial-memory", 600000000, "initial memory bytes request&limit in kubernetes yaml, (same value)")
	memoryLimitRequestRatio = flag.Float64("memory-limit-request-ratio", 1.04, "memory limit = this value*memory request(recommended)")
	exitWhenLargeOverrun    = flag.Int("exit-memory-large-overrun", 0, "exit when memory overrun > this value, to save time, 0 is disable")
)

var (
	loopCounter      int64                = 0
	curCPUUsage      model.ResourceAmount = model.ResourceAmount(0)
	curCPURequest    model.ResourceAmount = model.ResourceAmount(0)
	curMemoryUsage   model.ResourceAmount = model.ResourceAmount(0)
	curMemoryRequest model.ResourceAmount = model.ResourceAmount(0)
)

func main() {
	// ML 模型的参数也会通过这里传入
	flag.Parse()
	executor := NewRecommenderExecutor()

	traceParser := NewTraceParser()
	defer traceParser.Close()

	curCPURequest = model.CPUAmountFromCores(*initialCPU)
	curMemoryRequest = model.MemoryAmountFromBytes(float64(*initialMemory))

	initialTime := time.Now()

	metricsCollector := NewMetricsCollector(*memoryLimitRequestRatio)
	start := time.Now()
	memoryOverrunSeconds := 0
	for {
		ok, period, cpu, memory := traceParser.ScanNext()
		if !ok {
			break
		}
		// fmt.Printf("Read data: %v %v %v\n", period, cpu, memory)
		for period > 0 {
			curCPUSample := cpu
			curMemorySample := memory
			// fmt.Printf("Round start, loopCounter: %v, period: %v\n", loopCounter, period)
			loopCounter++
			curTime := initialTime.Add(time.Duration(loopCounter) * time.Second)
			cpuOverrun := false
			oom := false
			if curCPUSample > curCPURequest {
				// CPU overrun (but no limit)
				cpuOverrun = true
			}
			memoryOverrun := false
			if curMemorySample > curMemoryRequest {
				// Memory overrun (only for record)
				memoryOverrun = true
			}
			memoryLimit := model.MemoryAmountFromBytes(model.BytesFromMemoryAmount(curMemoryRequest) * (*memoryLimitRequestRatio))
			if curMemorySample > memoryLimit {
				// OOM
				executor.AddOOM(memoryLimit, curMemoryUsage)
				// When OOM, the workload crash, no CPU and memory usage
				curMemorySample = 0
				curCPUSample = 0
				oom = true
			} else {
				period--
				executor.AddSample(curTime, curCPUSample, model.ResourceCPU)
				executor.AddSample(curTime, curMemorySample, model.ResourceMemory)
			}
			curCPUUsage = curCPUSample
			curMemoryUsage = curMemorySample

			if *exitWhenLargeOverrun > 0 {
				if memoryOverrun {
					memoryOverrunSeconds += 1
				}
				if memoryOverrunSeconds > *exitWhenLargeOverrun {
					return
				}
			}
			metricsCollector.Record(metricPoint{
				CpuUsage:      curCPUUsage,
				CpuRequest:    curCPURequest,
				MemoryUsage:   curMemoryUsage,
				MemoryRequest: curMemoryRequest,
				MemoryOverrun: memoryOverrun,
				CpuOverrun:    cpuOverrun,
				Oom:           oom,
				ts:            curTime,
			})

			var recommendedResources logic.RecommendedContainerResources
			var ok bool
			if loopCounter%300 == 0 { // 5min
				executor.HistogramAggregate(curTime)
				recommendedResources, ok = executor.Recommend(true)
				if ok {
					curCPURequest = recommendedResources.Target[model.ResourceCPU]
					curMemoryRequest = recommendedResources.Target[model.ResourceMemory]
					fmt.Printf("Recommendation: %+v\n", recommendedResources.Target)
				}
			} else if oom {
				recommendedResources, ok = executor.Recommend(false)
				if ok {
					curCPURequest = recommendedResources.Target[model.ResourceCPU]
					curMemoryRequest = recommendedResources.Target[model.ResourceMemory]
					fmt.Printf("OOM Recommendation: %+v\n", recommendedResources.Target)
				}
			}

		}
	}
	elapsed := time.Since(start)

	// 打印运行时间
	fmt.Printf("Time: %s. Summary file: %s\n", elapsed, *metricsSummaryFile)
	metricsCollector.Dump()
	metricsCollector.DumpSummary()
}
