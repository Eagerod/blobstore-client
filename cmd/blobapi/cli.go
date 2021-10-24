package blobapi

import (
	"errors"
	"fmt"
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

	var appendString string
	var recursive bool

	baseCommand := &cobra.Command{
		Use:   "blob",
		Short: "Blobstore CLI",
		Long:  "Download, upload or append data to the blobstore",
	}

	appendCommand := &cobra.Command{
		Use:   "append <BlobPath>",
		Short: "Append to blobstore",
		Long:  "Append to an existing file in the blobstore",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if appendString == "" {
				return errors.New("Nothing to append")
			}

			appendArg := newBlobParsedArg(args[0])

			if !appendArg.isRemote {
				return errors.New("Cannot append to local file")
			}

			return b.AppendString(appendArg.path, appendString)
		},
	}

	lsCommand := &cobra.Command{
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

			files, err := b.ListPrefix(prefix, recursive)
			if err != nil {
				return err
			}

			for i := range files {
				fmt.Println(files[i])
			}

			return nil
		},
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

	appendCommand.Flags().StringVarP(&appendString, "string", "s", "", "String to append")
	lsCommand.Flags().BoolVarP(&recursive, "recursive", "r", false, "List all files and folders recursively")

	baseCommand.AddCommand(newCpCommand(b))
	baseCommand.AddCommand(appendCommand)
	baseCommand.AddCommand(lsCommand)
	baseCommand.AddCommand(rmCommand)

	return baseCommand.Execute()
}
