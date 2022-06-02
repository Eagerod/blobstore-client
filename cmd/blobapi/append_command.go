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

func newAppendCommand(client blob.IBlobStoreClient) *cobra.Command {
	var appendString string

	command := &cobra.Command{
		Use:   "append <BlobPath>",
		Short: "Append to blobstore",
		Long:  "Append to an existing file in the blobstore",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if appendString == "" {
				return errors.New("Nothing to append")
			}

			appendArg, err := newBlobParsedArg(args[0])
			if err != nil {
				return err
			}

			if appendArg.Scheme != BlobStoreUrlScheme {
				return errors.New("Cannot append to local file")
			}

			return client.AppendString(appendArg, appendString)
		},
	}

	command.Flags().StringVarP(&appendString, "string", "s", "", "String to append")

	return command
}