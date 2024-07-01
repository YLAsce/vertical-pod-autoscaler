package model

import (
	"fmt"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/util"
)

type ContainerNameToAggregateGPUMap map[string]*AggregateGPUState

type AggregateGPUState struct {
	AggregateSMUsage     util.AutopilotHisto
	AggregateMemoryUsage util.AutopilotHisto
	MLRecommenderSM      *Recommender
	MLRecommenderMemory  *Recommender
}

func NewAggregateGPUState() *AggregateGPUState {
	config := GetAggregationsGPUConfig()
	agSM := util.NewAutopilotHisto(config.SMHistogramOptions, config.SMHistogramDecayHalfLife, config.SMLastSamplesN, config.SMDefaultAggregationDuration, util.AutopilotAddSampleModeDistribution)
	agMemory := util.NewAutopilotHisto(config.MemoryHistogramOptions, config.MemoryHistogramDecayHalfLife, config.MemoryLastSamplesN, config.MemoryDefaultAggregationDuration, util.AutopilotAddSampleModeMax)
	return &AggregateGPUState{
		AggregateSMUsage:     agSM,
		AggregateMemoryUsage: agMemory,
		MLRecommenderSM:      NewGPUSMRecommender(agSM),
		MLRecommenderMemory:  NewGPUMemoryRecommender(agMemory),
	}
}

func (a *AggregateGPUState) HistogramAggregate(now time.Time) {
	a.AggregateSMUsage.Aggregate(now)
	a.AggregateMemoryUsage.Aggregate(now)
}

func (a *AggregateGPUState) AddSample(sample *ContainerUsageSample) {
	switch sample.Resource {
	case ResourceGPUSM:
		a.AggregateSMUsage.AddSample(GIsFromGPUSMAmount(sample.Usage))
	case ResourceGPUMemory:
		a.AggregateMemoryUsage.AddSample(BytesFromMemoryAmount(sample.Usage))
	default:
		panic(fmt.Sprintf("AddSample doesn't support resource '%s'", sample.Resource))
	}
}
