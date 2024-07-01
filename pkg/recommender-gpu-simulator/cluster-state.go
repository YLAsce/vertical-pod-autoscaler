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
	gpumHistogramDecayHalfLife = flag.Duration("ap-gpum-histogram-decay-half-life", model.DefaultGPUMemoryHistogramDecayHalfLife, `The amount of time it takes a historical memory usage sample to lose half of its weight. In other words, a fresh usage sample is twice as 'important' as one with age equal to the half life period.`)
	gpusHistogramDecayHalfLife = flag.Duration("ap-gpus-histogram-decay-half-life", model.DefaultGPUSMHistogramDecayHalfLife, `The amount of time it takes a historical CPU usage sample to lose half of its weight.`)
	gpusHistogramMaxValue      = flag.Float64("ap-gpus-histogram-max", model.DefaultGPUSMHistogramMaxValue, `CPU End of the last bucket >= this value`)
	gpumHistogramMaxValue      = flag.Float64("ap-gpum-histogram-max", model.DefaultGPUMemoryHistogramMaxValue, `Memory End of the last bucket >= this value`)
	gpusHistogramBucketNum     = flag.Int("ap-gpus-histogram-bucket-num", model.DefaultGPUSMHistogramBucketNum, `CPU num of buckets`)
	gpumHistogramBucketNum     = flag.Int("ap-gpum-histogram-bucket-num", model.DefaultGPUMemoryHistogramBucketNum, `Mem num of buckets`)

	gpusLastSamplesN = flag.Int("ap-gpus-histogram-n", model.DefaultLastSamplesN, `N last CPU samples in Autopilot paper`)
	gpumLastSamplesN = flag.Int("ap-gpum-histogram-n", model.DefaultLastSamplesN, `N last memory samples in Autopilot paper`)

	isML = flag.Bool("ap-algorithm-ml", true, "Boolean value, is ML? or rule based")
)

type ClusterState struct {
	podStateMap       map[string]*PodState
	podGPURecommender logic.PodGPURecommender
	constRecommender  *ConstRecommender
	nameToAggregate   model.ContainerNameToAggregateGPUMap
	nodePool          *NodePool
}

func NewClusterState(traceInfoMap map[string]*TraceInfo) *ClusterState {
	model.InitializeAggregationsGPUConfig(model.NewAggregationsGPUConfig(
		*gpumHistogramDecayHalfLife, *gpusHistogramDecayHalfLife,
		5*time.Minute, 5*time.Minute,
		*gpusLastSamplesN, *gpumLastSamplesN,
		*gpusHistogramMaxValue, *gpusHistogramBucketNum,
		*gpumHistogramMaxValue, *gpumHistogramBucketNum,
	))

	clusterState := ClusterState{
		podStateMap:       make(map[string]*PodState),
		podGPURecommender: logic.CreatePodGPURecommender(*gpusLastSamplesN, *gpumLastSamplesN),
		constRecommender:  CreateConstRecommender(traceInfoMap),
		nameToAggregate:   make(model.ContainerNameToAggregateGPUMap),
		nodePool:          NewNodePool(),
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
		gpuCores := result[name].(map[string]interface{})["gpusm"].(float64)
		memoryBytes := result[name].(map[string]interface{})["gpumemory"].(float64)
		fmt.Printf("Read Init: %s GPUSM: %v GPUMemory: %v\n", name, gpuCores, memoryBytes)
		clusterState.podStateMap[name] = NewPodState(traceInfo, name, model.GPUSMAmountFromGIs(gpuCores), model.GPUMemoryAmountFromBytes(memoryBytes), clusterState.nodePool)
	}

	for name, podState := range clusterState.podStateMap {
		clusterState.nameToAggregate[name] = podState.aggregateGPUState
	}

	return &clusterState
}

func (c *ClusterState) Record(t time.Time, isConst bool) {
	for _, podState := range c.podStateMap {
		podState.Record(t, isConst)
	}
}

func (c *ClusterState) CollectMetrics(t time.Time) {
	for _, podState := range c.podStateMap {
		podState.CollectMetrics(t)
	}
}

func (c *ClusterState) AdjustOverrun() {
	for _, podState := range c.podStateMap {
		podState.AdjustOverrun()
	}
}

func (c *ClusterState) Recommend(t time.Time, isConst bool) {
	if !isConst {
		for _, podState := range c.podStateMap {
			podState.aggregateGPUState.HistogramAggregate(t)
		}
	}
	var recommendedResources logic.RecommendedPodResources
	if isConst {
		recommendedResources = c.constRecommender.GetRecommendedPodResources(c.nameToAggregate)
	} else {
		recommendedResources = c.podGPURecommender.GetRecommendedPodResources(c.nameToAggregate)
	}
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
