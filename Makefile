SHELL := /bin/bash

APP       := tfm
PKG       := ./cmd/$(APP)
BIN_DIR   := bin
BIN       := $(BIN_DIR)/$(APP)
GO        ?= go
GOFLAGS   ?=
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.1.0)
LDFLAGS   ?= -X 'main.version=$(VERSION)'

.PHONY: all build build-release install run clean fmt vet test help install-config config-path

all: build

$(BIN):
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN) $(PKG)

build: $(BIN)

build-release: GOFLAGS += -trimpath
build-release: LDFLAGS += -s -w
build-release:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN) $(PKG)

install:
	CGO_ENABLED=0 $(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS)" $(PKG)

run: build
	$(BIN) $(ARGS)

clean:
	rm -rf $(BIN_DIR)
	$(GO) clean -cache -testcache

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

help:
	@echo "make build          - Build $(BIN)"
	@echo "make build-release  - Build optimized static binary"
	@echo "make install        - Install to $$GOBIN (or $$GOPATH/bin)"
	@echo "make run ARGS='...' - Build and run with args"
	@echo "make fmt vet test   - Format, vet, and test"
	@echo "make clean          - Clean build artifacts"
	@echo "make install-config - Install example config to XDG path"
	@echo "make config-path    - Print resolved config path"

# XDG config path resolution
CONFIG_HOME ?= $(XDG_CONFIG_HOME)
ifeq ($(CONFIG_HOME),)
CONFIG_HOME := $(HOME)/.config
endif
CONFIG_DIR  := $(CONFIG_HOME)/tfm
CONFIG_FILE := $(CONFIG_DIR)/config.toml

config-path:
	@echo $(CONFIG_FILE)

install-config:
	@mkdir -p $(CONFIG_DIR)
	@if [ -f "$(CONFIG_FILE)" ]; then \
	  echo "Config already exists at $(CONFIG_FILE). Skipping."; \
	else \
	  cp configs/config.example.toml "$(CONFIG_FILE)" && echo "Installed config to $(CONFIG_FILE)"; \
	fi
