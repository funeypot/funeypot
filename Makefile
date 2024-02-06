.PHONY: help tidy format generate test lint

help:
	@cat Makefile

tidy:
	go mod tidy

format: tidy
	go fmt

generate: tidy
	go generate ./...

test:
	go test -race -coverpkg=./... -coverprofile=coverage.out -covermode=atomic ./...

lint:
	golangci-lint run
