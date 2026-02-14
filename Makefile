.PHONY: all build clean test linux darwin windows keygen server client gui gui-dev

BINARY_DIR=build
MODULE=github.com/gavsh/ShikVPN

ifeq ($(OS),Windows_NT)
  VERSION ?= $(shell git describe --tags --always --dirty 2>NUL || echo dev)
  COMMIT  ?= $(shell git rev-parse --short HEAD 2>NUL || echo unknown)
  BUILD_DATE ?= unknown
else
  VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
  COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
  BUILD_DATE ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || echo "unknown")
endif

LDFLAGS=-s -w \
	-X $(MODULE)/internal/version.Version=$(VERSION) \
	-X $(MODULE)/internal/version.Commit=$(COMMIT) \
	-X $(MODULE)/internal/version.BuildDate=$(BUILD_DATE)

all: build

build: keygen server client

keygen:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/vpn-keygen ./cmd/vpn-keygen

server:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/vpn-server ./cmd/vpn-server

client:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/vpn-client ./cmd/vpn-client

# Cross-compilation targets
linux: export GOOS = linux
linux: export GOARCH = amd64
linux:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/vpn-keygen-linux ./cmd/vpn-keygen
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/vpn-server-linux ./cmd/vpn-server
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/vpn-client-linux ./cmd/vpn-client

darwin: export GOOS = darwin
darwin: export GOARCH = amd64
darwin:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/vpn-keygen-darwin ./cmd/vpn-keygen
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/vpn-server-darwin ./cmd/vpn-server
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/vpn-client-darwin ./cmd/vpn-client

windows: export GOOS = windows
windows: export GOARCH = amd64
windows:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/vpn-keygen.exe ./cmd/vpn-keygen
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/vpn-server.exe ./cmd/vpn-server
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/vpn-client.exe ./cmd/vpn-client

test:
	go test ./...

test-integration:
	go test -tags integration ./...

ifeq ($(OS),Windows_NT)
clean:
	if exist $(BINARY_DIR) del /q /s $(BINARY_DIR)\* >NUL 2>NUL
else
clean:
	rm -rf $(BINARY_DIR)/*
endif

# GUI targets (requires wails CLI: go install github.com/wailsapp/wails/v2/cmd/wails@latest)
WAILS=$(shell go env GOPATH)/bin/wails
GUI_DIR=cmd/vpn-client-gui

gui:
	cd $(GUI_DIR) && "$(WAILS)" build
ifeq ($(OS),Windows_NT)
	copy $(subst /,\,$(GUI_DIR))\build\bin\shikvpn-gui.exe $(BINARY_DIR)\shikvpn-gui.exe
else
	cp $(GUI_DIR)/build/bin/shikvpn-gui* $(BINARY_DIR)/
endif

gui-dev:
	cd $(GUI_DIR) && "$(WAILS)" dev

# Docker testing
docker-test: export GOOS = linux
docker-test: export GOARCH = amd64
docker-test:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/vpn-server ./cmd/vpn-server
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/vpn-client ./cmd/vpn-client
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/vpn-keygen ./cmd/vpn-keygen
	cd test && docker compose up --build
