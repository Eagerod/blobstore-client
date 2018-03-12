package blobapi;

import (
    "bufio"
    "errors"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "strings"
    "time"
)

type BlobStoreApiClient struct {
    DefaultUrl string
    DefaultReadAcl string
    DefaultWriteAcl string

    http http.Client
}


type IBlobStoreApiClient interface {
    UploadStream(path string, stream *bufio.Reader, contentType string) error
    UploadFile(path string, source string, contentType string) error

    GetFileContents(path string) (string, error)
    DownloadFile(path string, dest string) error
    CatFile(path string) error
}


func NewBlobStoreApiClient(url, readAcl, writeAcl string) *BlobStoreApiClient {
    return &BlobStoreApiClient{
        url,
        readAcl,
        writeAcl,
        http.Client{Timeout: time.Second * 30},
    }
}


func (b *BlobStoreApiClient) route(path string) string {
    defaultClash := strings.HasSuffix(b.DefaultUrl, "/")
    pathClash := strings.HasPrefix(path, "/")

    if defaultClash && pathClash {
        return fmt.Sprintf("%s/%s", b.DefaultUrl[:len(b.DefaultUrl)-1], path[1:])
    } else if !(defaultClash && pathClash) {
        return fmt.Sprintf("%s/%s", b.DefaultUrl, path)
    }

    return fmt.Sprintf("%s%s", b.DefaultUrl, path)
}

func (b *BlobStoreApiClient) UploadStream(path string, stream *bufio.Reader, contentType string) error {
    response, err := b.http.Post(b.route(path), contentType, stream)
    if err != nil {
        return err
    }

    if response.StatusCode != 200 {
        body, err := ioutil.ReadAll(response.Body)
        if err != nil {
           return err
        }

        return errors.New(fmt.Sprintf("Blobstore Upload Failed (%d): %s", response.StatusCode, string(body)))
    }

    return nil
}

func (b *BlobStoreApiClient) UploadFile(path string, source string, contentType string) error {
    file, err := os.Open(source)
    if err != nil {
        return err
    }

    fileReader := bufio.NewReader(file)
    return b.UploadStream(path, fileReader, contentType)
}

func (b *BlobStoreApiClient) GetFileContents(path string) (string, error) {
    response, err := b.http.Get(b.route(path))
    if err != nil {
        return "", err
    }

    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return "", err
    }

    if response.StatusCode != 200 {
        return "", errors.New(fmt.Sprintf("Blobstore Download Failed (%d): %s", response.StatusCode, string(body)))
    }

    return string(body), nil
}

func (b *BlobStoreApiClient) DownloadFile(path, dest string) error {
    str, err := b.GetFileContents(path)
    if err != nil {
        return err
    }

    err = ioutil.WriteFile(dest, []byte(str), 0644)
    return err
}

func (b *BlobStoreApiClient) CatFile(path string) error {
    str, err := b.GetFileContents(path)
    if err != nil {
        return err
    }

    fmt.Println(str)
    return nil
}
