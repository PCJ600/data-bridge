BINARY_NAME := cloud-agent
IMAGE_NAME := cloud-agent
IMAGE_TAG := 1.0.0.10003
GO_BUILD_FLAGS := -ldflags="-s -w"

PROXY_ADDR ?= 
DOCKER_BUILD_ARGS := --network host
ifneq ($(PROXY_ADDR),)
	DOCKER_BUILD_ARGS += --build-arg HTTP_PROXY=$(PROXY_ADDR) \
	                     --build-arg HTTPS_PROXY=$(PROXY_ADDR)
endif

.PHONY: all build clean dist

all: build

build:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o ./$(BINARY_NAME) ./cmd/main.go
	@echo "Binary built: ./$(BINARY_NAME)"

clean:
	rm -rf ./$(BINARY_NAME)
	@echo "Cleaned build artifacts"

dist:
	docker build $(DOCKER_BUILD_ARGS) -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@echo "Docker image built: $(IMAGE_NAME):$(IMAGE_TAG)"
