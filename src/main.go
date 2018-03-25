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

type blobParsedArg struct {
    isRemote bool
    path string
}

func newBlobParsedArg(arg string) *blobParsedArg {
    rv := blobParsedArg{false, arg}

    if strings.HasPrefix(arg, "blob:/") {
        rv.isRemote = true
        rv.path = strings.Replace(rv.path, "blob:/", "", 1)
    }

    return &rv
}

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
                return b.DownloadFile(cpArg0.path, cpArg1.path)
            } else {
                return b.UploadFile(cpArg1.path, cpArg0.path, contentType)
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

            appendArg := newBlobParsedArg(args[0])

            if !appendArg.isRemote {
                return errors.New("Cannot append to local file")
            }

            return b.AppendString(appendArg.path, appendString)
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
