GO := go

MAIN_FILE := main.go

BUILD_DIR := build
EXECUTABLE := blob
BIN_NAME := $(BUILD_DIR)/$(EXECUTABLE)
INSTALLED_NAME := /usr/local/bin/$(EXECUTABLE)

PACKAGE_DIR := ./cmd/blobapi
PACKAGE_PATHS := $(PACKAGE_DIR)

SRC := $(shell find . -iname "*.go" -and -not -name "*_test.go")
SRC_TEST := $(shell find . -iname "*_test.go")

COVERAGE_FILE=coverage.out

PUBLISH = publish/$(EXECUTABLE)-linux-amd64 publish/$(EXECUTABLE)-darwin-amd64 publish/$(EXECUTABLE)-darwin-arm64


.PHONY: all
all: $(BIN_NAME)

$(BIN_NAME): $(SRC) go.mod go.sum
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BIN_NAME) $(MAIN_FILE)


.PHONY: publish
publish: $(PUBLISH)

.PHONY: publish/$(EXECUTABLE)-linux-amd64
publish/$(EXECUTABLE)-linux-amd64:
	@# Force build; don't let existing versions interfere.
	rm -f $(BIN_NAME)
	GOOS=linux GOARCH=amd64 $(MAKE) $(BIN_NAME)
	mkdir -p $$(dirname "$@")
	mv $(BIN_NAME) $@

.PHONY: publish/$(EXECUTABLE)-darwin-amd64
publish/$(EXECUTABLE)-darwin-amd64:
	@# Force build; don't let existing versions interfere.
	rm -f $(BIN_NAME)
	GOOS=darwin GOARCH=amd64 $(MAKE) $(BIN_NAME)
	mkdir -p $$(dirname "$@")
	mv $(BIN_NAME) $@

.PHONY: publish/$(EXECUTABLE)-darwin-arm64
publish/$(EXECUTABLE)-darwin-arm64:
	@# Force build; don't let existing versions interfere.
	rm -f $(BIN_NAME)
	GOOS=darwin GOARCH=arm64 $(MAKE) $(BIN_NAME)
	mkdir -p $$(dirname "$@")
	mv $(BIN_NAME) $@


.PHONY: install isntall
install isntall: $(INSTALLED_NAME)
$(INSTALLED_NAME): $(BIN_NAME)
	cp $(BIN_NAME) $(INSTALLED_NAME)


.PHONY: test
test: $(BIN_NAME) $(SRC_TEST)
	@if [ -z $$T ]; then \
		$(GO) test -v ./...; \
	else \
		$(GO) test -v ./... -run $$T; \
	fi


$(COVERAGE_FILE): $(BIN_NAME) $(SRC_TEST)
	$(GO) test -v --coverprofile=$(COVERAGE_FILE) ./...

.PHONY: coverage
coverage: $(COVERAGE_FILE)
	$(GO) tool cover -func=$(COVERAGE_FILE)

.PHONY: pretty-coverage
pretty-coverage: $(COVERAGE_FILE)
	$(GO) tool cover -html=$(COVERAGE_FILE)

.PHONY: fmt
fmt:
	@$(GO) fmt ./...

.PHONY: clean
clean:
	rm -rf $(COVERAGE_FILE) $(BUILD_DIR)
