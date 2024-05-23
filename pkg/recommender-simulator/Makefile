all: clean build

GOPATH := $(shell go env GOPATH)

clean:
	rm -rf ${GOPATH}/src/k8s.io/autoscaler/vertical-pod-autoscaler/pkg

build:
	cd .. && cp -a . ${GOPATH}/src/k8s.io/autoscaler/vertical-pod-autoscaler/pkg
	cd ${GOPATH}/src/k8s.io/autoscaler/vertical-pod-autoscaler && \
	CGO_ENABLED=0 LD_FLAGS=-s go build -C pkg/recommender-simulator -mod vendor -o recommender-simulator
	cp ${GOPATH}/src/k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender-simulator/recommender-simulator bin/

run-rule:
	bin/recommender-simulator \
		-ap-algorithm-ml=false \
		-ap-cpu-histogram-bucket-num=400 \
		-ap-cpu-histogram-decay-half-life=12h \
		-ap-cpu-histogram-max=1.9 \
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
		-initial-cpu=0.6 \
		-initial-memory=600000000 \
		-metrics-file=metrics-rule \
		-oom-bump-up-ratio=1.2 \
		-oom-min-bump-up-bytes=100000000 \
		-recommender-interval=5m \
		-trace-file=trace \
		-metrics-summary-file=metrics-summary-rule \
		-metrics-summary-ignore-head=1800 \
		-memory-limit-request-ratio=1.04

run-ml:
	bin/recommender-simulator \
		-ap-algorithm-ml=true \
		-ap-cpu-histogram-bucket-num=400 \
		-ap-cpu-histogram-decay-half-life=12h \
		-ap-cpu-histogram-max=1.9 \
		-ap-cpu-histogram-n=5 \
		-ap-cpu-recommend-policy=sp_95 \
		-ap-fluctuation-reducer-duration=1h \
		-ap-memory-histogram-bucket-num=400 \
		-ap-memory-histogram-decay-half-life=48h \
		-ap-memory-histogram-max=3483278000 \
		-ap-memory-histogram-n=5 \
		-ap-memory-recommend-policy=sp_98 \
		-ap-ml-cpu-hyperparam-d=0.8571428571428571 \
		-ap-ml-cpu-hyperparam-wdeltal=0.0 \
		-ap-ml-cpu-hyperparam-wdeltam=0.0 \
		-ap-ml-cpu-hyperparam-wo=0.14285714285714285 \
		-ap-ml-cpu-hyperparam-wu=0.0 \
		-ap-ml-cpu-size-buckets-mm=1 \
		-ap-ml-cpu-num-dm=50 \
		-ap-ml-cpu-num-mm=400 \
		-ap-ml-memory-hyperparam-d=0.07612065994954909 \
		-ap-ml-memory-hyperparam-wdeltal=0.4129785051687559 \
		-ap-ml-memory-hyperparam-wdeltam=0.08667298207063878 \
		-ap-ml-memory-hyperparam-wo=0.39 \
		-ap-ml-memory-hyperparam-wu=0.7975485606173869 \
		-ap-ml-memory-size-buckets-mm=1 \
		-ap-ml-memory-num-dm=50 \
		-ap-ml-memory-num-mm=500 \
		-initial-cpu=0.6 \
		-initial-memory=600000000 \
		-metrics-file=metrics-mlv2-test \
		-oom-bump-up-ratio=1.2 \
		-oom-min-bump-up-bytes=100000000 \
		-recommender-interval=5m \
		-trace-file=trace \
		-metrics-summary-file=metrics-summary-mlv2-test \
		-metrics-summary-ignore-head=1800 \
		-memory-limit-request-ratio=1.04 \
		-exit-memory-large-overrun=2000
