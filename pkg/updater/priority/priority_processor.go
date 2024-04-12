/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package priority

import (
	"math"

	apiv1 "k8s.io/api/core/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	"k8s.io/klog/v2"
)

// PriorityProcessor calculates priority for pod updates.
type PriorityProcessor interface {
	GetUpdatePriority(pod *apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler,
		recommendation *vpa_types.RecommendedPodResources) PodPriority
}

// NewProcessor creates a new default PriorityProcessor.
func NewProcessor() PriorityProcessor {
	return &defaultPriorityProcessor{}
}

type defaultPriorityProcessor struct {
}

// 核心逻辑在这里
func (*defaultPriorityProcessor) GetUpdatePriority(pod *apiv1.Pod, _ *vpa_types.VerticalPodAutoscaler,
	recommendation *vpa_types.RecommendedPodResources) PodPriority {
	outsideRecommendedRange := false
	scaleUp := false
	// Sum of requests over all containers, per resource type.
	totalRequestPerResource := make(map[apiv1.ResourceName]int64)
	// Sum of recommendations over all containers, per resource type.
	totalRecommendedPerResource := make(map[apiv1.ResourceName]int64)

	hasObservedContainers, vpaContainerSet := parseVpaObservedContainers(pod)

	for _, podContainer := range pod.Spec.Containers {
		if hasObservedContainers && !vpaContainerSet.Has(podContainer.Name) {
			klog.V(4).Infof("Not listed in %s:%s. Skipping container %s priority calculations",
				annotations.VpaObservedContainersLabel, pod.GetAnnotations()[annotations.VpaObservedContainersLabel], podContainer.Name)
			continue
		}
		//对每个container获取推荐
		recommendedRequest := vpa_api_util.GetRecommendationForContainer(podContainer.Name, recommendation)
		if recommendedRequest == nil {
			continue
		}
		//现在的实际资源小于Lower bound或者大于Upper bound，则更新
		for resourceName, recommended := range recommendedRequest.Target {
			totalRecommendedPerResource[resourceName] += recommended.MilliValue()
			lowerBound, hasLowerBound := recommendedRequest.LowerBound[resourceName]
			upperBound, hasUpperBound := recommendedRequest.UpperBound[resourceName]
			if request, hasRequest := podContainer.Resources.Requests[resourceName]; hasRequest {
				totalRequestPerResource[resourceName] += request.MilliValue()
				if recommended.MilliValue() > request.MilliValue() {
					scaleUp = true
				}
				if (hasLowerBound && request.Cmp(lowerBound) < 0) ||
					(hasUpperBound && request.Cmp(upperBound) > 0) {
					outsideRecommendedRange = true
				}
			} else {
				// Note: if the request is not specified, the container will use the
				// namespace default request. Currently we ignore it and treat such
				// containers as if they had 0 request. A more correct approach would
				// be to always calculate the 'effective' request.
				scaleUp = true
				outsideRecommendedRange = true
			}
		}
	}

	resourceDiff := 0.0

	for resource, totalRecommended := range totalRecommendedPerResource {
		totalRequest := math.Max(float64(totalRequestPerResource[resource]), 1.0)
		resourceDiff += math.Abs(totalRequest-float64(totalRecommended)) / totalRequest
	}
	return PodPriority{
		OutsideRecommendedRange: outsideRecommendedRange, //只要有一个container的一个当前资源请求在推荐的limit之外
		ScaleUp:                 scaleUp,                 //只要有一个container的一个当前资源的推荐target大于当前请求
		ResourceDiff:            resourceDiff,            //对每个资源算所有container的总和的diff。然后把这些diff相加
	}
}
