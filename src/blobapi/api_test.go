package blobapi;

import (
    "bufio"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
    "os"
    "strings"
    "time"
    "testing"
)

import (
    "github.com/stretchr/testify/assert"
)

type HttpMockedMethod func(params ...interface{})(*http.Response, error)

type TestDrivenHttpClient struct {
    t *testing.T
    mockedCalls []HttpMockedMethod
}

func (t *TestDrivenHttpClient) Do(a *http.Request) (*http.Response, error) { return t.r(a) }
func (t *TestDrivenHttpClient) Get(a string) (*http.Response, error) { return t.r(a) }
func (t *TestDrivenHttpClient) Head(a string) (*http.Response, error) { return t.r(a) }
func (t *TestDrivenHttpClient) Post(a string, b string, c io.Reader) (*http.Response, error) { return t.r(a, b, c) }
func (t *TestDrivenHttpClient) PostForm(a string, b url.Values) (*http.Response, error) { return t.r(a, b) }

func (t *TestDrivenHttpClient) r(args ...interface{}) (*http.Response, error) {
    if len(t.mockedCalls) == 0 {
        panic("Not enough mocked calls to http.client to facilitate request.")
    }

    defer func() {
        t.mockedCalls = t.mockedCalls[1:]
    }()

    return t.mockedCalls[0](args...)
}

func TestCreation(t *testing.T) {
    var api *BlobStoreApiClient = NewBlobStoreApiClient("a", "b", "c")

    assert.Equal(t, "a/", api.DefaultUrl)
    assert.Equal(t, "b", api.DefaultReadAcl)
    assert.Equal(t, "c", api.DefaultWriteAcl)

    httpClient := api.http.(*http.Client)
    assert.Equal(t,  time.Second * 30, httpClient.Timeout)
}

func TestRoute(t *testing.T) {
    resolutions := []struct {
        BaseUrl string
        PathComponent string
        FinalUrl string
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

    for _, ti := range resolutions {
        api := NewBlobStoreApiClient(ti.BaseUrl, "", "")

        assert.Equal(t, ti.FinalUrl, api.route(ti.PathComponent))
    }

    panics := []struct {
        BaseUrl string
        PathComponent string
        PanicMessage string
    }{
        {":broken", "", "parse :broken/: missing protocol scheme"},
        {"https://example.org", ":broken", "parse :broken: missing protocol scheme"},
    }

    for _, ti := range panics {
        api := NewBlobStoreApiClient(ti.BaseUrl, "", "")
        func () {
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

func TestUploadRequest(t *testing.T) {
    api := NewBlobStoreApiClient("https://example.org/deeper", "read secret", "write secret")

    httpMock := func(params ...interface{}) (*http.Response, error) {
        request := params[0].(*http.Request)

        assert.Equal(t, "POST", request.Method)
        assert.Equal(t, "https://example.org/deeper/remote_filename", request.URL.String())
        assert.Equal(t, "text/plain", request.Header.Get("Content-Type"))
        assert.Equal(t, "read secret", request.Header.Get("X-BlobStore-Read-Acl"))
        assert.Equal(t, "write secret", request.Header.Get("X-BlobStore-Write-Acl"))

        body, err := ioutil.ReadAll(request.Body)
        assert.Nil(t, err)

        file, err := os.Open("../../Makefile")
        assert.Nil(t, err)

        expectedBody, err := ioutil.ReadAll(bufio.NewReader(file))
        assert.Nil(t, err)

        assert.Equal(t, expectedBody, body)

        response := http.Response{
            StatusCode: 200,
        }
        return &response, nil
    }

    api.http = &TestDrivenHttpClient{t, []HttpMockedMethod{httpMock}}
    err := api.UploadFile("remote_filename", "../../Makefile", "text/plain")
    assert.Nil(t, err)
}

func TestUploadRequestNoContentType(t *testing.T) {
    api := NewBlobStoreApiClient("https://example.org/deeper", "read secret", "write secret")

    httpMock := func(params ...interface{}) (*http.Response, error) {
        request := params[0].(*http.Request)

        assert.Equal(t, "POST", request.Method)
        assert.Equal(t, "https://example.org/deeper/remote_filename", request.URL.String())
        assert.Equal(t, "text/plain; charset=utf-8", request.Header.Get("Content-Type"))
        assert.Equal(t, "read secret", request.Header.Get("X-BlobStore-Read-Acl"))
        assert.Equal(t, "write secret", request.Header.Get("X-BlobStore-Write-Acl"))

        body, err := ioutil.ReadAll(request.Body)
        assert.Nil(t, err)

        file, err := os.Open("../../Makefile")
        assert.Nil(t, err)

        expectedBody, err := ioutil.ReadAll(bufio.NewReader(file))
        assert.Nil(t, err)

        assert.Equal(t, expectedBody, body)

        response := http.Response{
            StatusCode: 200,
        }
        return &response, nil
    }

    api.http = &TestDrivenHttpClient{t, []HttpMockedMethod{httpMock}}
    err := api.UploadFile("remote_filename", "../../Makefile", "")
    assert.Nil(t, err)
}

func TestUploadRequestFails(t *testing.T) {
     api := NewBlobStoreApiClient("https://example.org/deeper", "read secret", "write secret")

    httpMock := func(params ...interface{}) (*http.Response, error) {
        bodyReader := strings.NewReader("{\"code\":\"NotFound\",\"message\":\"File not found\"}")

        response := http.Response{
            StatusCode: 404,
            Body: ioutil.NopCloser(bodyReader),
        }
        return &response, nil
    }

    api.http = &TestDrivenHttpClient{t, []HttpMockedMethod{httpMock}}
    err := api.UploadFile("remote_filename", "../../Makefile", "text/plain")
    assert.Equal(t, "Blobstore Upload Failed (404): {\"code\":\"NotFound\",\"message\":\"File not found\"}", err.Error())
}

func TestDownloadRequest(t *testing.T) {
    api := NewBlobStoreApiClient("https://example.org/deeper", "read secret", "write secret")

    httpMock := func(params ...interface{}) (*http.Response, error) {
        request := params[0].(*http.Request)

        assert.Equal(t, "GET", request.Method)
        assert.Equal(t, "https://example.org/deeper/remote_filename", request.URL.String())
        assert.Equal(t, "read secret", request.Header.Get("X-BlobStore-Read-Acl"))
        assert.Equal(t, "", request.Header.Get("X-BlobStore-Write-Acl"))

        file, err := os.Open("../../Makefile")
        assert.Nil(t, err)

        bodyReader := bufio.NewReader(file)

        response := http.Response{
            StatusCode: 200,
            Body: ioutil.NopCloser(bodyReader),
        }
        return &response, nil
    }

    api.http = &TestDrivenHttpClient{t, []HttpMockedMethod{httpMock}}
    tempFile, err := ioutil.TempFile("", "")
    assert.Nil(t, err)
    tempFile.Close()

    err = api.DownloadFile("remote_filename", tempFile.Name())
    assert.Nil(t, err)

    tempFile, err = os.Open(tempFile.Name())
    assert.Nil(t, err)

    body, err := ioutil.ReadAll(bufio.NewReader(tempFile))
    assert.Nil(t, err)

    file, err := os.Open("../../Makefile")
    assert.Nil(t, err)

    expectedBody, err := ioutil.ReadAll(bufio.NewReader(file))
    assert.Nil(t, err)

    assert.Equal(t, expectedBody, body)
}

func TestDownloadRequestFails(t *testing.T) {
    api := NewBlobStoreApiClient("https://example.org/deeper", "read secret", "write secret")

    httpMock := func(params ...interface{}) (*http.Response, error) {
        bodyReader := strings.NewReader("{\"code\":\"NotFound\",\"message\":\"File not found\"}")

        response := http.Response{
            StatusCode: 404,
            Body: ioutil.NopCloser(bodyReader),
        }
        return &response, nil
    }

    api.http = &TestDrivenHttpClient{t, []HttpMockedMethod{httpMock}}
    tempFile, err := ioutil.TempFile("", "")
    assert.Nil(t, err)
    tempFile.Close()

    err = api.DownloadFile("remote_filename", tempFile.Name())
    assert.Equal(t, "Blobstore Download Failed (404): {\"code\":\"NotFound\",\"message\":\"File not found\"}", err.Error())
}

// Append functions are a bit more difficult to test, because they need to mock
// both a download and an upload function and return the appropriate responses
// for both calls.
func TestAppendStringRequest(t *testing.T) {
    api := NewBlobStoreApiClient("https://example.org/deeper", "read secret", "write secret")
    stringToAppend := "This is some text that should appear after the rest."

    getMock := func(params ...interface{}) (*http.Response, error) {
        request := params[0].(*http.Request)

        assert.Equal(t, "GET", request.Method)
        assert.Equal(t, "https://example.org/deeper/remote_filename", request.URL.String())
        assert.Equal(t, "read secret", request.Header.Get("X-BlobStore-Read-Acl"))
        assert.Equal(t, "", request.Header.Get("X-BlobStore-Write-Acl"))

        file, err := os.Open("../../Makefile")
        assert.Nil(t, err)

        bodyReader := bufio.NewReader(file)

        response := http.Response{
            StatusCode: 200,
            Body: ioutil.NopCloser(bodyReader),
        }

        response.Header = make(map[string][]string)
        response.Header.Set("Content-Type", "text/plain")

        return &response, nil
    }

    postMock := func(params ...interface{}) (*http.Response, error) {
        request := params[0].(*http.Request)

        assert.Equal(t, "POST", request.Method)
        assert.Equal(t, "https://example.org/deeper/remote_filename", request.URL.String())
        assert.Equal(t, "text/plain", request.Header.Get("Content-Type"))
        assert.Equal(t, "read secret", request.Header.Get("X-BlobStore-Read-Acl"))
        assert.Equal(t, "write secret", request.Header.Get("X-BlobStore-Write-Acl"))

        body, err := ioutil.ReadAll(request.Body)
        assert.Nil(t, err)

        file, err := os.Open("../../Makefile")
        assert.Nil(t, err)

        expectedBody, err := ioutil.ReadAll(bufio.NewReader(file))
        assert.Nil(t, err)

        appendBody, err := ioutil.ReadAll(strings.NewReader(stringToAppend))
        assert.Nil(t, err)

        expectedBody = append(expectedBody, appendBody...)
        assert.Equal(t, expectedBody, body)

        response := http.Response{
            StatusCode: 200,
        }
        return &response, nil
    }

    api.http = &TestDrivenHttpClient{t, []HttpMockedMethod{getMock, postMock}}
    tempFile, err := ioutil.TempFile("", "")
    assert.Nil(t, err)
    tempFile.Close()

    err = api.AppendString("remote_filename", stringToAppend)
    assert.Nil(t, err)
}

// Append functions are a bit more difficult to test, because they need to mock
// both a download and an upload function and return the appropriate responses
// for both calls.
func TestAppendFileRequest(t *testing.T) {
    api := NewBlobStoreApiClient("https://example.org/deeper", "read secret", "write secret")

    getMock := func(params ...interface{}) (*http.Response, error) {
        request := params[0].(*http.Request)

        assert.Equal(t, "GET", request.Method)
        assert.Equal(t, "https://example.org/deeper/remote_filename", request.URL.String())
        assert.Equal(t, "read secret", request.Header.Get("X-BlobStore-Read-Acl"))
        assert.Equal(t, "", request.Header.Get("X-BlobStore-Write-Acl"))

        file, err := os.Open("../../Makefile")
        assert.Nil(t, err)

        bodyReader := bufio.NewReader(file)

        response := http.Response{
            StatusCode: 200,
            Body: ioutil.NopCloser(bodyReader),
        }

        response.Header = make(map[string][]string)
        response.Header.Set("Content-Type", "text/plain")

        return &response, nil
    }

    postMock := func(params ...interface{}) (*http.Response, error) {
        request := params[0].(*http.Request)

        assert.Equal(t, "POST", request.Method)
        assert.Equal(t, "https://example.org/deeper/remote_filename", request.URL.String())
        assert.Equal(t, "text/plain", request.Header.Get("Content-Type"))
        assert.Equal(t, "read secret", request.Header.Get("X-BlobStore-Read-Acl"))
        assert.Equal(t, "write secret", request.Header.Get("X-BlobStore-Write-Acl"))

        body, err := ioutil.ReadAll(request.Body)
        assert.Nil(t, err)

        file, err := os.Open("../../Makefile")
        assert.Nil(t, err)

        expectedBody, err := ioutil.ReadAll(bufio.NewReader(file))
        assert.Nil(t, err)

        expectedBody = append(expectedBody, expectedBody...)
        assert.Equal(t, expectedBody, body)

        response := http.Response{
            StatusCode: 200,
        }
        return &response, nil
    }

    api.http = &TestDrivenHttpClient{t, []HttpMockedMethod{getMock, postMock}}
    tempFile, err := ioutil.TempFile("", "")
    assert.Nil(t, err)
    tempFile.Close()

    err = api.AppendFile("remote_filename", "../../Makefile")
    assert.Nil(t, err)
}
