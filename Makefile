.PHONY: build test test-race test-e2e test-integration fmt check

build:
	mkdir -p .build
	CGO_ENABLED=0 go build -trimpath -o .build/agent ./examples/basic-agent

test:
	go test ./...

test-race:
	go test -race ./...

test-e2e:
	go test -tags=e2e ./tests/e2e -count=1 -timeout=10m

test-integration:
	go test -tags=integration ./... -count=1 -timeout=10m

fmt:
	gofumpt -w agent cli config core modules examples internal tests

check:
	test -z "$$(gofumpt -l agent cli config core modules examples internal tests)"
	go vet ./...
	go test ./...
