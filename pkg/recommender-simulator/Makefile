all: clean build

GOPATH=/Users/yuqiangma/go

clean:
	rm -rf ${GOPATH}/src/k8s.io/autoscaler/vertical-pod-autoscaler/pkg

build:
	cd .. && cp -a . ${GOPATH}/src/k8s.io/autoscaler/vertical-pod-autoscaler/pkg
	cd ${GOPATH}/src/k8s.io/autoscaler/vertical-pod-autoscaler && \
	CGO_ENABLED=0 LD_FLAGS=-s go build -C pkg/recommender-simulator -mod vendor -o recommender-simulator
	cp ${GOPATH}/src/k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender-simulator/recommender-simulator bin/

