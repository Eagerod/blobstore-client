package main

import (
	"os"
)

import (
	"github.com/Eagerod/blobstore-client/cmd/blobapi"
)

func main() {
	if err := blobapi.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
