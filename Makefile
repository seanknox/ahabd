.DEFAULT: all
.PHONY: all clean image publish-image minikube-publish

DH_ORG=jpangms
VERSION=$(shell git symbolic-ref --short HEAD)-$(shell git rev-parse --short HEAD)

all: image

clean:
	go clean
	rm -f cmd/ahabd/ahabd
	rm -rf ./build

cmd/ahabd/ahabd:
cmd/ahabd/ahabd: cmd/ahabd/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/ahabd/*.go

build/.image.done: cmd/ahabd/Dockerfile cmd/ahabd/ahabd
	mkdir -p build
	cp $^ build
	docker build -t docker.io/$(DH_ORG)/ahabd -f build/Dockerfile ./build
	docker tag docker.io/$(DH_ORG)/ahabd docker.io/$(DH_ORG)/ahabd:$(VERSION)
	touch $@

image: build/.image.done

publish-image: image
	docker push docker.io/$(DH_ORG)/ahabd:$(VERSION)

minikube-publish: image
	docker save docker.io/$(DH_ORG)/ahabd | (eval $$(minikube docker-env) && docker load)
