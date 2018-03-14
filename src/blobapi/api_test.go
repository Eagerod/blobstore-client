package blobapi;

import (
    "net/http"
    "time"
    "testing"
)

import (
    "github.com/stretchr/testify/assert"
)

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
}
