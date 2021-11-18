GODEPS := $(shell find . -name '*.go')
GOOS ?= linux
GOARCH ?= amd64
VERSION := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GOVARS += -X main.version=$(VERSION)
GOVARS += -X main.buildtime=$(BUILDTIME)
GOFLAGS := -ldflags "$(GOVARS)"

ll: $(GODEPS) Makefile
	go vet ./...
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GOFLAGS) cmd/ll/ll.go

.PHONY: clean
clean:
	rm -f ll
