ENV_PREFIX := GOPATH=`pwd`:$$HOME/go
PREFIX := $(ENV_PREFIX)

SOURCES := src/main.go

BIN_ROOT := bin
BIN_FILE := blob
BIN_NAME := $(BIN_ROOT)/$(BIN_FILE)

UPLOAD_PATH := clientlib

$(BIN_NAME):
	mkdir -p $(BIN_ROOT)
	$(PREFIX) go build -o $(BIN_NAME) $(SOURCES)

all: $(BIN_NAME)

install: $(BIN_NAME)
	cp $(BIN_NAME) /usr/local/bin/$(BIN_FILE)

build/%.zip:
	mkdir -p build
	git archive --format zip $* -o build/$*.zip

upload/%.zip: $(BIN_NAME) build/%.zip
	$(BIN_NAME) upload -f "$(UPLOAD_PATH)/$*.zip" -t "application/zip" -s "build/$*.zip"

release:
	echo "thoon"

clean:
	rm $(BIN_NAME) || true
