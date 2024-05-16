package main

import (
	"bufio"
	"flag"
	"os"
	"strconv"
	"strings"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

var (
	traceFile = flag.String("trace-file", "trace", "trace file name")
)

type traceParser struct {
	traceFile *os.File
	scanner   *bufio.Scanner
}

func NewTraceParser() *traceParser {
	filePath := "trace/" + *traceFile + ".data"
	traceFile, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	return &traceParser{
		traceFile: traceFile,
		scanner:   bufio.NewScanner(traceFile),
	}
}

func (p *traceParser) ScanNext() (bool, int, model.ResourceAmount, model.ResourceAmount) {
	if !p.scanner.Scan() {
		return false, 0, model.ResourceAmount(0), model.ResourceAmount(0)
	}
	line := p.scanner.Text()
	nums := strings.Fields(line)
	if len(nums) != 3 {
		panic("Error in trace file, not 3 numbers in one line: " + line)
	}

	var period int = 0
	var cpu float64 = 0.0
	var memory int64 = 0
	var err error

	if period, err = strconv.Atoi(nums[0]); err != nil {
		panic(err)
	}

	if cpu, err = strconv.ParseFloat(nums[1], 64); err != nil {
		panic(err)
	}

	if memory, err = strconv.ParseInt(nums[2], 10, 64); err != nil {
		panic(err)
	}
	return true, period, model.CPUAmountFromCores(cpu), model.MemoryAmountFromBytes(float64(memory))
}

func (p *traceParser) Close() {
	p.traceFile.Close()
}
