package blob

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
)

import (
	"gitea.internal.aleemhaji.com/aleem/blobstore-cli/pkg/credential_provider"
)

func testApiClient() *BlobStoreApiClient {
	cred := credential_provider.DirectCredentialProvider{
		ReadAcl: RemoteTestReadSecret,
		WriteAcl: RemoteTestWriteSecret,
	}
	return NewBlobStoreApiClient(RemoteTestBaseUrl, &cred)
}

func TestNewApiClient(t *testing.T) {
	client := testApiClient()

	assert.Equal(t, fmt.Sprintf("%s/", RemoteTestBaseUrl), client.baseUrl)

	cred := client.credentialProvider.(*credential_provider.DirectCredentialProvider)
	assert.Equal(t, RemoteTestReadSecret, cred.ReadAcl)
	assert.Equal(t, RemoteTestWriteSecret, cred.WriteAcl)

	httpClient := client.http.(*http.Client)
	assert.Equal(t, time.Second*30, httpClient.Timeout)
}

func TestRoute(t *testing.T) {
	happyCases := []struct {
		BaseUrl       string
		PathComponent string
		FinalUrl      string
	}{
		{"https://example.org", "/path/to/object", "https://example.org/path/to/object"},
		{"https://example.org", "path/to/object", "https://example.org/path/to/object"},
		{"https://example.org/", "/path/to/object", "https://example.org/path/to/object"},
		{"https://example.org/", "path/to/object", "https://example.org/path/to/object"},
		{"https://example.org/deeper", "/path/to/object", "https://example.org/deeper/path/to/object"},
		{"https://example.org/deeper", "path/to/object", "https://example.org/deeper/path/to/object"},
		{"https://example.org/deeper/", "/path/to/object", "https://example.org/deeper/path/to/object"},
		{"https://example.org/deeper/", "path/to/object", "https://example.org/deeper/path/to/object"},
	}

	for _, ti := range happyCases {
		api := NewBlobStoreApiClient(ti.BaseUrl, nil)

		assert.Equal(t, ti.FinalUrl, api.route(ti.PathComponent))
	}
}

func TestRouteErrors(t *testing.T) {
	errorCases := []struct {
		BaseUrl       string
		PathComponent string
		PanicMessage  string
	}{
		{":broken", "", "parse \":broken/\": missing protocol scheme"},
		{"https://example.org", ":broken", "parse \":broken\": missing protocol scheme"},
	}

	for _, ti := range errorCases {
		api := NewBlobStoreApiClient(ti.BaseUrl, nil)
		func() {
			defer func() {
				r := recover()
				if r == nil {
					t.Errorf("Failed to produce panic: %s", ti.PanicMessage)
				} else {
					assert.Equal(t, ti.PanicMessage, r.(error).Error())
				}
			}()
			api.route(ti.PathComponent)
		}()
	}
}

func TestUploadStream(t *testing.T) {
	client := testApiClient()

	httpMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		assert.Equal(t, RemoteTestUploadHttpMethod, request.Method)
		assert.Equal(t, RemoteTestFileUrl, request.URL.String())
		assert.Equal(t, RemoteTestFileManualMimeType, request.Header.Get("Content-Type"))
		assert.Equal(t, RemoteTestReadSecret, request.Header.Get("X-BlobStore-Read-Acl"))
		assert.Equal(t, RemoteTestWriteSecret, request.Header.Get("X-BlobStore-Write-Acl"))

		response := http.Response{
			StatusCode: 200,
		}
		return &response, nil
	}

	client.http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}

	file, err := os.Open(LocalTestFilePath)
	assert.Nil(t, err)

	err = client.UploadStream(RemoteTestFilename, bufio.NewReader(file), RemoteTestFileManualMimeType)
	assert.Nil(t, err)
}

func TestUploadStreamNoContentType(t *testing.T) {
	client := testApiClient()

	httpMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		assert.Equal(t, RemoteTestUploadHttpMethod, request.Method)
		assert.Equal(t, RemoteTestFileUrl, request.URL.String())
		assert.Equal(t, RemoteTestFileAutomaticMimeType, request.Header.Get("Content-Type"))
		assert.Equal(t, RemoteTestReadSecret, request.Header.Get("X-BlobStore-Read-Acl"))
		assert.Equal(t, RemoteTestWriteSecret, request.Header.Get("X-BlobStore-Write-Acl"))

		response := http.Response{
			StatusCode: 200,
		}
		return &response, nil
	}

	client.http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}

	file, err := os.Open(LocalTestFilePath)
	assert.Nil(t, err)

	err = client.UploadStream(RemoteTestFilename, bufio.NewReader(file), "")
	assert.Nil(t, err)
}

func TestUploadStreamFails(t *testing.T) {
	client := testApiClient()

	httpMock := func(params ...interface{}) (*http.Response, error) {
		bodyReader := strings.NewReader("{\"code\":\"NotFound\",\"message\":\"File not found\"}")

		response := http.Response{
			StatusCode: 404,
			Body:       ioutil.NopCloser(bodyReader),
		}
		return &response, nil
	}

	client.http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}

	file, err := os.Open(LocalTestFilePath)
	assert.Nil(t, err)

	err = client.UploadStream(RemoteTestFilename, bufio.NewReader(file), RemoteTestFileManualMimeType)
	assert.Equal(t, "Blobstore Upload Failed (404): {\"code\":\"NotFound\",\"message\":\"File not found\"}", err.Error())
}

func TestGetFile(t *testing.T) {
	client := testApiClient()

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

	client.http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}

	blobFile, err := client.GetFile(RemoteTestFilename)
	assert.Nil(t, err)

	file, err := os.Open(LocalTestFilePath)
	assert.Nil(t, err)

	expectedBody, err := ioutil.ReadAll(bufio.NewReader(file))
	assert.Nil(t, err)

	body, err := ioutil.ReadAll(*blobFile.contents)
	assert.Nil(t, err)

	assert.Equal(t, expectedBody, body)
}

func TestGetFileFails(t *testing.T) {
	client := testApiClient()

	httpMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		bodyReader := strings.NewReader("{\"code\":\"NotFound\",\"message\":\"File not found\"}")

		response := http.Response{
			StatusCode: 404,
			Body:       ioutil.NopCloser(bodyReader),
			Request:    request,
		}
		return &response, nil
	}

	client.http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}
	tempFile, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	tempFile.Close()

	_, err = client.GetFile(RemoteTestFilename)
	assert.Equal(t, "Blobstore Download Failed (404): {\"code\":\"NotFound\",\"message\":\"File not found\"}", err.Error())
}

func TestGetStat(t *testing.T) {
	client := testApiClient()

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

	client.http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}

	fileStat, err := client.GetStat(RemoteTestFilename)
	assert.Nil(t, err)

	assert.Equal(t, "", fileStat.Path)
	assert.Equal(t, RemoteTestFilename, fileStat.Name)
	assert.Equal(t, RemoteTestFileManualMimeType, fileStat.MimeType)
	assert.Equal(t, 1024, fileStat.SizeBytes)
	assert.Equal(t, true, fileStat.Exists)
}

func TestGetStatLongerFilename(t *testing.T) {
	client := testApiClient()

	httpMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		assert.Equal(t, RemoteTestStatHttpMethod, request.Method)
		assert.Equal(t, RemoteTestDeepFileUrl, request.URL.String())
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

	client.http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}

	fileStat, err := client.GetStat(RemoteTestDeepFilename)
	assert.Nil(t, err)

	assert.Equal(t, "path/to/", fileStat.Path)
	assert.Equal(t, RemoteTestFilename, fileStat.Name)
	assert.Equal(t, RemoteTestFileManualMimeType, fileStat.MimeType)
	assert.Equal(t, 1024, fileStat.SizeBytes)
	assert.Equal(t, true, fileStat.Exists)
}

func TestGetStatDoesNotExist(t *testing.T) {
	client := testApiClient()

	httpMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		assert.Equal(t, RemoteTestStatHttpMethod, request.Method)
		assert.Equal(t, RemoteTestFileUrl, request.URL.String())
		assert.Equal(t, RemoteTestReadSecret, request.Header.Get("X-BlobStore-Read-Acl"))
		assert.Equal(t, RemoteTestWriteSecret, request.Header.Get("X-BlobStore-Write-Acl"))

		response := http.Response{
			StatusCode: 404,
			Request:    request,
		}

		return &response, nil
	}

	client.http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}
	fileStat, err := client.GetStat(RemoteTestFilename)
	assert.Nil(t, err)

	assert.Equal(t, "", fileStat.Path)
	assert.Equal(t, RemoteTestFilename, fileStat.Name)
	assert.Equal(t, "", fileStat.MimeType)
	assert.Equal(t, 0, fileStat.SizeBytes)
	assert.Equal(t, false, fileStat.Exists)
}

func TestGetStatFails(t *testing.T) {
	client := testApiClient()

	httpMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		response := http.Response{
			StatusCode: 403,
			Request:    request,
		}
		return &response, nil
	}

	client.http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}
	fileStat, err := client.GetStat(RemoteTestFilename)
	assert.Equal(t, "Blobstore Stat Failed (403)", err.Error())
	assert.Nil(t, fileStat)
}

func TestListPrefix(t *testing.T) {
	client := testApiClient()

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

	client.http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}
	filenames, err := client.ListPrefix("/", false)
	assert.Nil(t, err)

	assert.Equal(t, []string{"file-1", "file-2", "file-3"}, filenames)
}

func TestListPrefixRecursive(t *testing.T) {
	client := testApiClient()

	httpMock := func(params ...interface{}) (*http.Response, error) {
		request := params[0].(*http.Request)

		assert.Equal(t, RemoteTestListDirHttpMethod, request.Method)
		assert.Equal(t, RemoteTestListDirRecursiveUrl, request.URL.String())
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

	client.http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}
	filenames, err := client.ListPrefix("/", true)
	assert.Nil(t, err)

	assert.Equal(t, []string{"file-1", "file-2", "file-3"}, filenames)
}

func TestListPrefixFails(t *testing.T) {
	client := testApiClient()

	httpMock := func(params ...interface{}) (*http.Response, error) {
		bodyReader := strings.NewReader("{\"code\":\"BigProblem\",\"message\":\"The code is broken\"}")

		response := http.Response{
			StatusCode: 500,
			Body:       ioutil.NopCloser(bodyReader),
		}
		return &response, nil
	}

	client.http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}
	_, err := client.ListPrefix("/", false)
	assert.Equal(t, "Blobstore List Failed (500): {\"code\":\"BigProblem\",\"message\":\"The code is broken\"}", err.Error())
}

func TestDeleteFile(t *testing.T) {
	client := testApiClient()

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

	client.http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}
	err := client.DeleteFile(RemoteTestFilename)
	assert.Nil(t, err)
}

func TestDeleteFileFails(t *testing.T) {
	client := testApiClient()

	httpMock := func(params ...interface{}) (*http.Response, error) {
		bodyReader := strings.NewReader("{\"code\":\"PermissionDenied\",\"message\":\"Cannot delete\"}")

		response := http.Response{
			StatusCode: 403,
			Body:       ioutil.NopCloser(bodyReader),
		}

		return &response, nil
	}

	client.http = &TestDrivenHttpClient{[]HttpMockedMethod{httpMock}}
	err := client.DeleteFile(RemoteTestFilename)
	assert.Equal(t, "Blobstore Delete Failed (403): {\"code\":\"PermissionDenied\",\"message\":\"Cannot delete\"}", err.Error())
}
