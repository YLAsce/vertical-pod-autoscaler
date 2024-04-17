package util

import (
	"testing"
	"time"
)

func Test_autopilotHisto_Average(t *testing.T) {
	type fields struct {
		options                            HistogramOptions
		halfLife                           time.Duration
		lastSamplesN                       int
		cumulativeWeightedAverageLower     float64
		lastAggregationTime                time.Time
		aggregationDuration                time.Duration
		currentBucketWeight                []int
		cumulativeWeightedAverageUpper     float64
		cumulativeAdjustedUsage            []float64
		cumulativeAdjustedUsageWeightTotal float64
		cumulativeMaxWindow                []float64
		cumulativeMaxHeadPosition          int
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ah := &autopilotHisto{
				options:                            tt.fields.options,
				halfLife:                           tt.fields.halfLife,
				lastSamplesN:                       tt.fields.lastSamplesN,
				cumulativeWeightedAverageLower:     tt.fields.cumulativeWeightedAverageLower,
				lastAggregationTime:                tt.fields.lastAggregationTime,
				aggregationDuration:                tt.fields.aggregationDuration,
				currentBucketWeight:                tt.fields.currentBucketWeight,
				cumulativeWeightedAverageUpper:     tt.fields.cumulativeWeightedAverageUpper,
				cumulativeAdjustedUsage:            tt.fields.cumulativeAdjustedUsage,
				cumulativeAdjustedUsageWeightTotal: tt.fields.cumulativeAdjustedUsageWeightTotal,
				cumulativeMaxWindow:                tt.fields.cumulativeMaxWindow,
				cumulativeMaxHeadPosition:          tt.fields.cumulativeMaxHeadPosition,
			}
			if got := ah.Average(); got != tt.want {
				t.Errorf("autopilotHisto.Average() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_autopilotHisto_Max(t *testing.T) {
	type fields struct {
		options                            HistogramOptions
		halfLife                           time.Duration
		lastSamplesN                       int
		cumulativeWeightedAverageLower     float64
		lastAggregationTime                time.Time
		aggregationDuration                time.Duration
		currentBucketWeight                []int
		cumulativeWeightedAverageUpper     float64
		cumulativeAdjustedUsage            []float64
		cumulativeAdjustedUsageWeightTotal float64
		cumulativeMaxWindow                []float64
		cumulativeMaxHeadPosition          int
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ah := &autopilotHisto{
				options:                            tt.fields.options,
				halfLife:                           tt.fields.halfLife,
				lastSamplesN:                       tt.fields.lastSamplesN,
				cumulativeWeightedAverageLower:     tt.fields.cumulativeWeightedAverageLower,
				lastAggregationTime:                tt.fields.lastAggregationTime,
				aggregationDuration:                tt.fields.aggregationDuration,
				currentBucketWeight:                tt.fields.currentBucketWeight,
				cumulativeWeightedAverageUpper:     tt.fields.cumulativeWeightedAverageUpper,
				cumulativeAdjustedUsage:            tt.fields.cumulativeAdjustedUsage,
				cumulativeAdjustedUsageWeightTotal: tt.fields.cumulativeAdjustedUsageWeightTotal,
				cumulativeMaxWindow:                tt.fields.cumulativeMaxWindow,
				cumulativeMaxHeadPosition:          tt.fields.cumulativeMaxHeadPosition,
			}
			if got := ah.Max(); got != tt.want {
				t.Errorf("autopilotHisto.Max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_autopilotHisto_Percentile(t *testing.T) {
	type fields struct {
		options                            HistogramOptions
		halfLife                           time.Duration
		lastSamplesN                       int
		cumulativeWeightedAverageLower     float64
		lastAggregationTime                time.Time
		aggregationDuration                time.Duration
		currentBucketWeight                []int
		cumulativeWeightedAverageUpper     float64
		cumulativeAdjustedUsage            []float64
		cumulativeAdjustedUsageWeightTotal float64
		cumulativeMaxWindow                []float64
		cumulativeMaxHeadPosition          int
	}
	type args struct {
		percentile float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ah := &autopilotHisto{
				options:                            tt.fields.options,
				halfLife:                           tt.fields.halfLife,
				lastSamplesN:                       tt.fields.lastSamplesN,
				cumulativeWeightedAverageLower:     tt.fields.cumulativeWeightedAverageLower,
				lastAggregationTime:                tt.fields.lastAggregationTime,
				aggregationDuration:                tt.fields.aggregationDuration,
				currentBucketWeight:                tt.fields.currentBucketWeight,
				cumulativeWeightedAverageUpper:     tt.fields.cumulativeWeightedAverageUpper,
				cumulativeAdjustedUsage:            tt.fields.cumulativeAdjustedUsage,
				cumulativeAdjustedUsageWeightTotal: tt.fields.cumulativeAdjustedUsageWeightTotal,
				cumulativeMaxWindow:                tt.fields.cumulativeMaxWindow,
				cumulativeMaxHeadPosition:          tt.fields.cumulativeMaxHeadPosition,
			}
			if got := ah.Percentile(tt.args.percentile); got != tt.want {
				t.Errorf("autopilotHisto.Percentile() = %v, want %v", got, tt.want)
			}
		})
	}
}
