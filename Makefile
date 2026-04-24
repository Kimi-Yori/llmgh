GO := /usr/local/go/bin/go
BINARY := llmgh
VERSION := 0.1.0
LDFLAGS := -s -w -X main.version=$(VERSION)

DIST := dist
PLATFORMS := linux/amd64 darwin/amd64 darwin/arm64

.PHONY: build test clean install release

build:
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY) .

test:
	$(GO) test ./... -v -count=1

release:
	@mkdir -p $(DIST)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output=$(DIST)/$(BINARY)-$${os}-$${arch}; \
		echo "Building $${output}..."; \
		CGO_ENABLED=0 GOOS=$${os} GOARCH=$${arch} $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $${output} . ; \
	done
	@echo "Done. Binaries in $(DIST)/"
	@ls -lh $(DIST)/

clean:
	rm -f $(BINARY)
	rm -rf $(DIST)

install: build
	cp $(BINARY) /usr/local/bin/$(BINARY)
