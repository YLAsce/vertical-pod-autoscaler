package model

import (
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/util"
	"k8s.io/klog/v2"
)

func NewCPURecommender(backHisto util.AutopilotHisto) *Recommender {
	numModels := (*numDmCPU) * (*numMmCPU)
	ret := Recommender{
		modelPool:       make([]*Model, numModels),
		numModels:       numModels,
		wdm:             *wdmCPU,
		wdl:             *wdlCPU,
		selectedModelId: -1,
		recommendation:  ResourceAmount(-1),
	}
	for i := 0; i <= *numDmCPU-1; i++ {
		for j := 0; j <= *numMmCPU-1; j++ {
			ret.modelPool[i*(*numDmCPU)+j] = NewCPUModel(backHisto, float64(i)/float64(*numDmCPU-1), float64(j)*(*maxMmCPU)/float64(*numMmCPU-1))
		}
	}
	return &ret
}

func NewMemoryRecommender(backHisto util.AutopilotHisto) *Recommender {
	numModels := (*numDmMemory) * (*numMmMemory)
	ret := Recommender{
		modelPool:       make([]*Model, numModels),
		numModels:       numModels,
		wdm:             *wdmMemory,
		wdl:             *wdlMemory,
		selectedModelId: -1,
		recommendation:  ResourceAmount(-1),
	}
	for i := 0; i <= *numDmMemory-1; i++ {
		for j := 0; j <= *numMmMemory-1; j++ {
			ret.modelPool[i*(*numDmMemory)+j] = NewMemoryModel(backHisto, float64(i)/float64(*numDmMemory-1), float64(j)*(*maxMmMemory)/float64(*numMmMemory-1))
		}
	}
	return &ret
}

type Recommender struct {
	modelPool       []*Model
	numModels       int
	wdm             float64
	wdl             float64
	selectedModelId int

	recommendation ResourceAmount
}

func deltaResource(x, y ResourceAmount) int {
	if x != y {
		return 1
	}
	return 0
}

func (r *Recommender) CalculateOnce() {
<<<<<<< HEAD
	for _, m := range r.modelPool {
		// klog.V(4).Infof("NICO ML Model %v before r: dm %v, mm %v, lmId %v, lm %v, cm %v", i, m.dm, m.mm, m.lmId, m.lm, m.cm)
		m.CalculateOnce()
		// klog.V(4).Infof("NICO ML Model %v finished: dm %v, mm %v, lmId %v, lm %v, cm %v", i, m.dm, m.mm, m.lmId, m.lm, m.cm)
=======
	for i, m := range r.modelPool {
		klog.V(4).Infof("NICO ML Model %v before r: dm %v, mm %v, lmId %v, lm %v, cm %v", i, m.dm, m.mm, m.lmId, m.lm, m.cm)
		m.CalculateOnce()
		klog.V(4).Infof("NICO ML Model %v finished: dm %v, mm %v, lmId %v, lm %v, cm %v", i, m.dm, m.mm, m.lmId, m.lm, m.cm)
>>>>>>> 83b4a7b2995f5cda7ade0e061d546bffbdfb3724
	}

	minVal := r.modelPool[0].GetCurCm() +
		r.wdm*float64(delta(r.selectedModelId, 0)) +
		r.wdl*float64(deltaResource(r.recommendation, r.modelPool[0].GetCurLm()))
	minId := 0
	for id, m := range r.modelPool {
		curVal := m.GetCurCm() +
			r.wdm*float64(delta(r.selectedModelId, id)) +
			r.wdl*float64(deltaResource(r.recommendation, m.GetCurLm()))
		if curVal < minVal {
			minVal = curVal
			minId = id
		}
	}
	r.selectedModelId = minId
	r.recommendation = r.modelPool[r.selectedModelId].GetCurLm()

	klog.V(4).Infof("NICO ML Recommender finished: Use Model %v", r.selectedModelId)
}

func (r *Recommender) GetRecommendation() ResourceAmount {
	return r.recommendation
}

func (r *Recommender) Merge(other *Recommender) {
	if r.recommendation != ResourceAmount(-1) {
		panic("ML Recommender can't merge two ongoing instances. Merge will be discarded later")
	}
	r.modelPool = other.modelPool
	r.numModels = other.numModels
	r.wdm = other.wdm
	r.wdl = other.wdl
	r.selectedModelId = other.selectedModelId
	r.recommendation = other.recommendation
}
