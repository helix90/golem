# Makefile for Golem AIML interpreter

.PHONY: all test lint build

all: build

build:
	go build -o golem ./cmd/golem

test:
	go test ./...

lint:
	go vet ./... 