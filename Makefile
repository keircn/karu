.PHONY: all
all: clean build

BINARY_NAME=karu
MAIN_PATH=cmd/karu/main.go
BUILD_DIR=bin
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
GOFMT=gofmt

.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: clean
clean:
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)

.PHONY: lint
lint:
	$(GOVET) ./...
	@if [ -n "$$($(GOFMT) -l .)" ]; then \
		echo "Code is not formatted. Run 'make fmt' to fix."; \
		$(GOFMT) -l .; \
		exit 1; \
	fi
	@echo "Lint checks passed"

.PHONY: fmt
fmt:
	$(GOFMT) -w .
	@echo "Code formatted"

.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	@echo "Dependencies tidied"

.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Dependencies downloaded"

.PHONY: install
install: build
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) installed to /usr/local/bin/"

.PHONY: uninstall
uninstall:
	@rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) uninstalled"

.PHONY: build-all
build-all: clean
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@ls -la $(BUILD_DIR)/

.PHONY: release
release: build-all
	@cd $(BUILD_DIR) && \
	tar -czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
	tar -czf $(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
	tar -czf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
	tar -czf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
	zip $(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@echo "Release archives created in $(BUILD_DIR)/"

