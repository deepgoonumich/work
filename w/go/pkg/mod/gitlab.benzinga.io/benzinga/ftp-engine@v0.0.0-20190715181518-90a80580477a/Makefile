sources := $(shell find "$(PWD)" -name '*.go')

.PHONY: all get check test install

all: build

# Get dependencies
get:
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

# Check
check:
	golangci-lint run -c .golangci.yml

deps:
	docker-compose down
	docker-compose up -d jaeger zk1 zk2 zk3 kafka-1 kafka-2 kafka-3 redis

# Test
test: deps check
	go test ./...

# Build
build:
	docker-compose --build