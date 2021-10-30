package blob

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

import (
	"gitea.internal.aleemhaji.com/aleem/blobapi/pkg/credential_provider"
)

const BlobStoreUrlScheme string = "blob"

type BlobStoreClient struct {
	apiClient IBlobStoreApiClient
}

type IBlobStoreClient interface {
	Cat(src *url.URL) error
	Copy(src *url.URL, dst *url.URL, force bool) error

	UploadFile(path string, source string, contentType string) error

	GetFileContents(path string) (string, error)
	DownloadFile(path string, dest string) error

	StatFile(path string) (*BlobFileStat, error)

	AppendStream(path string, stream *bufio.Reader) error
	AppendString(path string, value string) error
	AppendFile(path string, source string) error

	ListPrefix(prefix string, recursive bool) ([]string, error)

	DeleteFile(path string) error

	Exists(url_ url.URL) (bool, error)
}

func NewBlobStoreClient(url string, credentialProvider credential_provider.ICredentialProvider) *BlobStoreClient {
	apiClient := NewBlobStoreApiClient(url, credentialProvider)

	return &BlobStoreClient{
		apiClient,
	}
}

func (b *BlobStoreClient) Copy(src *url.URL, dst *url.URL, force bool) error {
	cpArg0 := src

	// Determine if this is an upload or download command based on which
	// order the parameters came in.
	cpArg1 := dst

	if cpArg0.Scheme == BlobStoreUrlScheme && cpArg1.Scheme == BlobStoreUrlScheme {
		return errors.New("No support for copying files in the blobstore directly")
	}

	if cpArg0.Scheme != BlobStoreUrlScheme && cpArg1.Scheme != BlobStoreUrlScheme {
		return errors.New("Must provide at least one blob:/ path to upload to or download from")
	}

	if cpArg0.Scheme == BlobStoreUrlScheme {
		if force == false {
			if _, err := os.Stat(cpArg1.Path); err == nil {
				return errors.New("Destination file already exists on local machine; use --force to overwrite")
			}
		}
		return b.DownloadFile(cpArg0.Path, cpArg1.Path)
	} else {
		if force == false {
			fileStat, err := b.StatFile(cpArg1.Path)
			if err != nil {
				return err
			}
			if fileStat.Exists {
				return errors.New("Destination file already exists on blobstore; use --force to overwrite")
			}
		}
		return b.UploadFile(cpArg1.Path, cpArg0.Path, "")
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

func (b *BlobStoreClient) Cat(src *url.URL) error {
	if src.Scheme != BlobStoreUrlScheme {
		return errors.New("Must download files from blob:/")
	}

	str, err := b.GetFileContents(src.Path)
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

func (b *BlobStoreClient) Exists(url_ url.URL) (bool, error) {
	if url_.Scheme == "blob" {
		f, err := b.StatFile(url_.Path)
		if err != nil {
			return false, err
		}
		return f.Exists, nil
	} else {
		if _, err := os.Stat(url_.Path); err == nil {
			return true, nil
		} else {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, err
		}
	}
}
