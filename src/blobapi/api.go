package blobapi;

import (
    "bufio"
    "encoding/json"
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

    GetFileReadStream(path string) (*io.Reader, error)
    GetFileContents(path string) (string, error)
    DownloadFile(path string, dest string) error
    CatFile(path string) error

    AppendStream(path string, stream *bufio.Reader) error
    AppendString(path string, value string) error
    AppendFile(path string, source string) error

    ListPrefix(prefix string) ([]string, error)
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

    if contentType == "" {
        buffer, err := stream.Peek(512)
        if err != nil && err != io.EOF {
            return err
        }

        contentType = http.DetectContentType(buffer)
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

type getFileReadStreamResponse struct {
    reader *io.Reader
    contentType string
}

func (b *BlobStoreApiClient) getFileReadStream(path string) (*getFileReadStreamResponse, error) {
    request, err := http.NewRequest("GET", b.route(path), nil)
    if err != nil {
        return nil, err
    }

    request.Header.Add("X-BlobStore-Read-Acl", b.DefaultReadAcl)

    response, err := b.http.Do(request)
    if err != nil {
        return nil, err
    }

    if response.StatusCode != 200 {
        bodyBytes, err := ioutil.ReadAll(response.Body)
        if err != nil {
            return nil, err
        }

        return nil, errors.New(fmt.Sprintf("Blobstore Download Failed (%d): %s", response.StatusCode, string(bodyBytes)))
    }

    body := response.Body.(io.Reader)
    r := getFileReadStreamResponse{&body, response.Header.Get("Content-Type")}

    return &r, nil
}

func (b *BlobStoreApiClient) GetFileReadStream(path string) (*io.Reader, error) {
    response, err := b.getFileReadStream(path)
    if err != nil {
        return nil, err
    }
    return response.reader, err
}

func (b *BlobStoreApiClient) GetFileContents(path string) (string, error) {
    body, err := b.GetFileReadStream(path)
    if err != nil {
        return "", err
    }

    bodyBytes, err := ioutil.ReadAll(*body)
    if err != nil {
        return "", err
    }

    return string(bodyBytes), nil
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

func (b *BlobStoreApiClient) AppendStream(path string, stream *bufio.Reader) error {
    getFileReadStreamResponse, err := b.getFileReadStream(path)
    if err != nil {
        return err
    }

    multiStream := bufio.NewReader(io.MultiReader(*getFileReadStreamResponse.reader, stream))
    return b.UploadStream(path, multiStream, getFileReadStreamResponse.contentType)
}

func (b *BlobStoreApiClient) AppendString(path string, value string) error {
    stringReader := bufio.NewReader(strings.NewReader(value))
    return b.AppendStream(path, stringReader)
}

func (b *BlobStoreApiClient) AppendFile(path string, source string) error {
    file, err := os.Open(source)
    if err != nil {
        return err
    }

    fileReader := bufio.NewReader(file)
    return b.AppendStream(path, fileReader)
}

func (b *BlobStoreApiClient) ListPrefix(prefix string) ([]string, error) {
    paths := make([]string, 0)

    for strings.HasPrefix(prefix, "/") {
        prefix = prefix[1:]
    }

    request, err := http.NewRequest("GET", b.route("_dir/" + prefix), nil)
    if err != nil {
        return paths, err
    }

    request.Header.Add("X-BlobStore-Read-Acl", b.DefaultReadAcl)

    response, err := b.http.Do(request)
    if err != nil {
        return paths, err
    }

    if response.StatusCode != 200 {
        body, err := ioutil.ReadAll(response.Body)
        if err != nil {
            return paths, err
        }

        return paths, errors.New(fmt.Sprintf("Blobstore List Failed (%d): %s", response.StatusCode, string(body)))
    }

    err = json.NewDecoder(response.Body).Decode(&paths)
    if err != nil {
        return paths, err
    }

    return paths, nil
}
