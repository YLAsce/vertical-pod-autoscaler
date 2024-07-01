package model

import (
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/util"
)

type CombinedOLULLmPrime struct {
	OL        []float64
	UL        []float64
	LmPrimeId int
}

func NewCombinedOLULLmPrime(arrSize int) *CombinedOLULLmPrime {
	return &CombinedOLULLmPrime{
		OL:        make([]float64, arrSize),
		UL:        make([]float64, arrSize),
		LmPrimeId: -1,
	}
}

func NewCPURecommender(backHisto util.AutopilotHisto) *Recommender {
	numModels := (*numDmCPU) * (*numMmCPU)
	ret := Recommender{
		modelPool:               make([]*Model, numModels),
		combinedOLULLmPrimePool: make([]*CombinedOLULLmPrime, *numDmCPU),
		numModels:               numModels,
		numDm:                   *numDmCPU,
		wo:                      *woCPU,
		wu:                      *wuCPU,
		wdm:                     *wdmCPU,
		wdl:                     *wdlCPU,
		selectedModelId:         -1,
		backHisto:               backHisto,
		recommendation:          ResourceAmount(-1),
		recommendationId:        -1,
	}
	for i := 0; i <= *numDmCPU-1; i++ {
		ret.combinedOLULLmPrimePool[i] = NewCombinedOLULLmPrime(backHisto.GetMaxIdL() + 1)
	}

	for i := 0; i <= *numDmCPU-1; i++ {
		for j := 0; j <= *numMmCPU-1; j++ {
			ret.modelPool[i*(*numMmCPU)+j] = NewCPUModel(backHisto, ret.combinedOLULLmPrimePool[i], j*(*sizeMmBucketsCPU))
		}
	}
	return &ret
}

func NewMemoryRecommender(backHisto util.AutopilotHisto) *Recommender {
	numModels := (*numDmMemory) * (*numMmMemory)
	ret := Recommender{
		modelPool:               make([]*Model, numModels),
		combinedOLULLmPrimePool: make([]*CombinedOLULLmPrime, *numDmMemory),
		numModels:               numModels,
		numDm:                   *numDmMemory,
		wo:                      *woMemory,
		wu:                      *wuMemory,
		wdm:                     *wdmMemory,
		wdl:                     *wdlMemory,
		selectedModelId:         -1,
		backHisto:               backHisto,
		recommendation:          ResourceAmount(-1),
		recommendationId:        -1,
	}
	for i := 0; i <= *numDmMemory-1; i++ {
		ret.combinedOLULLmPrimePool[i] = NewCombinedOLULLmPrime(backHisto.GetMaxIdL() + 1)
	}

	for i := 0; i <= *numDmMemory-1; i++ {
		for j := 0; j <= *numMmMemory-1; j++ {
			ret.modelPool[i*(*numMmMemory)+j] = NewMemoryModel(backHisto, ret.combinedOLULLmPrimePool[i], j*(*sizeMmBucketsMemory))
		}
	}
	return &ret
}

func NewGPUSMRecommender(backHisto util.AutopilotHisto) *Recommender {
	numModels := (*numDmGPUS) * (*numMmGPUS)
	ret := Recommender{
		modelPool:               make([]*Model, numModels),
		combinedOLULLmPrimePool: make([]*CombinedOLULLmPrime, *numDmGPUS),
		numModels:               numModels,
		numDm:                   *numDmGPUM,
		wo:                      *woGPUS,
		wu:                      *wuGPUS,
		wdm:                     *wdmGPUS,
		wdl:                     *wdlGPUS,
		selectedModelId:         -1,
		backHisto:               backHisto,
		recommendation:          ResourceAmount(-1),
		recommendationId:        -1,
	}
	for i := 0; i <= *numDmGPUS-1; i++ {
		ret.combinedOLULLmPrimePool[i] = NewCombinedOLULLmPrime(backHisto.GetMaxIdL() + 1)
	}

	for i := 0; i <= *numDmGPUS-1; i++ {
		for j := 0; j <= *numMmGPUS-1; j++ {
			ret.modelPool[i*(*numMmGPUS)+j] = NewGPUSModel(backHisto, ret.combinedOLULLmPrimePool[i], j)
		}
	}
	return &ret
}

func NewGPUMemoryRecommender(backHisto util.AutopilotHisto) *Recommender {
	numModels := (*numDmGPUM) * (*numMmGPUM)
	ret := Recommender{
		modelPool:               make([]*Model, numModels),
		combinedOLULLmPrimePool: make([]*CombinedOLULLmPrime, *numDmGPUM),
		numModels:               numModels,
		numDm:                   *numDmGPUM,
		wo:                      *woGPUM,
		wu:                      *wuGPUM,
		wdm:                     *wdmGPUM,
		wdl:                     *wdlGPUM,
		selectedModelId:         -1,
		backHisto:               backHisto,
		recommendation:          ResourceAmount(-1),
		recommendationId:        -1,
	}
	for i := 0; i <= *numDmGPUM-1; i++ {
		ret.combinedOLULLmPrimePool[i] = NewCombinedOLULLmPrime(backHisto.GetMaxIdL() + 1)
	}

	for i := 0; i <= *numDmGPUM-1; i++ {
		for j := 0; j <= *numMmGPUM-1; j++ {
			ret.modelPool[i*(*numMmGPUM)+j] = NewGPUMModel(backHisto, ret.combinedOLULLmPrimePool[i], j)
		}
	}
	return &ret
}

type Recommender struct {
	modelPool               []*Model
	combinedOLULLmPrimePool []*CombinedOLULLmPrime
	numModels               int
	numDm                   int
	wo                      float64
	wu                      float64
	wdm                     float64
	wdl                     float64
	selectedModelId         int
	backHisto               util.AutopilotHisto

	recommendation   ResourceAmount
	recommendationId int
}

// 总体时间复杂度 O(dM * max(bucketNum, mM))
// 总体空间复杂度 O(dM * max(bucketNum, mM))
func (r *Recommender) CalculateOnce() {
	//复杂度 O(dM*bucketNum)
	//在外部更新所有pool中的mL和uL, 更新Dm*Mm次优化到更新Dm次
	for i := 0; i <= r.numDm-1; i++ {
		dm := float64(i) / float64(r.numDm-1)
		for idL := 0; idL <= r.backHisto.GetMaxIdL(); idL++ {
			r.combinedOLULLmPrimePool[i].OL[idL] = (1.0-dm)*r.combinedOLULLmPrimePool[i].OL[idL] + dm*float64(r.backHisto.NumSamplesWithValueMoreThan(idL))
			r.combinedOLULLmPrimePool[i].UL[idL] = (1.0-dm)*r.combinedOLULLmPrimePool[i].UL[idL] + dm*float64(r.backHisto.NumSamplesWithValueLessThan(idL))
		}

		prevLmPrimeId := r.combinedOLULLmPrimePool[i].LmPrimeId
		minSubL := r.wo*r.combinedOLULLmPrimePool[i].OL[0] + r.wu*r.combinedOLULLmPrimePool[i].UL[0] + r.wdl*float64(delta(0, prevLmPrimeId))
		r.combinedOLULLmPrimePool[i].LmPrimeId = 0
		for j := 1; j <= r.backHisto.GetMaxIdL(); j++ {
			curMin := r.wo*r.combinedOLULLmPrimePool[i].OL[j] + r.wu*r.combinedOLULLmPrimePool[i].UL[j] + r.wdl*float64(delta(j, prevLmPrimeId))
			if curMin < minSubL {
				minSubL = curMin
				r.combinedOLULLmPrimePool[i].LmPrimeId = i
			}
		}
	}

	//复杂度 O(numModels) = O(dM * mM)
	for _, m := range r.modelPool {
		m.CalculateOnce()
		// klog.V(4).Infof("NICO ML Model %v finished: dm %v, mm %v, lmId %v, lm %v, cm %v", i, m.dm, m.mm, m.lmId, m.lm, m.cm)
	}

	minVal := r.modelPool[0].GetCurCm() +
		r.wdm*float64(delta(r.selectedModelId, 0)) +
		r.wdl*float64(delta(r.recommendationId, r.modelPool[0].GetCurLmId()))
	minModelId := 0
	//复杂度 O(dM * mM)
	for id, m := range r.modelPool {
		curVal := m.GetCurCm() +
			r.wdm*float64(delta(r.selectedModelId, id)) +
			r.wdl*float64(delta(r.recommendationId, m.GetCurLmId()))
		if curVal < minVal {
			minVal = curVal
			minModelId = id
		}
	}
	r.selectedModelId = minModelId

	r.recommendation = r.modelPool[r.selectedModelId].GetCurLm()
	r.recommendationId = r.modelPool[r.selectedModelId].GetCurLmId()

	// klog.V(4).Infof("NICO ML Recommender finished: Use Model %v", r.selectedModelId)
}

func (r *Recommender) GetRecommendation() ResourceAmount {
	return r.recommendation
}

func (r *Recommender) Merge(other *Recommender) {
	if r.recommendation != ResourceAmount(-1) {
		panic("ML Recommender can't merge two ongoing instances. Merge will be discarded later")
	}
	r.modelPool = other.modelPool
	r.combinedOLULLmPrimePool = other.combinedOLULLmPrimePool
	r.numModels = other.numModels
	r.wo = other.wo
	r.wu = other.wu
	r.wdm = other.wdm
	r.wdl = other.wdl
	r.selectedModelId = other.selectedModelId
	r.backHisto = other.backHisto

	r.recommendation = other.recommendation
	r.recommendationId = other.recommendationId
}
