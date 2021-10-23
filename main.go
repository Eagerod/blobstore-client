package main

import (
	"os"
)

import (
	"gitea.internal.aleemhaji.com/aleem/blobapi/cmd/blobapi"
)

func main() {
	if err := blobapi.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
