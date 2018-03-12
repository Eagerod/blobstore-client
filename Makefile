ENV_PREFIX := GOPATH=`pwd`:$$HOME/go
PREFIX := $(ENV_PREFIX)

SOURCES := src/main.go

BIN_ROOT := bin
BIN_FILE := blob
BIN_NAME := $(BIN_ROOT)/$(BIN_FILE)

UPLOAD_PATH := clientlib

BLOB_LATEST_VERSION := $(shell git tag -ln | tail -1 | awk '{print $$1}')

$(BIN_NAME):
	mkdir -p $(BIN_ROOT)
	$(PREFIX) go build -o $(BIN_NAME) $(SOURCES)

all: $(BIN_NAME)

install: $(BIN_NAME)
	cp $(BIN_NAME) /usr/local/bin/$(BIN_FILE)

build/%.zip:
	mkdir -p build
	git archive --format zip $* -o build/$*.zip

build/installer.sh:
	mkdir -p build
	sed 's/^\(BLOB_LATEST_VERSION=\).*/\1"'$(BLOB_LATEST_VERSION)'"/' blober.sh > build/installer.sh

upload/%.zip: $(BIN_NAME) build/%.zip
	$(BIN_NAME) upload -f "$(UPLOAD_PATH)/$*.zip" -t "application/zip" -s "build/$*.zip"

release: upload/$(BLOB_LATEST_VERSION).zip build/installer.sh
	$(BIN_NAME) upload -f "$(UPLOAD_PATH)/installer.sh" -t "text/x-shellscript" -s build/installer.sh

clean:
	rm -rf build || true
	rm $(BIN_NAME) || true
