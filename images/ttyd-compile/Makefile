# Image URL to use all building/pushing image targets
IMG ?= eu.gcr.io/cloud-native-coding/ttyd-build:latest

CID := $(shell docker create ${IMG})

# Build the docker image
docker-build:
	docker build . -t ${IMG}

# Get the binary
get: docker-build
	docker cp ${CID}:/opt/stage/x86_64-linux-musl/bin/ttyd ./
	docker rm ${CID}
	
