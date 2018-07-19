.DEFAULT: all
.PHONY: all clean lint image publish-image

DOCKER_REGISTRY ?= docker.io
DOCKER_ORG ?= jpangms
VERSION ?= $(shell git rev-parse --short HEAD)
TAG ?= git-$(VERSION)
IMAGE_REPO ?= ahabd
IMAGE ?= $(DOCKER_REGISTRY)/$(DOCKER_ORG)/$(IMAGE_REPO)

PKG=github.com/juan-lee/ahabd

LINT_FLAGS ?= --deadline=5m --cyclo-over=40 --fast --vendor

.PHONY: info
info: 
	@echo "Version: $(VERSION)"
	@echo "Tag: $(TAG)"
	@echo "DOCKER_ORG: $(DOCKER_ORG)"
	@echo "DOCKER_REGISTRY: $(DOCKER_REGISTRY)"
	@echo "Image: $(IMAGE):$(TAG)"

all: image

clean:
	go clean
	rm -f ./ahabd
	rm -rf ./build

ahabd:
ahabd: *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X $(PKG)/pkg/version.Version=$(TAG)" -o $@ *.go

lint:
	@gometalinter $(LINT_FLAGS) ./...

build/.image.done: Dockerfile ahabd
	mkdir -p build
	cp $^ build
	docker build -t $(IMAGE) -f build/Dockerfile ./build
	docker tag $(IMAGE) $(IMAGE):$(TAG)
	touch $@

image: build/.image.done

publish-image: image
	docker push $(IMAGE):$(TAG)
	docker push $(IMAGE)

publish-immutable-image: image
	docker push $(IMAGE):$(TAG)
