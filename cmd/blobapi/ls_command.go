package blobapi

import (
	"errors"
	"fmt"
)

import (
	"github.com/spf13/cobra"
)

import (
	"gitea.internal.aleemhaji.com/aleem/blobapi/pkg/blob"
)

func newLsCommand(client blob.IBlobStoreClient) *cobra.Command {
	var recursive bool

	command := &cobra.Command{
		Use:   "ls [BlobPath]",
		Short: "List files on blobstore",
		Long:  "List existing files in the blobstore",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := ""

			if len(args) == 1 {
				lsArg := newBlobParsedArg(args[0])
				if !lsArg.isRemote {
					return errors.New("Must start remote ls path with blob:/")
				}

				prefix = lsArg.path
			}

			files, err := client.ListPrefix(prefix, recursive)
			if err != nil {
				return err
			}

			for i := range files {
				fmt.Println(files[i])
			}

			return nil
		},
	}

	command.Flags().BoolVarP(&recursive, "recursive", "r", false, "List all files and folders recursively")

	return command
}