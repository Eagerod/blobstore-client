package blobapi;

import (
    "time"
    "testing"
)

import (
    "github.com/stretchr/testify/assert"
)

func TestCreation(t *testing.T) {
    var api *BlobStoreApiClient = NewBlobStoreApiClient("a", "b", "c")

    assert.Equal(t, "a", api.DefaultUrl)
    assert.Equal(t, "b", api.DefaultReadAcl)
    assert.Equal(t, "c", api.DefaultWriteAcl)
    assert.Equal(t,  time.Second * 30, api.http.Timeout)
}

func TestRoute(t *testing.T) {
    var api *BlobStoreApiClient = NewBlobStoreApiClient("https://example.org", "", "")

    assert.Equal(t, "https://example.org/path/to/object", api.route("/path/to/object"))
    assert.Equal(t, "https://example.org/path/to/object", api.route("path/to/object"))

    api = NewBlobStoreApiClient("https://example.org/", "", "")

    assert.Equal(t, "https://example.org/path/to/object", api.route("/path/to/object"))
    assert.Equal(t, "https://example.org/path/to/object", api.route("path/to/object"))
}
