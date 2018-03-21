package main;

import (
    "errors"
    "fmt"
    "os"
)

import (
    "github.com/akamensky/argparse"
)

import (
    "blobapi"
)

const BlobStoreReadAclEnvironmentVariable = "BLOBSTORE_READ_ACL"
const BlobStoreWriteAclEnvironmentVariable = "BLOBSTORE_WRITE_ACL"
const BlobStoreDefaultUrlBase = "https://aleem.haji.ca/blob"

func main() {
    parser := argparse.NewParser("blob", "Upload and download from blobstore.")

    uploadCommand := parser.NewCommand("upload", "Upload file to blobstore")
    uploadFilename := uploadCommand.String("f", "filename", &argparse.Options{Help: "Name of file uploaded to blobstore", Required: true})
    cType := uploadCommand.String("t", "type", &argparse.Options{Help: "Content type of uploaded file", Required: true})
    source := uploadCommand.String("s", "source", &argparse.Options{Help: "Local file to upload", Required: true})

    downloadCommand := parser.NewCommand("download", "Download file from blobstore")
    downloadFilename := downloadCommand.String("f", "filename", &argparse.Options{Help: "Name of file downloaded from blobstore", Required: true})    
    destination := downloadCommand.String("d", "dest", &argparse.Options{Help: "Local file to write", Required: true})    

    appendCommand := parser.NewCommand("append", "Append to an existing file on blobstore")
    appendFilename := appendCommand.String("f", "filename", &argparse.Options{Help: "Name of file on blobstore", Required: true})    
    appendString := appendCommand.String("s", "string", &argparse.Options{Help: "String to append to existing file", Required: true})    

    err := parser.Parse(os.Args)
    if err != nil {
        fmt.Println(parser.Usage(err))
        os.Exit(1)
    }

    readAcl, ok := os.LookupEnv(BlobStoreReadAclEnvironmentVariable)
    if !ok {
        readAcl = ""
    }

    writeAcl, ok := os.LookupEnv(BlobStoreWriteAclEnvironmentVariable)
    if !ok {
        writeAcl = ""
    }

    var b blobapi.IBlobStoreApiClient = blobapi.NewBlobStoreApiClient(BlobStoreDefaultUrlBase, readAcl, writeAcl)

    switch {
    case uploadCommand.Happened():
        err = b.UploadFile(*uploadFilename, *source, *cType)
    case downloadCommand.Happened():
        err = b.DownloadFile(*downloadFilename, *destination)
    case appendCommand.Happened():
        err = b.AppendString(*appendFilename, *appendString)
    default:
        err = errors.New("Failed to identify command to run")
    }

    if err != nil {
        fmt.Println(parser.Usage(err))
        os.Exit(1)
    }
}
