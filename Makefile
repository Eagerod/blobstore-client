SHELL := $(shell which bash)

ENV_PREFIX := GOPATH=`pwd`:`pwd`/deps
PREFIX := $(ENV_PREFIX)

SOURCES := src/main.go

BIN_ROOT := bin
BIN_FILE := blob
BIN_NAME := $(BIN_ROOT)/$(BIN_FILE)

DEPS_DIR := deps/src
DEPS := $(foreach f,$(shell cat deps.txt | grep -v "^\s*\#"),$(DEPS_DIR)/$(f))
DEV_DEPS := $(foreach f,$(shell cat dev_deps.txt | grep -v "^\s*\#"),$(DEPS_DIR)/$(f))

UPLOAD_PATH := clientlib

BLOB_LATEST_VERSION := $(shell git tag | sort -n | tail -1 | awk '{print $$1}')

BUILD_DEPS_PREFIX := build_deps

UNAME := $(shell uname -s)
ifeq ($(UNAME),Linux)
	DEP_CHECK := dpkg -s
	SYS_INSTALL := sudo apt install -y
	BUILD_DEPS_P := git golang-go
else ifeq ($(UNAME),Darwin)
	DEP_CHECK := brew ls --versions
	SYS_INSTALL := brew install
	BUILD_DEPS_P := git go
endif

BUILD_DEPS := $(foreach f,$(BUILD_DEPS_P),$(BUILD_DEPS_PREFIX)/$(f))

$(BIN_NAME): dependencies
	mkdir -p $(BIN_ROOT)
	$(PREFIX) go build -o $(BIN_NAME) $(SOURCES)

.PHONY: $(DEPS_DIR)/%
$(DEPS_DIR)/%:
	mkdir -p $(DEPS_DIR)
	$(eval DEP_SRC := $(shell echo $* | awk -F '@' '{print $$1}'))
	$(eval DEP_TAG := $(shell echo $* | awk -F '@' '{print $$2}'))

	if [ ! -d "$(DEPS_DIR)/$(DEP_SRC)" ]; then \
		git clone "https://$(DEP_SRC)" "$(DEPS_DIR)/$(DEP_SRC)"; \
	fi
	if [ "$$(git -C $(DEPS_DIR)/$(DEP_SRC) log --format=%H -1)" != "$(DEP_TAG)" ]; then \
		git -C $(DEPS_DIR)/$(DEP_SRC) remote update; \
		git -C $(DEPS_DIR)/$(DEP_SRC) checkout $(DEP_TAG); \
	fi

.PHONY: dependencies
dependencies: $(DEPS)

.PHONY: dev-dependencies
dev-dependencies: $(DEV_DEPS)

.PHONY: $(BUILD_DEPS_PREFIX)/%
$(BUILD_DEPS_PREFIX)/%:
	if ! $(DEP_CHECK) $*; then \
		$(SYS_INSTALL) $*; \
	fi

.PHONY: build-dependencies
build-dependencies: $(BUILD_DEPS)

.PHONY: all
all: $(BIN_NAME)

.PHONY: install
install: build-dependencies dependencies $(BIN_NAME)
	cp $(BIN_NAME) /usr/local/bin/$(BIN_FILE)

.PHONY: test
test: dev-dependencies dependencies
	$(PREFIX) go test -v 'blobapi'

.PHONY: system-test
system-test: install dev-dependencies
	$(PREFIX) go test -v src/main_test.go 

.PHONY: test-cover
test-cover: 
	$(PREFIX) go test -v --coverprofile=coverage.out 'blobapi'

.PHONY: coverage
cover: test-cover
	$(PREFIX) go tool cover -func=coverage.out

.PHONY: pretty-coverage
pretty-coverage: test-cover
	$(PREFIX) go tool cover -html=coverage.out

build/%.zip:
	mkdir -p build
	git archive --format zip $* -o build/$*.zip

build/installer.sh:
	mkdir -p build
	sed 's/^\(BLOB_LATEST_VERSION=\).*/\1"'$(BLOB_LATEST_VERSION)'"/' blober.sh > build/installer.sh

.PHONY: release
release: $(BIN_NAME) build/$(BLOB_LATEST_VERSION).zip build/installer.sh
	source .env && if [ -z "$$BLOBSTORE_WRITE_ACL" ]; then \
		echo >&2 "Write ACL not present in environment; aborting release."; \
		exit -1; \
	fi;
	source .env && $(BIN_NAME) cp "build/$(BLOB_LATEST_VERSION).zip" "blob:/$(UPLOAD_PATH)/$(BLOB_LATEST_VERSION).zip"
	source .env && $(BIN_NAME) cp -f "build/installer.sh" "blob:/$(UPLOAD_PATH)/installer.sh"

.PHONY: clean
clean:
	rm coverage.out || true
	rm -rf build || true
	rm -rf $(DEPS_DIR) || true
	rm $(BIN_NAME) || true
