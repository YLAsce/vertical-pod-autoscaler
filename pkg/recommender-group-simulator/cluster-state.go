package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

var (
	memoryHistogramDecayHalfLife = flag.Duration("ap-memory-histogram-decay-half-life", model.DefaultMemoryHistogramDecayHalfLife, `The amount of time it takes a historical memory usage sample to lose half of its weight. In other words, a fresh usage sample is twice as 'important' as one with age equal to the half life period.`)
	cpuHistogramDecayHalfLife    = flag.Duration("ap-cpu-histogram-decay-half-life", model.DefaultCPUHistogramDecayHalfLife, `The amount of time it takes a historical CPU usage sample to lose half of its weight.`)
	cpuHistogramMaxValue         = flag.Float64("ap-cpu-histogram-max", model.DefaultCPUHistogramMaxValue, `CPU End of the last bucket >= this value`)
	memoryHistogramMaxValue      = flag.Float64("ap-memory-histogram-max", model.DefaultMemoryHistogramMaxValue, `Memory End of the last bucket >= this value`)
	cpuHistogramBucketNum        = flag.Int("ap-cpu-histogram-bucket-num", model.DefaultCPUHistogramBucketNum, `CPU num of buckets`)
	memoryHistogramBucketNum     = flag.Int("ap-memory-histogram-bucket-num", model.DefaultMemoryHistogramBucketNum, `Mem num of buckets`)

	cpuLastSamplesN    = flag.Int("ap-cpu-histogram-n", model.DefaultLastSamplesN, `N last CPU samples in Autopilot paper`)
	memoryLastSamplesN = flag.Int("ap-memory-histogram-n", model.DefaultLastSamplesN, `N last memory samples in Autopilot paper`)

	oomBumpUpRatio = flag.Float64("oom-bump-up-ratio", model.DefaultOOMBumpUpRatio, `The memory bump up ratio when OOM occurred, default is 1.2.`)
	oomMinBumpUp   = flag.Float64("oom-min-bump-up-bytes", model.DefaultOOMMinBumpUp, `The minimal increase of memory when OOM occurred in bytes, default is 100 * 1024 * 1024`)

	isML = flag.Bool("ap-algorithm-ml", true, "Boolean value, is ML? or rule based")

	memoryLimitRequestRatio = flag.Float64("memory-limit-request-ratio", 1.04, "memory limit = this value*memory request(recommended)")
	nodeSizeCPU             = flag.Float64("node-size-cpu", 3.9, "node size cpu")
	nodeSizeMemory          = flag.Int64("node-size-memory", 3483278000, "node size memory")
)

type ClusterState struct {
	podStateMap            map[string]*PodState
	podResourceRecommender logic.PodResourceRecommender
	nameToAggregate        model.ContainerNameToAggregateStateMap
	nodePool               *NodePool
}

func NewClusterState(traceInfoMap map[string]*TraceInfo) *ClusterState {
	model.InitializeAggregationsConfig(model.NewAggregationsConfig(5*time.Minute,
		model.DefaultMemoryAggregationIntervalCount,
		*memoryHistogramDecayHalfLife, *cpuHistogramDecayHalfLife,
		*oomBumpUpRatio, *oomMinBumpUp,
		5*time.Minute, 5*time.Minute,
		*cpuLastSamplesN, *memoryLastSamplesN,
		*cpuHistogramMaxValue, *cpuHistogramBucketNum, *memoryHistogramMaxValue, *memoryHistogramBucketNum))

	clusterState := ClusterState{
		podStateMap:            make(map[string]*PodState),
		podResourceRecommender: logic.CreatePodResourceRecommender(*cpuHistogramMaxValue, *memoryHistogramMaxValue, 5*time.Minute, *cpuLastSamplesN, *memoryLastSamplesN, *isML),
		nameToAggregate:        make(model.ContainerNameToAggregateStateMap),
		nodePool:               NewNodePool(model.CPUAmountFromCores(*nodeSizeCPU), model.MemoryAmountFromBytes(float64(*nodeSizeMemory))),
	}

	filePath := "trace/init.json"
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		panic(err.Error())
	}

	var result map[string]interface{}
	err = json.Unmarshal(fileContent, &result)
	if err != nil {
		panic(err.Error())
	}

	for name, traceInfo := range traceInfoMap {
		cpuCores := result[name].(map[string]interface{})["cpu"].(float64)
		memoryBytes := result[name].(map[string]interface{})["memory"].(float64)
		fmt.Printf("Read Init: %s CPU: %v Memory: %v\n", name, cpuCores, memoryBytes)
		clusterState.podStateMap[name] = NewPodState(traceInfo, name, *memoryLimitRequestRatio, model.CPUAmountFromCores(cpuCores), model.MemoryAmountFromBytes(memoryBytes), clusterState.nodePool)
	}

	for name, podState := range clusterState.podStateMap {
		clusterState.nameToAggregate[name] = podState.aggregateContainerState
	}

	return &clusterState
}

func (c *ClusterState) Record(t time.Time) bool {
	hasOom := false
	for _, podState := range c.podStateMap {
		podState.Record(t)
		if podState.Oom {
			hasOom = true
		}
	}
	return hasOom
}

func (c *ClusterState) CollectMetrics(t time.Time) {
	for _, podState := range c.podStateMap {
		podState.CollectMetrics(t)
	}
}

func (c *ClusterState) Recommend(t time.Time, algorithmRun bool) {
	if algorithmRun {
		for _, podState := range c.podStateMap {
			podState.aggregateContainerState.HistogramAggregate(t)
		}
	}
	recommendedResources := c.podResourceRecommender.GetRecommendedPodResources(c.nameToAggregate, algorithmRun)
	for name, resources := range recommendedResources {
		c.podStateMap[name].UpdateRequest(resources)
	}
	fmt.Printf("Finished rec resources at %v\n", t)
}

func (c *ClusterState) DumpMetrics() {
	for _, podState := range c.podStateMap {
		podState.DumpMetrics()
	}
	c.nodePool.Dump()
}
