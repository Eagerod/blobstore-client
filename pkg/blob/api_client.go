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
	"strconv"
	"strings"
	"time"
)

import (
	"gitea.internal.aleemhaji.com/aleem/blobapi/pkg/credential_provider"
)

type BlobFileStat struct {
	Path      string
	Name      string
	MimeType  string
	SizeBytes int
	Exists    bool
}

type BlobFile struct {
	info 	 BlobFileStat
	contents *io.Reader
}

type IHttpClient interface {
	Do(*http.Request) (*http.Response, error)
	Get(string) (*http.Response, error)
	Head(string) (*http.Response, error)
	Post(string, string, io.Reader) (*http.Response, error)
	PostForm(string, url.Values) (*http.Response, error)
}

type IBlobStoreApiClient interface {
	UploadStream(path string, stream *bufio.Reader, contentType string) error

	GetStat(path string) (*BlobFileStat, error)
	GetFile(path string) (*BlobFile, error)

	ListPrefix(prefix string, recursive bool) ([]string, error)

	DeleteFile(path string) error
}

type BlobStoreApiClient struct {
	baseUrl string
	credentialProvider credential_provider.ICredentialProvider

	http IHttpClient
}

func NewBlobStoreApiClient(baseUrl string, credentialProvider credential_provider.ICredentialProvider) *BlobStoreApiClient {
	// Make sure that the base url looks like a path, so that url resolution
	// always uses the full base url as the prefix.
	if !strings.HasSuffix(baseUrl, "/") {
		baseUrl = baseUrl + "/"
	}

	return &BlobStoreApiClient{
		baseUrl,
		credentialProvider,
		&http.Client{Timeout: time.Second * 30},
	}
}

func NewBlobFileStatFromResponse(basePathComponent string, response *http.Response) BlobFileStat {
	val := BlobFileStat{
		MimeType: response.Header.Get("Content-Type"),
		Exists:   true,
	}

	// Have to remove components based on the API base URL.
	requestPath := response.Request.URL.Path
	relativePath := strings.Replace(requestPath, basePathComponent, "", 1)

	finalSlash := strings.LastIndex(relativePath, "/")
	if finalSlash == -1 {
		val.Path = ""
		val.Name = relativePath
	} else {
		val.Path = relativePath[0 : finalSlash+1]
		val.Name = relativePath[finalSlash+1:]
	}

	size, err := strconv.Atoi(response.Header.Get("Content-Length"))
	if err == nil {
		val.SizeBytes = size
	}

	if response.StatusCode == 404 {
		val.Exists = false
	}

	return val
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

// This should be adapted to return an error, rather than panicing.
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

	baseUrlComponent, err := url.Parse(b.baseUrl)
	if err != nil {
		panic(err)
	}

	return baseUrlComponent.ResolveReference(pathUrlComponent).String()
}

func (b *BlobStoreApiClient) newAuthorizedRequest(method, path string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, b.route(path), body)
	if err != nil {
		return request, err
	}

	err = b.credentialProvider.AuthorizeRequest(request)
	return request, err
}

func (b *BlobStoreApiClient) UploadStream(path string, stream *bufio.Reader, contentType string) error {
	request, err := b.newAuthorizedRequest("POST", path, stream)
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

	response, err := b.http.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return NewBlobStoreHttpError("Upload", response)
	}

	return nil
}

func (b *BlobStoreApiClient) GetStat(path string) (*BlobFileStat, error) {
	request, err := b.newAuthorizedRequest("HEAD", path, nil)
	if err != nil {
		return nil, err
	}

	response, err := b.http.Do(request)
	if err != nil {
		return nil, err
	}

	baseUrlComponent, err := url.Parse(b.baseUrl)
	if err != nil {
		return nil, err
	}

	stat := NewBlobFileStatFromResponse(baseUrlComponent.Path, response)

	if response.StatusCode != 404 && response.StatusCode != 200 {
		return nil, NewBlobStoreHttpError("Stat", response)
	}

	return &stat, nil
}

func (b *BlobStoreApiClient) GetFile(path string) (*BlobFile, error) {
	request, err := b.newAuthorizedRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	response, err := b.http.Do(request)
	if err != nil {
		return nil, err
	}

	baseUrlComponent, err := url.Parse(b.baseUrl)
	if err != nil {
		return nil, err
	}

	stat := NewBlobFileStatFromResponse(baseUrlComponent.Path, response)

	if response.StatusCode != 200 {
		return nil, NewBlobStoreHttpError("Download", response)
	}

	body := response.Body.(io.Reader)
	rv := BlobFile{
		stat,
		&body,
	}

	return &rv, nil
}

func (b *BlobStoreApiClient) ListPrefix(prefix string, recursive bool) ([]string, error) {
	paths := make([]string, 0)

	for strings.HasPrefix(prefix, "/") {
		prefix = prefix[1:]
	}

	requestUrl := b.route("_dir/" + prefix)
	if recursive {
		requestUrl += "?recursive=true"
	}

	request, err := b.newAuthorizedRequest("GET", requestUrl, nil)
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

func (b *BlobStoreApiClient) DeleteFile(path string) error {
	request, err := b.newAuthorizedRequest("DELETE", path, nil)
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
