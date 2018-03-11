ENV_PREFIX := GOPATH=`pwd`
PREFIX := $(ENV_PREFIX)
SOURCES := main.go
BIN_NAME := blob

all:
	$(PREFIX) go build $(SOURCE) -o $(BIN_NAME)

go:
	./$(BIN_NAME)

release:
	echo "thoon"