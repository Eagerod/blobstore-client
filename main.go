package main

import (
	"os"
)

import (
	"gitea.internal.aleemhaji.com/aleem/blobstore-cli/cmd/blobapi"
)

func main() {
	if err := blobapi.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
