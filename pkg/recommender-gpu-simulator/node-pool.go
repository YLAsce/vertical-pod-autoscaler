package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
)

var (
	nodeFolder = flag.String("nodepool-folder", "group", "node pool dump parent folder")
)

// nodes in single mode
// one pool for each size

type NodePool struct {
	mode            int // 1 is recording, 2 is assigning. Decisions are made before every first assigning
	batchRequestIds map[string]int

	nodeIdMap  map[string]int
	maxNodeNum int

	// nodeTotalUseCPU    []model.ResourceAmount
	// nodeTotalUseMemory []model.ResourceAmount
	// numRecords         int64
}

func NewNodePool() *NodePool {

	return &NodePool{
		mode:            0,
		batchRequestIds: make(map[string]int),

		nodeIdMap:  make(map[string]int),
		maxNodeNum: 0,

		// nodeTotalUseCPU:    make([]model.ResourceAmount, 0),
		// nodeTotalUseMemory: make([]model.ResourceAmount, 0),
		// numRecords:         0,
	}
}

func (n *NodePool) batchAssignNode() {
	//input: batchRequestIds
	//output: nodeIdMap
	//现在的算法:先按使用量排序，再按名称（先daemon后serverless排序）
	// 创建一个切片来保存 map 的键值对
	type kv struct {
		Key   string
		Value int
	}

	var kvSlice []kv
	for k, v := range n.batchRequestIds {
		kvSlice = append(kvSlice, kv{k, v})
	}

	// 按值排序，如果值相同再按键第一个字符排序，最后按键排序
	sort.Slice(kvSlice, func(i, j int) bool {
		if kvSlice[i].Value == kvSlice[j].Value {
			if kvSlice[i].Key[0] == kvSlice[j].Key[0] {
				return kvSlice[i].Key < kvSlice[j].Key
			}
			return kvSlice[i].Key[0] < kvSlice[j].Key[0]
		}
		return kvSlice[i].Value < kvSlice[j].Value
	})

	curNode := -1
	curNodeSize := 0
	curNodeMode := 0
	for _, s := range kvSlice {
		if curNodeMode != s.Value || curNodeSize == 42 {
			curNode++
			curNodeSize = 0
			curNodeMode = s.Value
		}
		n.nodeIdMap[s.Key] = curNode
		curNodeSize += s.Value
	}
	n.maxNodeNum = max(n.maxNodeNum, curNode+1)
}

func (n *NodePool) CollectInfo(name string, CurRequestId int) {
	n.mode = 1
	n.batchRequestIds[name] = CurRequestId
}

func (n *NodePool) AssignNode(name string) int {
	if n.mode != 2 {
		// first assign in the batch, calculate here
		n.batchAssignNode()
		// clear batch, wait for next records
		n.batchRequestIds = make(map[string]int)
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

	_, err = fmt.Fprintf(file, "%d\n", n.maxNodeNum)
	if err != nil {
		panic(err)
	}

	// for _, u := range n.nodeTotalUseCPU {
	// 	fmt.Fprintf(file, "%f\n", 1.0-model.CoresFromCPUAmount(u)/(float64(n.numRecords)*model.CoresFromCPUAmount(n.cpuSize)))
	// }
	// for _, u := range n.nodeTotalUseMemory {
	// 	fmt.Fprintf(file, "%f\n", 1.0-model.BytesFromMemoryAmount(u)/(float64(n.numRecords)*model.BytesFromMemoryAmount(n.memSize)))
	// }
}
