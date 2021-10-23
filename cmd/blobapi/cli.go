package blobapi

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

import (
	"github.com/spf13/cobra"
)

const BlobStoreReadAclEnvironmentVariable = "BLOBSTORE_READ_ACL"
const BlobStoreWriteAclEnvironmentVariable = "BLOBSTORE_WRITE_ACL"
const BlobStoreDefaultUrlBase = "https://blob.internal.aleemhaji.com"

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
	var b IBlobStoreApiClient = NewBlobStoreApiClient(
		BlobStoreDefaultUrlBase,
		DefaultCredentialProviderChain(),
	)

	var contentType string
	var appendString string
	var recursive bool
	var force bool

	baseCommand := &cobra.Command{
		Use:   "blob",
		Short: "Blobstore CLI",
		Long:  "Download, upload or append data to the blobstore",
	}

	cpCommand := &cobra.Command{
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
				return b.CatFile(cpArg.path)
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
				return b.DownloadFile(cpArg0.path, cpArg1.path)
			} else {
				if force == false {
					fileStat, err := b.StatFile(cpArg1.path)
					if err != nil {
						return err
					}
					if fileStat.Exists {
						return errors.New("Destination file already exists on blobstore; use --force to overwrite")
					}
				}
				return b.UploadFile(cpArg1.path, cpArg0.path, contentType)
			}
		},
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

	cpCommand.Flags().StringVarP(&contentType, "type", "t", "", "Content type of uploaded file")
	appendCommand.Flags().StringVarP(&appendString, "string", "s", "", "String to append")
	lsCommand.Flags().BoolVarP(&recursive, "recursive", "r", false, "List all files and folders recursively")
	cpCommand.Flags().BoolVarP(&force, "force", "f", false, "Force the copy if the destination already exists")

	baseCommand.AddCommand(cpCommand)
	baseCommand.AddCommand(appendCommand)
	baseCommand.AddCommand(lsCommand)
	baseCommand.AddCommand(rmCommand)

	return baseCommand.Execute()
}
