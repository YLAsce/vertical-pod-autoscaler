package util

import (
	"fmt"
	"math"
	"strings"
	"time"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

var valueDelta = 1e-11

type AutopilotHisto interface {
	// Returns an approximation of the given percentile of the distribution.
	// Note: the argument passed to Percentile() is a number between
	// 0 and 1. For example 0.5 corresponds to the median and 0.9 to the
	// 90th percentile.
	// If the histogram is empty, Percentile() returns 0.0.
	Percentile(percentile float64) float64

	Average() float64

	Max() float64

	// Add a sample with a given value, weight = 1.
	AddSample(value float64)

	// Remove a sample with a given value, weight = 1.
	SubtractSample(value float64)

	// Add all samples from another histogram. Requires the histograms to be
	// of the exact same type.
	Merge(other AutopilotHisto)

	// Return the number of aggregations made.
	AggregateNums() int

	// Has at least 1 aggration which is not 0
	HasValidAggregation() bool

	// Aggregate histogram in time window statistics
	Aggregate(operationTime time.Time)

	// Returns true if the histogram is equal to another one. The two
	// histograms must use the same HistogramOptions object (not two
	// different copies).
	// If the two histograms are not of the same runtime type returns false.
	Equals(other AutopilotHisto) bool

	// Returns a human-readable text description of the histogram.
	String() string

	// SaveToChekpoint returns a representation of the histogram as a
	// HistogramCheckpoint. During conversion buckets with small weights
	// can be omitted.
	SaveToChekpoint() (*vpa_types.HistogramCheckpoint, error)

	// LoadFromCheckpoint loads data from the checkpoint into the histogram
	// by appending samples.
	LoadFromCheckpoint(*vpa_types.HistogramCheckpoint) error
}

func NewAutopilotHisto(options HistogramOptions, halfLife time.Duration, n int, defaultAggregationDuration time.Duration) AutopilotHisto {
	a := autopilotHisto{
		options:      options,
		halfLife:     halfLife,
		lastSamplesN: n,

		cumulativeWeightedAverageLower: 0.0, // To Calculate below

		lastAggregationTime: time.Time{},
		aggregationDuration: defaultAggregationDuration,
		aggregateNums:       0,

		currentBucketWeight:                make([]int, options.NumBuckets()),
		cumulativeWeightedAverageUpper:     0.0,
		cumulativeAdjustedUsage:            make([]float64, options.NumBuckets()),
		cumulativeAdjustedUsageWeightTotal: 0.0,
		cumulativeMaxWindow:                make([]float64, n), // sliding window, N is small, no need to improve algorithm..
		cumulativeMaxHeadPosition:          0,
	}
	a.calCumulativeWeightedAverageLower()
	return &a
}

type autopilotHisto struct {
	options      HistogramOptions
	halfLife     time.Duration
	lastSamplesN int

	cumulativeWeightedAverageLower float64

	lastAggregationTime time.Time
	aggregationDuration time.Duration
	aggregateNums       int

	currentBucketWeight                []int
	cumulativeWeightedAverageUpper     float64
	cumulativeAdjustedUsage            []float64
	cumulativeAdjustedUsageWeightTotal float64
	cumulativeMaxWindow                []float64
	cumulativeMaxHeadPosition          int
}

func (ah *autopilotHisto) calCumulativeWeightedAverageLower() {
	ah.cumulativeWeightedAverageLower = 0.0
	for i := 0; i <= ah.lastSamplesN; i++ { // 0 to N inclusive
		ah.cumulativeWeightedAverageLower += ah.calExponentialDecayingWeight(time.Duration(i) * ah.aggregationDuration)
	}
}

func (ah *autopilotHisto) calExponentialDecayingWeight(t time.Duration) float64 {
	up := -float64(t) / float64(ah.halfLife)
	return math.Pow(2, up)
}

func (ah *autopilotHisto) currentAverageUsage() float64 {
	upper := 0.0
	lower := 0
	for j := 0; j < ah.options.NumBuckets(); j++ {
		upper += ah.options.GetBucketEnd(j) * float64(ah.currentBucketWeight[j])
		lower += ah.currentBucketWeight[j]
	}
	if lower == 0 {
		return 0.0
	}
	return upper / float64(lower)
}

func (ah *autopilotHisto) AddSample(value float64) {
	bucket := ah.options.FindBucket(value)
	ah.currentBucketWeight[bucket] += 1
}

func (ah *autopilotHisto) SubtractSample(value float64) {
	bucket := ah.options.FindBucket(value)
	ah.currentBucketWeight[bucket] -= 1
}

func (ah *autopilotHisto) Merge(other AutopilotHisto) {
	o := other.(*autopilotHisto)
	if ah.options != o.options || ah.halfLife != o.halfLife || ah.lastSamplesN != o.lastSamplesN {
		panic("Can't merge histograms with different options / halflife / n")
	}
	// Merge current weights
	for i := 0; i < ah.options.NumBuckets(); i++ {
		ah.currentBucketWeight[i] += o.currentBucketWeight[i]
	}

	// Merge aggragation data
	if ah.AggregateNums() == 0 {
		// If Not aggragated, direct copy
		ah.lastAggregationTime = o.lastAggregationTime
		ah.aggregationDuration = o.aggregationDuration
		ah.aggregateNums = o.aggregateNums

		ah.cumulativeWeightedAverageUpper = o.cumulativeWeightedAverageUpper
		ah.cumulativeAdjustedUsage = o.cumulativeAdjustedUsage
		ah.cumulativeAdjustedUsageWeightTotal = o.cumulativeAdjustedUsageWeightTotal
		ah.cumulativeMaxWindow = o.cumulativeMaxWindow
		ah.cumulativeMaxHeadPosition = o.cumulativeMaxHeadPosition
		return
	}

	// If Already aggregated, should be in the same aggregation batch
	if ah.lastAggregationTime != o.lastAggregationTime || ah.aggregationDuration != o.aggregationDuration {
		panic("Can't merge histograms with different aggregation time")
	}

	// S_avg = Average of the two, because the number of samples are always the same, so the sum of weights are the same. Calculation in pink folder.
	ah.cumulativeWeightedAverageUpper = (ah.cumulativeWeightedAverageUpper + o.cumulativeWeightedAverageUpper) / 2
	// S_pj buckets weights = the sum of the two
	for i := 0; i < ah.options.NumBuckets(); i++ {
		ah.cumulativeAdjustedUsage[i] += o.cumulativeAdjustedUsage[i]
	}
	ah.cumulativeAdjustedUsageWeightTotal += o.cumulativeAdjustedUsageWeightTotal
	// S_max = max of the two
	ahHead, oHead := ah.cumulativeMaxHeadPosition, o.cumulativeMaxHeadPosition
	for i := 0; i < ah.lastSamplesN; i++ {
		ah.cumulativeMaxWindow[ahHead] = math.Max(ah.cumulativeMaxWindow[ahHead], o.cumulativeMaxWindow[oHead])
		ahHead = (ahHead + 1) % ah.lastSamplesN
		oHead = (oHead + 1) % o.lastSamplesN
	}
}

func (ah *autopilotHisto) Percentile(percentile float64) float64 {
	if ah.AggregateNums() == 0 {
		return 0.0
	}
	partialSum := 0.0
	threshold := percentile * ah.cumulativeAdjustedUsageWeightTotal
	bucket := 0
	for ; bucket < ah.options.NumBuckets(); bucket++ {
		partialSum += ah.cumulativeAdjustedUsage[bucket]
		if partialSum >= threshold {
			break
		}
	}
	return ah.options.GetBucketEnd(bucket)
}

func (ah *autopilotHisto) Average() float64 {
	return ah.cumulativeWeightedAverageUpper / ah.cumulativeWeightedAverageLower
}

func (ah *autopilotHisto) Max() float64 {
	maxB := 0.0
	for i := 0; i < ah.lastSamplesN; i++ {
		maxB = math.Max(maxB, ah.cumulativeMaxWindow[i])
	}
	return maxB
}

func (ah *autopilotHisto) AggregateNums() int {
	return ah.aggregateNums
}

func (ah *autopilotHisto) HasValidAggregation() bool {
	return ah.Max() > valueDelta
}

func (ah *autopilotHisto) Aggregate(operationTime time.Time) {
	// Process the time
	if ah.AggregateNums() >= 1 {
		ah.aggregationDuration = operationTime.Sub(ah.lastAggregationTime)
	}
	ah.lastAggregationTime = operationTime
	ah.aggregateNums++

	// Process Max
	maxBucket := ah.options.NumBuckets() - 1
	for ; maxBucket >= 0; maxBucket-- {
		if ah.currentBucketWeight[maxBucket] > 0 {
			break
		}
	}
	maxVal := 0.0
	if maxBucket >= 0 {
		maxVal = ah.options.GetBucketEnd(maxBucket)
	}
	ah.cumulativeMaxWindow[ah.cumulativeMaxHeadPosition] = maxVal
	ah.cumulativeMaxHeadPosition = (ah.cumulativeMaxHeadPosition + 1) % ah.lastSamplesN

	// Process Average
	ah.cumulativeWeightedAverageUpper *= ah.calExponentialDecayingWeight(ah.aggregationDuration)
	ah.cumulativeWeightedAverageUpper += ah.calExponentialDecayingWeight(time.Duration(0)) * ah.currentAverageUsage()
	ah.calCumulativeWeightedAverageLower()

	// Process Adjusted Usage
	ah.cumulativeAdjustedUsageWeightTotal = 0.0
	for i := 0; i < ah.options.NumBuckets(); i++ {
		ah.cumulativeAdjustedUsage[i] *= ah.calExponentialDecayingWeight(ah.aggregationDuration)
		ah.cumulativeAdjustedUsage[i] += ah.calExponentialDecayingWeight(time.Duration(0)) * float64(ah.currentBucketWeight[i]) * ah.options.GetBucketEnd(i)
		ah.cumulativeAdjustedUsageWeightTotal += ah.cumulativeAdjustedUsage[i]
	}

	// Clear current bucket, ready for the samples in next window ...
	for i := 0; i < ah.options.NumBuckets(); i++ {
		ah.currentBucketWeight[i] = 0
	}
}

func (ah *autopilotHisto) String() string {
	lines := []string{
		"",
		"+++++++++++++++++++++++++++",
		fmt.Sprintf("Time: Halflife: %v, N: %v, LastAggregation: %v, Aggregation duration: %v", ah.halfLife, ah.lastSamplesN, ah.lastAggregationTime, ah.aggregationDuration),
		fmt.Sprintf("Current buckets: %+v", ah.currentBucketWeight),
		fmt.Sprintf("Max: Window: %+v, HeadPosition: %v", ah.cumulativeMaxWindow, ah.cumulativeMaxHeadPosition),
		fmt.Sprintf("Average: Upper: %v, Lower: %v", ah.cumulativeWeightedAverageUpper, ah.cumulativeWeightedAverageLower),
		fmt.Sprintf("Adjusted Usage: Weights: %+v, Total: %v", ah.cumulativeAdjustedUsage, ah.cumulativeAdjustedUsageWeightTotal),
		"---------------------------",
	}
	return strings.Join(lines, "\n")
}

func (ah *autopilotHisto) Equals(other AutopilotHisto) bool {
	// TODO
	return true
}

func (ah *autopilotHisto) SaveToChekpoint() (*vpa_types.HistogramCheckpoint, error) {
	result := vpa_types.HistogramCheckpoint{
		BucketWeights: make(map[int]uint32),
	}

	// TODO
	return &result, nil
}

func (ah *autopilotHisto) LoadFromCheckpoint(checkpoint *vpa_types.HistogramCheckpoint) error {
	// TODO
	return nil
}
