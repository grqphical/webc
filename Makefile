.PHONY: build format

export CGO_ENABLED=0

build:
	@go build -o webc .

format:
	@go fmt ./...