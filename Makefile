.PHONY: build format test

export CGO_ENABLED=0

build:
	@go build -o webc .

format:
	@go fmt ./...
	@npx prettier ./templates --write

test:
	@go test -v ./...