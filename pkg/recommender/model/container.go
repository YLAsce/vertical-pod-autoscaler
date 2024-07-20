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

package model

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	metrics_quality "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/quality"
	"k8s.io/klog/v2"
)

// ContainerUsageSample is a measure of resource usage of a container over some
// interval.
type ContainerUsageSample struct {
	// Start of the measurement interval.
	MeasureStart time.Time
	// Average CPU usage in cores or memory usage in bytes.
	Usage ResourceAmount
	// CPU or memory request at the time of measurement.
	Request ResourceAmount
	// Which resource is this sample for.
	Resource ResourceName
}

// ContainerState stores information about a single container instance.
// Each ContainerState has a pointer to the aggregation that is used for
// aggregating its usage samples.
// It holds the recent history of CPU and memory utilization.
//
//	Note: samples are added to intervals based on their start timestamps.
type ContainerState struct {
	// Current request.
	Request Resources
	// Start of the latest CPU usage sample that was aggregated. Discard in Autopilot
	// LastCPUSampleStart time.Time
	// Max memory usage observed in the current aggregation interval.
	memoryRecent ResourceAmount
	// Max memory usage estimated from an OOM event in the current aggregation interval.
	oomPeak ResourceAmount
	// End time of the current memory aggregation interval (not inclusive).
	WindowEnd time.Time
	// Start of the latest memory usage sample that was aggregated. Discard in Autopilot
	// lastMemorySampleStart time.Time
	// Aggregation to add usage samples to.
	aggregator ContainerStateAggregator

	memoryPeak ResourceAmount
	cpuPeak    ResourceAmount
	memoryPrev ResourceAmount
	cpuPrev    ResourceAmount

	idlePercentage float64
}

// NewContainerState returns a new ContainerState.
func NewContainerState(request Resources, aggregator ContainerStateAggregator, idlePercentage float64) *ContainerState {
	return &ContainerState{
		Request: request,
		// LastCPUSampleStart:    time.Time{},
		memoryRecent: 0,
		WindowEnd:    time.Time{},
		// lastMemorySampleStart: time.Time{},
		aggregator: aggregator,

		memoryPeak:     ResourceAmount(0),
		cpuPeak:        ResourceAmount(0),
		memoryPrev:     ResourceAmount(0),
		cpuPrev:        ResourceAmount(0),
		idlePercentage: idlePercentage,
	}
}

func (sample *ContainerUsageSample) isValid(expectedResource ResourceName) bool {
	return sample.Usage >= 0 && sample.Resource == expectedResource
}

func (container *ContainerState) addCPUSample(sample *ContainerUsageSample) bool {
	// Order should not matter for the histogram, other than deduplication.
	if !sample.isValid(ResourceCPU) {
		return false // Discard invalid, In Autopilot Keep duplicate or out-of-order samples here.
	}
	container.observeQualityMetrics(sample.Usage, false, corev1.ResourceCPU)

	if sample.Usage < ResourceAmount(container.idlePercentage*float64(container.cpuPeak)) {
		klog.V(5).Infof("NICO sample CPU is idle, %v %v %v %v", sample.Usage, container.cpuPeak, container.cpuPrev, container.idlePercentage)
		sample.Usage = container.cpuPrev
	}
	container.aggregator.AddSample(sample)
	container.cpuPeak = ResourceAmountMax(sample.Usage, container.cpuPeak)
	container.cpuPrev = sample.Usage

	// Discard time sequence...
	// container.LastCPUSampleStart = sample.MeasureStart
	return true
}

func (container *ContainerState) observeQualityMetrics(usage ResourceAmount, isOOM bool, resource corev1.ResourceName) {
	if !container.aggregator.NeedsRecommendation() {
		return
	}
	updateMode := container.aggregator.GetUpdateMode()
	var usageValue float64
	switch resource {
	case corev1.ResourceCPU:
		usageValue = CoresFromCPUAmount(usage)
	case corev1.ResourceMemory:
		usageValue = BytesFromMemoryAmount(usage)
	}
	if container.aggregator.GetLastRecommendation() == nil {
		metrics_quality.ObserveQualityMetricsRecommendationMissing(usageValue, isOOM, resource, updateMode)
		return
	}
	recommendation := container.aggregator.GetLastRecommendation()[resource]
	if recommendation.IsZero() {
		metrics_quality.ObserveQualityMetricsRecommendationMissing(usageValue, isOOM, resource, updateMode)
		return
	}
	var recommendationValue float64
	switch resource {
	case corev1.ResourceCPU:
		recommendationValue = float64(recommendation.MilliValue()) / 1000.0
	case corev1.ResourceMemory:
		recommendationValue = float64(recommendation.Value())
	default:
		klog.Warningf("Unknown resource: %v", resource)
		return
	}
	metrics_quality.ObserveQualityMetrics(usageValue, recommendationValue, isOOM, resource, updateMode)
}

// GetMaxMemoryPeak returns maximum memory usage in the sample, possibly estimated from OOM
func (container *ContainerState) GetMaxMemoryPeak() ResourceAmount {
	// return ResourceAmountMax(container.memoryPeak, container.oomPeak)
	panic("not implemented")
}

func (container *ContainerState) HistogramAggregate(now time.Time) {
	container.aggregator.HistogramAggregate(now)
}

func (container *ContainerState) addMemorySample(sample *ContainerUsageSample) bool {
	if !sample.isValid(ResourceMemory) {
		return false // Discard invalid, In Autopilot Keep duplicate or out-of-order samples here.
	}
	// container.observeQualityMetrics(sample.Usage, isOOM, corev1.ResourceMemory)
	if sample.Usage < ResourceAmount(container.idlePercentage*float64(container.memoryPeak)) {
		klog.V(5).Infof("NICO sample MEM is idle, %v %v %v %v", sample.Usage, container.memoryPeak, container.memoryPrev, container.idlePercentage)
		sample.Usage = container.memoryPrev
	}
	container.aggregator.AddSample(sample)
	container.memoryPeak = ResourceAmountMax(sample.Usage, container.memoryPeak)
	container.memoryPrev = sample.Usage

	container.memoryRecent = sample.Usage

	// In Autopilot Does not process OOM...
	// ts := sample.MeasureStart
	// // We always process OOM samples.
	// if !sample.isValid(ResourceMemory) ||
	// 	(!isOOM && ts.Before(container.lastMemorySampleStart)) {
	// 	return false // Discard invalid or outdated samples.
	// }
	// container.lastMemorySampleStart = ts
	// if container.WindowEnd.IsZero() { // This is the first sample.
	// 	container.WindowEnd = ts
	// }

	// // Each container aggregates one peak per aggregation interval. If the timestamp of the
	// // current sample is earlier than the end of the current interval (WindowEnd) and is larger
	// // than the current peak, the peak is updated in the aggregation by subtracting the old value
	// // and adding the new value.
	// addNewPeak := false
	// if ts.Before(container.WindowEnd) {
	// 	oldMaxMem := container.GetMaxMemoryPeak()
	// 	if oldMaxMem != 0 && sample.Usage > oldMaxMem {
	// 		// Remove the old peak.
	// 		oldPeak := ContainerUsageSample{
	// 			MeasureStart: container.WindowEnd,
	// 			Usage:        oldMaxMem,
	// 			Request:      sample.Request,
	// 			Resource:     ResourceMemory,
	// 		}
	// 		container.aggregator.SubtractSample(&oldPeak)
	// 		addNewPeak = true
	// 	}
	// } else {
	// 	// Shift the memory aggregation window to the next interval.
	// 	memoryAggregationInterval := GetAggregationsConfig().MemoryAggregationInterval
	// 	shift := ts.Sub(container.WindowEnd).Truncate(memoryAggregationInterval) + memoryAggregationInterval
	// 	container.WindowEnd = container.WindowEnd.Add(shift)
	// 	container.memoryPeak = 0
	// 	container.oomPeak = 0
	// 	addNewPeak = true
	// }
	// container.observeQualityMetrics(sample.Usage, isOOM, corev1.ResourceMemory)
	// if addNewPeak {
	// 	newPeak := ContainerUsageSample{
	// 		MeasureStart: container.WindowEnd,
	// 		Usage:        sample.Usage,
	// 		Request:      sample.Request,
	// 		Resource:     ResourceMemory,
	// 	}
	// 	container.aggregator.AddSample(&newPeak)
	// 	if isOOM {
	// 		container.oomPeak = sample.Usage
	// 	} else {
	// 		container.memoryPeak = sample.Usage
	// 	}
	// }
	return true
}

// RecordOOM adds info regarding OOM event in the model as an artificial memory sample.
func (container *ContainerState) RecordOOM(timestamp time.Time, requestedMemory ResourceAmount) {
	// Discard old OOM
	// if timestamp.Before(container.WindowEnd.Add(-1 * GetAggregationsConfig().MemoryAggregationInterval)) {
	// 	return fmt.Errorf("OOM event will be discarded - it is too old (%v)", timestamp)
	// }
	// Get max of the request and the recent usage-based memory peak.
	// Omitting oomPeak here to protect against recommendation running too high on subsequent OOMs.
	// NICO 新逻辑：和最近的分配成功的memory取最大值，然后再bump up
	memoryUsed := ResourceAmountMax(requestedMemory, container.memoryRecent)
	memoryNeeded := ResourceAmountMax(memoryUsed+MemoryAmountFromBytes(GetAggregationsConfig().OOMMinBumpUp),
		ScaleResource(memoryUsed, GetAggregationsConfig().OOMBumpUpRatio))
	container.aggregator.AddOOM(memoryNeeded)
}

// AddSample adds a usage sample to the given ContainerState. Requires samples
// for a single resource to be passed in chronological order (i.e. in order of
// growing MeasureStart). Invalid samples (out of order or measure out of legal
// range) are discarded. Returns true if the sample was aggregated, false if it
// was discarded.
// Note: usage samples don't hold their end timestamp / duration. They are
// implicitly assumed to be disjoint when aggregating.
func (container *ContainerState) AddSample(sample *ContainerUsageSample) bool {
	switch sample.Resource {
	case ResourceCPU:
		return container.addCPUSample(sample)
	case ResourceMemory:
		return container.addMemorySample(sample)
	default:
		return false
	}
}
