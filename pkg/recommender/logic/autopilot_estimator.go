package logic

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/util"
	"k8s.io/klog/v2"
)

const (
	spikePercentileValue                      = 60
	minSafetyMargin                   float64 = 0.1
	maxSafetyMargin                   float64 = 0.15
	defaultFluctuationReducerDuration         = time.Hour
)

type AutopilotResourceEstimator interface {
	GetResourceEstimation(containerName string, s *model.AggregateContainerState) (model.Resources, error)
}

type AutopilotGPUResourceEstimator interface {
	GetResourceEstimation(containerName string, s *model.AggregateGPUState) (model.Resources, error)
}

type mlGPUEstimator struct {
	smLastSamplesN, memoryLastSamplesN int
}

func NewMLGPUEstimator(smLastSamplesN, memoryLastSamplesN int) AutopilotGPUResourceEstimator {
	return &mlGPUEstimator{
		smLastSamplesN:     smLastSamplesN,
		memoryLastSamplesN: memoryLastSamplesN,
	}
}

func (e *mlGPUEstimator) GetResourceEstimation(containerName string, s *model.AggregateGPUState) (model.Resources, error) {
	klog.V(5).Infof("Algorithm Enter GPU" + containerName)
	s.MLRecommenderSM.CalculateOnce()
	s.MLRecommenderMemory.CalculateOnce()
	SMResult := s.MLRecommenderSM.GetRecommendation()
	MemoryResult := s.MLRecommenderMemory.GetRecommendation()

	var err error = nil
	if s.AggregateSMUsage.AggregateNums() <= e.smLastSamplesN || s.AggregateMemoryUsage.AggregateNums() <= e.memoryLastSamplesN {
		err = errors.New("Too early for ML to give recommendation, wait for 5N time")
	}
	return model.Resources{
		model.ResourceGPUSM:     SMResult,
		model.ResourceGPUMemory: MemoryResult,
	}, err
}

type mlEstimator struct {
	cpuLastSamplesN, memoryLastSamplesN int
}

func NewMLEstimator(cpuLastSamplesN, memoryLastSamplesN int) AutopilotResourceEstimator {
	return &mlEstimator{
		cpuLastSamplesN:    cpuLastSamplesN,
		memoryLastSamplesN: memoryLastSamplesN,
	}
}

func (e *mlEstimator) GetResourceEstimation(containerName string, s *model.AggregateContainerState) (model.Resources, error) {
	klog.V(5).Infof("Algorithm Enter ML" + containerName)
	s.MLRecommenderCPU.CalculateOnce()
	s.MLRecommenderMemory.CalculateOnce()
	CPUResult := s.MLRecommenderCPU.GetRecommendation()
	MemoryResult := s.MLRecommenderMemory.GetRecommendation()

	var err error = nil
	if s.AggregateCPUUsage.AggregateNums() <= e.cpuLastSamplesN || s.AggregateMemoryUsage.AggregateNums() <= e.memoryLastSamplesN {
		err = errors.New("Too early for ML to give recommendation, wait for 5N time")
	}
	return model.Resources{
		model.ResourceCPU:    CPUResult,
		model.ResourceMemory: MemoryResult,
	}, err
}

type autopilotEstimator struct {
	cpuEstimator    AutopilotSingleEstimator
	memoryEstimator AutopilotSingleEstimator
}

func getSingleEstimatorFromPolicy(policy string, N int) AutopilotSingleEstimator {
	switch policy {
	case "avg":
		return NewAutopilotAverageEstimator(N)
	case "max":
		return NewAutopilotMaxEstimator(N)
	case "spike":
		return NewAutopilotSpikeEstimator(spikePercentileValue, N)
	default:
		if strings.HasPrefix(policy, "sp_") {
			percentileInt, err := strconv.Atoi(policy[len("sp_"):])
			if err == nil {
				return NewAutopilotPercentileEstimator(percentileInt, N)
			}
		}
	}
	// Error decoding int, or unrecognized policy string
	panic("Wrong autopilot recommender config")
}
func NewAutopilotEstimator(cpuRecommendPolicy string, memoryRecommendPolicy string, cpuLastSamplesN, memoryLastSamplesN int) AutopilotResourceEstimator {
	return &autopilotEstimator{
		cpuEstimator:    getSingleEstimatorFromPolicy(cpuRecommendPolicy, cpuLastSamplesN),
		memoryEstimator: getSingleEstimatorFromPolicy(memoryRecommendPolicy, memoryLastSamplesN),
	}
}

func (e *autopilotEstimator) GetResourceEstimation(containerName string, s *model.AggregateContainerState) (model.Resources, error) {
	klog.V(5).Infof("Algorithm Enter Autopilot" + containerName)
	klog.V(3).Info("Start Estimate CPU:" + containerName)
	rawCPUResult, err1 := e.cpuEstimator.GetRawEstimation(s.AggregateCPUUsage)
	klog.V(3).Info("Start Estimate RAM:" + containerName)
	rawMemoryResult, err2 := e.memoryEstimator.GetRawEstimation(s.AggregateMemoryUsage)
	return model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(rawCPUResult),
		model.ResourceMemory: model.MemoryAmountFromBytes(rawMemoryResult),
	}, errors.Join(err1, err2)
}

type autopilotSafetyMarginEstimator struct {
	cpuHistogramMaxValue    model.ResourceAmount
	memoryHistogramMaxValue model.ResourceAmount
	baseEstimator           AutopilotResourceEstimator
}

func WithAutopilotSafetyMargin(cpuHistogramMaxValue, memoryHistogramMaxValue float64, baseEstimator AutopilotResourceEstimator) AutopilotResourceEstimator {
	return &autopilotSafetyMarginEstimator{
		cpuHistogramMaxValue:    model.CPUAmountFromCores(cpuHistogramMaxValue),
		memoryHistogramMaxValue: model.MemoryAmountFromBytes(memoryHistogramMaxValue),
		baseEstimator:           baseEstimator,
	}
}

func calMarginedValue(curValue, histogramMaxValue model.ResourceAmount) model.ResourceAmount {
	proportion := float64(curValue) / float64(histogramMaxValue)
	// The more is the resource, the lower is the margin
	scale := proportion*minSafetyMargin + (1-proportion)*maxSafetyMargin + 1.0
	return model.ScaleResource(curValue, scale)
}

func (e *autopilotSafetyMarginEstimator) GetResourceEstimation(containerName string, s *model.AggregateContainerState) (model.Resources, error) {
	klog.V(5).Infof("Algorithm Enter SafetyMargin" + containerName)
	originalResources, err := e.baseEstimator.GetResourceEstimation(containerName, s)
	// klog.V(4).Infof("NICONICO Resource input in Margin: %+v", originalResources)
	return model.Resources{
		model.ResourceCPU:    calMarginedValue(originalResources[model.ResourceCPU], e.cpuHistogramMaxValue),
		model.ResourceMemory: calMarginedValue(originalResources[model.ResourceMemory], e.memoryHistogramMaxValue),
	}, err
}

type bufferBody struct {
	bufferCPU    []model.ResourceAmount
	bufferMemory []model.ResourceAmount
	bufPtr       int64
}

type autopilotFluctuationReducer struct {
	bufSize       int64
	buffer        map[string]bufferBody
	baseEstimator AutopilotResourceEstimator
}

func WithAutopilotFluctuationReducer(fluctuationReducerDuration, recommenderInterval time.Duration, baseEstimator AutopilotResourceEstimator) AutopilotResourceEstimator {
	bufSize := fluctuationReducerDuration.Nanoseconds() / recommenderInterval.Nanoseconds()
	return &autopilotFluctuationReducer{
		bufSize: bufSize,
		buffer:  make(map[string]bufferBody),
		// bufferMemory:  make([]model.ResourceAmount, bufSize),
		// bufPtr:        0,
		baseEstimator: baseEstimator,
	}
}

func (e *autopilotFluctuationReducer) GetResourceEstimation(containerName string, s *model.AggregateContainerState) (model.Resources, error) {
	klog.V(5).Infof("Algorithm Enter FluctionReducer" + containerName)
	originalResources, err := e.baseEstimator.GetResourceEstimation(containerName, s)
	klog.V(4).Infof("NICONICO Resource input in Fluctuation: %+v", originalResources)
	klog.V(4).Infof("NICONICO Resource status in Fluctuation: %+v", e.buffer)
	// Only data without error is authorized into the buffer
	if err == nil {
		b, exist := e.buffer[containerName]

		if !exist {
			e.buffer[containerName] = bufferBody{
				bufferCPU:    make([]model.ResourceAmount, e.bufSize),
				bufferMemory: make([]model.ResourceAmount, e.bufSize),
				bufPtr:       0,
			}
			b = e.buffer[containerName]
		}
		b.bufferCPU[b.bufPtr] = originalResources[model.ResourceCPU]
		b.bufferMemory[b.bufPtr] = originalResources[model.ResourceMemory]
		b.bufPtr = (b.bufPtr + 1) % e.bufSize
	} else {
		klog.V(3).Infof("Error input fluctuation originalResources: %s", err.Error())
	}

	maxCPU := model.ResourceAmount(0)
	maxMem := model.ResourceAmount(0)
	if b, exist := e.buffer[containerName]; exist {
		for i := 0; i < int(e.bufSize); i++ {
			maxCPU = model.ResourceAmountMax(maxCPU, b.bufferCPU[i])
			maxMem = model.ResourceAmountMax(maxMem, b.bufferMemory[i])
		}
	}

	var err0 error = nil
	if maxCPU == model.ResourceAmount(0) || maxMem == model.ResourceAmount(0) {
		err0 = errors.New("No Available CPU or Memory data in past 1h buffer. This may because of coldstart, need to wait for enough samples...")
	}

	klog.V(4).Infof("NICONICO Resource output Fluctuation: %v %v", maxCPU, maxMem)
	return model.Resources{
		model.ResourceCPU:    maxCPU,
		model.ResourceMemory: maxMem,
	}, err0
}

type AutopilotSingleEstimator interface {
	GetRawEstimation(h util.AutopilotHisto) (float64, error)
}

type autopilotMaxEstimator struct {
	N int
}

func NewAutopilotMaxEstimator(N int) AutopilotSingleEstimator {
	return &autopilotMaxEstimator{
		N: N,
	}
}

func (e *autopilotMaxEstimator) GetRawEstimation(h util.AutopilotHisto) (float64, error) {
	var err error = nil
	if h.AggregateNums() <= e.N || !h.HasValidAggregation() {
		err = errors.New("Max: No enough valid aggregations, estimation could be small")
	}
	return h.Max(), err
}

type autopilotAverageEstimator struct {
	N int
}

func NewAutopilotAverageEstimator(N int) AutopilotSingleEstimator {
	return &autopilotAverageEstimator{
		N: N,
	}
}

func (e *autopilotAverageEstimator) GetRawEstimation(h util.AutopilotHisto) (float64, error) {
	var err error = nil
	if h.AggregateNums() <= e.N || !h.HasValidAggregation() {
		klog.V(3).Infof("Average Err: %v, %v, %v", h.AggregateNums(), e.N, h.HasValidAggregation())
		klog.V(3).Infof(h.String())
		err = errors.New("Average: No enough valid aggregations, estimation could be small")
	}
	return h.Average(), err
}

type autopilotPercentileEstimator struct {
	percentile float64
	N          int
}

func NewAutopilotPercentileEstimator(percentileInt int, N int) AutopilotSingleEstimator {
	return &autopilotPercentileEstimator{
		percentile: float64(percentileInt) / 100.0,
		N:          N,
	}
}

func (e *autopilotPercentileEstimator) GetRawEstimation(h util.AutopilotHisto) (float64, error) {
	var err error = nil
	if h.AggregateNums() <= e.N || !h.HasValidAggregation() {
		klog.V(3).Infof("Percentile Err: %v, %v, %v", h.AggregateNums(), e.N, h.HasValidAggregation())
		klog.V(3).Infof(h.String())
		err = errors.New("Percentile: No enough valid aggregations, estimation could be small")
	}
	return h.Percentile(e.percentile), err
}

type autopilotSpikeEstimator struct {
	basePercentileEstimator AutopilotSingleEstimator
	baseMaxEstimator        AutopilotSingleEstimator
	N                       int
}

func NewAutopilotSpikeEstimator(percentileInt int, N int) AutopilotSingleEstimator {
	return &autopilotSpikeEstimator{
		basePercentileEstimator: NewAutopilotPercentileEstimator(percentileInt, N),
		baseMaxEstimator:        NewAutopilotMaxEstimator(N),
	}
}

func (e *autopilotSpikeEstimator) GetRawEstimation(h util.AutopilotHisto) (float64, error) {
	percentileResult, err1 := e.basePercentileEstimator.GetRawEstimation(h)
	maxResult, err2 := e.baseMaxEstimator.GetRawEstimation(h)
	return math.Max(percentileResult, 0.5*maxResult), errors.Join(err1, err2)
}
