package blob

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

import (
	"gitea.internal.aleemhaji.com/aleem/blobapi/pkg/credential_provider"
)

type BlobStoreClient struct {
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
	apiClient := NewBlobStoreApiClient(url, credentialProvider)

	return &BlobStoreClient{
		apiClient,
	}
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
