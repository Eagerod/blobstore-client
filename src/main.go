package main;

import (
    "errors"
    "os"
)

import (
    "github.com/spf13/cobra"
)

import (
    "blobapi"
)

const BlobStoreReadAclEnvironmentVariable = "BLOBSTORE_READ_ACL"
const BlobStoreWriteAclEnvironmentVariable = "BLOBSTORE_WRITE_ACL"
const BlobStoreDefaultUrlBase = "https://aleem.haji.ca/blob"

var emptyAcl string = ""
var readAcl *string = &emptyAcl
var writeAcl *string = &emptyAcl

func init() {    
    if acl, ok := os.LookupEnv(BlobStoreReadAclEnvironmentVariable); ok {
        readAcl = &acl
    }

    if acl, ok := os.LookupEnv(BlobStoreWriteAclEnvironmentVariable); ok {
        writeAcl = &acl
    }
}

func main() {
    var b blobapi.IBlobStoreApiClient = blobapi.NewBlobStoreApiClient(BlobStoreDefaultUrlBase, *readAcl, *writeAcl)

    var contentType string
    var appendString string

    baseCommand := &cobra.Command{
        Use: "blob",
        Short: "Blobstore CLI",
        Long: "Download, upload or append data to the blobstore",
        SilenceUsage: false,
        Run: func(cmd *cobra.Command, args []string) {
            cmd.Usage()
        },
    }

    uploadCommand := &cobra.Command{
        Use: "upload",
        Short: "Upload to blobstore",
        Long: "Upload a file to blobstore from the local machine",
        Args: cobra.ExactArgs(2),
        RunE: func(cmd *cobra.Command, args []string) error {
            uploadFilename := args[0]
            source := args[1]
            return b.UploadFile(uploadFilename, source, contentType)
        },
    }

    downloadCommand := &cobra.Command{
        Use: "download",
        Short: "Download from blobstore",
        Long: "Download a file from blobstore to the local machine",
        Args: cobra.RangeArgs(1, 2),
        RunE: func(cmd *cobra.Command, args []string) error {
            downloadFilename := args[0]
            dest := ""

            if len(args) == 2 {
                dest = args[1]                
            }

            if dest == "" {
                return b.CatFile(downloadFilename)
            } else {
                return b.DownloadFile(downloadFilename, dest)
            }
        },
    }

    appendCommand := &cobra.Command{
        Use: "append",
        Short: "Append to blobstore",
        Long: "Append to an existing file in the blobstore",
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            if appendString == "" {
                return errors.New("Nothing to append")
            }

            sourceFilename := args[0]
            return b.AppendString(sourceFilename, appendString)
        },
    }

    uploadCommand.Flags().StringVarP(&contentType, "type", "t", "", "Content type of uploaded file")
    appendCommand.Flags().StringVarP(&appendString, "string", "s", "", "String to append")

    baseCommand.AddCommand(uploadCommand)
    baseCommand.AddCommand(downloadCommand)
    baseCommand.AddCommand(appendCommand)

    if err := baseCommand.Execute(); err != nil {
        os.Exit(1)
    }
    os.Exit(0)
}
