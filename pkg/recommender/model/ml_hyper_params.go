package model

import (
	"flag"
)

var (
	dCPU   = flag.Float64("ap-ml-cpu-hyperparam-d", 0.5, "Range: 0 to 1")
	woCPU  = flag.Float64("ap-ml-cpu-hyperparam-wo", 0.5, "Range: >= 0 (TODO)")
	wuCPU  = flag.Float64("ap-ml-cpu-hyperparam-wu", 0.5, "Range: >= 0 (TODO)")
	wdlCPU = flag.Float64("ap-ml-cpu-hyperparam-wdeltal", 0.5, "Range: >= 0 (TODO)")
	wdmCPU = flag.Float64("ap-ml-cpu-hyperparam-wdeltam", 0.5, "Range: >= 0 (TODO)")

	numDmCPU         = flag.Int("ap-ml-cpu-num-dm", 100, "Number of different models = num-dm * num-mm")
	numMmCPU         = flag.Int("ap-ml-cpu-num-mm", 100, "Number of different models = num-dm * num-mm")
	sizeMmBucketsCPU = flag.Int("ap-ml-cpu-size-buckets-mm", 1, "Each Mm equals to the total size of how many buckets")

	dMemory   = flag.Float64("ap-ml-memory-hyperparam-d", 0.5, "Range: 0 to 1")
	woMemory  = flag.Float64("ap-ml-memory-hyperparam-wo", 0.5, "Range: >= 0 (TODO)")
	wuMemory  = flag.Float64("ap-ml-memory-hyperparam-wu", 0.5, "Range: >= 0 (TODO)")
	wdlMemory = flag.Float64("ap-ml-memory-hyperparam-wdeltal", 0.5, "Range: >= 0 (TODO)")
	wdmMemory = flag.Float64("ap-ml-memory-hyperparam-wdeltam", 0.5, "Range: >= 0 (TODO)")

	numDmMemory         = flag.Int("ap-ml-memory-num-dm", 100, "Number of different models = num-dm * num-mm")
	numMmMemory         = flag.Int("ap-ml-memory-num-mm", 100, "Number of different models = num-dm * num-mm")
	sizeMmBucketsMemory = flag.Int("ap-ml-memory-size-buckets-mm", 1, "Each Mm equals to the total size of how many buckets")
)

func TestGetParam() int {
	return *numDmCPU
}
