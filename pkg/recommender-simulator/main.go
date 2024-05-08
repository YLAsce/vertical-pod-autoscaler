package main

import (
	"fmt"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/util"
)

var startTime = time.Unix(1234567890, 0)

func main() {
	options, _ := util.NewLinearHistogramOptions(1.0, 10, 1e-15)
	ah := util.NewAutopilotHisto(options, time.Minute, 1, time.Minute, util.AutopilotAddSampleModeDistribution)
	ah.AddSample(0.5)
	ah.Aggregate(startTime)
	fmt.Println(ah.Max())
}
