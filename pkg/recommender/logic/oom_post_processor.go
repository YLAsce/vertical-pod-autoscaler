package logic

import (
	"errors"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/klog/v2"
)

func NewOOMPostProcessor() *OOMPostProcessor {
	return &OOMPostProcessor{
		recordedResources: make(map[string]model.Resources),
		hasError:          make(map[string]bool),
	}
}

type OOMPostProcessor struct {
	recordedResources map[string]model.Resources
	hasError          map[string]bool
}

func (p *OOMPostProcessor) RecordBaseEstimation(containerName string, baseEstimation model.Resources, err error) {
	p.recordedResources[containerName] = baseEstimation
	if err != nil {
		p.hasError[containerName] = true
	} else {
		p.hasError[containerName] = false
	}
}

func (p *OOMPostProcessor) GetOOMPostProcessedEstimation(containerName string, s *model.AggregateContainerState) (model.Resources, error) {
	result := make(model.Resources)
	resources, ok := p.recordedResources[containerName]
	if !ok {
		return result, errors.New("OOM Post processor: Cannot find base estimation for container" + containerName)
	}

	if p.hasError[containerName] && s.OOMAmountToDo == 0 {
		return result, errors.New("When no OOM happen, do not tolerent error in base estimation. container:" + containerName)
	}

	// If there is oom, replace the corresponding resource
	for name, resource := range resources {
		if name == model.ResourceMemory && s.OOMAmountToDo > 0 && s.OOMAmountToDo > resource {
			klog.V(4).Infof("NICO Applied OOM Post Process for container %s, OLD: %v, NEW %v", containerName, resource, s.OOMAmountToDo)
			result[name] = s.OOMAmountToDo
		} else {
			result[name] = resource
		}
	}
	// Reset OOM Amount To Do
	// OOM event is only kept until its next recommendation.
	// Then when new OOM occurs, this post processing will work with new OOMAmountToDo
	s.OOMAmountToDo = 0
	return result, nil
}
