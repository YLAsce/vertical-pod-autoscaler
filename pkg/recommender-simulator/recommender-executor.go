package main

import (
	"flag"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

const mockContainerName = "container"

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

	recommenderInterval = flag.Duration("recommender-interval", 5*time.Minute, `How often recommender should run, prefer integer times of fetching interval`)

	isML = flag.Bool("ap-algorithm-ml", true, "Boolean value, is ML? or rule based")
)

type recommenderExecutor struct {
	podResourceRecommender  logic.PodResourceRecommender
	aggregateContainerState *model.AggregateContainerState
}

func NewRecommenderExecutor() *recommenderExecutor {
	model.InitializeAggregationsConfig(model.NewAggregationsConfig(*recommenderInterval,
		model.DefaultMemoryAggregationIntervalCount,
		*memoryHistogramDecayHalfLife, *cpuHistogramDecayHalfLife,
		*oomBumpUpRatio, *oomMinBumpUp,
		*recommenderInterval, *recommenderInterval,
		*cpuLastSamplesN, *memoryLastSamplesN,
		*cpuHistogramMaxValue, *cpuHistogramBucketNum, *memoryHistogramMaxValue, *memoryHistogramBucketNum))

	return &recommenderExecutor{
		podResourceRecommender:  logic.CreatePodResourceRecommender(*cpuHistogramMaxValue, *memoryHistogramMaxValue, *recommenderInterval, *cpuLastSamplesN, *memoryLastSamplesN, *isML),
		aggregateContainerState: model.NewAggregateContainerState(),
	}
}

func (r *recommenderExecutor) AddSample(now time.Time, usage model.ResourceAmount, name model.ResourceName) {
	sample := model.ContainerUsageSample{
		MeasureStart: now,
		Usage:        usage,
		Request:      model.ResourceAmount(0), // Not Used, give 0
		Resource:     name,
	}
	r.aggregateContainerState.AddSample(&sample)
}

func (r *recommenderExecutor) AddOOM(memoryAmount, memoryRecent model.ResourceAmount) {
	memoryUsed := model.ResourceAmountMax(memoryAmount, memoryRecent)
	memoryNeeded := model.ResourceAmountMax(memoryUsed+model.MemoryAmountFromBytes(model.GetAggregationsConfig().OOMMinBumpUp),
		model.ScaleResource(memoryUsed, model.GetAggregationsConfig().OOMBumpUpRatio))
	// fmt.Printf("Add OOM Init: %v, Recent %v, After: %v\n", memoryAmount, memoryRecent, memoryNeeded)
	r.aggregateContainerState.AddOOM(memoryNeeded)
}

func (r *recommenderExecutor) HistogramAggregate(now time.Time) {
	r.aggregateContainerState.HistogramAggregate(now)
}

func (r *recommenderExecutor) Recommend(algorithmRun bool) (logic.RecommendedContainerResources, bool) {
	nameToAggregate := make(model.ContainerNameToAggregateStateMap)
	nameToAggregate[mockContainerName] = r.aggregateContainerState
	recommendedResources := r.podResourceRecommender.GetRecommendedPodResources(nameToAggregate, algorithmRun)
	ret, ok := recommendedResources[mockContainerName]
	return ret, ok
}
