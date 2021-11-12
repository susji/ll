GODEPS := $(shell find . -name '*.go')
GOOS ?= linux
GOARCH ?= amd64

ll: $(GODEPS) Makefile
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build cmd/ll/ll.go

.PHONY: clean
clean:
	rm -f ll
