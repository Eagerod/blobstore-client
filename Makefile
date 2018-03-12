ENV_PREFIX := GOPATH=`pwd`:$$HOME/go
PREFIX := $(ENV_PREFIX)

SOURCES := src/main.go

BIN_ROOT := bin
BIN_NAME := bin/blob

UPLOAD_PATH := clientlib

$(BIN_NAME):
	mkdir -p $(BIN_ROOT)
	$(PREFIX) go build -o $(BIN_NAME) $(SOURCES)

all: $(BIN_NAME)

build/%.zip:
	mkdir -p build
	git archive --format zip $* -o build/$*.zip

upload/%.zip: $(BIN_NAME) build/%.zip
	$(BIN_NAME) upload -f "$(UPLOAD_PATH)/$*.zip" -t "application/zip" -s "build/$*.zip"

release:
	echo "thoon"

clean:
	rm $(BIN_NAME) || true
