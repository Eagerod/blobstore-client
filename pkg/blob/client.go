package blob

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	UploadFile(path string, source string, contentType string) error

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

func (b *BlobStoreClient) UploadFile(path string, source string, contentType string) error {
	file, err := os.Open(source)
	defer file.Close()

	if err != nil {
		return err
	}

	fileReader := bufio.NewReader(file)
	return b.apiClient.UploadStream(path, fileReader, contentType)
}

func (b *BlobStoreClient) GetFileContents(path string) (string, error) {
	file, err := b.apiClient.GetFile(path)
	if err != nil {
		return "", err
	}

	bodyBytes, err := ioutil.ReadAll(*file.contents)
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
	return b.apiClient.GetStat(path)
}

func (b *BlobStoreClient) AppendStream(path string, stream *bufio.Reader) error {
	f, err := b.apiClient.GetFile(path)
	if err != nil {
		return err
	}

	multiStream := bufio.NewReader(io.MultiReader(*f.contents, stream))
	return b.apiClient.UploadStream(path, multiStream, f.info.MimeType)
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
	return b.apiClient.ListPrefix(prefix, recursive)
}

func (b *BlobStoreClient) DeleteFile(path string) error {
	return b.apiClient.DeleteFile(path)
}
