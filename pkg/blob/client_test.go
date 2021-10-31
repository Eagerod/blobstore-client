package blob

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)


import (
	"gitea.internal.aleemhaji.com/aleem/blobapi/pkg/credential_provider"
)

type HttpMockedMethod func(params ...interface{}) (*http.Response, error)

type TestDrivenHttpClient struct {
	mockedCalls []HttpMockedMethod
}

func (t *TestDrivenHttpClient) Do(a *http.Request) (*http.Response, error) { return t.r(a) }
func (t *TestDrivenHttpClient) Get(a string) (*http.Response, error)       { return t.r(a) }
func (t *TestDrivenHttpClient) Head(a string) (*http.Response, error)      { return t.r(a) }
func (t *TestDrivenHttpClient) Post(a string, b string, c io.Reader) (*http.Response, error) {
	return t.r(a, b, c)
}
func (t *TestDrivenHttpClient) PostForm(a string, b url.Values) (*http.Response, error) {
	return t.r(a, b)
}

func (t *TestDrivenHttpClient) r(args ...interface{}) (*http.Response, error) {
	if len(t.mockedCalls) == 0 {
		panic("Not enough mocked calls to http.client to facilitate request.")
	}

	defer func() {
		t.mockedCalls = t.mockedCalls[1:]
	}()

	return t.mockedCalls[0](args...)
}

const (
	LocalTestFilePath = "../../Makefile"

	RemoteTestBaseUrl               = "https://example.com/deeper"
	RemoteTestFilename              = "remote_filename"
	RemoteTestDeepFilename          = "path/to/remote_filename"
	RemoteTestFileManualMimeType    = "text/plain"
	RemoteTestFileAutomaticMimeType = "text/plain; charset=utf-8"
	RemoteTestFileUrl               = "https://example.com/deeper/remote_filename"
	RemoteTestDeepFileUrl           = "https://example.com/deeper/path/to/remote_filename"
	RemoteTestListDirUrl            = "https://example.com/deeper/_dir/"
	RemoteTestListDirRecursiveUrl   = "https://example.com/deeper/_dir/?recursive=true"
	RemoteTestReadSecret            = "read secret"
	RemoteTestWriteSecret           = "write secret"
	RemoteTestUploadHttpMethod      = "POST"
	RemoteTestDownloadHttpMethod    = "GET"
	RemoteTestStatHttpMethod        = "HEAD"
	RemoteTestListDirHttpMethod     = "GET"
	RemoteTestDeleteHttpMethod      = "DELETE"
)

var RemoteTestURL *url.URL

func testClient() *BlobStoreClient {
	cred := credential_provider.DirectCredentialProvider{
		ReadAcl: RemoteTestReadSecret,
		WriteAcl: RemoteTestWriteSecret,
	}
	return NewBlobStoreClient(RemoteTestBaseUrl, &cred)
}

func TestMain(m *testing.M) {
	var err error
	RemoteTestURL, err = url.Parse(fmt.Sprintf("%s:/%s", BlobStoreUrlScheme, RemoteTestFilename))
	if err != nil {
		panic(err)
	}
	code := m.Run()
	os.Exit(code)
}

func TestCreation(t *testing.T) {
	client := testClient()
	assert.NotNil(t, client)
	assert.NotNil(t, client.apiClient)
}

func TestUploadRequest(t *testing.T) {
	var api *BlobStoreClient = testClient()

	httpMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		body, err := ioutil.ReadAll(request.Body)
		assert.Nil(t, err)

		file, err := os.Open(LocalTestFilePath)
		assert.Nil(t, err)

		expectedBody, err := ioutil.ReadAll(bufio.NewReader(file))
		assert.Nil(t, err)

		assert.Equal(t, expectedBody, body)

		response := http.Response{
			StatusCode: 200,
		}
		return &response, nil
	}

	api.apiClient.(*BlobStoreApiClient).http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}
	err := api.UploadFile(RemoteTestFilename, LocalTestFilePath, RemoteTestFileManualMimeType)
	assert.Nil(t, err)
}

func TestDownloadRequest(t *testing.T) {
	var api *BlobStoreClient = testClient()

	httpMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		assert.Equal(t, RemoteTestDownloadHttpMethod, request.Method)
		assert.Equal(t, RemoteTestFileUrl, request.URL.String())
		assert.Equal(t, RemoteTestReadSecret, request.Header.Get("X-BlobStore-Read-Acl"))
		assert.Equal(t, RemoteTestWriteSecret, request.Header.Get("X-BlobStore-Write-Acl"))

		file, err := os.Open(LocalTestFilePath)
		assert.Nil(t, err)

		bodyReader := bufio.NewReader(file)

		response := http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bodyReader),
			Request:    request,
		}
		return &response, nil
	}

	api.apiClient.(*BlobStoreApiClient).http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}
	tempFile, err := ioutil.TempFile("", "")
	defer os.Remove(tempFile.Name())
	assert.Nil(t, err)
	tempFile.Close()

	err = api.DownloadFile(RemoteTestFilename, tempFile.Name())
	assert.Nil(t, err)

	tempFile, err = os.Open(tempFile.Name())
	assert.Nil(t, err)

	body, err := ioutil.ReadAll(bufio.NewReader(tempFile))
	assert.Nil(t, err)

	file, err := os.Open(LocalTestFilePath)
	assert.Nil(t, err)

	expectedBody, err := ioutil.ReadAll(bufio.NewReader(file))
	assert.Nil(t, err)

	assert.Equal(t, expectedBody, body)
}

func TestDownloadRequestNonExistentDirectory(t *testing.T) {
	var api *BlobStoreClient = testClient()

	httpMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		assert.Equal(t, RemoteTestDownloadHttpMethod, request.Method)
		assert.Equal(t, RemoteTestFileUrl, request.URL.String())
		assert.Equal(t, RemoteTestReadSecret, request.Header.Get("X-BlobStore-Read-Acl"))
		assert.Equal(t, RemoteTestWriteSecret, request.Header.Get("X-BlobStore-Write-Acl"))

		file, err := os.Open(LocalTestFilePath)
		assert.Nil(t, err)

		bodyReader := bufio.NewReader(file)

		response := http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bodyReader),
			Request:    request,
		}
		return &response, nil
	}

	api.apiClient.(*BlobStoreApiClient).http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}
	tempDir, err := ioutil.TempDir("", "")
	defer os.RemoveAll(tempDir)

	assert.Nil(t, err)

	tempFilePath := filepath.Join(tempDir, "nested_directory", "temp_file")
	err = api.DownloadFile(RemoteTestFilename, tempFilePath)
	assert.Nil(t, err)

	tempFile, err := os.Open(tempFilePath)
	assert.Nil(t, err)

	body, err := ioutil.ReadAll(bufio.NewReader(tempFile))
	assert.Nil(t, err)

	file, err := os.Open(LocalTestFilePath)
	assert.Nil(t, err)

	expectedBody, err := ioutil.ReadAll(bufio.NewReader(file))
	assert.Nil(t, err)

	assert.Equal(t, expectedBody, body)
}

func TestStatRequest(t *testing.T) {
	var api *BlobStoreClient = testClient()

	httpMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		assert.Equal(t, RemoteTestStatHttpMethod, request.Method)
		assert.Equal(t, RemoteTestFileUrl, request.URL.String())
		assert.Equal(t, RemoteTestReadSecret, request.Header.Get("X-BlobStore-Read-Acl"))
		assert.Equal(t, RemoteTestWriteSecret, request.Header.Get("X-BlobStore-Write-Acl"))

		response := http.Response{
			StatusCode: 200,
			Request:    request,
		}

		response.Header = make(map[string][]string)
		response.Header.Set("Content-Type", RemoteTestFileManualMimeType)
		response.Header.Set("Content-Length", "1024")

		return &response, nil
	}

	api.apiClient.(*BlobStoreApiClient).http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}

	fileStat, err := api.StatFile(RemoteTestURL)
	assert.Nil(t, err)

	assert.Equal(t, "", fileStat.Path)
	assert.Equal(t, RemoteTestFilename, fileStat.Name)
	assert.Equal(t, RemoteTestFileManualMimeType, fileStat.MimeType)
	assert.Equal(t, 1024, fileStat.SizeBytes)
	assert.Equal(t, true, fileStat.Exists)
}

// Append functions are a bit more difficult to test, because they need to mock
// both a download and an upload function and return the appropriate responses
// for both calls.
func TestAppendStringRequest(t *testing.T) {
	var api *BlobStoreClient = testClient()
	stringToAppend := "This is some text that should appear after the rest."

	getMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		assert.Equal(t, RemoteTestDownloadHttpMethod, request.Method)
		assert.Equal(t, RemoteTestFileUrl, request.URL.String())
		assert.Equal(t, RemoteTestReadSecret, request.Header.Get("X-BlobStore-Read-Acl"))
		assert.Equal(t, RemoteTestWriteSecret, request.Header.Get("X-BlobStore-Write-Acl"))

		file, err := os.Open(LocalTestFilePath)
		assert.Nil(t, err)

		bodyReader := bufio.NewReader(file)

		response := http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bodyReader),
			Request:    request,
		}

		response.Header = make(map[string][]string)
		response.Header.Set("Content-Type", RemoteTestFileManualMimeType)

		return &response, nil
	}

	postMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		assert.Equal(t, RemoteTestUploadHttpMethod, request.Method)
		assert.Equal(t, RemoteTestFileUrl, request.URL.String())
		assert.Equal(t, RemoteTestFileManualMimeType, request.Header.Get("Content-Type"))
		assert.Equal(t, RemoteTestReadSecret, request.Header.Get("X-BlobStore-Read-Acl"))
		assert.Equal(t, RemoteTestWriteSecret, request.Header.Get("X-BlobStore-Write-Acl"))

		body, err := ioutil.ReadAll(request.Body)
		assert.Nil(t, err)

		file, err := os.Open(LocalTestFilePath)
		assert.Nil(t, err)

		expectedBody, err := ioutil.ReadAll(bufio.NewReader(file))
		assert.Nil(t, err)

		appendBody, err := ioutil.ReadAll(strings.NewReader(stringToAppend))
		assert.Nil(t, err)

		expectedBody = append(expectedBody, appendBody...)
		assert.Equal(t, expectedBody, body)

		response := http.Response{
			StatusCode: 200,
			Request:    request,
		}
		return &response, nil
	}

	api.apiClient.(*BlobStoreApiClient).http = &TestDrivenHttpClient{[]HttpMockedMethod{getMock, postMock}}
	tempFile, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	tempFile.Close()

	err = api.AppendString(RemoteTestFilename, stringToAppend)
	assert.Nil(t, err)
}

// Append functions are a bit more difficult to test, because they need to mock
// both a download and an upload function and return the appropriate responses
// for both calls.
func TestAppendFileRequest(t *testing.T) {
	var api *BlobStoreClient = testClient()

	getMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		assert.Equal(t, RemoteTestDownloadHttpMethod, request.Method)
		assert.Equal(t, RemoteTestFileUrl, request.URL.String())
		assert.Equal(t, RemoteTestReadSecret, request.Header.Get("X-BlobStore-Read-Acl"))
		assert.Equal(t, RemoteTestWriteSecret, request.Header.Get("X-BlobStore-Write-Acl"))

		file, err := os.Open(LocalTestFilePath)
		assert.Nil(t, err)

		bodyReader := bufio.NewReader(file)

		response := http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bodyReader),
			Request:    request,
		}

		response.Header = make(map[string][]string)
		response.Header.Set("Content-Type", RemoteTestFileManualMimeType)

		return &response, nil
	}

	postMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		assert.Equal(t, RemoteTestUploadHttpMethod, request.Method)
		assert.Equal(t, RemoteTestFileUrl, request.URL.String())
		assert.Equal(t, RemoteTestFileManualMimeType, request.Header.Get("Content-Type"))
		assert.Equal(t, RemoteTestReadSecret, request.Header.Get("X-BlobStore-Read-Acl"))
		assert.Equal(t, RemoteTestWriteSecret, request.Header.Get("X-BlobStore-Write-Acl"))

		body, err := ioutil.ReadAll(request.Body)
		assert.Nil(t, err)

		file, err := os.Open(LocalTestFilePath)
		assert.Nil(t, err)

		expectedBody, err := ioutil.ReadAll(bufio.NewReader(file))
		assert.Nil(t, err)

		expectedBody = append(expectedBody, expectedBody...)
		assert.Equal(t, expectedBody, body)

		response := http.Response{
			StatusCode: 200,
			Request:    request,
		}
		return &response, nil
	}

	api.apiClient.(*BlobStoreApiClient).http = &TestDrivenHttpClient{[]HttpMockedMethod{getMock, postMock}}
	tempFile, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	tempFile.Close()

	err = api.AppendFile(RemoteTestFilename, LocalTestFilePath)
	assert.Nil(t, err)
}

func TestListRequest(t *testing.T) {
	var api *BlobStoreClient = testClient()

	httpMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		assert.Equal(t, RemoteTestListDirHttpMethod, request.Method)
		assert.Equal(t, RemoteTestListDirUrl, request.URL.String())
		assert.Equal(t, RemoteTestReadSecret, request.Header.Get("X-BlobStore-Read-Acl"))
		assert.Equal(t, RemoteTestWriteSecret, request.Header.Get("X-BlobStore-Write-Acl"))

		files := []string{"file-1", "file-2", "file-3"}
		filesBytes, err := json.Marshal(files)
		assert.Nil(t, err)

		response := http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(filesBytes)),
		}
		return &response, nil
	}

	api.apiClient.(*BlobStoreApiClient).http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}
	filenames, err := api.ListPrefix("/", false)
	assert.Nil(t, err)

	assert.Equal(t, []string{"file-1", "file-2", "file-3"}, filenames)
}

func TestDeleteRequest(t *testing.T) {
	var api *BlobStoreClient = testClient()

	httpMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		assert.Equal(t, RemoteTestDeleteHttpMethod, request.Method)
		assert.Equal(t, RemoteTestFileUrl, request.URL.String())
		assert.Equal(t, RemoteTestReadSecret, request.Header.Get("X-BlobStore-Read-Acl"))
		assert.Equal(t, RemoteTestWriteSecret, request.Header.Get("X-BlobStore-Write-Acl"))

		response := http.Response{
			StatusCode: 200,
		}

		return &response, nil
	}

	api.apiClient.(*BlobStoreApiClient).http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}

	err := api.DeleteFile(RemoteTestURL)
	assert.Nil(t, err)
}
