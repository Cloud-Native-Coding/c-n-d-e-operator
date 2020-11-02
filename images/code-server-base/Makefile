VERSION=3.4.1

# Image URL to use all building/pushing image targets
IMG ?= eu.gcr.io/cloud-native-coding/code-server-example:${VERSION}

# Build the docker image
docker-build:
	cp ../ttyd-compile/ttyd .
	docker build . -t ${IMG} --build-arg VERSION=${VERSION}

# Push the docker image
docker-push:
	docker push ${IMG}
