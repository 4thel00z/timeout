BINARY  := timeout
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

GOLANGCI_LINT_VERSION := v2.12.2
GOFUMPT_VERSION       := v0.10.0
LEFTHOOK_VERSION      := v1.13.6

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build bin/timeout
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

.PHONY: install
install: ## Install timeout into GOBIN
	go install -trimpath -ldflags "$(LDFLAGS)" .

.PHONY: test
test: ## Run the test suite with the race detector
	go test -race -count=1 ./...

.PHONY: fmt
fmt: ## Format the code with gofumpt
	gofumpt -l -w .

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: check
check: ## Run the same checks as CI
	lefthook run ci
	$(MAKE) test

.PHONY: tools
tools: ## Install the dev tools at pinned versions
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)
	go install github.com/evilmartians/lefthook@$(LEFTHOOK_VERSION)

.PHONY: hooks
hooks: ## Install the git hooks
	lefthook install

.PHONY: demo
demo: build ## Record assets/demo.gif with vhs and ffmpeg
	rm -rf .demo-frames
	PATH="$(CURDIR)/bin:$$PATH" vhs assets/demo.tape
	ffmpeg -y -loglevel error -framerate 50 \
		-i .demo-frames/frame-text-%05d.png -i .demo-frames/frame-cursor-%05d.png \
		-filter_complex "[0][1]overlay,pad=iw+48:ih+48:24:24:color=0x1e1e2e,fps=20,split[a][b];[a]palettegen=stats_mode=diff[p];[b][p]paletteuse=dither=none" \
		assets/demo.gif
	rm -rf .demo-frames

.PHONY: snapshot
snapshot: ## Build a local release snapshot with goreleaser
	goreleaser release --snapshot --clean

.PHONY: release-check
release-check: ## Validate .goreleaser.yaml
	goreleaser check

.PHONY: clean
clean: ## Remove build output
	rm -rf bin dist
