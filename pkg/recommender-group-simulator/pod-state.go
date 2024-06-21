package main

import (
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

// 单个pod的状态集合，包含所有的aggregate container state

type PodState struct {
	aggregateContainerState *model.AggregateContainerState
	metricCollector         *MetricsCollector
	traceInfo               *TraceInfo

	CurCPUUsage      model.ResourceAmount
	CurCPURequest    model.ResourceAmount
	CurMemoryUsage   model.ResourceAmount
	CurMemoryRequest model.ResourceAmount
	CpuOverrun       bool
	MemoryOverrun    bool
	Oom              bool

	IsAlive        bool
	curTraceDataId int

	memoryLimitRequestRatio float64

	nodePool *NodePool
	name     string
}

func NewPodState(traceInfo *TraceInfo, name string, memoryLimitRequest float64, initCPURequest, initMemoryRequest model.ResourceAmount, nodePool *NodePool) *PodState {
	return &PodState{
		aggregateContainerState: model.NewAggregateContainerState(),
		metricCollector:         NewMetricsCollector(name),
		traceInfo:               traceInfo,

		CurCPUUsage:      model.ResourceAmount(0),
		CurCPURequest:    initCPURequest,
		CurMemoryUsage:   model.ResourceAmount(0),
		CurMemoryRequest: initMemoryRequest,
		CpuOverrun:       false,
		MemoryOverrun:    false,
		Oom:              false,

		IsAlive:        false,
		curTraceDataId: 0,

		memoryLimitRequestRatio: memoryLimitRequest,
		nodePool:                nodePool,
		name:                    name,
	}
}

func (p *PodState) Record(t time.Time, isConst bool) {
	p.IsAlive = false
	p.CpuOverrun = false
	p.MemoryOverrun = false
	p.Oom = false
	// if end of cur slot < t, it's time to consider next slot. To get end of slot >= t
	for p.curTraceDataId < len(p.traceInfo.traceData) && p.traceInfo.traceData[p.curTraceDataId].startTime.Add(p.traceInfo.traceData[p.curTraceDataId].duration).Before(t) {
		p.curTraceDataId++
	}
	// if start of cur slot <= t, ok
	if p.curTraceDataId < len(p.traceInfo.traceData) && !p.traceInfo.traceData[p.curTraceDataId].startTime.After(t) {
		p.IsAlive = true

		curCPUSample := p.traceInfo.traceData[p.curTraceDataId].cpuUsage
		curMemorySample := p.traceInfo.traceData[p.curTraceDataId].memoryUsage
		// fmt.Printf("%v (%v %v) %v %v %v %v %v\n", p.curTraceDataId, p.traceInfo.traceData[p.curTraceDataId].startTime, p.traceInfo.traceData[p.curTraceDataId].duration, t, curCPUSample, curMemorySample, p.curCPURequest, p.curMemoryRequest)

		if curCPUSample > p.CurCPURequest {
			// CPU overrun (but no limit)
			p.CpuOverrun = true
		}

		if curMemorySample > p.CurMemoryRequest {
			// Memory overrun (only for record)
			p.MemoryOverrun = true
		}

		memoryLimit := model.MemoryAmountFromBytes(model.BytesFromMemoryAmount(p.CurMemoryRequest) * p.memoryLimitRequestRatio)
		if curMemorySample > memoryLimit {
			// OOM
			memoryUsed := model.ResourceAmountMax(memoryLimit, p.CurMemoryUsage)
			memoryNeeded := model.ResourceAmountMax(memoryUsed+model.MemoryAmountFromBytes(model.GetAggregationsConfig().OOMMinBumpUp),
				model.ScaleResource(memoryUsed, model.GetAggregationsConfig().OOMBumpUpRatio))
			// fmt.Printf("Add OOM Init: %v, Recent %v, After: %v\n", memoryAmount, memoryRecent, memoryNeeded)
			p.aggregateContainerState.AddOOM(memoryNeeded)
			// When OOM, the workload crash, no CPU and memory usage
			curMemorySample = 0
			curCPUSample = 0
			p.Oom = true
		} else if !isConst {
			p.aggregateContainerState.AddSample(&model.ContainerUsageSample{
				MeasureStart: t,
				Usage:        curCPUSample,
				Request:      model.ResourceAmount(0), // Not Used, give 0
				Resource:     model.ResourceCPU,
			})

			p.aggregateContainerState.AddSample(&model.ContainerUsageSample{
				MeasureStart: t,
				Usage:        curMemorySample,
				Request:      model.ResourceAmount(0), // Not Used, give 0
				Resource:     model.ResourceMemory,
			})
		}
		p.CurCPUUsage = curCPUSample
		p.CurMemoryUsage = curMemorySample

		p.nodePool.CollectInfo(p.name, p.CurCPURequest, p.CurMemoryRequest)

	}
}

func (p *PodState) CollectMetrics(t time.Time) {
	if !p.IsAlive {
		return
	}

	p.metricCollector.Record(metricPoint{
		CpuUsage:      p.CurCPUUsage,
		CpuRequest:    p.CurCPURequest,
		MemoryUsage:   p.CurMemoryUsage,
		MemoryRequest: p.CurMemoryRequest,
		MemoryOverrun: p.MemoryOverrun,
		CpuOverrun:    p.CpuOverrun,
		Oom:           p.Oom,
		Ts:            t,
		NodeId:        p.nodePool.AssignNode(p.name),
	})
}

func (p *PodState) UpdateRequest(resources logic.RecommendedContainerResources) {
	p.CurCPURequest = resources.Target[model.ResourceCPU]
	p.CurMemoryRequest = resources.Target[model.ResourceMemory]
}

func (p *PodState) DumpMetrics() {
	p.metricCollector.Dump()
}
