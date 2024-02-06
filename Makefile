help:
	@cat Makefile

tidy:
	go mod tidy

format: tidy
	go fmt

generate: tidy
	go generate ./...

lint:
	golangci-lint run


