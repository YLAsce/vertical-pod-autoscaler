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
type AggregationsConfig struct {
	// MemoryAggregationInterval is the length of a single interval, for
	// which the peak memory usage is computed.
	// Memory usage peaks are aggregated in multiples of this interval. In other words
	// there is one memory usage sample per interval (the maximum usage over that
	// interval).
	MemoryAggregationInterval time.Duration
	// MemoryAggregationWindowIntervalCount is the number of consecutive MemoryAggregationIntervals
	// which make up the MemoryAggregationWindowLength which in turn is the period for memory
	// usage aggregation by VPA.
	MemoryAggregationIntervalCount int64
	// CPUHistogramOptions are options to be used by histograms that store
	// CPU measures expressed in cores.
	CPUHistogramOptions util.HistogramOptions
	// MemoryHistogramOptions are options to be used by histograms that
	// store memory measures expressed in bytes.
	MemoryHistogramOptions util.HistogramOptions
	// HistogramBucketSizeGrowth defines the growth rate of the histogram buckets.
	// Each bucket is wider than the previous one by this fraction.
	HistogramBucketSizeGrowth float64

	// OOMBumpUpRatio specifies the memory bump up ratio when OOM occurred.
	OOMBumpUpRatio float64
	// OOMMinBumpUp specifies the minimal increase of memory when OOM occurred in bytes.
	OOMMinBumpUp float64

	//Below are Autopilot configs
	CPUHistogramDecayHalfLife        time.Duration
	MemoryHistogramDecayHalfLife     time.Duration
	CPULastSamplesN                  int
	MemoryLastSamplesN               int
	CPUDefaultAggregationDuration    time.Duration
	MemoryDefaultAggregationDuration time.Duration
}

const (
	// minSampleWeight is the minimal weight of any sample (prior to including decaying factor)
	minSampleWeight = 0.1
	// epsilon is the minimal weight kept in histograms, it should be small enough that old samples
	// (just inside MemoryAggregationWindowLength) added with minSampleWeight are still kept
	epsilon = 0.001 * minSampleWeight
	// DefaultMemoryAggregationIntervalCount is the default value for MemoryAggregationIntervalCount.
	DefaultMemoryAggregationIntervalCount = 8
	// DefaultMemoryAggregationInterval is the default value for MemoryAggregationInterval.
	// which the peak memory usage is computed.
	DefaultMemoryAggregationInterval = time.Hour * 24
	// DefaultHistogramBucketSizeGrowth is the default value for HistogramBucketSizeGrowth.
	DefaultHistogramBucketSizeGrowth = 0.05 // Make each bucket 5% larger than the previous one.
	// DefaultMemoryHistogramDecayHalfLife is the default value for MemoryHistogramDecayHalfLife.
	DefaultMemoryHistogramDecayHalfLife = time.Hour * 24
	// DefaultCPUHistogramDecayHalfLife is the default value for CPUHistogramDecayHalfLife.
	// CPU usage sample to lose half of its weight.
	DefaultCPUHistogramDecayHalfLife = time.Hour * 24
	// DefaultOOMBumpUpRatio is the default value for OOMBumpUpRatio.
	DefaultOOMBumpUpRatio float64 = 1.2 // Memory is increased by 20% after an OOMKill.
	// DefaultOOMMinBumpUp is the default value for OOMMinBumpUp.
	DefaultOOMMinBumpUp float64 = 100 * 1024 * 1024 // Memory is increased by at least 100MB after an OOMKill.

	// Autopilot special
	DefaultAggregationDuration      = time.Minute * 5
	DefaultLastSamplesN             = 5
	DefaultCPUHistogramMaxValue     = 2.0          // In Cores
	DefaultMemoryHistogramMaxValue  = 1000000000.0 // Assume 1GB, Unit to be confirmed
	DefaultCPUHistogramBucketNum    = 400
	DefaultMemoryHistogramBucketNum = 400
)

// GetMemoryAggregationWindowLength returns the total length of the memory usage history aggregated by VPA.
func (a *AggregationsConfig) GetMemoryAggregationWindowLength() time.Duration {
	return a.MemoryAggregationInterval * time.Duration(a.MemoryAggregationIntervalCount)
}

func (a *AggregationsConfig) cpuHistogramOptions(maxValue float64, bucketNum int) util.HistogramOptions {
	// CPU histograms use exponential bucketing scheme with the smallest bucket
	// size of 0.01 core, max of 1000.0 cores and the relative error of HistogramRelativeError.
	//
	// When parameters below are changed SupportedCheckpointVersion has to be bumped.
	options, err := util.NewLinearHistogramOptions(maxValue, bucketNum, epsilon) // Epsilon is unused in Autopilot...
	if err != nil {
		panic("Invalid CPU histogram options") // Should not happen.
	}
	return options
}

func (a *AggregationsConfig) memoryHistogramOptions(maxValue float64, bucketNum int) util.HistogramOptions {
	// Memory histograms use exponential bucketing scheme with the smallest
	// bucket size of 10MB, max of 1TB and the relative error of HistogramRelativeError.
	//
	// When parameters below are changed SupportedCheckpointVersion has to be bumped.
	options, err := util.NewLinearHistogramOptions(maxValue, bucketNum, epsilon) // Epsilon is unused in Autopilot...
	if err != nil {
		panic("Invalid memory histogram options") // Should not happen.
	}
	return options
}

// NewAggregationsConfig creates a new AggregationsConfig based on the supplied parameters and default values.
func NewAggregationsConfig(memoryAggregationInterval time.Duration,
	memoryAggregationIntervalCount int64,
	memoryHistogramDecayHalfLife, cpuHistogramDecayHalfLife time.Duration,
	oomBumpUpRatio float64, oomMinBumpUp float64,
	cpuDefaultAggregationDuration, memoryDefaultAggregationDuration time.Duration,
	cpuLastSamplesN, memoryLastSamplesN int,
	cpuHistogramMaxValue float64, cpuHistogramBucketNum int, memoryHistogramMaxValue float64, memoryHistogramBucketNum int) *AggregationsConfig {
	a := &AggregationsConfig{
		MemoryAggregationInterval:      memoryAggregationInterval,
		MemoryAggregationIntervalCount: memoryAggregationIntervalCount,
		HistogramBucketSizeGrowth:      DefaultHistogramBucketSizeGrowth,
		OOMBumpUpRatio:                 oomBumpUpRatio,
		OOMMinBumpUp:                   oomMinBumpUp,

		CPUHistogramDecayHalfLife:        cpuHistogramDecayHalfLife,
		MemoryHistogramDecayHalfLife:     memoryHistogramDecayHalfLife,
		CPUDefaultAggregationDuration:    cpuDefaultAggregationDuration,
		MemoryDefaultAggregationDuration: memoryDefaultAggregationDuration,
		CPULastSamplesN:                  cpuLastSamplesN,
		MemoryLastSamplesN:               memoryLastSamplesN,
	}
	// Calculate per-histogram size and set
	a.CPUHistogramOptions = a.cpuHistogramOptions(cpuHistogramMaxValue, cpuHistogramBucketNum)
	a.MemoryHistogramOptions = a.memoryHistogramOptions(memoryHistogramMaxValue, memoryHistogramBucketNum)
	return a
}

var aggregationsConfig *AggregationsConfig

// GetAggregationsConfig gets the aggregations config. Initializes to default values if not initialized already.
func GetAggregationsConfig() *AggregationsConfig {
	if aggregationsConfig == nil {
<<<<<<< HEAD
		klog.V(4).Infof("Aggregation config Not initialized!")
=======
>>>>>>> 83b4a7b2995f5cda7ade0e061d546bffbdfb3724
		aggregationsConfig = NewAggregationsConfig(DefaultMemoryAggregationInterval,
			DefaultMemoryAggregationIntervalCount,
			DefaultMemoryHistogramDecayHalfLife, DefaultCPUHistogramDecayHalfLife,
			DefaultOOMBumpUpRatio, DefaultOOMMinBumpUp,
			DefaultAggregationDuration, DefaultAggregationDuration,
			DefaultLastSamplesN, DefaultLastSamplesN,
			DefaultCPUHistogramMaxValue, DefaultCPUHistogramBucketNum, DefaultMemoryHistogramMaxValue, DefaultMemoryHistogramBucketNum,
		)
	}

	return aggregationsConfig
}

// InitializeAggregationsConfig initializes the global aggregations configuration. Not thread-safe.
func InitializeAggregationsConfig(config *AggregationsConfig) {
	aggregationsConfig = config
}
