
# Image URL to use all building/pushing image targets
IMG ?=  eu.gcr.io/cloud-native-coding/create-kubeconfig:latest

# Build the docker image
docker-build:
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}
