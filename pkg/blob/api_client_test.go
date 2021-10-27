package blob

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
)


import (
	"gitea.internal.aleemhaji.com/aleem/blobapi/pkg/credential_provider"
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
