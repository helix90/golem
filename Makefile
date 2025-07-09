# Makefile for Golem AIML interpreter

.PHONY: all test lint build clean

all: build

build:
	go build -o golem ./cmd/golem

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f golem 