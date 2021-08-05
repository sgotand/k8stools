


.PHONY: build
build: k8stools

.PHONY: container
container:
	docker build . -t ghcr.io/sgotand/k8stools

k8stools: main.go
	go build .
