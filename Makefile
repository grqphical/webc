.PHONY: build format test

export CGO_ENABLED=0

build:
	@go build -o webc .

format:
	@go fmt ./...

test:
	@go test -v ./...