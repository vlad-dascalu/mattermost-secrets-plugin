.PHONY: all dist build clean test server webapp

PLUGIN_ID ?= com.mattermost.secrets-plugin
PLUGIN_VERSION ?= 0.1.0
BUNDLE_NAME ?= $(PLUGIN_ID)-$(PLUGIN_VERSION).tar.gz

# Define commands with proper quoting for Windows paths
GO ?= "$(shell command -v go 2> /dev/null)"
NPM ?= "$(shell command -v npm 2> /dev/null)"
CURL ?= "$(shell command -v curl 2> /dev/null)"

# Get OS and architecture for conditional building
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

all: dist

dist: build
	mkdir -p dist
	# Make Linux and macOS binaries executable
	chmod +x build/server/dist/plugin-linux-amd64
	chmod +x build/server/dist/plugin-darwin-amd64
	cd build && tar -czf ../dist/$(BUNDLE_NAME) .

build: server webapp

server:
	mkdir -p build/server/dist
	# Build for Windows
	cd server && GOOS=windows GOARCH=amd64 $(GO) build -o "../build/server/dist/plugin-windows-amd64.exe" .
	# Build for Linux
	cd server && GOOS=linux GOARCH=amd64 $(GO) build -o "../build/server/dist/plugin-linux-amd64" .
	# Build for macOS (optional)
	cd server && GOOS=darwin GOARCH=amd64 $(GO) build -o "../build/server/dist/plugin-darwin-amd64" .

webapp:
	cd webapp && $(NPM) install --legacy-peer-deps
	cd webapp && $(NPM) install --save ajv@^8.0.0
	cd webapp && $(NPM) run build
	mkdir -p build/webapp
	cp -r webapp/dist build/webapp
	cp plugin.json build/

clean:
	rm -rf build
	rm -rf dist

test: test-server test-webapp

test-server:
	cd server && CGO_ENABLED=1 $(GO) test -race -coverprofile=coverage.txt -covermode=atomic ./...

test-webapp:
	cd webapp && $(NPM) run test

# Checks the code style, but doesn't fix it
lint: lint-server lint-webapp

lint-server:
	cd server && $(GO) vet ./...
	cd server && $(GO) fmt ./...

lint-webapp:
	cd webapp && $(NPM) run lint

# Applies code style fixes
fix: fix-server fix-webapp

fix-server:
	cd server && $(GO) fmt ./...

fix-webapp:
	cd webapp && $(NPM) run fix 