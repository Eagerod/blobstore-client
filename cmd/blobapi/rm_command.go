package blobapi

import (
	"errors"
)

import (
	"github.com/spf13/cobra"
)

import (
	"github.com/Eagerod/blobstore-client/pkg/blob"
)

func newRmCommand(client blob.IBlobStoreClient) *cobra.Command {
	command := &cobra.Command{
		Use:   "rm <BlobPath>",
		Short: "Remove from blobstore",
		Long:  "Delete a file from the blobstore",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rmArg, err := newBlobParsedArg(args[0])
			if err != nil {
				return err
			}

			if rmArg.Scheme != BlobStoreUrlScheme {
				return errors.New("Cannot delete a local file")
			}

			return client.DeleteFile(rmArg)
		},
	}

	return command
}