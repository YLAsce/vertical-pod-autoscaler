package main

import (
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

type ConstRecommender struct {
	constResources logic.RecommendedPodResources
}

func CreateConstRecommender(traceInfoMap map[string]*TraceInfo) *ConstRecommender {
	constResources := make(logic.RecommendedPodResources)
	for k, v := range traceInfoMap {
		mv := v.Max()
		constResources[k] = logic.RecommendedContainerResources{
			Target:     mv,
			LowerBound: mv,
			UpperBound: mv,
		}
	}
	return &ConstRecommender{
		constResources: constResources,
	}
}

func (c *ConstRecommender) GetRecommendedPodResources(nameToAggregate model.ContainerNameToAggregateGPUMap) logic.RecommendedPodResources {
	ret := make(logic.RecommendedPodResources)
	for k, _ := range nameToAggregate {
		ret[k] = c.constResources[k]
	}
	return ret
}
