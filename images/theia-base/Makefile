# Image URL to use all building/pushing image targets
IMG ?= eu.gcr.io/cloud-native-coding/theia-example:latest

# Build the docker image
docker-build:
	cp ../ttyd-compile/ttyd .
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}
