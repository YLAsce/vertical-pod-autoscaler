package main

import (
	"flag"
	"time"
)

var isConst = flag.Bool("is-const", false, "If constant request values")

func main() {
	// ML 模型的参数也会通过这里传入
	flag.Parse()

	initialTime := time.Now()
	traceInfoMap := GetTraceInfoMap(initialTime)
	maxTraceTime := GetMaxTraceTime(traceInfoMap, initialTime)
	DumpTraceInfoSummary(traceInfoMap, initialTime)
	clusterState := NewClusterState(traceInfoMap)
	for t := initialTime; !t.After(maxTraceTime); t = t.Add(time.Second) {
		hasOom := clusterState.Record(t, *isConst)
		clusterState.CollectMetrics(t)
		if t.Sub(initialTime)%(5*time.Minute) == 0 {
			clusterState.Recommend(t, true, *isConst)
		} else if hasOom {
			clusterState.Recommend(t, false, *isConst)
		}
	}

	clusterState.DumpMetrics()
}
