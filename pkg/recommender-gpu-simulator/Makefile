all: clean build

GOPATH := $(shell go env GOPATH)

clean:
	rm -rf ${GOPATH}/src/k8s.io/autoscaler/vertical-pod-autoscaler/pkg

build:
	cd .. && cp -a . ${GOPATH}/src/k8s.io/autoscaler/vertical-pod-autoscaler/pkg
	cd ${GOPATH}/src/k8s.io/autoscaler/vertical-pod-autoscaler && \
	CGO_ENABLED=0 LD_FLAGS=-s go build -C pkg/recommender-gpu-simulator -mod vendor -o recommender-gpu-simulator
	cp ${GOPATH}/src/k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender-gpu-simulator/recommender-gpu-simulator bin/

del-rule:

run-rule:
	rm metrics/group/*.csv
	rm metrics/group/*.info
	bin/recommender-group-simulator \
		-ap-algorithm-ml=false \
		-ap-cpu-histogram-bucket-num=400 \
		-ap-cpu-histogram-decay-half-life=12h \
		-ap-cpu-histogram-max=3.9 \
		-ap-cpu-histogram-n=5 \
		-ap-cpu-recommend-policy=sp_95 \
		-ap-fluctuation-reducer-duration=1h \
		-ap-memory-histogram-bucket-num=400 \
		-ap-memory-histogram-decay-half-life=48h \
		-ap-memory-histogram-max=3483278000 \
		-ap-memory-histogram-n=5 \
		-ap-memory-recommend-policy=sp_98 \
		-ap-ml-cpu-hyperparam-d=0.5 \
		-ap-ml-cpu-hyperparam-wdeltal=0.5 \
		-ap-ml-cpu-hyperparam-wdeltam=0.5 \
		-ap-ml-cpu-hyperparam-wo=0.5 \
		-ap-ml-cpu-hyperparam-wu=0.5 \
		-ap-ml-cpu-size-buckets-mm=1 \
		-ap-ml-cpu-num-dm=50 \
		-ap-ml-cpu-num-mm=400 \
		-ap-ml-memory-hyperparam-d=0.7 \
		-ap-ml-memory-hyperparam-wdeltal=0.05 \
		-ap-ml-memory-hyperparam-wdeltam=0.05 \
		-ap-ml-memory-hyperparam-wo=0.5 \
		-ap-ml-memory-hyperparam-wu=0.5 \
		-ap-ml-memory-size-buckets-mm=1 \
		-ap-ml-memory-num-dm=50 \
		-ap-ml-memory-num-mm=500 \
		-oom-bump-up-ratio=1.2 \
		-oom-min-bump-up-bytes=100000000 \
		-metrics-summary-ignore-head=1800 \
		-memory-limit-request-ratio=1.04 \
		-metrics-folder=group \
		-node-size-cpu=3.9 \
		-node-size-memory=3483278000 \
		-nodepool-folder=group

run-ml:
	# rm metrics/group-ml/*.csv
	# rm metrics/group-ml/*.info
	bin/recommender-gpu-simulator \
		-ap-algorithm-ml=true \
		-ap-ml-gpus-hyperparam-d=0.8571428571428571 \
		-ap-ml-gpus-hyperparam-wdeltal=0.0 \
		-ap-ml-gpus-hyperparam-wdeltam=0.0 \
		-ap-ml-gpus-hyperparam-wo=0.14285714285714285 \
		-ap-ml-gpus-hyperparam-wu=0.0 \
		-ap-ml-gpus-num-dm=50 \
		-ap-ml-gpus-num-mm=3 \
		-ap-ml-gpum-hyperparam-d=0.6666666666666666 \
		-ap-ml-gpum-hyperparam-wdeltal=0.009 \
		-ap-ml-gpum-hyperparam-wdeltam=0.000999 \
		-ap-ml-gpum-hyperparam-wo=0.5 \
		-ap-ml-gpum-hyperparam-wu=0.01 \
		-ap-ml-gpum-num-dm=50 \
		-ap-ml-gpum-num-mm=4 \
		-metrics-summary-ignore-head=1800 \
		-metrics-folder=gpu-ml \
		-nodepool-folder=gpu-ml


run-const:
	# rm metrics/group-const/*.csv
	# rm metrics/group-const/*.info
	bin/recommender-gpu-simulator \
		-ap-algorithm-ml=true \
		-ap-ml-gpus-hyperparam-d=0.8571428571428571 \
		-ap-ml-gpus-hyperparam-wdeltal=0.0 \
		-ap-ml-gpus-hyperparam-wdeltam=0.0 \
		-ap-ml-gpus-hyperparam-wo=0.14285714285714285 \
		-ap-ml-gpus-hyperparam-wu=0.0 \
		-ap-ml-gpus-num-dm=50 \
		-ap-ml-gpus-num-mm=3 \
		-ap-ml-gpum-hyperparam-d=0.6666666666666666 \
		-ap-ml-gpum-hyperparam-wdeltal=0.009 \
		-ap-ml-gpum-hyperparam-wdeltam=0.000999 \
		-ap-ml-gpum-hyperparam-wo=0.5 \
		-ap-ml-gpum-hyperparam-wu=0.01 \
		-ap-ml-gpum-num-dm=50 \
		-ap-ml-gpum-num-mm=4 \
		-metrics-summary-ignore-head=1800 \
		-metrics-folder=gpu-const \
		-nodepool-folder=gpu-const \
		-is-const=true
