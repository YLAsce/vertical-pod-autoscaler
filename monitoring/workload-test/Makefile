.PHONY: all
all: docker-build docker-push uptag

TAG?=test-11
REGISTRY?=yuqiang302
COMPONENT=workload

ALL_ARCHITECTURES?=amd64  # add arm32v7 arm64v8 but extra string processing needed
VPA ?= 

.PHONY: docker-build
docker-build: $(addprefix docker-build-,$(ALL_ARCHITECTURES))

.PHONY: docker-build-*
docker-build-%:
ifndef REGISTRY
	ERR = $(error REGISTRY is undefined)
	$(ERR)
endif
ifndef TAG
	ERR = $(error TAG is undefined)
	$(ERR)
endif

	docker buildx build --pull --load --platform linux/$* -t ${REGISTRY}/${COMPONENT}-$*:${TAG} --build-arg ARCH=$* .

.PHONY: docker-push
docker-push: $(addprefix do-push-,$(ALL_ARCHITECTURES)) push-multi-arch;

.PHONY: do-push-*
do-push-%:
ifndef REGISTRY
	ERR = $(error REGISTRY is undefined)
	$(ERR)
endif
ifndef TAG
	ERR = $(error TAG is undefined)
	$(ERR)
endif
	docker push ${REGISTRY}/${COMPONENT}-$*:${TAG}

.PHONY: push-multi-arch
push-multi-arch:
	docker manifest create --amend $(REGISTRY)/${COMPONENT}:$(TAG) $(shell echo $(ALL_ARCHITECTURES) | sed -e "s~[^ ]*~$(REGISTRY)/${COMPONENT}\-&:$(TAG)~g")
	@for arch in $(ALL_ARCHITECTURES); do docker manifest annotate --arch $${arch} $(REGISTRY)/${COMPONENT}:$(TAG) $(REGISTRY)/${COMPONENT}-$${arch}:${TAG}; done
	docker manifest push --purge $(REGISTRY)/${COMPONENT}:$(TAG)


.PHONY: uptag
uptag:
	sed -i '' '19s/.*/          image: docker.io\/${REGISTRY}\/${COMPONENT}:${TAG}/' workload_test.yaml

.PHONY: restart
restart: stop start

.PHONY: stop
stop:
	kubectl delete -f workload_test.yaml
	sleep 15

.PHONY: start
start:
	echo 0 | aws s3 cp - s3://mlintra/workload-test.state
	kubectl apply -f workload_test.yaml


.PHONY: reset
reset:
	echo 0 | aws s3 cp - s3://mlintra/stress.state

.PHONY: state
state:
	aws s3 cp s3://mlintra/workload-test.state -