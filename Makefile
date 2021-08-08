GO := go
IMAGEMAGICK := convert

MAIN_FILE := main.go

BUILD_DIR := build
EXECUTABLE := blob
BIN_NAME := $(BUILD_DIR)/$(EXECUTABLE)
INSTALLED_NAME := /usr/local/bin/$(EXECUTABLE)

WP_PACKAGE_DIR := ./cmd/blobapi
PACKAGE_PATHS := $(WP_PACKAGE_DIR)

AUTOGEN_VERSION_FILENAME=$(WP_PACKAGE_DIR)/version-temp.go

SRC := $(shell find . -iname "*.go" -and -not -name "*_test.go") $(AUTOGEN_VERSION_FILENAME)

COVERAGE_FILE=coverage.out

PUBLISH = publish/wp-linux-amd64 publish/wp-darwin-amd64


.PHONY: all
all: $(BIN_NAME)

$(BIN_NAME): $(SRC)
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BIN_NAME) $(MAIN_FILE)


.PHONY: publish
publish: $(PUBLISH)

.PHONY: publish/wp-linux-amd64
publish/wp-linux-amd64:
	# Force build; don't let existing versions interfere.
	rm -f $(BIN_NAME)
	GOOS=linux GOARCH=amd64 $(MAKE) $(BIN_NAME)
	mkdir -p $$(dirname "$@")
	mv $(BIN_NAME) $@

.PHONY: publish/wp-darwin-amd64
publish/wp-darwin-amd64:
	# Force build; don't let existing versions interfere.
	rm -f $(BIN_NAME)
	GOOS=darwin GOARCH=amd64 $(MAKE) $(BIN_NAME)
	mkdir -p $$(dirname "$@")
	mv $(BIN_NAME) $@


.PHONY: install isntall
install isntall: $(BIN_NAME)
	cp $(BIN_NAME) $(INSTALLED_NAME)

.PHONY: test
test: $(AUTOGEN_VERSION_FILENAME) $(BIN_NAME)
	@if [ -z $$T ]; then \
		$(GO) test -v ./...; \
	else \
		$(GO) test -v ./... -run $$T; \
	fi


$(COVERAGE_FILE): $(AUTOGEN_VERSION_FILENAME) $(BIN_NAME)
	$(GO) test -v --coverprofile=$(COVERAGE_FILE) ./...

.PHONY: coverage
coverage: $(COVERAGE_FILE)
	$(GO) tool cover -func=$(COVERAGE_FILE)

.INTERMEDIATE: $(AUTOGEN_VERSION_FILENAME)
$(AUTOGEN_VERSION_FILENAME):
	@version="v$$(cat VERSION)" && \
	build="$$(if [ "$$(git describe)" != "$$version" ]; then echo "-$$(git rev-parse --short HEAD)"; fi)" && \
	dirty="$$(if [ ! -z "$$(git diff)" ]; then echo "-dirty"; fi)" && \
	printf "package blobapi\n\nconst VersionBuild = \"%s%s%s\"" $$version $$build $$dirty > $@

.PHONY: pretty-coverage
pretty-coverage: $(COVERAGE_FILE)
	$(GO) tool cover -html=$(COVERAGE_FILE)

.PHONY: fmt
fmt:
	@$(GO) fmt ./...

.PHONY: clean
clean:
	rm -rf $(COVERAGE_FILE) $(BUILD_DIR)
