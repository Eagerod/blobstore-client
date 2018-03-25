package main;

import (
    "errors"
    "os"
    "strings"
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

    cpCommand := &cobra.Command{
        Use: "cp",
        Short: "Copy files to and from blobstore",
        Long: "Upload files to or download files from the blobstore",
        Args: cobra.RangeArgs(1, 2),
        RunE: func(cmd *cobra.Command, args []string) error {
            if len(args) == 1 {
                if !strings.HasPrefix(args[0], "blob:/") {
                    return errors.New("Must download files from blob:/")
                }
                actualDownloadPath := strings.Replace(args[0], "blob:/", "", 1)
                return b.CatFile(actualDownloadPath)
            }

            // Determine if this is an upload or download command based on which 
            // order the parameters came in.
            if strings.HasPrefix(args[0], "blob:/") {
                if strings.HasPrefix(args[1], "blob:/") {
                    return errors.New("No support for copying files in the blobstore directly")
                }

                // Download to local file
                actualDownloadPath := strings.Replace(args[0], "blob:/", "", 1)
                return b.DownloadFile(actualDownloadPath, args[1])
            } else if strings.HasPrefix(args[1], "blob:/") {
                actualUploadPath := strings.Replace(args[1], "blob:/", "", 1)
                return b.UploadFile(actualUploadPath, args[0], contentType)
            } else {
                return errors.New("Must provide at least one blob:/ path to upload to or download from")
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

            if !strings.HasPrefix(args[0], "blob:/") {
                return errors.New("Cannot append to local file")
            }

            sourceFilename := strings.Replace(args[0], "blob:/", "", 1)
            return b.AppendString(sourceFilename, appendString)
        },
    }

    cpCommand.Flags().StringVarP(&contentType, "type", "t", "", "Content type of uploaded file")
    appendCommand.Flags().StringVarP(&appendString, "string", "s", "", "String to append")

    baseCommand.AddCommand(cpCommand)
    baseCommand.AddCommand(appendCommand)

    if err := baseCommand.Execute(); err != nil {
        os.Exit(1)
    }
    os.Exit(0)
}
