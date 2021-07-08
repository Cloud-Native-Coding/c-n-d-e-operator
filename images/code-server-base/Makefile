VERSION=3.10.2

# Image URL to use all building/pushing image targets
IMG ?= eu.gcr.io/cloud-native-coding/code-server-example:${VERSION}

# Build the docker image
docker-build:
	cp ../ttyd-compile/ttyd .
	docker build . -t ${IMG} --build-arg VERSION=${VERSION}

# Push the docker image
docker-push:
	docker push ${IMG}

docker-run:
	docker run --rm -p 8080:8080 ${IMG}