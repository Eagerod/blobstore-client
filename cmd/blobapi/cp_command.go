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
			if len(args) == 1 {
				cpArg := newBlobParsedArg(args[0])
				if !cpArg.isRemote {
					return errors.New("Must download files from blob:/")
				}
				return client.CatFile(cpArg.path)
			}

			// Determine if this is an upload or download command based on which
			// order the parameters came in.
			cpArg0 := newBlobParsedArg(args[0])
			cpArg1 := newBlobParsedArg(args[1])

			if cpArg0.isRemote && cpArg1.isRemote {
				return errors.New("No support for copying files in the blobstore directly")
			}

			if !cpArg0.isRemote && !cpArg1.isRemote {
				return errors.New("Must provide at least one blob:/ path to upload to or download from")
			}

			if cpArg0.isRemote {
				if force == false {
					if _, err := os.Stat(cpArg1.path); err == nil {
						return errors.New("Destination file already exists on local machine; use --force to overwrite")
					}
				}
				return client.DownloadFile(cpArg0.path, cpArg1.path)
			} else {
				if force == false {
					fileStat, err := client.StatFile(cpArg1.path)
					if err != nil {
						return err
					}
					if fileStat.Exists {
						return errors.New("Destination file already exists on blobstore; use --force to overwrite")
					}
				}
				return client.UploadFile(cpArg1.path, cpArg0.path, contentType)
			}
		},
	}

	command.Flags().StringVarP(&contentType, "type", "t", "", "Content type of uploaded file")
	command.Flags().BoolVarP(&force, "force", "f", false, "Force the copy if the destination already exists")

	return command
}