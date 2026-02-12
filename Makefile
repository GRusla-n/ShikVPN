.PHONY: all build clean test linux darwin windows keygen server client

BINARY_DIR=build
MODULE=github.com/gavsh/simplevpn

all: build

build: keygen server client

keygen:
	go build -o $(BINARY_DIR)/vpn-keygen ./cmd/vpn-keygen

server:
	go build -o $(BINARY_DIR)/vpn-server ./cmd/vpn-server

client:
	go build -o $(BINARY_DIR)/vpn-client ./cmd/vpn-client

# Cross-compilation targets
linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_DIR)/vpn-keygen-linux ./cmd/vpn-keygen
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_DIR)/vpn-server-linux ./cmd/vpn-server
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_DIR)/vpn-client-linux ./cmd/vpn-client

darwin:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_DIR)/vpn-keygen-darwin ./cmd/vpn-keygen
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_DIR)/vpn-server-darwin ./cmd/vpn-server
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_DIR)/vpn-client-darwin ./cmd/vpn-client

windows:
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_DIR)/vpn-keygen.exe ./cmd/vpn-keygen
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_DIR)/vpn-server.exe ./cmd/vpn-server
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_DIR)/vpn-client.exe ./cmd/vpn-client

test:
	go test ./...

test-integration:
	go test -tags integration ./...

clean:
	rm -rf $(BINARY_DIR)/*

# Docker testing
docker-test:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_DIR)/vpn-server ./cmd/vpn-server
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_DIR)/vpn-client ./cmd/vpn-client
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_DIR)/vpn-keygen ./cmd/vpn-keygen
	cd test && docker compose up --build
