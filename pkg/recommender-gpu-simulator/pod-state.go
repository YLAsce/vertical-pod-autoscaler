package main

import (
	"fmt"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

var smids = []model.ResourceAmount{model.ResourceAmount(1000), model.ResourceAmount(2000), model.ResourceAmount(3000)}
var memoryids = []model.ResourceAmount{model.ResourceAmount(10000000000), model.ResourceAmount(20000000000), model.ResourceAmount(40000000000)}

var magicmemoryId = model.ResourceAmount(30000000000)

func alignRequest(req model.ResourceAmount, target []model.ResourceAmount) model.ResourceAmount {
	for i := 0; i < len(target); i++ {
		if target[i] >= req {
			return target[i]
		}
	}
	panic("req too large")
}

func getRequestId(smRequest, memoryRequest model.ResourceAmount) int {
	smid := -1
	memoryid := -1
	for i := 0; i < len(smids); i++ {
		if smids[i] >= smRequest {
			smid = i
		}
	}

	for i := 0; i < len(memoryids); i++ {
		if memoryids[i] >= memoryRequest {
			memoryid = i
		}
	}

	if memoryid < 0 || smid < 0 {
		panic("smrequest or memoryrequest")
	}
	return max(smid, memoryid)
}

func getRequestIdEqual(smRequest, memoryRequest model.ResourceAmount) int {
	smid := -1
	memoryid := -1
	for i := 0; i < len(smids); i++ {
		if smids[i] == smRequest {
			smid = i
		}
	}

	for i := 0; i < len(memoryids); i++ {
		if memoryids[i] == memoryRequest {
			memoryid = i
		}
	}
	if memoryRequest == magicmemoryId {
		memoryid = 2
	}

	if memoryid < 0 || smid < 0 {
		fmt.Println(smRequest, memoryRequest)
		panic("smrequest or memoryrequest floating error")
	}
	return max(smid, memoryid)
}

// 单个pod的状态集合，包含所有的aggregate container state

type PodState struct {
	aggregateGPUState *model.AggregateGPUState
	metricCollector   *MetricsCollector
	traceInfo         *TraceInfo

	CurSMUsage       model.ResourceAmount
	CurSMRequest     model.ResourceAmount
	CurMemoryUsage   model.ResourceAmount
	CurMemoryRequest model.ResourceAmount
	CurRequestId     int
	SMOverrun        bool
	MemoryOverrun    bool
	Oom              bool

	IsAlive        bool
	curTraceDataId int

	nodePool *NodePool
	name     string
}

func NewPodState(traceInfo *TraceInfo, name string, initSMRequest, initMemoryRequest model.ResourceAmount, nodePool *NodePool) *PodState {
	return &PodState{
		aggregateGPUState: model.NewAggregateGPUState(),
		metricCollector:   NewMetricsCollector(name),
		traceInfo:         traceInfo,

		CurSMUsage:       model.ResourceAmount(0),
		CurSMRequest:     initSMRequest,
		CurMemoryUsage:   model.ResourceAmount(0),
		CurMemoryRequest: initMemoryRequest,
		CurRequestId:     getRequestIdEqual(initSMRequest, initMemoryRequest),
		SMOverrun:        false,
		MemoryOverrun:    false,
		Oom:              false,

		IsAlive:        false,
		curTraceDataId: 0,

		nodePool: nodePool,
		name:     name,
	}
}

func (p *PodState) Record(t time.Time, isConst bool) {
	p.IsAlive = false
	p.SMOverrun = false
	p.MemoryOverrun = false
	p.Oom = false
	// if end of cur slot < t, it's time to consider next slot. To get end of slot >= t
	for p.curTraceDataId < len(p.traceInfo.traceData) && p.traceInfo.traceData[p.curTraceDataId].startTime.Add(p.traceInfo.traceData[p.curTraceDataId].duration).Before(t) {
		p.curTraceDataId++
	}
	// if start of cur slot <= t, ok
	if p.curTraceDataId < len(p.traceInfo.traceData) && !p.traceInfo.traceData[p.curTraceDataId].startTime.After(t) {
		p.IsAlive = true

		curSMSample := p.traceInfo.traceData[p.curTraceDataId].smUsage
		curMemorySample := p.traceInfo.traceData[p.curTraceDataId].memoryUsage
		// fmt.Printf("%v (%v %v) %v %v %v %v %v\n", p.curTraceDataId, p.traceInfo.traceData[p.curTraceDataId].startTime, p.traceInfo.traceData[p.curTraceDataId].duration, t, curCPUSample, curMemorySample, p.curCPURequest, p.curMemoryRequest)

		if curSMSample > p.CurSMRequest {
			p.SMOverrun = true
			curSMSample = p.CurSMRequest
		}

		if curMemorySample > p.CurMemoryRequest {
			p.MemoryOverrun = true
		}

		if !isConst {
			p.aggregateGPUState.AddSample(&model.ContainerUsageSample{
				MeasureStart: t,
				Usage:        curSMSample,
				Request:      model.ResourceAmount(0), // Not Used, give 0
				Resource:     model.ResourceGPUSM,
			})
			if !p.MemoryOverrun {
				p.aggregateGPUState.AddSample(&model.ContainerUsageSample{
					MeasureStart: t,
					Usage:        curMemorySample,
					Request:      model.ResourceAmount(0), // Not Used, give 0
					Resource:     model.ResourceGPUMemory,
				})
			}
		}
		p.CurSMUsage = curSMSample
		p.CurMemoryUsage = curMemorySample

		p.nodePool.CollectInfo(p.name, convertRequestId(p.CurRequestId))

	}
}

func convertRequestId(id int) int {
	if id == 2 {
		return 4
	}
	if id == 1 {
		return 2
	}
	if id == 0 {
		return 1
	}
	panic("illegal ID")
}

func (p *PodState) CollectMetrics(t time.Time) {
	if !p.IsAlive {
		return
	}
	p.metricCollector.Record(metricPoint{
		SMUsage:       p.CurSMUsage,
		MemoryUsage:   p.CurMemoryUsage,
		MemoryOverrun: p.MemoryOverrun,
		SMOverrun:     p.SMOverrun,
		Ts:            t,
		RequestId:     convertRequestId(p.CurRequestId),
		NodeId:        p.nodePool.AssignNode(p.name),
	})
}

func (p *PodState) AdjustOverrun() {
	if p.SMOverrun || p.MemoryOverrun {
		p.CurRequestId++
		if p.CurRequestId >= len(smids) {
			fmt.Println(p.CurMemoryUsage, p.CurSMUsage, p.CurMemoryRequest, p.CurSMRequest, p.CurRequestId)
			panic("request id too large")
		}
		p.CurSMRequest = smids[p.CurRequestId]
		p.CurMemoryRequest = memoryids[p.CurRequestId]
	}
}

func (p *PodState) UpdateRequest(resources logic.RecommendedContainerResources) {
	p.CurSMRequest = resources.Target[model.ResourceGPUSM]
	p.CurMemoryRequest = resources.Target[model.ResourceGPUMemory]
	p.CurRequestId = getRequestIdEqual(p.CurSMRequest, p.CurMemoryRequest)
	p.CurSMRequest = smids[p.CurRequestId]
	p.CurMemoryRequest = memoryids[p.CurRequestId]
}

func (p *PodState) DumpMetrics() {
	p.metricCollector.Dump()
}
