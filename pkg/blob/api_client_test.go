package blob

import (
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

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
