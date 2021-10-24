package blobapi

import (
	"errors"
	"strings"
)

import (
	"github.com/spf13/cobra"
)

import (
	"gitea.internal.aleemhaji.com/aleem/blobapi/pkg/blob"
	"gitea.internal.aleemhaji.com/aleem/blobapi/pkg/credential_provider"
)

const BlobStoreDefaultUrlBase = "https://blob.aleemhaji.com"

type blobParsedArg struct {
	isRemote bool
	path     string
}

func newBlobParsedArg(arg string) *blobParsedArg {
	rv := blobParsedArg{false, arg}

	if strings.HasPrefix(arg, "blob:/") {
		rv.isRemote = true
		rv.path = strings.Replace(rv.path, "blob:/", "", 1)
	}

	return &rv
}

func Execute() error {
	var b blob.IBlobStoreApiClient = blob.NewBlobStoreApiClient(
		BlobStoreDefaultUrlBase,
		credential_provider.DefaultCredentialProviderChain(),
	)


	baseCommand := &cobra.Command{
		Use:   "blob",
		Short: "Blobstore CLI",
		Long:  "Download, upload or append data to the blobstore",
	}

	rmCommand := &cobra.Command{
		Use:   "rm <BlobPath>",
		Short: "Remove from blobstore",
		Long:  "Delete a file from the blobstore",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rmArg := newBlobParsedArg(args[0])

			if !rmArg.isRemote {
				return errors.New("Cannot delete a local file")
			}

			return b.DeleteFile(rmArg.path)
		},
	}


	baseCommand.AddCommand(newCpCommand(b))
	baseCommand.AddCommand(newAppendCommand(b))
	baseCommand.AddCommand(newLsCommand(b))
	baseCommand.AddCommand(rmCommand)

	return baseCommand.Execute()
}
