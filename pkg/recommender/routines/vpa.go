/*
Copyright 2018 The Kubernetes Authors.

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

package routines

import (
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	api_utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// GetContainerNameToAggregateStateMap returns ContainerNameToAggregateStateMap for pods.
func GetContainerNameToAggregateStateMapAndOOMStatus(vpa *model.Vpa) (model.ContainerNameToAggregateStateMap, bool) {
	containerNameToAggregateStateMap := vpa.AggregateStateByContainerName()
	filteredContainerNameToAggregateStateMap := make(model.ContainerNameToAggregateStateMap)

	hasOOM := false
	for containerName, aggregatedContainerState := range containerNameToAggregateStateMap {
		// klog.V(4).Infof("[NICO]aggregatedContainerName: %+v:", containerName)
		// klog.V(4).Infof("[NICO]aggregatedContainerCPUHistogram: %+v:", aggregatedContainerState.AggregateCPUUsage.String())
		containerResourcePolicy := api_utils.GetContainerResourcePolicy(containerName, vpa.ResourcePolicy)
		// klog.V(4).Infof("[NICO]aggregatedContainerResourcePolicy: %+v:", containerResourcePolicy)
		autoscalingDisabled := containerResourcePolicy != nil && containerResourcePolicy.Mode != nil &&
			*containerResourcePolicy.Mode == vpa_types.ContainerScalingModeOff
		if !autoscalingDisabled {
			aggregatedContainerState.UpdateFromPolicy(containerResourcePolicy)
			filteredContainerNameToAggregateStateMap[containerName] = aggregatedContainerState
			if aggregatedContainerState.OOMAmountToDo > 0 {
				hasOOM = true
			}
		}
	}
	return filteredContainerNameToAggregateStateMap, hasOOM
}
