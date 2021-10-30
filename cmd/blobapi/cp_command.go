package blobapi

import (
	"errors"
	"os"
)

import (
	"github.com/spf13/cobra"
)

import (
	"gitea.internal.aleemhaji.com/aleem/blobapi/pkg/blob"
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
				if cpArg0.Scheme != BlobStoreUrlScheme {
					return errors.New("Must download files from blob:/")
				}
				return client.CatFile(cpArg0.Path)
			}

			// Determine if this is an upload or download command based on which
			// order the parameters came in.
			cpArg1, err := newBlobParsedArg(args[1])
			if err != nil {
				return err
			}

			if cpArg0.Scheme == BlobStoreUrlScheme && cpArg1.Scheme == BlobStoreUrlScheme {
				return errors.New("No support for copying files in the blobstore directly")
			}

			if cpArg0.Scheme != BlobStoreUrlScheme && cpArg1.Scheme != BlobStoreUrlScheme {
				return errors.New("Must provide at least one blob:/ path to upload to or download from")
			}

			if cpArg0.Scheme == BlobStoreUrlScheme {
				if force == false {
					if _, err := os.Stat(cpArg1.Path); err == nil {
						return errors.New("Destination file already exists on local machine; use --force to overwrite")
					}
				}
				return client.DownloadFile(cpArg0.Path, cpArg1.Path)
			} else {
				if force == false {
					fileStat, err := client.StatFile(cpArg1.Path)
					if err != nil {
						return err
					}
					if fileStat.Exists {
						return errors.New("Destination file already exists on blobstore; use --force to overwrite")
					}
				}
				return client.UploadFile(cpArg1.Path, cpArg0.Path, contentType)
			}
		},
	}

	command.Flags().StringVarP(&contentType, "type", "t", "", "Content type of uploaded file")
	command.Flags().BoolVarP(&force, "force", "f", false, "Force the copy if the destination already exists")

	return command
}