/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logic

import (
	"flag"
	"sort"
	"time"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/klog/v2"
)

var (
	// safetyMarginFraction = flag.Float64("recommendation-margin-fraction", 0.15, `Fraction of usage added as the safety margin to the recommended request`)
	// podMinCPUMillicores = flag.Float64("pod-recommendation-min-cpu-millicores", 25, `Minimum CPU recommendation for a pod`)
	// podMinMemoryMb      = flag.Float64("pod-recommendation-min-memory-mb", 250, `Minimum memory recommendation for a pod`)
	// targetCPUPercentile  = flag.Float64("target-cpu-percentile", 0.9, "CPU usage percentile that will be used as a base for CPU target recommendation. Doesn't affect CPU lower bound, CPU upper bound nor memory recommendations.")
	cpuRecommendPolicy         = flag.String("ap-cpu-recommend-policy", "avg", "choice among`: 'avg', 'max', 'sp_xx' where xx is the percentile, 'spike' which is max(sp_60 , 0.5*max)")
	memRecommendPolicy         = flag.String("ap-memory-recommend-policy", "avg", "choice among`: 'avg', 'max', 'sp_xx' where xx is the percentile, 'spike' which is max(sp_60 , 0.5*max)")
	fluctuationReducerDuration = flag.Duration("ap-fluctuation-reducer-duration", defaultFluctuationReducerDuration, "Period for fluctuation reducer, to choose the max value of recommendation in this period.")
)

// PodResourceRecommender computes resource recommendation for a Vpa object.
type PodResourceRecommender interface {
	GetRecommendedPodResources(containerNameToAggregateStateMap model.ContainerNameToAggregateStateMap, algorithmRun bool) RecommendedPodResources
}

// RecommendedPodResources is a Map from container name to recommended resources.
type RecommendedPodResources map[string]RecommendedContainerResources

// RecommendedContainerResources is the recommendation of resources for a
// container.
type RecommendedContainerResources struct {
	// Recommended optimal amount of resources.
	Target model.Resources
	// Recommended minimum amount of resources.
	LowerBound model.Resources
	// Recommended maximum amount of resources.
	UpperBound model.Resources
}

// --------
type PodGPURecommender interface {
	GetRecommendedPodResources(containerNameToAggregateGPUMap model.ContainerNameToAggregateGPUMap) RecommendedPodResources
}
type podGPURecommender struct {
	targetEstimator AutopilotGPUResourceEstimator
}

func (r *podGPURecommender) GetRecommendedPodResources(containerNameToAggregateGPUMap model.ContainerNameToAggregateGPUMap) RecommendedPodResources {

	var recommendation = make(RecommendedPodResources)
	if len(containerNameToAggregateGPUMap) == 0 {
		return recommendation
	}

	// In Autopilot, recommender can refuse to give result, this can avoid code start problem
	for containerName, aggregatedGPUState := range containerNameToAggregateGPUMap {
		estimation, err := r.estimateGPUResources(containerName, aggregatedGPUState)
		if err != nil {
			klog.V(3).Infof("NICONICO Cannot give valid pod recommendation. Reason: %s", err.Error())
		} else {
			recommendation[containerName] = estimation
		}
	}
	return recommendation
}

func (r *podGPURecommender) estimateGPUResources(containerName string, s *model.AggregateGPUState) (RecommendedContainerResources, error) {
	res, err := r.targetEstimator.GetResourceEstimation(containerName, s)
	// The same target, lower bound and upper bound
	return RecommendedContainerResources{
		res,
		res,
		res,
	}, err
}

func CreatePodGPURecommender(smLastSamplesN, memoryLastSamplesN int) PodGPURecommender {
	targetEstimator := NewMLGPUEstimator(smLastSamplesN, memoryLastSamplesN)
	return &podGPURecommender{
		targetEstimator: targetEstimator,
	}

}

// --------------

type podResourceRecommender struct {
	targetEstimator  AutopilotResourceEstimator
	oomPostProcessor *OOMPostProcessor
	// Discarded in autopilot
	// lowerBoundEstimator ResourceEstimator
	// upperBoundEstimator ResourceEstimator
}

func (r *podResourceRecommender) GetRecommendedPodResources(containerNameToAggregateStateMap model.ContainerNameToAggregateStateMap, algorithmRun bool) RecommendedPodResources {

	// klog.V(4).Info("NICONICO============================================")
	// for name, state := range containerNameToAggregateStateMap {
	// fmt.Printf("NICONICO %s:\n%s", name, state.AggregateMemoryUsage.String())
	// 	klog.V(4).Infof("NICONICO CPU Max: %v, Avg: %v, Per95: %v", state.AggregateCPUUsage.Max(), state.AggregateCPUUsage.Average(), state.AggregateCPUUsage.Percentile(0.95))
	// 	klog.V(4).Infof("NICONICO MEM Max: %v, Avg: %v, Per95: %v", state.AggregateMemoryUsage.Max(), state.AggregateMemoryUsage.Average(), state.AggregateMemoryUsage.Percentile(0.95))
	// }
	// klog.V(4).Info("NICONICO============================================")
	// time.Sleep(20 * time.Millisecond)
	var recommendation = make(RecommendedPodResources)
	if len(containerNameToAggregateStateMap) == 0 {
		return recommendation
	}

	// In Autopilot, recommender can refuse to give result, this can avoid code start problem
	for containerName, aggregatedContainerState := range containerNameToAggregateStateMap {
		estimation, err := r.estimateContainerResources(containerName, aggregatedContainerState, algorithmRun)
		klog.V(4).Infof("NICONICO ESTIMATE %+v", estimation)
		if err != nil {
			klog.V(3).Infof("NICONICO Cannot give valid pod recommendation. Reason: %s", err.Error())
		} else {
			recommendation[containerName] = estimation
		}
	}
	return recommendation

	// fraction := 1.0 / float64(len(containerNameToAggregateStateMap))
	// minResources := model.Resources{
	// 	model.ResourceCPU:    model.ScaleResource(model.CPUAmountFromCores(*podMinCPUMillicores*0.001), fraction),
	// 	model.ResourceMemory: model.ScaleResource(model.MemoryAmountFromBytes(*podMinMemoryMb*1024*1024), fraction),
	// }

	// recommender := &podResourceRecommender{
	// 	WithMinResources(minResources, r.targetEstimator),
	// 	// WithMinResources(minResources, r.lowerBoundEstimator),
	// 	// WithMinResources(minResources, r.upperBoundEstimator),
	// }
}

// Takes AggregateContainerState and returns a container recommendation.
func (r *podResourceRecommender) estimateContainerResources(containerName string, s *model.AggregateContainerState, algorithmRun bool) (RecommendedContainerResources, error) {
	// return RecommendedContainerResources{
	// 	FilterControlledResources(r.targetEstimator.GetResourceEstimation(s), s.GetControlledResources()),
	// 	FilterControlledResources(r.lowerBoundEstimator.GetResourceEstimation(s), s.GetControlledResources()),
	// 	FilterControlledResources(r.upperBoundEstimator.GetResourceEstimation(s), s.GetControlledResources()),
	// }
	if algorithmRun {
		baseEstimation, err0 := r.targetEstimator.GetResourceEstimation(containerName, s)
		r.oomPostProcessor.RecordBaseEstimation(containerName, baseEstimation, err0)
		// fmt.Printf("Base: %+v\n", baseEstimation)
	}
	estimation, err := r.oomPostProcessor.GetOOMPostProcessedEstimation(containerName, s)
	res := FilterControlledResources(estimation, s.GetControlledResources())
	// The same target, lower bound and upper bound
	return RecommendedContainerResources{
		res,
		res,
		res,
	}, err
}

// FilterControlledResources returns estimations from 'estimation' only for resources present in 'controlledResources'.
func FilterControlledResources(estimation model.Resources, controlledResources []model.ResourceName) model.Resources {
	result := make(model.Resources)
	for _, resource := range controlledResources {
		if value, ok := estimation[resource]; ok {
			result[resource] = value
		}
	}
	return result
}

// CreatePodResourceRecommender take the config info, returns the primary recommender.
func CreatePodResourceRecommender(cpuHistogramMaxValue, memoryHistogramMaxValue float64, recommenderInterval time.Duration, cpuLastSamplesN, memoryLastSamplesN int, isML bool) PodResourceRecommender {
	oomPostProcessor := NewOOMPostProcessor()

	if isML {
		targetEstimator := NewMLEstimator(cpuLastSamplesN, memoryLastSamplesN)
		return &podResourceRecommender{
			targetEstimator:  targetEstimator,
			oomPostProcessor: oomPostProcessor,
		}
	}
	// Is Rule
	targetEstimator := NewAutopilotEstimator(*cpuRecommendPolicy, *memRecommendPolicy, cpuLastSamplesN, memoryLastSamplesN)
	targetEstimator = WithAutopilotSafetyMargin(cpuHistogramMaxValue, memoryHistogramMaxValue, targetEstimator)
	targetEstimator = WithAutopilotFluctuationReducer(*fluctuationReducerDuration, recommenderInterval, targetEstimator)
	return &podResourceRecommender{
		targetEstimator:  targetEstimator,
		oomPostProcessor: oomPostProcessor,
	}
}

// Discarded in autopilot
// CreatePodResourceRecommender take the config info, returns the primary recommender.
// func CreatePodResourceRecommender() PodResourceRecommender {
// lowerBoundCPUPercentile := 0.5
// upperBoundCPUPercentile := 0.95

// targetMemoryPeaksPercentile := 0.9
// lowerBoundMemoryPeaksPercentile := 0.5
// upperBoundMemoryPeaksPercentile := 0.95

// targetEstimator := NewPercentileEstimator(*targetCPUPercentile, targetMemoryPeaksPercentile)
// lowerBoundEstimator := NewPercentileEstimator(lowerBoundCPUPercentile, lowerBoundMemoryPeaksPercentile)
// upperBoundEstimator := NewPercentileEstimator(upperBoundCPUPercentile, upperBoundMemoryPeaksPercentile)

// targetEstimator = WithMargin(*safetyMarginFraction, targetEstimator)
// lowerBoundEstimator = WithMargin(*safetyMarginFraction, lowerBoundEstimator)
// upperBoundEstimator = WithMargin(*safetyMarginFraction, upperBoundEstimator)

// Apply confidence multiplier to the upper bound estimator. This means
// that the updater will be less eager to evict pods with short history
// in order to reclaim unused resources.
// Using the confidence multiplier 1 with exponent +1 means that
// the upper bound is multiplied by (1 + 1/history-length-in-days).
// See estimator.go to see how the history length and the confidence
// multiplier are determined. The formula yields the following multipliers:
// No history     : *INF  (do not force pod eviction)
// 12h history    : *3    (force pod eviction if the request is > 3 * upper bound)
// 24h history    : *2
// 1 week history : *1.14
// upperBoundEstimator = WithConfidenceMultiplier(1.0, 1.0, upperBoundEstimator)

// Apply confidence multiplier to the lower bound estimator. This means
// that the updater will be less eager to evict pods with short history
// in order to provision them with more resources.
// Using the confidence multiplier 0.001 with exponent -2 means that
// the lower bound is multiplied by the factor (1 + 0.001/history-length-in-days)^-2
// (which is very rapidly converging to 1.0).
// See estimator.go to see how the history length and the confidence
// multiplier are determined. The formula yields the following multipliers:
// No history   : *0   (do not force pod eviction)
// 5m history   : *0.6 (force pod eviction if the request is < 0.6 * lower bound)
// 30m history  : *0.9
// 60m history  : *0.95
// lowerBoundEstimator = WithConfidenceMultiplier(0.001, -2.0, lowerBoundEstimator)

// return &podResourceRecommender{
// 	targetEstimator,
// 	lowerBoundEstimator,
// 	upperBoundEstimator}
// }

// MapToListOfRecommendedContainerResources converts the map of RecommendedContainerResources into a stable sorted list
// This can be used to get a stable sequence while ranging on the data
func MapToListOfRecommendedContainerResources(resources RecommendedPodResources) *vpa_types.RecommendedPodResources {
	containerResources := make([]vpa_types.RecommendedContainerResources, 0, len(resources))
	// Sort the container names from the map. This is because maps are an
	// unordered data structure, and iterating through the map will return
	// a different order on every call.
	containerNames := make([]string, 0, len(resources))
	for containerName := range resources {
		containerNames = append(containerNames, containerName)
	}
	sort.Strings(containerNames)
	// Create the list of recommendations for each container.
	for _, name := range containerNames {
		containerResources = append(containerResources, vpa_types.RecommendedContainerResources{
			ContainerName:  name,
			Target:         model.ResourcesAsResourceList(resources[name].Target),
			LowerBound:     model.ResourcesAsResourceList(resources[name].LowerBound),
			UpperBound:     model.ResourcesAsResourceList(resources[name].UpperBound),
			UncappedTarget: model.ResourcesAsResourceList(resources[name].Target),
		})
	}
	recommendation := &vpa_types.RecommendedPodResources{
		ContainerRecommendations: containerResources,
	}
	return recommendation
}
