ENV_PREFIX := GOPATH=`pwd`:$$HOME/go
PREFIX := $(ENV_PREFIX)

SOURCES := src/main.go

BIN_ROOT := bin
BIN_NAME := bin/blob

$(BIN_NAME):
	mkdir -p $(BIN_ROOT)
	$(PREFIX) go build -o $(BIN_NAME) $(SOURCES)

all: $(BIN_NAME)

release:
	echo "thoon"

clean:
	rm $(BIN_NAME) || true
