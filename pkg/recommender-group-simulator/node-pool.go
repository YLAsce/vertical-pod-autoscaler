package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

var (
	nodeFolder = flag.String("nodepool-folder", "group", "node pool dump parent folder")
)

type NodePool struct {
	cpuSize model.ResourceAmount
	memSize model.ResourceAmount

	mode                int // 1 is recording, 2 is assigning. Decisions are made before every first assigning
	batchCPURequests    map[string]model.ResourceAmount
	batchMemoryRequests map[string]model.ResourceAmount

	nodeIdMap     map[string]int
	nodeUseCPU    []model.ResourceAmount
	nodeUseMemory []model.ResourceAmount

	nodeTotalUseCPU    []model.ResourceAmount
	nodeTotalUseMemory []model.ResourceAmount
	numRecords         int64
}

func NewNodePool(cpuSize, memSize model.ResourceAmount) *NodePool {

	return &NodePool{
		cpuSize: cpuSize,
		memSize: memSize,

		mode:                0,
		batchCPURequests:    make(map[string]model.ResourceAmount),
		batchMemoryRequests: make(map[string]model.ResourceAmount),

		nodeIdMap:     make(map[string]int),
		nodeUseCPU:    make([]model.ResourceAmount, 1),
		nodeUseMemory: make([]model.ResourceAmount, 1),

		nodeTotalUseCPU:    make([]model.ResourceAmount, 0),
		nodeTotalUseMemory: make([]model.ResourceAmount, 0),
		numRecords:         0,
	}
}

func (n *NodePool) batchAssignNode() {
	//input: batchCPURequests, batchMemRequests
	//output: nodeIdMap
	//现在的算法:每一秒把所有pod从小到大排序，同时daemon在前 serverless在后
	keys_serverless := make([]string, 0)
	keys_daemon := make([]string, 0)
	for k, _ := range n.batchCPURequests {
		if strings.HasPrefix(k, "daemon") {
			keys_daemon = append(keys_daemon, k)
		} else {
			keys_serverless = append(keys_serverless, k)
		}
	}

	sort.Strings(keys_daemon)
	sort.Strings(keys_serverless)

	keys := append(keys_daemon, keys_serverless...)

	for i, _ := range n.nodeUseCPU {
		n.nodeUseCPU[i] = model.ResourceAmount(0)
		n.nodeUseMemory[i] = model.ResourceAmount(0)
	}

	n.nodeIdMap = make(map[string]int)

	curNode := 0
	for _, k := range keys {
		if (n.batchCPURequests[k]+n.nodeUseCPU[curNode] > n.cpuSize) || (n.batchMemoryRequests[k]+n.nodeUseMemory[curNode] > n.memSize) {
			curNode++
		}
		if len(n.nodeUseCPU) <= curNode {
			n.nodeUseCPU = append(n.nodeUseCPU, model.ResourceAmount(0))
			n.nodeUseMemory = append(n.nodeUseMemory, model.ResourceAmount(0))
		}
		if (n.batchCPURequests[k]+n.nodeUseCPU[curNode] > n.cpuSize) || (n.batchMemoryRequests[k]+n.nodeUseMemory[curNode] > n.memSize) {
			panic("size of a pod request may be too large. Check")
		}

		n.nodeIdMap[k] = curNode
		n.nodeUseCPU[curNode] += n.batchCPURequests[k]
		n.nodeUseMemory[curNode] += n.batchMemoryRequests[k]
	}

	n.numRecords++
	for len(n.nodeTotalUseCPU) < len(n.nodeUseCPU) {
		n.nodeTotalUseCPU = append(n.nodeTotalUseCPU, model.ResourceAmount(0))
		n.nodeTotalUseMemory = append(n.nodeTotalUseMemory, model.ResourceAmount(0))
	}
	for i, _ := range n.nodeUseCPU {
		if n.nodeUseCPU[i] > n.cpuSize {
			panic(fmt.Sprint("Error node assign", i, n.nodeTotalUseCPU[i]))
		}
		if n.nodeUseMemory[i] > n.memSize {
			panic("Error node assign")
		}
		n.nodeTotalUseCPU[i] += n.nodeUseCPU[i]
		n.nodeTotalUseMemory[i] += n.nodeUseMemory[i]
		if n.nodeTotalUseMemory[i] < 0 {
			panic("int64 overflow")
		}
	}
}

func (n *NodePool) CollectInfo(name string, curCPURequest, curMemoryRequest model.ResourceAmount) {
	n.mode = 1
	n.batchCPURequests[name] = curCPURequest
	n.batchMemoryRequests[name] = curMemoryRequest
}

func (n *NodePool) AssignNode(name string) int {
	if n.mode != 2 {
		// first assign in the batch, calculate here
		n.batchAssignNode()
		// clear batch, wait for next records
		n.batchCPURequests = make(map[string]model.ResourceAmount)
		n.batchMemoryRequests = make(map[string]model.ResourceAmount)
	}
	n.mode = 2
	r, ok := n.nodeIdMap[name]
	if !ok {
		panic("Node Pool id not exist")
	}
	return r
}

func (n *NodePool) Dump() {
	fullPath := "metrics/" + *nodeFolder + "/nodepool.info"
	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "%f %f %d\n", model.CoresFromCPUAmount(n.cpuSize), model.BytesFromMemoryAmount(n.memSize), len(n.nodeUseCPU))
	if err != nil {
		panic(err)
	}

	for _, u := range n.nodeTotalUseCPU {
		fmt.Fprintf(file, "%f\n", 1.0-model.CoresFromCPUAmount(u)/(float64(n.numRecords)*model.CoresFromCPUAmount(n.cpuSize)))
	}
	for _, u := range n.nodeTotalUseMemory {
		fmt.Fprintf(file, "%f\n", 1.0-model.BytesFromMemoryAmount(u)/(float64(n.numRecords)*model.BytesFromMemoryAmount(n.memSize)))
	}
}
