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

func main() {
    parser := argparse.NewParser("blob", "Upload and download from blobstore.")

    uploadCommand := parser.NewCommand("upload", "Upload file to blobstore")
    uploadFilename := uploadCommand.String("f", "filename", &argparse.Options{Help: "Name of file uploaded to blobstore", Required: true})
    cType := uploadCommand.String("t", "type", &argparse.Options{Help: "Content type of uploaded file", Required: true})
    source := uploadCommand.String("s", "source", &argparse.Options{Help: "Local file to upload", Required: true})

    downloadCommand := parser.NewCommand("download", "Download file from blobstore")
    downloadFilename := downloadCommand.String("f", "filename", &argparse.Options{Help: "Name of file downloaded from blobstore", Required: true})    
    destination := downloadCommand.String("d", "dest", &argparse.Options{Help: "Local file to write", Required: true})    

    err := parser.Parse(os.Args)
    if err != nil {
        fmt.Println(parser.Usage(err))
        os.Exit(1)
    }

    var b blobapi.IBlobStoreApiClient = blobapi.NewBlobStoreApiClient("https://aleem.haji.ca/blob", "", "")

    switch {
    case uploadCommand.Happened():
        err = b.UploadFile(*uploadFilename, *source, *cType)
    case downloadCommand.Happened():
        err = b.DownloadFile(*downloadFilename, *destination)
    default:
        err = errors.New("Failed to identify command to run")
    }

    if err != nil {
        fmt.Println(parser.Usage(err))
        os.Exit(1)
    }

}
