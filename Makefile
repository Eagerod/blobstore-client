ENV_PREFIX := GOPATH=`pwd`:`pwd`/deps
PREFIX := $(ENV_PREFIX)

SOURCES := src/main.go

BIN_ROOT := bin
BIN_FILE := blob
BIN_NAME := $(BIN_ROOT)/$(BIN_FILE)

DEPS_DIR := deps/src
DEPS := $(foreach f,$(shell cat deps.txt),$(DEPS_DIR)/$(f))

UPLOAD_PATH := clientlib

BLOB_LATEST_VERSION := $(shell git tag -ln | tail -1 | awk '{print $$1}')

$(BIN_NAME):
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

.PHONY: all
all: $(BIN_NAME)

.PHONY: install
install: dependencies $(BIN_NAME)
	cp $(BIN_NAME) /usr/local/bin/$(BIN_FILE)

build/%.zip:
	mkdir -p build
	git archive --format zip $* -o build/$*.zip

build/installer.sh:
	mkdir -p build
	sed 's/^\(BLOB_LATEST_VERSION=\).*/\1"'$(BLOB_LATEST_VERSION)'"/' blober.sh > build/installer.sh

upload/%.zip: $(BIN_NAME) build/%.zip
	$(BIN_NAME) upload -f "$(UPLOAD_PATH)/$*.zip" -t "application/zip" -s "build/$*.zip"

.PHONY: release
release: upload/$(BLOB_LATEST_VERSION).zip build/installer.sh
	$(BIN_NAME) upload -f "$(UPLOAD_PATH)/installer.sh" -t "text/x-shellscript" -s build/installer.sh

.PHONY: clean
clean:
	rm -rf build || true
	rm -rf $(DEPS_DIR) || true
	rm $(BIN_NAME) || true
