package model

import (
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/util"
)

func NewCPUModel(backHisto util.AutopilotHisto, combinedOLULLmPrime *CombinedOLULLmPrime, mmIdNums int) *Model {
	return &Model{
		usageHistogram:      backHisto,
		combinedOLULLmPrime: combinedOLULLmPrime,

		mmIdNums: mmIdNums,

		wo:  *woCPU,
		wu:  *wuCPU,
		wdl: *wdlCPU,
		d:   *dCPU,

		lmId: -1,
		lm:   0.0,
		cm:   0.0,

		resourceAmountFunction: CPUAmountFromCores,
	}
}

func NewMemoryModel(backHisto util.AutopilotHisto, combinedOLULLmPrime *CombinedOLULLmPrime, mmIdNums int) *Model {
	return &Model{
		usageHistogram:      backHisto,
		combinedOLULLmPrime: combinedOLULLmPrime,

		mmIdNums: mmIdNums,

		wo:  *woMemory,
		wu:  *wuMemory,
		wdl: *wdlMemory,
		d:   *dMemory,

		lmId: -1,
		lm:   0.0,
		cm:   0.0,

		resourceAmountFunction: MemoryAmountFromBytes,
	}
}

func NewGPUSModel(backHisto util.AutopilotHisto, combinedOLULLmPrime *CombinedOLULLmPrime, mmIdNums int) *Model {
	return &Model{
		usageHistogram:      backHisto,
		combinedOLULLmPrime: combinedOLULLmPrime,

		mmIdNums: mmIdNums,

		wo:  *woGPUS,
		wu:  *wuGPUS,
		wdl: *wdlGPUS,
		d:   *dGPUS,

		lmId: -1,
		lm:   0.0,
		cm:   0.0,

		resourceAmountFunction: GPUSMAmountFromGIs,
	}
}

func NewGPUMModel(backHisto util.AutopilotHisto, combinedOLULLmPrime *CombinedOLULLmPrime, mmIdNums int) *Model {
	return &Model{
		usageHistogram:      backHisto,
		combinedOLULLmPrime: combinedOLULLmPrime,

		mmIdNums: mmIdNums,

		wo:  *woGPUM,
		wu:  *wuGPUM,
		wdl: *wdlGPUM,
		d:   *dGPUM,

		lmId: -1,
		lm:   0.0,
		cm:   0.0,

		resourceAmountFunction: GPUMemoryAmountFromBytes,
	}
}

type Model struct {
	usageHistogram      util.AutopilotHisto
	combinedOLULLmPrime *CombinedOLULLmPrime //oL uL minSubLid改为外部传入结构体指针，节省内存,避免重复计算，CPU和内存都优化了mM倍

	// dm 在外边计算了
	mmIdNums int // Mm的大小相当于多少个bucket的大小

	wo  float64
	wu  float64
	wdl float64
	d   float64

	lmId int
	lm   float64
	cm   float64

	resourceAmountFunction func(float64) ResourceAmount
}

func delta(x, y int) int {
	if x != y {
		return 1
	}
	return 0
}

func minint(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func (m *Model) CalculateOnce() {
	// 原复杂度 O(NumBuckets),这里优化到了外面

	// Only the ID is useful here.
	// Compare the LID is ok, Because in the same model the limits are set in the same configuration set.
	prevlmId := m.lmId
	m.lmId = minint(m.combinedOLULLmPrime.LmPrimeId+m.mmIdNums, m.usageHistogram.GetMaxIdL())
	m.lm = m.usageHistogram.GetLValWithId(m.lmId)
	// 复杂度 O(NumBuckets) -> 优化为O(1)
	m.cm = m.d*(m.wo*float64(m.usageHistogram.NumSamplesWithValueMoreThan(m.lmId))+
		m.wu*float64(m.usageHistogram.NumSamplesWithValueLessThan(m.lmId))+
		m.wdl*float64(delta(prevlmId, m.lmId))) + (1.0-m.d)*m.cm
}

func (m *Model) GetCurLm() ResourceAmount {
	return m.resourceAmountFunction(m.lm)
}

func (m *Model) GetCurLmId() int {
	return m.lmId
}

func (m *Model) GetCurCm() float64 {
	return m.cm
}
