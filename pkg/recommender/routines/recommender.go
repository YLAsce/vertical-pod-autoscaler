/*
Copyright 2017 The Kubernetes Authors.

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
	"context"
	"flag"
	"time"

	"k8s.io/klog/v2"

	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/checkpoint"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	metrics_recommender "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/recommender"
	vpa_utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

var (
	checkpointsWriteTimeout = flag.Duration("checkpoints-timeout", time.Minute, `Timeout for writing checkpoints since the start of the recommender's main loop`)
	minCheckpointsPerRun    = flag.Int("min-checkpoints", 10, "Minimum number of checkpoints to write per recommender's main loop")
	idlePercentage          = flag.Float64("idle-percentage", 0.1, "Usage under this percentage will be raised up to previous usage")
)

// Recommender recommend resources for certain containers, based on utilization periodically got from metrics api.
type Recommender interface {
	// RunOnce performs one iteration of recommender duties followed by update of recommendations in VPA objects.
	RunOnce()
	// New API
	CollectOnce()
	SetupOnce()
	RecommendOnce(algorithmRun bool)
	// GetClusterState returns ClusterState used by Recommender
	GetClusterState() *model.ClusterState
	// GetClusterStateFeeder returns ClusterStateFeeder used by Recommender
	GetClusterStateFeeder() input.ClusterStateFeeder
	// UpdateVPAs computes recommendations and sends VPAs status updates to API Server
	UpdateVPAs(algorithmRun bool)
	// MaintainCheckpoints stores current checkpoints in API Server and garbage collect old ones
	// MaintainCheckpoints writes at least minCheckpoints if there are more checkpoints to write.
	// Checkpoints are written until ctx permits or all checkpoints are written.
	MaintainCheckpoints(ctx context.Context, minCheckpoints int)
}

type recommender struct {
	clusterState                  *model.ClusterState
	clusterStateFeeder            input.ClusterStateFeeder
	checkpointWriter              checkpoint.CheckpointWriter
	checkpointsGCInterval         time.Duration
	controllerFetcher             controllerfetcher.ControllerFetcher
	lastCheckpointGC              time.Time
	vpaClient                     vpa_api.VerticalPodAutoscalersGetter
	podResourceRecommender        logic.PodResourceRecommender
	useCheckpoints                bool
	lastAggregateContainerStateGC time.Time
	recommendationPostProcessor   []RecommendationPostProcessor
}

func (r *recommender) GetClusterState() *model.ClusterState {
	return r.clusterState
}

func (r *recommender) GetClusterStateFeeder() input.ClusterStateFeeder {
	return r.clusterStateFeeder
}

// Updates VPA CRD objects' statuses.
func (r *recommender) UpdateVPAs(algorithmRun bool) {
	// 防止性能问题，这个版本先不放这个
	// cnt := metrics_recommender.NewObjectCounter()
	// defer cnt.Observe()

	// 只有在OOM或者算法需要执行的时候，才走这个流程
	// if !(r.clusterState.OOMToDo || algorithmRun) {
	// 	return
	// }

	for _, observedVpa := range r.clusterState.ObservedVpas {
		key := model.VpaID{
			Namespace: observedVpa.Namespace,
			VpaName:   observedVpa.Name,
		}
		vpa, found := r.clusterState.Vpas[key]
		if !found {
			continue
		}
		containerNameToAggregateStateMap, hasOOMAmongContainers := GetContainerNameToAggregateStateMapAndOOMStatus(vpa)
		updateCache := false
		if hasOOMAmongContainers || algorithmRun {
			updateCache = true
		}

		klog.V(4).Infof("Execute Update VPA, Reason: ClusterOOMTODO? %v, VPAOOMTodo? %v, Algorithm? %v", r.clusterState.OOMToDo, hasOOMAmongContainers, algorithmRun)

		resources := r.podResourceRecommender.GetRecommendedPodResources(containerNameToAggregateStateMap, algorithmRun, updateCache)
		// 这里获得推荐结果了，是个map，map[string]RecommendedContainerResources

		had := vpa.HasRecommendation() //VPA中已经有recommendation了吗？

		// 转成list，按照container name排序的
		listOfResourceRecommendation := logic.MapToListOfRecommendedContainerResources(resources)

		//限制整数，限制上限下限（根据config）
		for _, postProcessor := range r.recommendationPostProcessor {
			listOfResourceRecommendation = postProcessor.Process(observedVpa, listOfResourceRecommendation)
		}

		//更新VPA中的recommendation，更新prometheus指标
		vpa.UpdateRecommendation(listOfResourceRecommendation)
		if vpa.HasRecommendation() && !had {
			metrics_recommender.ObserveRecommendationLatency(vpa.Created)
		}
		//更新显示的condition
		// hasMatchingPods := vpa.PodCount > 0
		// 这里强制推荐为True
		vpa.UpdateConditions(true)

		// 更新cluster state，谁是空的vpa
		if err := r.clusterState.RecordRecommendation(vpa, time.Now()); err != nil {
			klog.Warningf("%v", err)
			if klog.V(4).Enabled() {
				klog.Infof("VPA dump")
				klog.Infof("%+v", vpa)
				klog.Infof("HasMatchingPods: %v", true)
				klog.Infof("PodCount: %v", vpa.PodCount)
				pods := r.clusterState.GetMatchingPods(vpa)
				klog.Infof("MatchingPods: %+v", pods)
				if len(pods) != vpa.PodCount {
					klog.Errorf("ClusterState pod count and matching pods disagree for vpa %v/%v", vpa.ID.Namespace, vpa.ID.VpaName)
				}
			}
		}
		// cnt.Add(vpa)
		//更新status界面
		_, err := vpa_utils.UpdateVpaStatusIfNeeded(
			r.vpaClient.VerticalPodAutoscalers(vpa.ID.Namespace), vpa.ID.VpaName, vpa.AsStatus(), &observedVpa.Status)
		if err != nil {
			klog.Errorf(
				"Cannot update VPA %v/%v object. Reason: %+v", vpa.ID.Namespace, vpa.ID.VpaName, err)
		}
	}

	// Reset the state mark, because if there is OOM, this OOM event MUST be solved above
	r.clusterState.OOMToDo = false
}

func (r *recommender) MaintainCheckpoints(ctx context.Context, minCheckpointsPerRun int) {
	now := time.Now()
	if r.useCheckpoints {
		if err := r.checkpointWriter.StoreCheckpoints(ctx, now, minCheckpointsPerRun); err != nil {
			klog.Warningf("Failed to store checkpoints. Reason: %+v", err)
		}
		if time.Since(r.lastCheckpointGC) > r.checkpointsGCInterval {
			r.lastCheckpointGC = now
			r.clusterStateFeeder.GarbageCollectCheckpoints()
		}
	}
}

// 1 second
func (r *recommender) CollectOnce() {
	r.clusterStateFeeder.LoadVPAs()
	r.clusterStateFeeder.LoadPods(*idlePercentage)

	r.clusterStateFeeder.LoadRealTimeMetrics()
}

// 5 minute
func (r *recommender) SetupOnce() {
	// r.clusterStateFeeder.LoadVPAs()
}

func (r *recommender) RecommendOnce(algorithmRun bool) {
	// NICO 这个版本不提供checkpoint功能
	//自己也提供一个prometheus数据源
	// timer := metrics_recommender.NewExecutionTimer()
	// defer timer.ObserveTotal()

	// // 管理远程请求
	// ctx := context.Background()
	// //默认1分钟写checkpoint
	// ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(*checkpointsWriteTimeout))
	// defer cancelFunc()

	// Autopilot Histogram aggragate
	if algorithmRun {
		r.clusterStateFeeder.HistogramAggregate(time.Now())
		// timer.ObserveStep("HistogramAggregate")
		klog.V(3).Infof("ClusterState is tracking %v PodStates and %v VPAs", len(r.clusterState.Pods), len(r.clusterState.Vpas))
	}

	// 最终结果是container级别的recommendation
	r.UpdateVPAs(algorithmRun)
	// timer.ObserveStep("UpdateVPAs")

	if algorithmRun {
		// NICO 这个版本不提供checkpoint功能
		// r.MaintainCheckpoints(ctx, *minCheckpointsPerRun)
		// timer.ObserveStep("MaintainCheckpoints")
		r.clusterState.RateLimitedGarbageCollectAggregateCollectionStates(time.Now(), r.controllerFetcher)
		// timer.ObserveStep("GarbageCollect")
		klog.V(3).Infof("ClusterState is tracking %d aggregated container states", r.clusterState.StateMapSize())
	}
}

func (r *recommender) RunOnce() {
	//自己也提供一个prometheus数据源
	timer := metrics_recommender.NewExecutionTimer()
	defer timer.ObserveTotal()

	// 管理远程请求
	ctx := context.Background()
	//默认1分钟写checkpoint
	ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(*checkpointsWriteTimeout))
	defer cancelFunc()

	klog.V(3).Infof("Recommender Run")

	r.clusterStateFeeder.LoadVPAs()
	timer.ObserveStep("LoadVPAs")

	r.clusterStateFeeder.LoadPods(*idlePercentage)
	timer.ObserveStep("LoadPods")

	// 把实时的container级别数据传入到feeder.clusterState
	r.clusterStateFeeder.LoadRealTimeMetrics()
	timer.ObserveStep("LoadMetrics")
	klog.V(3).Infof("ClusterState is tracking %v PodStates and %v VPAs", len(r.clusterState.Pods), len(r.clusterState.Vpas))

	// 最终结果是container级别的recommendation. 这个true是后加的参数
	r.UpdateVPAs(true)
	timer.ObserveStep("UpdateVPAs")

	r.MaintainCheckpoints(ctx, *minCheckpointsPerRun)
	timer.ObserveStep("MaintainCheckpoints")

	r.clusterState.RateLimitedGarbageCollectAggregateCollectionStates(time.Now(), r.controllerFetcher)
	timer.ObserveStep("GarbageCollect")
	klog.V(3).Infof("ClusterState is tracking %d aggregated container states", r.clusterState.StateMapSize())
}

// RecommenderFactory makes instances of Recommender.
type RecommenderFactory struct {
	ClusterState *model.ClusterState

	ClusterStateFeeder     input.ClusterStateFeeder
	ControllerFetcher      controllerfetcher.ControllerFetcher
	CheckpointWriter       checkpoint.CheckpointWriter
	PodResourceRecommender logic.PodResourceRecommender
	VpaClient              vpa_api.VerticalPodAutoscalersGetter

	RecommendationPostProcessors []RecommendationPostProcessor

	CheckpointsGCInterval time.Duration
	UseCheckpoints        bool
}

// Make creates a new recommender instance,
// which can be run in order to provide continuous resource recommendations for containers.
func (c RecommenderFactory) Make() Recommender {
	recommender := &recommender{
		clusterState:                  c.ClusterState,
		clusterStateFeeder:            c.ClusterStateFeeder,
		checkpointWriter:              c.CheckpointWriter,
		checkpointsGCInterval:         c.CheckpointsGCInterval,
		controllerFetcher:             c.ControllerFetcher,
		useCheckpoints:                c.UseCheckpoints,
		vpaClient:                     c.VpaClient,
		podResourceRecommender:        c.PodResourceRecommender,
		recommendationPostProcessor:   c.RecommendationPostProcessors,
		lastAggregateContainerStateGC: time.Now(),
		lastCheckpointGC:              time.Now(),
	}
	klog.V(3).Infof("New Recommender created %+v", recommender)
	return recommender
}
