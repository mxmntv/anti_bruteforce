BIN := "./bin/antibruteforce"
DOCKER_IMG="antibruteforce:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/app
.PHONY: build

run-bin: build
	$(BIN) -config ./config/config.yml
.PHONY: run-bin

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f ./Dockerfile .
.PHONY: build-img

run-img: build-img
	docker run $(DOCKER_IMG)
.PHONY: run-img

test:
	go test -v -race ./internal/usecase/repository/...
.PHONY: test

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.41.1
.PHONY: install-lint-deps

lint: install-lint-deps
	golangci-lint run ./...
.PHONY: lint

run:
	docker-compose -f ./docker-compose.yml up
.PHONY: run

stop:
	docker-compose down -v --remove-orphans
.PHONY: stop
