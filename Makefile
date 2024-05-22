.PHONY: all fmt tidy lint
all: fmt tidy lint

fmt:
	go fmt ./...

tidy:
	go mod tidy -v

lint:
	golangci-lint run
