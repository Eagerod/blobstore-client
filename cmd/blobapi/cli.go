package blobapi

import (
	"net/url"
)

import (
	"github.com/spf13/cobra"
)

import (
	"github.com/Eagerod/blobstore-client/pkg/blob"
	"github.com/Eagerod/blobstore-client/pkg/credential_provider"
)

const BlobStoreDefaultUrlBase = "https://blob.aleemhaji.com"

type blobParsedArg struct {
	isRemote bool
	path     string
}

const BlobStoreUrlScheme string = "blob"

func newBlobParsedArg(arg string) (*url.URL, error) {
	return url.Parse(arg)
}

func Execute() error {
	var b blob.IBlobStoreClient = blob.NewBlobStoreClient(
		BlobStoreDefaultUrlBase,
		credential_provider.DefaultCredentialProviderChain(),
	)

	baseCommand := &cobra.Command{
		Use:   "blob",
		Short: "Blobstore CLI",
		Long:  "Download, upload or append data to the blobstore",
	}

	baseCommand.AddCommand(newCpCommand(b))
	baseCommand.AddCommand(newAppendCommand(b))
	baseCommand.AddCommand(newLsCommand(b))
	baseCommand.AddCommand(newRmCommand(b))

	return baseCommand.Execute()
}
