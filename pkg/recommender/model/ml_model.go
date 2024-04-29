package model

import (
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/util"
)

func NewCPUModel(backHisto util.AutopilotHisto, dm, mm float64) *Model {
	return &Model{
		dm: dm,
		mm: mm,
		oL: make([]float64, backHisto.GetMaxIdL()+1),
		uL: make([]float64, backHisto.GetMaxIdL()+1),

		wo:  *woCPU,
		wu:  *wuCPU,
		wdl: *wdlCPU,
		d:   *dCPU,

		usageHistogram: backHisto,

		lmId: 0,
		lm:   0.0,
		cm:   0.0,

		resourceAmountFunction: CPUAmountFromCores,
	}
}

func NewMemoryModel(backHisto util.AutopilotHisto, dm, mm float64) *Model {
	return &Model{
		dm: dm,
		mm: mm,
		oL: make([]float64, backHisto.GetMaxIdL()+1),
		uL: make([]float64, backHisto.GetMaxIdL()+1),

		wo:  *woMemory,
		wu:  *wuMemory,
		wdl: *wdlMemory,
		d:   *dMemory,

		usageHistogram: backHisto,

		lmId: 0,
		lm:   0.0,
		cm:   0.0,

		resourceAmountFunction: MemoryAmountFromBytes,
	}
}

type Model struct {
	dm float64
	mm float64
	oL []float64
	uL []float64

	wo  float64
	wu  float64
	wdl float64
	d   float64

	usageHistogram util.AutopilotHisto

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

func (m *Model) CalculateOnce() {
	for lId := 0; lId <= m.usageHistogram.GetMaxIdL(); lId++ {
		m.oL[lId] = (1.0-m.dm)*m.oL[lId] + m.dm*float64(m.usageHistogram.NumSamplesWithValueMoreThan(lId))
		m.uL[lId] = (1.0-m.dm)*m.uL[lId] + m.dm*float64(m.usageHistogram.NumSamplesWithValueLessThan(lId))
	}

	minSubL := m.wo*m.oL[0] + m.wu*m.uL[0] + m.wdl*float64(delta(0, m.lmId))
	minSubLId := 0
	for i := 1; i <= m.usageHistogram.GetMaxIdL(); i++ {
		curMin := m.wo*m.oL[i] + m.wu*m.uL[i] + m.wdl*float64(delta(i, m.lmId))
		if curMin < minSubL {
			minSubL = curMin
			minSubLId = i
		}
	}
	// Only the ID is useful here.
	// Compare the LID is ok, Because in the same model the limits are set in the same configuration set.
	m.lmId = minSubLId
	m.lm = m.usageHistogram.GetLValWithId(m.lmId) + m.mm

	m.cm = m.d*(m.wo*float64(m.usageHistogram.NumSamplesWithValueMoreThanValue(m.lm))+
		m.wu*float64(m.usageHistogram.NumSamplesWithValueLessThanValue(m.lm))+
		m.wdl*float64(delta(minSubLId, m.lmId))) + (1.0-m.d)*m.cm
}

func (m *Model) GetCurLm() ResourceAmount {
	return m.resourceAmountFunction(m.lm)
}

func (m *Model) GetCurCm() float64 {
	return m.cm
}
