SHELL := $(shell which bash)

ENV_PREFIX := GOPATH=`pwd`:`pwd`/deps
PREFIX := $(ENV_PREFIX)

SOURCES := src/main.go

BIN_ROOT := bin
BIN_FILE := blob
BIN_NAME := $(BIN_ROOT)/$(BIN_FILE)

DEPS_DIR := deps/src
DEPS := $(foreach f,$(shell cat deps.txt),$(DEPS_DIR)/$(f))
DEV_DEPS := $(foreach f,$(shell cat dev_deps.txt),$(DEPS_DIR)/$(f))

UPLOAD_PATH := clientlib

BLOB_LATEST_VERSION := $(shell git tag | sort -n | tail -1 | awk '{print $$1}')

$(BIN_NAME): dependencies
	mkdir -p $(BIN_ROOT)
	$(PREFIX) go build -o $(BIN_NAME) $(SOURCES)

.PHONY: $(DEPS_DIR)/%
$(DEPS_DIR)/%:
	mkdir -p $(DEPS_DIR)
	if [ ! -d "$(DEPS_DIR)/$$(echo $* | awk -F '@' '{print $$1}')" ]; then \
		git clone "https://$$(echo $* | awk -F '@' '{print $$1}')" "$(DEPS_DIR)/$$(echo $* | awk -F '@' '{print $$1}')"; \
	fi
	cd $(DEPS_DIR)/$$(echo $* | awk -F '@' '{print $$1}') && git remote update && git checkout $$(echo $* | awk -F '@' '{print $$2}')

.PHONY: dependencies
dependencies: $(DEPS)

.PHONY: dev_dependencies
dev_dependencies: $(DEV_DEPS)

.PHONY: build-dependencies
build-dependencies:
	if ! type git > /dev/null 2> /dev/null; then \
	    if [ "$$(uname)" -eq "Darwin" ]; then \
	        brew install git; \
	    elif [ "$$(uname)" -eq "Linux" ]; then \
	        apt install git; \
	    fi; \
	fi; \
	if ! type go > /dev/null 2> /dev/null; then \
	    if [ "$$(uname)" -eq "Darwin" ]; then \
	        brew install go; \
	    elif [ "$$(uname)" -eq "Linux" ]; then \
	        apt install golang-go; \
	    fi; \
	fi;

.PHONY: all
all: $(BIN_NAME)

.PHONY: install
install: build-dependencies dependencies $(BIN_NAME)
	cp $(BIN_NAME) /usr/local/bin/$(BIN_FILE)

.PHONY: test
test:
	$(PREFIX) go test -v --coverprofile=coverage.out 'blobapi'

.PHONY: coverage
cover: test
	$(PREFIX) go tool cover -func=coverage.out

.PHONY: pretty-coverage
pretty-coverage: test
	$(PREFIX) go tool cover -html=coverage.out

build/%.zip:
	mkdir -p build
	git archive --format zip $* -o build/$*.zip

build/installer.sh:
	mkdir -p build
	sed 's/^\(BLOB_LATEST_VERSION=\).*/\1"'$(BLOB_LATEST_VERSION)'"/' blober.sh > build/installer.sh

.PHONY: release
release: $(BIN_NAME) build/$(BLOB_LATEST_VERSION).zip build/installer.sh
	if [ -z "$$BLOBSTORE_WRITE_ACL" ]; then \
		echo >&2 "Write ACL not present in environment; aborting release."; \
		exit -1; \
	fi;
	$(BIN_NAME) upload -f "$(UPLOAD_PATH)/$(BLOB_LATEST_VERSION).zip" -t "application/zip" -s "build/$(BLOB_LATEST_VERSION).zip"
	$(BIN_NAME) upload -f "$(UPLOAD_PATH)/installer.sh" -t "text/x-shellscript" -s build/installer.sh

.PHONY: clean
clean:
	rm -rf build || true
	rm -rf $(DEPS_DIR) || true
	rm $(BIN_NAME) || true
