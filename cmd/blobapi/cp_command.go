package blobapi

import (
	"github.com/spf13/cobra"
)

import (
	"gitea.internal.aleemhaji.com/aleem/blobstore-cli/pkg/blob"
)


func newCpCommand(client blob.IBlobStoreClient) *cobra.Command {
	var contentType string
	var force bool

	command := &cobra.Command{
		Use:   "cp <LocalPath> <BlobPath> or <BlobPath> <LocalPath>",
		Short: "Copy files to and from blobstore",
		Long:  "Upload files to or download files from the blobstore",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cpArg0, err := newBlobParsedArg(args[0])
			if err != nil {
				return err
			}

			if len(args) == 1 {
				return client.Cat(cpArg0)
			}

			cpArg1, err := newBlobParsedArg(args[1])
			if err != nil {
				return err
			}

			return client.Copy(cpArg0, cpArg1, force)
		},
	}

	command.Flags().StringVarP(&contentType, "type", "t", "", "Content type of uploaded file")
	command.Flags().BoolVarP(&force, "force", "f", false, "Force the copy if the destination already exists")

	return command
}