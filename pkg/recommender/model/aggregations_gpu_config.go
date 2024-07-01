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

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/util"
	"k8s.io/klog/v2"
)

// AggregationsConfig is used to configure aggregation behaviour.
type AggregationsGPUConfig struct {

	// CPUHistogramOptions are options to be used by histograms that store
	// CPU measures expressed in cores.
	SMHistogramOptions util.HistogramOptions
	// MemoryHistogramOptions are options to be used by histograms that
	// store memory measures expressed in bytes.
	MemoryHistogramOptions util.HistogramOptions

	//Below are Autopilot configs
	SMHistogramDecayHalfLife         time.Duration
	MemoryHistogramDecayHalfLife     time.Duration
	SMLastSamplesN                   int
	MemoryLastSamplesN               int
	SMDefaultAggregationDuration     time.Duration
	MemoryDefaultAggregationDuration time.Duration
}

const (
	// DefaultMemoryAggregationIntervalCount is the default value for MemoryAggregationIntervalCount.
	DefaultGPUMemoryAggregationIntervalCount = 8
	// DefaultMemoryAggregationInterval is the default value for MemoryAggregationInterval.
	// which the peak memory usage is computed.
	DefaultGPUMemoryAggregationInterval = time.Hour * 24

	// DefaultMemoryHistogramDecayHalfLife is the default value for MemoryHistogramDecayHalfLife.
	DefaultGPUMemoryHistogramDecayHalfLife = time.Hour * 24
	// DefaultCPUHistogramDecayHalfLife is the default value for CPUHistogramDecayHalfLife.
	// CPU usage sample to lose half of its weight.
	DefaultGPUSMHistogramDecayHalfLife = time.Hour * 24

	// Autopilot special
	DefaultGPUSMHistogramMaxValue      = 3.0           // In GIs
	DefaultGPUMemoryHistogramMaxValue  = 40000000000.0 // 40GB (1000 not 1024 here)
	DefaultGPUSMHistogramBucketNum     = 3
	DefaultGPUMemoryHistogramBucketNum = 4
)

func (a *AggregationsGPUConfig) gpuSmHistogramOptions(maxValue float64, bucketNum int) util.HistogramOptions {
	// CPU histograms use exponential bucketing scheme with the smallest bucket
	// size of 0.01 core, max of 1000.0 cores and the relative error of HistogramRelativeError.
	//
	// When parameters below are changed SupportedCheckpointVersion has to be bumped.
	options, err := util.NewLinearHistogramOptions(maxValue, bucketNum, epsilon) // Epsilon is unused in Autopilot...
	if err != nil {
		panic("Invalid GPU SM histogram options") // Should not happen.
	}
	return options
}

func (a *AggregationsGPUConfig) gpuMemoryHistogramOptions(maxValue float64, bucketNum int) util.HistogramOptions {
	// Memory histograms use exponential bucketing scheme with the smallest
	// bucket size of 10MB, max of 1TB and the relative error of HistogramRelativeError.
	//
	// When parameters below are changed SupportedCheckpointVersion has to be bumped.
	options, err := util.NewLinearHistogramOptions(maxValue, bucketNum, epsilon) // Epsilon is unused in Autopilot...
	if err != nil {
		panic("Invalid GPU memory histogram options") // Should not happen.
	}
	return options
}

// NewAggregationsConfig creates a new AggregationsConfig based on the supplied parameters and default values.
func NewAggregationsGPUConfig(memoryHistogramDecayHalfLife, smHistogramDecayHalfLife time.Duration,
	smDefaultAggregationDuration, memoryDefaultAggregationDuration time.Duration,
	smLastSamplesN, memoryLastSamplesN int,
	smHistogramMaxValue float64, smHistogramBucketNum int, memoryHistogramMaxValue float64, memoryHistogramBucketNum int) *AggregationsGPUConfig {
	a := &AggregationsGPUConfig{
		SMHistogramDecayHalfLife:         smHistogramDecayHalfLife,
		MemoryHistogramDecayHalfLife:     memoryHistogramDecayHalfLife,
		SMLastSamplesN:                   smLastSamplesN,
		MemoryLastSamplesN:               memoryLastSamplesN,
		SMDefaultAggregationDuration:     smDefaultAggregationDuration,
		MemoryDefaultAggregationDuration: memoryDefaultAggregationDuration,
	}
	// Calculate per-histogram size and set
	a.SMHistogramOptions = a.gpuSmHistogramOptions(smHistogramMaxValue, smHistogramBucketNum)
	a.MemoryHistogramOptions = a.gpuMemoryHistogramOptions(memoryHistogramMaxValue, memoryHistogramBucketNum)
	return a
}

var aggregationsGPUConfig *AggregationsGPUConfig

// GetAggregationsConfig gets the aggregations config. Initializes to default values if not initialized already.
func GetAggregationsGPUConfig() *AggregationsGPUConfig {
	if aggregationsGPUConfig == nil {
		klog.V(4).Infof("Aggregation config Not initialized!")
		aggregationsGPUConfig = NewAggregationsGPUConfig(
			DefaultGPUMemoryHistogramDecayHalfLife, DefaultGPUSMHistogramDecayHalfLife,
			DefaultAggregationDuration, DefaultAggregationDuration,
			DefaultLastSamplesN, DefaultLastSamplesN,
			DefaultGPUSMHistogramMaxValue, DefaultGPUSMHistogramBucketNum, DefaultGPUMemoryHistogramMaxValue, DefaultGPUMemoryHistogramBucketNum,
		)
	}

	return aggregationsGPUConfig
}

// InitializeAggregationsConfig initializes the global aggregations configuration. Not thread-safe.
func InitializeAggregationsGPUConfig(config *AggregationsGPUConfig) {
	aggregationsGPUConfig = config
}
