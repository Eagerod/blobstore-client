ENV_PREFIX := GOPATH=`pwd`
PREFIX := $(ENV_PREFIX)
SOURCES := main.go
BIN_NAME := blob

$(BIN_NAME):
	$(PREFIX) go build $(SOURCE) -o $(BIN_NAME)

all: $(BIN_NAME)

release:
	echo "thoon"

clean:
	rm $(BIN_NAME)
