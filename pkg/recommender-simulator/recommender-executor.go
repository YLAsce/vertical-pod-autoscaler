package main

import (
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

type recommenderExecutor struct {
	podResourceRecommender  logic.PodResourceRecommender
	aggregateContainerState *model.AggregateContainerState
}

func NewMemoryRecommender
