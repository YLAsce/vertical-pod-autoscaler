.PHONY: all
all: build run


.PHONY: build
build: 
	sbt package

.PHONY: run
run: 
	spark-submit target/scala-2.12/clusterdata_2.12-1.0.jar task10 local 330587160469

.PHONY: upload
upload:
	gsutil cp target/scala-2.12/clusterdata_2.12-1.0.jar  gs://clusterdata_autopilot/clusterdata.jar
	gsutil ls gs://clusterdata_autopilot/ 