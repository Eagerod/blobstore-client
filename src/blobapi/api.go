package blobapi;

import (
    "bufio"
    "errors"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
    "os"
    "strings"
    "time"
)

type IHttpClient interface {
    Do(*http.Request) (*http.Response, error)
    Get(string) (*http.Response, error)
    Head(string) (*http.Response, error)
    Post(string, string, io.Reader) (*http.Response, error)
    PostForm(string, url.Values) (*http.Response, error)
}

type BlobStoreApiClient struct {
    DefaultUrl string
    DefaultReadAcl string
    DefaultWriteAcl string

    http IHttpClient
}


type IBlobStoreApiClient interface {
    UploadStream(path string, stream *bufio.Reader, contentType string) error
    UploadFile(path string, source string, contentType string) error

    GetFileContents(path string) (string, error)
    DownloadFile(path string, dest string) error
    CatFile(path string) error
}


func NewBlobStoreApiClient(url, readAcl, writeAcl string) *BlobStoreApiClient {
    // Make sure that the base url looks like a path, so that url resolution
    // always uses the full base url as the prefix.
    if !strings.HasSuffix(url, "/") {
        url = url + "/"
    }

    return &BlobStoreApiClient{
        url,
        readAcl,
        writeAcl,
        &http.Client{Timeout: time.Second * 30},
    }
}


func (b *BlobStoreApiClient) route(path string) string {
    // Always remove a / prefix on `path`, since it will resolve itself down to
    // the host, rather than whatever additional pathing we want to add to the
    // BlobStore default URL.
    for strings.HasPrefix(path, "/") {
        path = path[1:]
    }

    pathUrlComponent, err := url.Parse(path)
    if err != nil {
        panic(err)
    }

    baseUrlComponent, err := url.Parse(b.DefaultUrl)
    if err != nil {
        panic(err)
    }

    return baseUrlComponent.ResolveReference(pathUrlComponent).String()
}

func (b *BlobStoreApiClient) UploadStream(path string, stream *bufio.Reader, contentType string) error {
    request, err := http.NewRequest("POST", b.route(path), stream)
    if err != nil {
        return err
    }

    request.Header.Add("Content-Type", contentType)
    request.Header.Add("X-BlobStore-Read-Acl", b.DefaultReadAcl)
    request.Header.Add("X-BlobStore-Write-Acl", b.DefaultWriteAcl)

    response, err := b.http.Do(request)
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
    request, err := http.NewRequest("GET", b.route(path), nil)
    if err != nil {
        return "", err
    }

    request.Header.Add("X-BlobStore-Read-Acl", b.DefaultReadAcl)

    response, err := b.http.Do(request)
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
