package blob

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
	"path/filepath"
	"strconv"
	"strings"
)

import (
	"gitea.internal.aleemhaji.com/aleem/blobapi/pkg/credential_provider"
)

type BlobStoreClient struct {
	DefaultUrl         string
	CredentialProvider credential_provider.ICredentialProvider

	http IHttpClient
	apiClient IBlobStoreApiClient
}

type IBlobStoreClient interface {
	// UploadStream(path string, stream *bufio.Reader, contentType string) error
	UploadFile(path string, source string, contentType string) error

	GetFileReadStream(path string) (*io.Reader, error)
	GetFileContents(path string) (string, error)
	DownloadFile(path string, dest string) error
	CatFile(path string) error

	StatFile(path string) (*BlobFileStat, error)

	AppendStream(path string, stream *bufio.Reader) error
	AppendString(path string, value string) error
	AppendFile(path string, source string) error

	ListPrefix(prefix string, recursive bool) ([]string, error)

	DeleteFile(path string) error
}

func NewBlobStoreClient(url string, credentialProvider credential_provider.ICredentialProvider) *BlobStoreClient {
	// Make sure that the base url looks like a path, so that url resolution
	// always uses the full base url as the prefix.
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}

	apiClient := NewBlobStoreApiClient(url, credentialProvider)

	return &BlobStoreClient{
		url,
		credentialProvider,
		apiClient.http,
		apiClient,
	}
}

func NewBlobStoreHttpError(operation string, response *http.Response) error {
	if response.Body == nil {
		return errors.New(fmt.Sprintf("Blobstore %s Failed (%d)", operation, response.StatusCode))
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return errors.New(fmt.Sprintf("Blobstore %s Failed (%d): %s", operation, response.StatusCode, string(body)))
}

func (b *BlobStoreClient) route(path string) string {
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

func (b *BlobStoreClient) NewAuthorizedRequest(method, path string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, b.route(path), body)
	if err != nil {
		return request, err
	}

	err = b.CredentialProvider.AuthorizeRequest(request)
	return request, err
}

// func (b *BlobStoreClient) UploadStream(path string, stream *bufio.Reader, contentType string) error {
// 	return b.apiClient.UploadStream(path, stream, contentType)
// }

func (b *BlobStoreClient) UploadFile(path string, source string, contentType string) error {
	file, err := os.Open(source)
	defer file.Close()

	if err != nil {
		return err
	}

	fileReader := bufio.NewReader(file)
	return b.apiClient.UploadStream(path, fileReader, contentType)
}

type getFileReadStreamResponse struct {
	reader      *io.Reader
	contentType string
}

func (b *BlobStoreClient) getFileReadStream(path string) (*getFileReadStreamResponse, error) {
	request, err := b.NewAuthorizedRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	response, err := b.http.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, NewBlobStoreHttpError("Download", response)
	}

	body := response.Body.(io.Reader)
	r := getFileReadStreamResponse{&body, response.Header.Get("Content-Type")}

	return &r, nil
}

func (b *BlobStoreClient) GetFileReadStream(path string) (*io.Reader, error) {
	response, err := b.getFileReadStream(path)
	if err != nil {
		return nil, err
	}
	return response.reader, err
}

func (b *BlobStoreClient) GetFileContents(path string) (string, error) {
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

func (b *BlobStoreClient) DownloadFile(path, dest string) error {
	str, err := b.GetFileContents(path)
	if err != nil {
		return err
	}

	destDirectory := filepath.Dir(dest)
	err = os.MkdirAll(destDirectory, 0755)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dest, []byte(str), 0644)
	return err
}

func (b *BlobStoreClient) CatFile(path string) error {
	str, err := b.GetFileContents(path)
	if err != nil {
		return err
	}

	fmt.Println(str)
	return nil
}

func (b *BlobStoreClient) StatFile(path string) (*BlobFileStat, error) {
	request, err := b.NewAuthorizedRequest("HEAD", path, nil)
	if err != nil {
		return nil, err
	}

	response, err := b.http.Do(request)
	if err != nil {
		return nil, err
	}

	rv := BlobFileStat{
		MimeType: response.Header.Get("Content-Type"),
		Exists:   true,
	}
	finalSlash := strings.LastIndex(path, "/")
	if finalSlash == -1 {
		rv.Path = ""
		rv.Name = path
	} else {
		rv.Path = path[0 : finalSlash+1]
		rv.Name = path[finalSlash+1:]
	}

	size, err := strconv.Atoi(response.Header.Get("Content-Length"))
	if err == nil {
		rv.SizeBytes = size
	}

	if response.StatusCode == 404 {
		rv.Exists = false
	} else if response.StatusCode != 200 {
		return nil, NewBlobStoreHttpError("Stat", response)
	}

	return &rv, nil
}

func (b *BlobStoreClient) AppendStream(path string, stream *bufio.Reader) error {
	getFileReadStreamResponse, err := b.getFileReadStream(path)
	if err != nil {
		return err
	}

	multiStream := bufio.NewReader(io.MultiReader(*getFileReadStreamResponse.reader, stream))
	return b.apiClient.UploadStream(path, multiStream, getFileReadStreamResponse.contentType)
}

func (b *BlobStoreClient) AppendString(path string, value string) error {
	stringReader := bufio.NewReader(strings.NewReader(value))
	return b.AppendStream(path, stringReader)
}

func (b *BlobStoreClient) AppendFile(path string, source string) error {
	file, err := os.Open(source)
	defer file.Close()

	if err != nil {
		return err
	}

	fileReader := bufio.NewReader(file)
	return b.AppendStream(path, fileReader)
}

func (b *BlobStoreClient) ListPrefix(prefix string, recursive bool) ([]string, error) {
	paths := make([]string, 0)

	for strings.HasPrefix(prefix, "/") {
		prefix = prefix[1:]
	}

	requestUrl := b.route("_dir/" + prefix)
	if recursive {
		requestUrl += "?recursive=true"
	}

	request, err := b.NewAuthorizedRequest("GET", requestUrl, nil)
	if err != nil {
		return paths, err
	}

	response, err := b.http.Do(request)
	if err != nil {
		return paths, err
	}

	if response.StatusCode != 200 {
		return paths, NewBlobStoreHttpError("List", response)
	}

	err = json.NewDecoder(response.Body).Decode(&paths)
	if err != nil {
		return paths, err
	}

	return paths, nil
}

func (b *BlobStoreClient) DeleteFile(path string) error {
	request, err := b.NewAuthorizedRequest("DELETE", path, nil)
	if err != nil {
		return err
	}

	response, err := b.http.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return NewBlobStoreHttpError("Delete", response)
	}

	return nil
}
