BIN := "./bin/antibruteforce"

build:
	go build -v -o $(BIN) ./cmd/app
.PHONY: build

test:
	go test -v -race ./internal/usecase/repository/...
.PHONY: test

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.51.2
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
