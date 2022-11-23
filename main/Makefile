all: build_bins build_images

build_bins: client node 

client:
	CGO_ENABLED=0 GOOS=linux go build -o bin/client client/client.go

node:
	CGO_ENABLED=0 GOOS=linux go build -o bin/node node/node.go

build_images:  image-python

image-python:
	podman build executor --tag executor
	rm executor/image.tar
	podman save --output executor/image.tar localhost/executor

.PHONY: build_bins build_images
