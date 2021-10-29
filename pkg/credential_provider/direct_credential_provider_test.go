package credential_provider

import (
	"net/http"
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestDirectCredentialProvider(t *testing.T) {
	dcp := DirectCredentialProvider{"abc", "bcd"}

	request, err := http.NewRequest("GET", "https://example.org", nil)
	assert.Nil(t, err)

	err = dcp.AuthorizeRequest(request)
	assert.Nil(t, err)
	assert.Equal(t, "abc", request.Header.Get("X-BlobStore-Read-Acl"))
	assert.Equal(t, "bcd", request.Header.Get("X-BlobStore-Write-Acl"))

	dcp2 := DirectCredentialProvider{"cde", "def"}
	err = dcp2.AuthorizeRequest(request)
	assert.Nil(t, err)
	assert.Equal(t, "abc", request.Header.Get("X-BlobStore-Read-Acl"))
	assert.Equal(t, "bcd", request.Header.Get("X-BlobStore-Write-Acl"))
}
