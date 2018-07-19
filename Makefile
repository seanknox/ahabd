.DEFAULT: all
.PHONY: all clean lint image publish-image

DOCKER_REGISTRY ?= docker.io
DOCKER_ORG ?= jpangms
VERSION=$(shell git symbolic-ref --short HEAD)-$(shell git rev-parse --short HEAD)
PKG=github.com/juan-lee/ahabd

LINT_FLAGS ?= --deadline=5m --cyclo-over=40 --fast --vendor

all: image

clean:
	go clean
	rm -f ./ahabd
	rm -rf ./build

ahabd:
ahabd: *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X $(PKG)/pkg/version.Version=$(VERSION)" -o $@ *.go

lint:
	@gometalinter $(LINT_FLAGS) ./...

build/.image.done: Dockerfile ahabd
	mkdir -p build
	cp $^ build
	docker build -t $(DOCKER_REGISTRY)/$(DOCKER_ORG)/ahabd -f build/Dockerfile ./build
	docker tag $(DOCKER_REGISTRY)/$(DOCKER_ORG)/ahabd $(DOCKER_REGISTRY)/$(DOCKER_ORG)/ahabd:$(VERSION)
	touch $@

image: build/.image.done

publish-image: image
	docker push $(DOCKER_REGISTRY)/$(DOCKER_ORG)/ahabd:$(VERSION)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_ORG)/ahabd
