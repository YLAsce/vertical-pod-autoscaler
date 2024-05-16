package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_autopilotHisto_Basic(t *testing.T) {
	options, err := NewLinearHistogramOptions(1.0, 10, weightEpsilon)
	assert.Nil(t, err)
	ah := NewAutopilotHisto(options, time.Minute, 1, time.Minute, AutopilotAddSampleModeDistribution)
	ah.AddSample(0.5)
	assert.InDelta(t, 0, ah.Max(), valueDelta)
	assert.InDelta(t, 0, ah.Average(), valueDelta)
	assert.InDelta(t, 0, ah.Percentile(0.5), valueDelta)

	ah.Aggregate(startTime)
	ah.AddSample(0.69)
	ah.AddSample(0.49)
	// t.Log(ah.String())
	assert.InDelta(t, 0.6, ah.Max(), valueDelta)
	assert.InDelta(t, 0.4, ah.Average(), valueDelta)
	assert.InDelta(t, 0.6, ah.Percentile(0.5), valueDelta)

	ah.Aggregate(startTime.Add(2 * time.Minute))
	// t.Log(ah.String())
	assert.InDelta(t, 0.7, ah.Max(), valueDelta)
	assert.InDelta(t, 0.6, ah.Average(), valueDelta)
	assert.InDelta(t, 0.7, ah.Percentile(0.5), valueDelta)
}

func Test_autopilotHisto_Begin(t *testing.T) {
	options, err := NewLinearHistogramOptions(1.0, 10, weightEpsilon)
	assert.Nil(t, err)
	ah := NewAutopilotHisto(options, time.Minute, 1, time.Minute, AutopilotAddSampleModeDistribution)
	ah.Aggregate(startTime)
	// t.Log(ah.String())
	assert.InDelta(t, 0, ah.Max(), valueDelta)
	assert.InDelta(t, 0, ah.Average(), valueDelta)
	assert.InDelta(t, 0.1, ah.Percentile(0.0), valueDelta) // The end of the minimum bucket,not 0!!!
}

func Test_autopilotHisto_Max_Window(t *testing.T) {
	options, err := NewLinearHistogramOptions(1.0, 10, weightEpsilon)
	assert.Nil(t, err)
	ah := NewAutopilotHisto(options, time.Minute, 2, time.Minute, AutopilotAddSampleModeDistribution)
	ah.AddSample(0.69)
	ah.AddSample(0.49)

	ah.Aggregate(startTime)
	ah.AddSample(0.5)
	// t.Log(ah.String())
	assert.InDelta(t, 0.7, ah.Max(), valueDelta)

	ah.Aggregate(startTime.Add(2 * time.Minute))
	ah.AddSample(0.4)
	// t.Log(ah.String())
	assert.InDelta(t, 0.7, ah.Max(), valueDelta)

	ah.Aggregate(startTime.Add(4 * time.Minute))
	// t.Log(ah.String())
	assert.InDelta(t, 0.6, ah.Max(), valueDelta)
}

func Test_autopilotHisto_Merge(t *testing.T) {
	options, err := NewLinearHistogramOptions(1.0, 10, weightEpsilon)
	assert.Nil(t, err)
	ah1 := NewAutopilotHisto(options, time.Minute, 2, time.Minute, AutopilotAddSampleModeDistribution)
	ah1.AddSample(0.69)
	ah1.AddSample(0.49)

	ah1.Aggregate(startTime)
	ah1.AddSample(0.5)
	// t.Log(ah1.String())

	ahEmpty := NewAutopilotHisto(options, time.Minute, 2, time.Minute, AutopilotAddSampleModeDistribution)
	ahEmpty.Merge(ah1)
	// t.Log(ahEmpty.String())
	assert.InDelta(t, 0.7, ahEmpty.Max(), valueDelta)

	ah3 := NewAutopilotHisto(options, time.Minute, 2, time.Minute, AutopilotAddSampleModeDistribution)
	ah3.AddSample(0.19)
	ah3.AddSample(0.39)
	ah3.Aggregate(startTime)
	ah3.AddSample(0.39)
	ah1.Merge(ah3)
	// t.Log(ah3.String())
	// t.Log(ah1.String())
	assert.InDelta(t, 0.7, ahEmpty.Max(), valueDelta)
}
