.PHONY: build build-initializer docs tools-openspec setup-specs check-knowledge check-specs check-artifacts test test-race test-e2e test-integration fmt check

VERSION ?= dev

build:
	mkdir -p .build
	CGO_ENABLED=0 go build -trimpath -o .build/agent ./examples/basic-agent

build-initializer:
	mkdir -p .build
	CGO_ENABLED=0 go build -trimpath -ldflags "-X main.version=$(VERSION)" -o .build/dome ./cmd/dome

docs:
	scripts/install-hugo.sh
	scripts/build-docs.sh next

tools-openspec:
	scripts/install-openspec.sh

setup-specs:
	scripts/setup-specs.sh

check-knowledge:
	go run scripts/check-knowledge.go -policy

check-specs:
	scripts/openspec.sh validate --all --strict

check-artifacts: check-knowledge check-specs

test:
	go test ./...

test-race:
	go test -race ./...

test-e2e:
	go test -tags=e2e ./tests/e2e -count=1 -timeout=10m

test-integration:
	go test -tags=integration ./... -count=1 -timeout=10m

fmt:
	gofumpt -w agent cli cmd config core modules examples internal tests scripts

check:
	test -z "$$(gofumpt -l agent cli cmd config core modules examples internal tests scripts)"
	go vet ./...
	go test ./...
