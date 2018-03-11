package main;

import (
    "flag"
    "fmt"
)

import (
    "blobapi"
)


func main() {
    var b blobapi.IBlobStoreApiClient = blobapi.NewBlobStoreApiClient("https://aleem.haji.ca/blob", "", "")

    str, err := b.GetFileContents("house.txt")
    if err != nil {
        panic(err)
    }

    fmt.Println(str)
    help := flag.Bool("help", false, "Print usage")

    if *help == true {
        flag.Usage()
    }
}
